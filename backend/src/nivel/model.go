package nivel

type Nivel struct {
	ID       int    `json:"id" db:"id"`
	Numero   int    `json:"numero" db:"numero"`
	Titulo   string `json:"titulo" db:"titulo"`
	MinimoXP int    `json:"minimoXp" db:"minimo_xp"`
}

func (n *Nivel) TableName() string {
	return "nivel"
}