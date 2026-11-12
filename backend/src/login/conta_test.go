package login

import (
	"os"
	"path/filepath"
	"testing"
)

// apagarUpload aceita qualquer pasta dentro de /uploads, mas nada fora dela.
func TestApagarUploadSoDentroDeUploads(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "uploads", "forum"), 0o755); err != nil {
		t.Fatal(err)
	}
	dentro := filepath.Join(dir, "uploads", "forum", "a.jpg")
	fora := filepath.Join(dir, "segredo.txt")
	for _, f := range []string{dentro, fora} {
		if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(dir)

	for _, u := range []string{"/segredo.txt", "/uploads/../segredo.txt", "uploads/forum/a.jpg", "/uploads//forum/a.jpg", ""} {
		apagarUpload(u)
	}
	if _, err := os.Stat(fora); err != nil {
		t.Error("apagarUpload removeu arquivo fora de /uploads")
	}
	if _, err := os.Stat(dentro); err != nil {
		t.Error("caminho não normalizado não deveria ter sido aceito")
	}

	apagarUpload("/uploads/forum/a.jpg")
	if _, err := os.Stat(dentro); !os.IsNotExist(err) {
		t.Error("apagarUpload deveria ter removido /uploads/forum/a.jpg")
	}
}
