package gamificacao

import "testing"

func TestProgressoNivel(t *testing.T) {
	semente := &Nivel{Numero: 1, Titulo: "Semente", MinimoXP: 0}
	broto := &Nivel{Numero: 2, Titulo: "Broto", MinimoXP: 100}
	separador := &Nivel{Numero: 3, Titulo: "Separador", MinimoXP: 300}
	lenda := &Nivel{Numero: 10, Titulo: "Lenda", MinimoXP: 10000}

	casos := []struct {
		nome          string
		xp            int
		atual, prox   *Nivel
		falta, porcem int
	}{
		{"início", 0, semente, broto, 100, 0},
		{"metade do nível 1", 50, semente, broto, 50, 50},
		{"acabou de subir", 100, broto, separador, 200, 0},
		{"3/4 do nível 2", 250, broto, separador, 50, 75},
		{"nível máximo", 12000, lenda, nil, 0, 100},
		{"sem nível atual", 30, nil, broto, 70, 30},
	}
	for _, c := range casos {
		falta, porcem := progressoNivel(c.xp, c.atual, c.prox)
		if falta != c.falta || porcem != c.porcem {
			t.Errorf("%s: esperava falta=%d %d%%, veio falta=%d %d%%", c.nome, c.falta, c.porcem, falta, porcem)
		}
	}
}
