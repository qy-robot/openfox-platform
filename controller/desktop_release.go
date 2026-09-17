package controller

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

const desktopReleaseFieldLimit = int64(64 * 1024)

func ListDesktopReleases(c *gin.Context) {
	view, err := service.ListDesktopReleases()
	if err != nil {
		writeDesktopReleaseError(c, http.StatusInternalServerError, "RELEASES_STORAGE_ERROR", "failed to read releases", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": view})
}

func UploadDesktopReleaseDraft(c *gin.Context) {
	actor, ok := desktopReleaseActor(c)
	if !ok {
		return
	}

	maxFileBytes := service.DesktopReleaseMaxFileBytes()
	maxRequestBytes := maxFileBytes*int64(len(service.DesktopReleaseTargets)) + 1024*1024
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBytes)
	reader, err := c.Request.MultipartReader()
	if err != nil {
		writeDesktopReleaseError(c, http.StatusBadRequest, "RELEASES_INVALID_MULTIPART", "multipart form data is required", err)
		return
	}

	root := service.DesktopReleaseStorageDir()
	if err := os.MkdirAll(root, 0o750); err != nil {
		writeDesktopReleaseError(c, http.StatusInternalServerError, "RELEASES_STORAGE_ERROR", "failed to prepare release storage", err)
		return
	}
	staging, err := os.MkdirTemp(root, ".upload-")
	if err != nil {
		writeDesktopReleaseError(c, http.StatusInternalServerError, "RELEASES_STORAGE_ERROR", "failed to prepare upload", err)
		return
	}
	defer os.RemoveAll(staging)

	version := ""
	changelog := ""
	type stagedArtifact struct {
		target   string
		fileName string
		path     string
	}
	staged := make([]stagedArtifact, 0, len(service.DesktopReleaseTargets))
	seen := make(map[string]struct{}, len(service.DesktopReleaseTargets))
	for {
		part, nextErr := reader.NextPart()
		if errors.Is(nextErr, io.EOF) {
			break
		}
		if nextErr != nil {
			var maxBytesErr *http.MaxBytesError
			if errors.As(nextErr, &maxBytesErr) {
				writeDesktopReleaseError(c, http.StatusRequestEntityTooLarge, "RELEASES_UPLOAD_TOO_LARGE", "release upload is too large", nextErr)
				return
			}
			writeDesktopReleaseError(c, http.StatusBadRequest, "RELEASES_INVALID_MULTIPART", "failed to read multipart upload", nextErr)
			return
		}

		fieldName := part.FormName()
		fileName := part.FileName()
		if fileName == "" {
			value, readErr := io.ReadAll(io.LimitReader(part, desktopReleaseFieldLimit+1))
			_ = part.Close()
			if readErr != nil || int64(len(value)) > desktopReleaseFieldLimit {
				writeDesktopReleaseError(c, http.StatusBadRequest, "RELEASES_INVALID_FIELD", "release metadata field is invalid", readErr)
				return
			}
			switch fieldName {
			case "version":
				version = strings.TrimSpace(string(value))
			case "changelog":
				changelog = strings.TrimSpace(string(value))
			default:
				writeDesktopReleaseError(c, http.StatusBadRequest, "RELEASES_UNKNOWN_FIELD", "unsupported release form field", nil)
				return
			}
			continue
		}

		if !service.IsDesktopReleaseTarget(fieldName) {
			_ = part.Close()
			writeDesktopReleaseError(c, http.StatusBadRequest, "RELEASES_UNSUPPORTED_TARGET", "unsupported release target", nil)
			return
		}
		if _, exists := seen[fieldName]; exists {
			_ = part.Close()
			writeDesktopReleaseError(c, http.StatusBadRequest, "RELEASES_DUPLICATE_TARGET", "a release target may only be uploaded once", nil)
			return
		}
		seen[fieldName] = struct{}{}
		path := filepath.Join(staging, fieldName)
		file, createErr := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
		if createErr != nil {
			_ = part.Close()
			writeDesktopReleaseError(c, http.StatusInternalServerError, "RELEASES_STORAGE_ERROR", "failed to stage artifact", createErr)
			return
		}
		written, copyErr := io.Copy(file, io.LimitReader(part, maxFileBytes+1))
		syncErr := file.Sync()
		closeErr := file.Close()
		_ = part.Close()
		if copyErr != nil || syncErr != nil || closeErr != nil {
			writeDesktopReleaseError(c, http.StatusInternalServerError, "RELEASES_STORAGE_ERROR", "failed to stage artifact", errors.Join(copyErr, syncErr, closeErr))
			return
		}
		if written > maxFileBytes {
			writeDesktopReleaseError(c, http.StatusRequestEntityTooLarge, "RELEASES_ARTIFACT_TOO_LARGE", fmt.Sprintf("each artifact is limited to %d bytes", maxFileBytes), nil)
			return
		}
		if written == 0 {
			writeDesktopReleaseError(c, http.StatusBadRequest, "RELEASES_EMPTY_ARTIFACT", "release artifacts cannot be empty", nil)
			return
		}
		staged = append(staged, stagedArtifact{target: fieldName, fileName: fileName, path: path})
	}

	if !service.ValidateDesktopReleaseVersion(version) {
		writeDesktopReleaseError(c, http.StatusBadRequest, "RELEASES_INVALID_VERSION", "version must be a semantic version such as 1.2.3 or 1.2.3-beta.1", nil)
		return
	}
	if changelog == "" {
		writeDesktopReleaseError(c, http.StatusBadRequest, "RELEASES_CHANGELOG_REQUIRED", "changelog is required", nil)
		return
	}
	if len(staged) == 0 {
		writeDesktopReleaseError(c, http.StatusBadRequest, "RELEASES_ARTIFACT_REQUIRED", "at least one release artifact is required", nil)
		return
	}

	uploads := make([]service.DesktopReleaseUpload, 0, len(staged))
	files := make([]*os.File, 0, len(staged))
	defer func() {
		for _, file := range files {
			_ = file.Close()
		}
	}()
	for _, artifact := range staged {
		file, openErr := os.Open(artifact.path)
		if openErr != nil {
			writeDesktopReleaseError(c, http.StatusInternalServerError, "RELEASES_STORAGE_ERROR", "failed to read staged artifact", openErr)
			return
		}
		files = append(files, file)
		uploads = append(uploads, service.DesktopReleaseUpload{Target: artifact.target, FileName: artifact.fileName, Reader: file})
	}

	release, err := service.SaveDesktopReleaseDraft(version, changelog, actor, uploads)
	if errors.Is(err, service.ErrDesktopReleaseAlreadyPublished) {
		writeDesktopReleaseError(c, http.StatusConflict, "RELEASES_IMMUTABLE", "published releases cannot be replaced", err)
		return
	}
	if err != nil {
		writeDesktopReleaseError(c, http.StatusInternalServerError, "RELEASES_STORAGE_ERROR", "failed to save release draft", err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": release})
}

func PublishDesktopRelease(c *gin.Context) {
	actor, ok := desktopReleaseActor(c)
	if !ok {
		return
	}
	version := strings.TrimSpace(c.Param("version"))
	if !service.ValidateDesktopReleaseVersion(version) {
		writeDesktopReleaseError(c, http.StatusBadRequest, "RELEASES_INVALID_VERSION", "invalid release version", nil)
		return
	}
	release, err := service.PublishDesktopRelease(version, actor)
	switch {
	case errors.Is(err, service.ErrDesktopReleaseNotFound):
		writeDesktopReleaseError(c, http.StatusNotFound, "RELEASES_NOT_FOUND", "release draft was not found", err)
		return
	case errors.Is(err, service.ErrDesktopReleaseAlreadyPublished):
		writeDesktopReleaseError(c, http.StatusConflict, "RELEASES_ALREADY_PUBLISHED", "release is already published", err)
		return
	case errors.Is(err, service.ErrDesktopReleaseNoArtifacts):
		writeDesktopReleaseError(c, http.StatusConflict, "RELEASES_ARTIFACT_REQUIRED", "release has no artifacts", err)
		return
	case err != nil:
		writeDesktopReleaseError(c, http.StatusInternalServerError, "RELEASES_STORAGE_ERROR", "failed to publish release", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": release})
}

func GetDesktopDownloadManifest(c *gin.Context) {
	manifest, err := service.DesktopDownloadManifestView()
	if err != nil {
		writeDesktopReleaseError(c, http.StatusInternalServerError, "RELEASES_STORAGE_ERROR", "failed to read download manifest", err)
		return
	}
	c.Header("Cache-Control", "no-cache")
	c.JSON(http.StatusOK, manifest)
}

func DownloadDesktopReleaseArtifact(c *gin.Context) {
	file, artifact, err := service.OpenDesktopReleaseArtifact(c.Param("version"), c.Param("target"))
	if err != nil {
		if !errors.Is(err, service.ErrDesktopReleaseNotFound) && !errors.Is(err, os.ErrNotExist) {
			common.SysLog("failed to open desktop release artifact: " + err.Error())
		}
		c.Status(http.StatusNotFound)
		return
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": artifact.FileName}))
	http.ServeContent(c.Writer, c.Request, artifact.FileName, stat.ModTime(), file)
}

func desktopReleaseActor(c *gin.Context) (service.DesktopReleaseActor, bool) {
	var actor service.DesktopReleaseActor
	raw := strings.TrimSpace(c.GetHeader("X-Robo-Actor"))
	if raw == "" || common.Unmarshal([]byte(raw), &actor) != nil {
		writeDesktopReleaseError(c, http.StatusBadRequest, "RELEASES_ACTOR_REQUIRED", "X-Robo-Actor must be JSON with subject and displayName", nil)
		return service.DesktopReleaseActor{}, false
	}
	actor.Subject = strings.TrimSpace(actor.Subject)
	actor.DisplayName = strings.TrimSpace(actor.DisplayName)
	if actor.Subject == "" || actor.DisplayName == "" || len(actor.Subject) > 256 || len(actor.DisplayName) > 256 {
		writeDesktopReleaseError(c, http.StatusBadRequest, "RELEASES_ACTOR_REQUIRED", "X-Robo-Actor must include a valid subject and displayName", nil)
		return service.DesktopReleaseActor{}, false
	}
	return actor, true
}

func writeDesktopReleaseError(c *gin.Context, status int, code, message string, err error) {
	if err != nil && status >= http.StatusInternalServerError {
		common.SysLog("desktop release error: " + err.Error())
	}
	c.JSON(status, gin.H{"success": false, "code": code, "message": message})
}
