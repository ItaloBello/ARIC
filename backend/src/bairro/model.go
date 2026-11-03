package bairro

import "strings"

type Bairro struct {
	ID              int    `json:"id" db:"id"`
	Nome            string `json:"nome" db:"nome"`
	NomeNormalizado string `json:"nomeNormalizado" db:"nome_normalizado"`
	CidadeID        int    `json:"cidadeId" db:"cidade_id"`
}

// Normalizar deixa o nome em minúsculas, sem acentos e sem espaços repetidos,
// para que "Jd. São Paulo" e "jd. sao  paulo" caiam no mesmo bairro.
func Normalizar(nome string) string {
	n := strings.ToLower(nome)
	n = strings.Join(strings.Fields(n), " ")

	return strings.Map(func(r rune) rune {
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
		case 'ñ':
			return 'n'
		default:
			return r
		}
	}, n)
}
