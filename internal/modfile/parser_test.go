package modfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFile(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		expectedErr  bool
		expectedPath string
	}{
		{
			name: "Valid module file",
			content: `module gno.land/r/dummy/foo

require (
	gno.land/p/demo/grc/grc20 v0.0.0-latest
	gno.land/p/demo/ownable v0.0.0-latest
	gno.land/p/demo/testutils v0.0.0-latest
	gno.land/p/demo/uassert v0.0.0-latest
	gno.land/p/demo/ufmt v0.0.0-latest
	gno.land/p/demo/users v0.0.0-latest
	gno.land/r/demo/users v0.0.0-latest
)
`,
			expectedErr:  false,
			expectedPath: "gno.land/r/dummy/foo",
		},
		{
			name: "Invalid module declaration",
			content: `invalid module declaration
require (
	gno.land/p/demo/grc/grc20 v0.0.0-latest
)
`,
			expectedErr: true,
		},
		{
			name: "Empty module declaration 2",
			content: `module gno.land /r/dummy/foo`,
			expectedErr: true,
		},
		{
			name: "No module declaration",
			content: `require (
	gno.land/p/demo/grc/grc20 v0.0.0-latest
)
`,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "gno-test")
			if err != nil {
				t.Fatalf("Failed to create temporary directory: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			tmpFile := filepath.Join(tmpDir, "gno.mod")
			err = os.WriteFile(tmpFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to write to temporary file: %v", err)
			}

			module, err := Parse(tmpFile)
			if tt.expectedErr {
				if err == nil {
					t.Error("Expected an error, but got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("ParseFile failed: %v", err)
				}
				if module.Path != tt.expectedPath {
					t.Errorf("Expected module path %s, but got %s", tt.expectedPath, module.Path)
				}
			}
		})
	}
}
