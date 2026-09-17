package service

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDesktopReleasePublishRetainsLatestThreeAndServesArtifacts(t *testing.T) {
	t.Setenv("ROBO_RELEASE_STORAGE_DIR", t.TempDir())
	t.Setenv("ROBO_RELEASE_MAX_FILE_BYTES", "1024")
	actor := DesktopReleaseActor{Subject: "admin-1", DisplayName: "Release Admin"}

	for _, version := range []string{"1.0.0", "1.1.0", "1.2.0", "1.3.0"} {
		content := "artifact-" + version
		draft, err := SaveDesktopReleaseDraft(version, "changes "+version, actor, []DesktopReleaseUpload{{
			Target: "windows-x64", FileName: "RoboCoding-" + version + ".exe", Reader: strings.NewReader(content),
		}})
		require.NoError(t, err)
		assert.Equal(t, "draft", draft.Status)
		published, err := PublishDesktopRelease(version, actor)
		require.NoError(t, err)
		assert.Equal(t, "published", published.Status)
	}

	view, err := ListDesktopReleases()
	require.NoError(t, err)
	require.NotNil(t, view.Latest)
	assert.Equal(t, "1.3.0", view.Latest.Version)
	assert.Len(t, view.Releases, 3)

	manifest, err := DesktopDownloadManifestView()
	require.NoError(t, err)
	require.NotNil(t, manifest.Version)
	require.NotNil(t, manifest.Changelog)
	assert.Equal(t, "1.3.0", *manifest.Version)
	assert.Equal(t, "changes 1.3.0", *manifest.Changelog)
	require.Len(t, manifest.Downloads, len(DesktopReleaseTargets))
	assert.Equal(t, "available", manifest.Downloads[0].Status)

	file, artifact, err := OpenDesktopReleaseArtifact("1.3.0", "windows-x64")
	require.NoError(t, err)
	defer file.Close()
	data, err := io.ReadAll(file)
	require.NoError(t, err)
	assert.Equal(t, "artifact-1.3.0", string(data))
	assert.Equal(t, "RoboCoding-1.3.0.exe", artifact.FileName)

	_, _, err = OpenDesktopReleaseArtifact("1.0.0", "windows-x64")
	assert.ErrorIs(t, err, ErrDesktopReleaseNotFound)
	_, statErr := os.Stat(desktopReleaseArtifactDir(DesktopReleaseStorageDir(), "1.0.0"))
	assert.ErrorIs(t, statErr, os.ErrNotExist)
}

func TestDesktopReleaseDraftRejectsPublishedReplacementAndOversizedArtifact(t *testing.T) {
	t.Setenv("ROBO_RELEASE_STORAGE_DIR", t.TempDir())
	t.Setenv("ROBO_RELEASE_MAX_FILE_BYTES", "4")
	actor := DesktopReleaseActor{Subject: "admin-1", DisplayName: "Release Admin"}

	_, err := SaveDesktopReleaseDraft("2.0.0", "changes", actor, []DesktopReleaseUpload{{
		Target: "linux-x64", FileName: "app.tar.gz", Reader: strings.NewReader("1234"),
	}})
	require.NoError(t, err)
	_, err = PublishDesktopRelease("2.0.0", actor)
	require.NoError(t, err)
	_, err = SaveDesktopReleaseDraft("2.0.0", "replacement", actor, []DesktopReleaseUpload{{
		Target: "linux-x64", FileName: "app.tar.gz", Reader: strings.NewReader("1234"),
	}})
	assert.ErrorIs(t, err, ErrDesktopReleaseAlreadyPublished)

	_, err = SaveDesktopReleaseDraft("2.1.0", "changes", actor, []DesktopReleaseUpload{{
		Target: "linux-x64", FileName: "app.tar.gz", Reader: strings.NewReader("12345"),
	}})
	require.Error(t, err)
}
