// Package updater keeps the asaguard binary up to date automatically.
//
// On each run asaguard checks the latest release on GitHub and, if a newer
// version is available, downloads the correct binary for the current OS and
// architecture, replaces the running executable in place, and re-execs into the
// new version — all transparently. Dev builds (version == "dev") skip the check
// entirely. This ensures that guard rail definitions and detection logic stay
// current without requiring manual update steps from engineers.
package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

const repo = "kaihendry/asaguard"

type release struct {
	TagName string  `json:"tag_name"`
	Assets  []asset `json:"assets"`
}

type asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// CheckAndUpdate is a no-op for dev builds. For release builds it fetches the
// latest GitHub release and, if newer, replaces the running binary and re-execs.
func CheckAndUpdate(version string) {
	if version == "dev" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rel, err := latestRelease(ctx)
	if err != nil || rel.TagName <= version {
		return
	}

	assetName := fmt.Sprintf("asaguard_%s_%s", runtime.GOOS, runtime.GOARCH)
	var downloadURL string
	for _, a := range rel.Assets {
		if strings.HasPrefix(a.Name, assetName) {
			downloadURL = a.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		return
	}

	if err := replaceAndExec(downloadURL); err != nil {
		// silently swallow — run with current binary
		_ = err
	}
}

func latestRelease(ctx context.Context) (*release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api: %s", resp.Status)
	}

	var rel release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

func replaceAndExec(url string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	resp, err := http.Get(url) //nolint:gosec // URL comes from GitHub API response
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	tmp, err := os.CreateTemp("", "asaguard-update-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() {
		tmp.Close()
		os.Remove(tmpPath)
	}()

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		return err
	}
	tmp.Close()

	if err := os.Chmod(tmpPath, 0o755); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, exe); err != nil {
		return err
	}

	return reexec(exe)
}
