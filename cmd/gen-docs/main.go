package main

import (
	"log"
	"os"
	"path/filepath"

	cmdpkg "github.com/auro/bitbucket_cli/internal/cmd"
)

func main() {
	writeGeneratedDoc(filepath.Join("docs", "cli-reference.md"), cmdpkg.GenerateCLIReference)
	writeGeneratedDoc(filepath.Join("docs", "flag-matrix.md"), cmdpkg.GenerateFlagMatrixDoc)
	writeGeneratedDoc(filepath.Join("docs", "error-index.md"), cmdpkg.GenerateErrorIndexDoc)
	writeGeneratedDoc(filepath.Join("docs", "json-shapes.md"), cmdpkg.GenerateJSONShapesDoc)
	writeGeneratedDoc(filepath.Join("docs", "json-fields.md"), cmdpkg.GenerateJSONFieldsDoc)
	writeGeneratedDoc(filepath.Join("docs", "recovery.md"), cmdpkg.GenerateRecoveryDoc)
}

func writeGeneratedDoc(path string, generate func() (string, error)) {
	content, err := generate()
	if err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		log.Fatal(err)
	}
}
