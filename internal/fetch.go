package internal

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gnoswap-labs/gnock/internal/modfile"
)

const gnoModFilename = "gno.mod"

var (
	ErrInvalidURL = errors.New("invalid URL")
)

var execCommand = executeGitCommand

func executeGitCommand(clone, url, dir string) error {
	cmd := exec.Command("git", clone, url, dir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %v", err)
	}
	return nil
}

// GetPackage fetches a package from the given URL (e.g. github.com/username/repo)
// and copies it to the target directory which declared in the gno.mod file.
func GetPackage(url string) error {
	// TODO: check valid URL
	parts := strings.Split(url, "/")

	// pats are contains at least 3 parts
	// ex: github.com/username/repo/...
	//     |---------|--------|-----|
	//       initial   owner    repo
	if len(parts) < 3 {
		return ErrInvalidURL
	}

	tempDir, err := os.MkdirTemp("", "gnock-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// TODO: use universal method to clone repository
	err = execCommand("clone", url, tempDir)
	if err != nil {
		return fmt.Errorf("failed to clone repository: %v", err)
	}

	return processDirectory(tempDir, "")
}

func processDirectory(dir, relpath string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to rea directory %s: %v", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			err := processDirectory(filepath.Join(dir, entry.Name()), filepath.Join(relpath, entry.Name()))
			if err != nil {
				return err
			}
		} else if entry.Name() == gnoModFilename {
			gnoModPath := filepath.Join(dir, gnoModFilename)
			module, err := modfile.Parse(gnoModPath)
			if err != nil {
				return fmt.Errorf("failed to parse gno.mod file in %s: %v", relpath, err)
			}

			// TODO: the name of the `gno` directory can be different for each user.
			// so need to set it as a variable and update via the CLI or find it automatically
			destDir := filepath.Join("gno", "examples", module.Path)
			if err := os.MkdirAll(destDir, 0755); err != nil {
				return fmt.Errorf("failed to create destination directory %s: %v", destDir, err)
			}

			if err := copyDir(dir, destDir); err != nil {
				return fmt.Errorf("failed to copy directory %s to %s: %v", dir, destDir, err)
			}

			fmt.Printf("Package %s installed successfully to %s\n", module.Path, destDir)
		}
	}

	return nil
}

func copyDir(src string, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(dstPath, data, info.Mode())
	})
}
