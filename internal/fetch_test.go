package internal

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

var mockExecCommand func(clone, url, dir string) error

func init() {
	execCommand = func(clone, url, dir string) error {
		return mockExecCommand(clone, url, dir)
	}
}

func TestGetPackage(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		mockGit func(clone, url, dir string) error
		wantErr bool
	}{
		{
			name: "Valid URL and successful clone",
			url:  "github.com/username/repo",
			mockGit: func(clone, url, dir string) error {
				gnoModPath := filepath.Join(dir, "gno.mod")
				content := []byte("module gno.land/r/demo/mypackage")
				if err := os.WriteFile(gnoModPath, content, 0644); err != nil {
					return err
				}
				return nil
			},
			wantErr: false,
		},
		{
			name:    "Invalid URL",
			url:     "invalid-url",
			mockGit: func(clone, url, dir string) error { return nil },
			wantErr: true,
		},
		{
			name: "Git clone fails",
			url:  "github.com/username/repo",
			mockGit: func(clone, url, dir string) error {
				return errors.New("git clone failed")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecCommand = tt.mockGit

			tmpDir, err := os.MkdirTemp("", "gnock-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// change working directory to temp directory
			oldWd, _ := os.Getwd()
			os.Chdir(tmpDir)
			defer os.Chdir(oldWd)

			err = GetPackage(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPackage() error = %v, wantErr %v", err, tt.wantErr)
			}

			// check if the package was installed correctly
			if !tt.wantErr {
				destDir := filepath.Join("gno", "examples", "gno.land", "r", "demo", "mypackage")
				if _, err := os.Stat(destDir); os.IsNotExist(err) {
					t.Errorf("Expected directory %s was not created", destDir)
				}
			}
		})
	}
}

func TestGetPackageComplexStructure(t *testing.T) {
	mockRepoDir, err := os.MkdirTemp("", "mock-repo-*")
	if err != nil {
		t.Fatalf("Failed to create mock repo dir: %v", err)
	}
	defer os.RemoveAll(mockRepoDir)

	dirs := []string{
		"example_pkg/nested1",
		"example_pkg/nested2",
	}
	for _, dir := range dirs {
		err := os.MkdirAll(filepath.Join(mockRepoDir, dir), 0755)
		if err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// mock files
	files := map[string]string{
		"example_pkg/nested1/foo.gno":      "// foo.gno content",
		"example_pkg/nested1/foo_test.gno": "// foo_test.gno content",
		"example_pkg/nested1/gno.mod":      "module gno.land/p/demo/nested1",
		"example_pkg/nested2/bar.gno":      "// bar.gno content",
		"example_pkg/nested2/bar_test.gno": "// bar_test.gno content",
		"example_pkg/nested2/gno.mod":      "module gno.land/r/nested2",
	}
	for filePath, content := range files {
		err := os.WriteFile(filepath.Join(mockRepoDir, filePath), []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", filePath, err)
		}
	}

	mockExecCommand = func(clone, url, dir string) error {
		return copyDir(mockRepoDir, dir)
	}

	// temporary directory for the test output
	tmpDir, err := os.MkdirTemp("", "gnock-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	err = GetPackage("github.com/example/repo")
	if err != nil {
		t.Fatalf("GetPackage failed: %v", err)
	}

	// Check if the package was installed correctly
	expectedStructure := map[string]bool{
		"gno/examples/gno.land/p/demo/nested1/foo.gno":      true,
		"gno/examples/gno.land/p/demo/nested1/foo_test.gno": true,
		"gno/examples/gno.land/p/demo/nested1/gno.mod":      true,
		"gno/examples/gno.land/r/nested2/bar.gno":           true,
		"gno/examples/gno.land/r/nested2/bar_test.gno":      true,
		"gno/examples/gno.land/r/nested2/gno.mod":           true,
	}

	err = filepath.Walk(filepath.Join(tmpDir, "gno"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(tmpDir, path)
		if err != nil {
			return err
		}
		if _, expected := expectedStructure[relPath]; !expected {
			t.Errorf("Unexpected file: %s", relPath)
		} else {
			delete(expectedStructure, relPath)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk directory structure: %v", err)
	}

	for missingFile := range expectedStructure {
		t.Errorf("Expected file not found: %s", missingFile)
	}

	// verify content of gno.mod files
	gnoModPaths := []struct {
		path    string
		content string
	}{
		{"gno/examples/gno.land/p/demo/nested1/gno.mod", "module gno.land/p/demo/nested1"},
		{"gno/examples/gno.land/r/nested2/gno.mod", "module gno.land/r/nested2"},
	}

	for _, gnoMod := range gnoModPaths {
		content, err := os.ReadFile(filepath.Join(tmpDir, gnoMod.path))
		if err != nil {
			t.Errorf("Failed to read gno.mod file %s: %v", gnoMod.path, err)
		} else if string(content) != gnoMod.content {
			t.Errorf("Incorrect content in gno.mod file %s. Expected: %s, Got: %s", gnoMod.path, gnoMod.content, string(content))
		}
	}
}

func TestCopyDir(t *testing.T) {
	srcDir, err := os.MkdirTemp("", "src-*")
	if err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}
	defer os.RemoveAll(srcDir)

	// create some files in the source directory
	files := []string{"file1.txt", "file2.txt", "subdir/file3.txt"}
	for _, file := range files {
		path := filepath.Join(srcDir, file)
		err := os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		err = os.WriteFile(path, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}

	// temporary destination directory
	dstDir, err := os.MkdirTemp("", "dst-*")
	if err != nil {
		t.Fatalf("Failed to create destination dir: %v", err)
	}
	defer os.RemoveAll(dstDir)

	err = copyDir(srcDir, dstDir)
	if err != nil {
		t.Fatalf("copyDir failed: %v", err)
	}

	// check if all files were copied correctly
	for _, file := range files {
		srcPath := filepath.Join(srcDir, file)
		dstPath := filepath.Join(dstDir, file)

		srcInfo, err := os.Stat(srcPath)
		if err != nil {
			t.Fatalf("Failed to stat source file: %v", err)
		}

		dstInfo, err := os.Stat(dstPath)
		if err != nil {
			t.Fatalf("Failed to stat destination file: %v", err)
		}

		if srcInfo.Mode() != dstInfo.Mode() {
			t.Errorf("File mode mismatch for %s", file)
		}

		srcContent, err := os.ReadFile(srcPath)
		if err != nil {
			t.Fatalf("Failed to read source file: %v", err)
		}

		dstContent, err := os.ReadFile(dstPath)
		if err != nil {
			t.Fatalf("Failed to read destination file: %v", err)
		}

		if string(srcContent) != string(dstContent) {
			t.Errorf("File content mismatch for %s", file)
		}
	}
}
