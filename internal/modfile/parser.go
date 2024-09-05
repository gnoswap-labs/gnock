package modfile

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Module struct {
	Path string
}

func Parse(path string) (*Module, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// module gno.land/{p,r}/...
		if strings.HasPrefix(line, "module") {
			parts := strings.Fields(line)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid module declaration: %s", line)
			}
			return &Module{Path: parts[1]}, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("module declaration not found")
}
