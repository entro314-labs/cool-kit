package templates

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed data/*
var templatesFS embed.FS

// GetTemplatesFS returns the embedded filesystem
func GetTemplatesFS() embed.FS {
	return templatesFS
}

// WriteTemplate writes a specific template file to the target path
func WriteTemplate(templateName, targetPath string) error {
	content, err := templatesFS.ReadFile("data/" + templateName)
	if err != nil {
		return fmt.Errorf("failed to read template %s: %w", templateName, err)
	}

	if err := os.WriteFile(targetPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", targetPath, err)
	}

	return nil
}

// ListTemplates lists all available templates
func ListTemplates() ([]string, error) {
	var files []string
	err := fs.WalkDir(templatesFS, "data", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			rel, _ := filepath.Rel("data", path)
			files = append(files, rel)
		}
		return nil
	})
	return files, err
}
