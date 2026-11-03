package bairro

import "testing"

func TestNormalizar(t *testing.T) {
	casos := map[string]string{
		"Jardim São Paulo":      "jardim sao paulo",
		"  JARDIM   SÃO PAULO ": "jardim sao paulo",
		"Vila Hortência":        "vila hortencia",
		"Éden":                  "eden",
	}
	for entrada, esperado := range casos {
		if got := Normalizar(entrada); got != esperado {
			t.Errorf("Normalizar(%q) = %q, esperado %q", entrada, got, esperado)
		}
	}
}
