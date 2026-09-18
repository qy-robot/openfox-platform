package service

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
)

const (
	desktopReleaseCatalogSchema = 1
	desktopReleaseRetention     = 3
	defaultReleaseMaxFileBytes  = int64(2 * 1024 * 1024 * 1024)
)

var (
	desktopReleaseMu      sync.Mutex
	desktopReleaseVersion = regexp.MustCompile(`^v?[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z][0-9A-Za-z.-]*)?(?:\+[0-9A-Za-z][0-9A-Za-z.-]*)?$`)
)

var DesktopReleaseTargets = []string{
	"windows-x64",
	"macos-arm64",
	"macos-x64",
	"linux-x64",
}

type DesktopReleaseActor struct {
	Subject     string `json:"subject"`
	DisplayName string `json:"displayName"`
}

type DesktopReleaseArtifact struct {
	Target    string `json:"target"`
	FileName  string `json:"fileName"`
	StoredAs  string `json:"-"`
	SizeBytes int64  `json:"sizeBytes"`
	SHA256    string `json:"sha256"`
}

type DesktopRelease struct {
	Version     string                   `json:"version"`
	Changelog   string                   `json:"changelog"`
	Status      string                   `json:"status"`
	CreatedAt   time.Time                `json:"createdAt"`
	UpdatedAt   time.Time                `json:"updatedAt"`
	PublishedAt *time.Time               `json:"publishedAt"`
	Actor       DesktopReleaseActor      `json:"actor"`
	Artifacts   []DesktopReleaseArtifact `json:"artifacts"`
}

type DesktopReleaseCatalogView struct {
	Latest   *DesktopRelease  `json:"latest"`
	Releases []DesktopRelease `json:"releases"`
}

type DesktopDownloadTarget struct {
	ID       string  `json:"id"`
	Status   string  `json:"status"`
	URL      *string `json:"url"`
	FileName *string `json:"fileName"`
	Size     *string `json:"size"`
	SHA256   *string `json:"sha256"`
}

type DesktopDownloadManifest struct {
	SchemaVersion int                     `json:"schemaVersion"`
	Product       string                  `json:"product"`
	Version       *string                 `json:"version"`
	PublishedAt   *time.Time              `json:"publishedAt"`
	Changelog     *string                 `json:"changelog"`
	Downloads     []DesktopDownloadTarget `json:"downloads"`
}

type DesktopReleaseUpload struct {
	Target   string
	FileName string
	Reader   io.Reader
}

type desktopReleaseCatalog struct {
	SchemaVersion int              `json:"schemaVersion"`
	Releases      []DesktopRelease `json:"releases"`
}

var (
	ErrDesktopReleaseNotFound         = errors.New("desktop release not found")
	ErrDesktopReleaseAlreadyPublished = errors.New("desktop release is already published")
	ErrDesktopReleaseNoArtifacts      = errors.New("desktop release has no artifacts")
)

func DesktopReleaseStorageDir() string {
	if configured := strings.TrimSpace(os.Getenv("ROBO_RELEASE_STORAGE_DIR")); configured != "" {
		return configured
	}
	return filepath.Join("data", "desktop-releases")
}

func DesktopReleaseMaxFileBytes() int64 {
	value := strings.TrimSpace(os.Getenv("ROBO_RELEASE_MAX_FILE_BYTES"))
	if value == "" {
		return defaultReleaseMaxFileBytes
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return defaultReleaseMaxFileBytes
	}
	return parsed
}

func IsDesktopReleaseTarget(target string) bool {
	for _, candidate := range DesktopReleaseTargets {
		if target == candidate {
			return true
		}
	}
	return false
}

func ValidateDesktopReleaseVersion(version string) bool {
	return desktopReleaseVersion.MatchString(version)
}

func ListDesktopReleases() (DesktopReleaseCatalogView, error) {
	desktopReleaseMu.Lock()
	defer desktopReleaseMu.Unlock()

	catalog, err := readDesktopReleaseCatalog()
	if err != nil {
		return DesktopReleaseCatalogView{}, err
	}
	return desktopReleaseCatalogView(catalog), nil
}

