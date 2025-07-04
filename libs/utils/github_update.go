package utils

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// 返回 (needDownload bool, err error)
func CheckUpdate(repoURL, targetDir string, forceDownload bool) (needUpdate bool, repo string, release *GitHubRelease, err error) {
	owner, repo, err := parseGitHubRepo(repoURL)
	if err != nil {
		return false, repo, nil, fmt.Errorf("invalid GitHub repo URL: %v", err)
	}
	release, err = fetchLatestRelease(owner, repo)
	if err != nil {
		return false, repo, nil, fmt.Errorf("failed to fetch latest release: %v", err)
	}
	if forceDownload {
		return true, repo, release, nil
	}
	md5File := filepath.Join(targetDir, ".release_md5")

	if _, err := os.Stat(md5File); os.IsNotExist(err) {
		return true, repo, release, nil
	} else if err != nil {
		return false, repo, release, fmt.Errorf("failed to check local release info: %v", err)
	}

	existingTag, err := os.ReadFile(md5File)
	if err != nil {
		return false, repo, release, fmt.Errorf("failed to read local release info: %v", err)
	}

	return string(existingTag) != release.TagName, repo, release, nil
}

func DownloadAndExtractLatestRelease(repoURL, targetDir string, forceDownload bool) error {
	needUpdate, repo, release, err := CheckUpdate(repoURL, targetDir, forceDownload)
	if err != nil {
		return err
	}
	if !needUpdate {
		return nil
	}

	zipPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s-%s.zip", repo, release.TagName))
	if err := downloadFile(release.ZipballURL, zipPath); err != nil {
		return fmt.Errorf("failed to download release: %v", err)
	}
	defer os.Remove(zipPath)

	if err := extractZip(zipPath, targetDir); err != nil {
		return fmt.Errorf("failed to extract release: %v", err)
	}

	md5File := filepath.Join(targetDir, ".release_md5")
	if err := os.WriteFile(md5File, []byte(release.TagName), 0644); err != nil {
		return fmt.Errorf("failed to save release info: %v", err)
	}

	return nil
}

type GitHubRelease struct {
	TagName    string `json:"tag_name"`
	ZipballURL string `json:"zipball_url"`
}

func parseGitHubRepo(repoURL string) (string, string, error) {
	parts := strings.Split(strings.TrimPrefix(repoURL, "https://github.com/"), "/")
	if len(parts) < 2 {
		return "", "", errors.New("invalid GitHub repo URL format")
	}
	return parts[0], parts[1], nil
}

func fetchLatestRelease(owner, repo string) (*GitHubRelease, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	if release.ZipballURL == "" {
		return nil, errors.New("no ZIP download URL found")
	}

	return &release, nil
}

func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func extractZip(zipPath, targetDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	var rootDir string
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			parts := strings.Split(f.Name, "/")
			if len(parts) > 0 {
				rootDir = parts[0] + "/"
				break
			}
		}
	}

	for _, f := range r.File {
		if strings.Contains(f.Name, "__MACOSX") {
			continue
		}

		relPath := strings.TrimPrefix(f.Name, rootDir)
		if relPath == "" {
			continue
		}

		destPath := filepath.Join(targetDir, relPath)

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, f.Mode()); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		defer destFile.Close()

		srcFile, err := f.Open()
		if err != nil {
			return err
		}
		defer srcFile.Close()

		if _, err := io.Copy(destFile, srcFile); err != nil {
			return err
		}
	}
	return nil
}
