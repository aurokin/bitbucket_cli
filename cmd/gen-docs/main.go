package main

import (
	"log"
	"os"
	"path/filepath"

	cmdpkg "github.com/auro/bitbucket_cli/internal/cmd"
)

func main() {
	content, err := cmdpkg.GenerateCLIReference()
	if err != nil {
		log.Fatal(err)
	}

	outputPath := filepath.Join("docs", "cli-reference.md")
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(outputPath, []byte(content), 0o644); err != nil {
		log.Fatal(err)
	}
}
