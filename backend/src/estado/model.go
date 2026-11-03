package estado

type Estado struct {
	ID    int    `json:"id" db:"id"`
	Nome  string `json:"nome" db:"nome"`
	Sigla string `json:"sigla" db:"sigla"`
}
