package bairro

import "strings"

type Bairro struct {
	ID              int    `json:"id" db:"id"`
	Nome            string `json:"nome" db:"nome"`
	NomeNormalizado string `json:"nomeNormalizado" db:"nome_normalizado"`
	CidadeID        int    `json:"cidadeId" db:"cidade_id"`
}

func (b *Bairro) TableName() string {
	return "bairros"
}

func (b *Bairro) NormalizarBairro() {
	// Normaliza o nome do bairro para facilitar buscas e comparações
	normalizado := b.Nome

	// Converte para minúsculas
	normalizado = strings.ToLower(normalizado)

	// Remove espaços extras
	normalizado = strings.TrimSpace(normalizado)

	// Remove acentos e caracteres especiais
	normalizado = strings.Map(func(r rune) rune {
		switch r {
		case 'á', 'à', 'ã', 'â', 'ä':
			return 'a'
		case 'é', 'è', 'ê', 'ë':
			return 'e'
		case 'í', 'ì', 'î', 'ï':
			return 'i'
		case 'ó', 'ò', 'õ', 'ô', 'ö':
			return 'o'
		case 'ú', 'ù', 'û', 'ü':
			return 'u'
		case 'ç':
			return 'c'
		default:
			return r
		}
	}, normalizado)

	b.NomeNormalizado = normalizado
}
