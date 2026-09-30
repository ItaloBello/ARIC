package cidade

type Cidade struct {
	ID int `json:"id" db:"id"`
	Nome string `json:"nome" db:"nome"`
	EstadoID int `json:"estadoId" db:"estado_id"`
}

func (c *Cidade) TableName() string {
	return "cidade"
}
