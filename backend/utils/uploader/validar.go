package uploader

import (
	"fmt"
	"os"
	"path"
	"strings"
)

// URLLocal confere se a URL aponta para um arquivo enviado com essa config
// (ex.: "/uploads/projetos/123.jpg"). Impede referenciar sites externos,
// outras pastas ("../") ou arquivos que não existem.
func URLLocal(u string, cfg UploadConfig) error {
	base := strings.TrimSuffix(cfg.BaseURL, "/")
	nome := strings.TrimPrefix(u, base+"/")
	if nome == u || nome == "" || nome != path.Base(nome) || strings.HasPrefix(nome, ".") {
		return fmt.Errorf("imagem inválida: %s (envie a imagem antes)", u)
	}
	if _, err := os.Stat(path.Join(cfg.DestFolder, nome)); err != nil {
		return fmt.Errorf("imagem não encontrada: %s", u)
	}
	return nil
}
