package router

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDesktopReleaseRoutesRequireServiceTokenAndPublishManifest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("ROBO_RELEASE_STORAGE_DIR", t.TempDir())
	t.Setenv("ROBO_RELEASE_MAX_FILE_BYTES", "1024")
	t.Setenv("ROBO_RELEASES_INTERNAL_TOKEN", "release-secret")
	router := gin.New()
	SetDesktopReleaseRouter(router)

	unauthorized := httptest.NewRecorder()
	router.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/v1/internal/releases", nil))
	assert.Equal(t, http.StatusUnauthorized, unauthorized.Code)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	require.NoError(t, writer.WriteField("version", "3.0.0"))
	require.NoError(t, writer.WriteField("changelog", "first public release"))
	file, err := writer.CreateFormFile("macos-arm64", "RoboCoding-3.0.0.dmg")
	require.NoError(t, err)
	_, err = file.Write([]byte("signed-package"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	upload := httptest.NewRequest(http.MethodPost, "/v1/internal/releases", &body)
	upload.Header.Set("Content-Type", writer.FormDataContentType())
	upload.Header.Set("X-Robo-Releases-Token", "release-secret")
	upload.Header.Set("X-Robo-Actor", `{"subject":"admin-1","displayName":"Release Admin"}`)
	uploadResponse := httptest.NewRecorder()
	router.ServeHTTP(uploadResponse, upload)
	require.Equal(t, http.StatusCreated, uploadResponse.Code, uploadResponse.Body.String())

	publish := httptest.NewRequest(http.MethodPost, "/v1/internal/releases/3.0.0/publish", nil)
	publish.Header.Set("X-Robo-Releases-Token", "release-secret")
	publish.Header.Set("X-Robo-Actor", `{"subject":"admin-1","displayName":"Release Admin"}`)
	publishResponse := httptest.NewRecorder()
	router.ServeHTTP(publishResponse, publish)
	require.Equal(t, http.StatusOK, publishResponse.Code, publishResponse.Body.String())

	manifestResponse := httptest.NewRecorder()
	router.ServeHTTP(manifestResponse, httptest.NewRequest(http.MethodGet, "/downloads.json", nil))
	require.Equal(t, http.StatusOK, manifestResponse.Code)
	assert.Contains(t, manifestResponse.Body.String(), `"version":"3.0.0"`)
	assert.Contains(t, manifestResponse.Body.String(), `"changelog":"first public release"`)
	assert.Contains(t, manifestResponse.Body.String(), `"id":"macos-arm64","status":"available"`)

	downloadResponse := httptest.NewRecorder()
	router.ServeHTTP(downloadResponse, httptest.NewRequest(http.MethodGet, "/release-artifacts/3.0.0/macos-arm64", nil))
	require.Equal(t, http.StatusOK, downloadResponse.Code)
	assert.Equal(t, "signed-package", downloadResponse.Body.String())
}
