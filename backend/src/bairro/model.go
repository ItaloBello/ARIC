package bairro

type Bairro struct {
	ID int `json:"id" db:"id"`
	Nome string `json:"nome" db:"nome"`
	CidadeID int `json:"cidadeId" db:"cidade_id"`
}

func (b *Bairro) TableName() string {
	return "bairro"
}