func SaveDesktopReleaseDraft(version, changelog string, actor DesktopReleaseActor, uploads []DesktopReleaseUpload) (DesktopRelease, error) {
	desktopReleaseMu.Lock()
	defer desktopReleaseMu.Unlock()

	catalog, err := readDesktopReleaseCatalog()
	if err != nil {
		return DesktopRelease{}, err
	}
	var existing *DesktopRelease
	for i := range catalog.Releases {
		release := &catalog.Releases[i]
		if release.Version == version && release.Status == "published" {
			return DesktopRelease{}, ErrDesktopReleaseAlreadyPublished
		}
		if release.Version == version {
			existing = release
		}
	}
	uploadTargets := make(map[string]struct{}, len(uploads))
	for _, upload := range uploads {
		if !IsDesktopReleaseTarget(upload.Target) {
			return DesktopRelease{}, fmt.Errorf("unsupported release target %q", upload.Target)
		}
		if _, exists := uploadTargets[upload.Target]; exists {
			return DesktopRelease{}, fmt.Errorf("duplicate release target %q", upload.Target)
		}
		uploadTargets[upload.Target] = struct{}{}
		fileName := filepath.Base(strings.TrimSpace(upload.FileName))
		if fileName == "." || fileName == "" || fileName != upload.FileName {
			return DesktopRelease{}, fmt.Errorf("invalid artifact filename")
		}
	}

	root := DesktopReleaseStorageDir()
	if err := os.MkdirAll(filepath.Join(root, "artifacts"), 0o750); err != nil {
		return DesktopRelease{}, fmt.Errorf("create release storage: %w", err)
	}
	staging, err := os.MkdirTemp(root, ".draft-")
	if err != nil {
		return DesktopRelease{}, fmt.Errorf("create release staging directory: %w", err)
	}
	defer os.RemoveAll(staging)

	artifacts := make([]DesktopReleaseArtifact, 0, len(uploads)+len(catalog.Releases))
	if existing != nil {
		existingDir := desktopReleaseArtifactDir(root, version)
		for _, artifact := range existing.Artifacts {
			if _, replaced := uploadTargets[artifact.Target]; replaced {
				continue
			}
			storedAs := desktopReleaseStoredName(artifact)
			source := filepath.Join(existingDir, storedAs)
			destination := filepath.Join(staging, storedAs)
			if err := os.Link(source, destination); err != nil {
				input, openErr := os.Open(source)
				if openErr != nil {
					return DesktopRelease{}, fmt.Errorf("open existing artifact: %w", openErr)
				}
				output, createErr := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
				if createErr != nil {
					_ = input.Close()
					return DesktopRelease{}, fmt.Errorf("copy existing artifact: %w", createErr)
				}
				_, copyErr := io.Copy(output, input)
				syncErr := output.Sync()
				closeOutputErr := output.Close()
				closeInputErr := input.Close()
				if err := errors.Join(copyErr, syncErr, closeOutputErr, closeInputErr); err != nil {
					return DesktopRelease{}, fmt.Errorf("copy existing artifact: %w", err)
				}
			}
			artifacts = append(artifacts, artifact)
		}
	}
	seen := make(map[string]struct{}, len(uploads))
	for _, upload := range uploads {
		seen[upload.Target] = struct{}{}

		fileName := filepath.Base(strings.TrimSpace(upload.FileName))
		temporaryPath := filepath.Join(staging, upload.Target)
		file, err := os.OpenFile(temporaryPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
		if err != nil {
			return DesktopRelease{}, fmt.Errorf("create staged artifact: %w", err)
		}
		hash := sha256.New()
		written, copyErr := io.Copy(io.MultiWriter(file, hash), io.LimitReader(upload.Reader, DesktopReleaseMaxFileBytes()+1))
		syncErr := file.Sync()
		closeErr := file.Close()
		if copyErr != nil {
			return DesktopRelease{}, fmt.Errorf("store artifact: %w", copyErr)
		}
		if written > DesktopReleaseMaxFileBytes() {
			return DesktopRelease{}, fmt.Errorf("artifact exceeds maximum size of %d bytes", DesktopReleaseMaxFileBytes())
		}
		if written == 0 {
			return DesktopRelease{}, fmt.Errorf("artifact %q is empty", upload.Target)
		}
		if syncErr != nil {
			return DesktopRelease{}, fmt.Errorf("sync artifact: %w", syncErr)
		}
		if closeErr != nil {
			return DesktopRelease{}, fmt.Errorf("close artifact: %w", closeErr)
		}
		digest := fmt.Sprintf("%x", hash.Sum(nil))
		storedAs := upload.Target + "-" + digest[:16] + strings.ToLower(filepath.Ext(fileName))
		if err := os.Rename(temporaryPath, filepath.Join(staging, storedAs)); err != nil {
			return DesktopRelease{}, fmt.Errorf("name staged artifact: %w", err)
		}
		artifacts = append(artifacts, DesktopReleaseArtifact{
			Target: upload.Target, FileName: fileName, StoredAs: storedAs, SizeBytes: written, SHA256: digest,
		})
	}
	if len(artifacts) == 0 {
		return DesktopRelease{}, ErrDesktopReleaseNoArtifacts
	}
	sort.Slice(artifacts, func(i, j int) bool { return artifacts[i].Target < artifacts[j].Target })

	now := time.Now().UTC()
	release := DesktopRelease{
		Version: version, Changelog: changelog, Status: "draft", CreatedAt: now, UpdatedAt: now, Actor: actor, Artifacts: artifacts,
	}
	for i := range catalog.Releases {
		if catalog.Releases[i].Version == version {
			release.CreatedAt = catalog.Releases[i].CreatedAt
			catalog.Releases = append(catalog.Releases[:i], catalog.Releases[i+1:]...)
			break
		}
	}

	releaseDir := desktopReleaseArtifactDir(root, version)
	backupDir := ""
	if _, statErr := os.Stat(releaseDir); statErr == nil {
		backupDir = releaseDir + ".old-" + fmt.Sprintf("%d", now.UnixNano())
		if err := os.Rename(releaseDir, backupDir); err != nil {
			return DesktopRelease{}, fmt.Errorf("stage existing release: %w", err)
		}
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return DesktopRelease{}, fmt.Errorf("inspect release directory: %w", statErr)
	}
	if err := os.Rename(staging, releaseDir); err != nil {
		if backupDir != "" {
			_ = os.Rename(backupDir, releaseDir)
		}
		return DesktopRelease{}, fmt.Errorf("commit release artifacts: %w", err)
	}

	catalog.Releases = append(catalog.Releases, release)
	if err := writeDesktopReleaseCatalog(catalog); err != nil {
		_ = os.RemoveAll(releaseDir)
		if backupDir != "" {
			_ = os.Rename(backupDir, releaseDir)
		}
		return DesktopRelease{}, err
	}
	if backupDir != "" {
		_ = os.RemoveAll(backupDir)
	}
	return release, nil
}

func PublishDesktopRelease(version string, actor DesktopReleaseActor) (DesktopRelease, error) {
	desktopReleaseMu.Lock()
	defer desktopReleaseMu.Unlock()

	catalog, err := readDesktopReleaseCatalog()
	if err != nil {
		return DesktopRelease{}, err
	}
	now := time.Now().UTC()
	publishedIndex := -1
	for i := range catalog.Releases {
		if catalog.Releases[i].Version != version {
			continue
		}
		if catalog.Releases[i].Status == "published" {
			return DesktopRelease{}, ErrDesktopReleaseAlreadyPublished
		}
		if len(catalog.Releases[i].Artifacts) == 0 {
			return DesktopRelease{}, ErrDesktopReleaseNoArtifacts
		}
		catalog.Releases[i].Status = "published"
		catalog.Releases[i].PublishedAt = &now
		catalog.Releases[i].UpdatedAt = now
		catalog.Releases[i].Actor = actor
		publishedIndex = i
		break
	}
	if publishedIndex < 0 {
		return DesktopRelease{}, ErrDesktopReleaseNotFound
	}

	published := catalog.Releases[publishedIndex]
	publishedReleases := make([]DesktopRelease, 0, len(catalog.Releases))
	for _, release := range catalog.Releases {
		if release.Status == "published" {
			publishedReleases = append(publishedReleases, release)
		}
	}
	sort.Slice(publishedReleases, func(i, j int) bool {
		return publishedReleases[i].PublishedAt.After(*publishedReleases[j].PublishedAt)
	})
	retained := make(map[string]struct{}, desktopReleaseRetention)
	for i := 0; i < min(desktopReleaseRetention, len(publishedReleases)); i++ {
		retained[publishedReleases[i].Version] = struct{}{}
	}
	removed := make([]string, 0)
	kept := catalog.Releases[:0]
	for _, release := range catalog.Releases {
		if release.Status != "published" {
			kept = append(kept, release)
			continue
		}
		if _, ok := retained[release.Version]; ok {
			kept = append(kept, release)
		} else {
			removed = append(removed, release.Version)
		}
	}
	catalog.Releases = kept
	if err := writeDesktopReleaseCatalog(catalog); err != nil {
		return DesktopRelease{}, err
	}

	root := DesktopReleaseStorageDir()
	trashRoot := filepath.Join(root, ".trash")
	for _, oldVersion := range removed {
		oldDir := desktopReleaseArtifactDir(root, oldVersion)
		if _, statErr := os.Stat(oldDir); errors.Is(statErr, os.ErrNotExist) {
			continue
		}
		if err := os.MkdirAll(trashRoot, 0o750); err != nil {
			common.SysLog("failed to create desktop release trash: " + err.Error())
			continue
		}
		trashDir := filepath.Join(trashRoot, filepath.Base(oldDir)+"-"+fmt.Sprintf("%d", now.UnixNano()))
		if err := os.Rename(oldDir, trashDir); err != nil {
			common.SysLog("failed to retire desktop release artifacts: " + err.Error())
			continue
		}
		if err := os.RemoveAll(trashDir); err != nil {
			common.SysLog("failed to delete retired desktop release artifacts: " + err.Error())
		}
	}
	return published, nil
}

func DesktopDownloadManifestView() (DesktopDownloadManifest, error) {
	view, err := ListDesktopReleases()
	if err != nil {
		return DesktopDownloadManifest{}, err
	}
	manifest := DesktopDownloadManifest{SchemaVersion: 1, Product: "RoboCoding"}
	artifacts := make(map[string]DesktopReleaseArtifact)
	if view.Latest != nil {
		manifest.Version = &view.Latest.Version
		manifest.PublishedAt = view.Latest.PublishedAt
		manifest.Changelog = &view.Latest.Changelog
		for _, artifact := range view.Latest.Artifacts {
			artifacts[artifact.Target] = artifact
		}
	}
	for _, target := range DesktopReleaseTargets {
		artifact, ok := artifacts[target]
		if !ok {
			manifest.Downloads = append(manifest.Downloads, DesktopDownloadTarget{ID: target, Status: "unavailable"})
			continue
		}
		url := "/release-artifacts/" + view.Latest.Version + "/" + target
		size := formatDesktopReleaseSize(artifact.SizeBytes)
		manifest.Downloads = append(manifest.Downloads, DesktopDownloadTarget{
			ID: target, Status: "available", URL: &url, FileName: &artifact.FileName, Size: &size, SHA256: &artifact.SHA256,
		})
	}
	return manifest, nil
}

func OpenDesktopReleaseArtifact(version, target string) (*os.File, DesktopReleaseArtifact, error) {
	desktopReleaseMu.Lock()
	defer desktopReleaseMu.Unlock()

	catalog, err := readDesktopReleaseCatalog()
	if err != nil {
		return nil, DesktopReleaseArtifact{}, err
	}
	for _, release := range catalog.Releases {
		if release.Version != version || release.Status != "published" {
			continue
		}
		for _, artifact := range release.Artifacts {
			if artifact.Target != target {
				continue
			}
			storedAs := artifact.StoredAs
			if storedAs == "" {
				storedAs = desktopReleaseStoredName(artifact)
			}
			file, err := os.Open(filepath.Join(desktopReleaseArtifactDir(DesktopReleaseStorageDir(), version), storedAs))
			return file, artifact, err
		}
	}
	return nil, DesktopReleaseArtifact{}, ErrDesktopReleaseNotFound
}

func desktopReleaseCatalogView(catalog desktopReleaseCatalog) DesktopReleaseCatalogView {
	releases := append([]DesktopRelease(nil), catalog.Releases...)
	sort.Slice(releases, func(i, j int) bool { return releases[i].UpdatedAt.After(releases[j].UpdatedAt) })
	view := DesktopReleaseCatalogView{Releases: releases}
	for i := range releases {
		if releases[i].Status != "published" {
			continue
		}
		if view.Latest == nil || releases[i].PublishedAt.After(*view.Latest.PublishedAt) {
			candidate := releases[i]
			view.Latest = &candidate
		}
	}
	return view
}

func readDesktopReleaseCatalog() (desktopReleaseCatalog, error) {
	catalog := desktopReleaseCatalog{SchemaVersion: desktopReleaseCatalogSchema, Releases: []DesktopRelease{}}
	data, err := os.ReadFile(filepath.Join(DesktopReleaseStorageDir(), "catalog.json"))
	if errors.Is(err, os.ErrNotExist) {
		return catalog, nil
	}
	if err != nil {
		return desktopReleaseCatalog{}, fmt.Errorf("read release catalog: %w", err)
	}
	if err := common.Unmarshal(data, &catalog); err != nil {
		return desktopReleaseCatalog{}, fmt.Errorf("decode release catalog: %w", err)
	}
	if catalog.SchemaVersion != desktopReleaseCatalogSchema {
		return desktopReleaseCatalog{}, fmt.Errorf("unsupported release catalog schema %d", catalog.SchemaVersion)
	}
	for _, release := range catalog.Releases {
		if !ValidateDesktopReleaseVersion(release.Version) || (release.Status != "draft" && release.Status != "published") {
			return desktopReleaseCatalog{}, fmt.Errorf("invalid release catalog entry")
		}
		if release.Status == "published" && release.PublishedAt == nil {
			return desktopReleaseCatalog{}, fmt.Errorf("published release is missing publication time")
		}
		seen := make(map[string]struct{}, len(release.Artifacts))
		for _, artifact := range release.Artifacts {
			if !IsDesktopReleaseTarget(artifact.Target) || len(artifact.SHA256) != 64 || artifact.SizeBytes <= 0 {
				return desktopReleaseCatalog{}, fmt.Errorf("invalid release artifact metadata")
			}
			if _, exists := seen[artifact.Target]; exists {
				return desktopReleaseCatalog{}, fmt.Errorf("duplicate release artifact target")
			}
			seen[artifact.Target] = struct{}{}
		}
	}
	return catalog, nil
}

func writeDesktopReleaseCatalog(catalog desktopReleaseCatalog) error {
	root := DesktopReleaseStorageDir()
	if err := os.MkdirAll(root, 0o750); err != nil {
		return fmt.Errorf("create release storage: %w", err)
	}
	data, err := common.Marshal(catalog)
	if err != nil {
		return fmt.Errorf("encode release catalog: %w", err)
	}
	temporary, err := os.CreateTemp(root, ".catalog-")
	if err != nil {
		return fmt.Errorf("create release catalog: %w", err)
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if err := temporary.Chmod(0o640); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("protect release catalog: %w", err)
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write release catalog: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("sync release catalog: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close release catalog: %w", err)
	}
	if err := os.Rename(temporaryName, filepath.Join(root, "catalog.json")); err != nil {
		return fmt.Errorf("commit release catalog: %w", err)
	}
	return nil
}

func desktopReleaseArtifactDir(root, version string) string {
	hash := sha256.Sum256([]byte(version))
	return filepath.Join(root, "artifacts", fmt.Sprintf("%x", hash[:16]))
}

func desktopReleaseStoredName(artifact DesktopReleaseArtifact) string {
	return artifact.Target + "-" + artifact.SHA256[:16] + strings.ToLower(filepath.Ext(artifact.FileName))
}

func formatDesktopReleaseSize(size int64) string {
	const (
		mib = int64(1024 * 1024)
		gib = int64(1024 * 1024 * 1024)
	)
	if size >= gib {
		return fmt.Sprintf("%.1f GB", float64(size)/float64(gib))
	}
	return fmt.Sprintf("%.1f MB", float64(size)/float64(mib))
}
