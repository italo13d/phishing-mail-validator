package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type FS struct {
	rawDir string
	anDir  string
}

func NewFS(raw, analysis string) *FS {
	_ = os.MkdirAll(raw, 0o755)
	_ = os.MkdirAll(analysis, 0o755)
	return &FS{raw, analysis}
}

func (fs *FS) SaveRaw(idx int, mail any) error {
	f, err := os.Create(filepath.Join(fs.rawDir, fmt.Sprintf("email_%02d.json", idx)))
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(mail)
}

func (fs *FS) SaveAnalysis(idx int, txt string) error {
	name := filepath.Join(fs.anDir, fmt.Sprintf("email_%02d_analysis.txt", idx))
	return os.WriteFile(name, []byte(txt), 0o644)
}
