package acesso

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"ARIC/utils/apperror"
)

func TestBloqueio(t *testing.T) {
	if err := bloqueio(sql.NullTime{}, sql.NullTime{}); err != nil {
		t.Errorf("sem liberação pendente não deveria bloquear: %v", err)
	}
	if err := bloqueio(sql.NullTime{Valid: true, Time: time.Now().Add(-time.Minute)}); err != nil {
		t.Errorf("liberação no passado não deveria bloquear: %v", err)
	}

	// usa a liberação mais distante (e-mail 3 min, IP 12 min -> 12 min)
	err := bloqueio(
		sql.NullTime{Valid: true, Time: time.Now().Add(3 * time.Minute)},
		sql.NullTime{Valid: true, Time: time.Now().Add(12*time.Minute - time.Second)},
	)
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("esperava AppError, veio %v", err)
	}
	if appErr.StatusCode != 429 {
		t.Errorf("status = %d, esperava 429", appErr.StatusCode)
	}
	if appErr.Message != "muitas tentativas, tente de novo em 12 minutos" {
		t.Errorf("mensagem inesperada: %q", appErr.Message)
	}
	if appErr.RetryAfter < 710 || appErr.RetryAfter > 720 {
		t.Errorf("Retry-After = %d, esperava ~719", appErr.RetryAfter)
	}

	err = bloqueio(sql.NullTime{Valid: true, Time: time.Now().Add(20 * time.Second)})
	if !errors.As(err, &appErr) || appErr.Message != "muitas tentativas, tente de novo em 1 minuto" {
		t.Errorf("menos de 1 minuto deveria dizer '1 minuto', veio %v", err)
	}
}

func TestLimparIP(t *testing.T) {
	casos := map[string]any{
		"200.1.1.1":       "200.1.1.1",
		" 10.0.0.1 ":      "10.0.0.1",
		"2001:db8::1":     "2001:db8::1",
		"::ffff:10.0.0.1": "10.0.0.1",
		"":                nil,
		"não-é-ip":        nil,
		"10.0.0.1; DROP":  nil,
	}
	for entrada, esperado := range casos {
		if got := limparIP(entrada); got != esperado {
			t.Errorf("limparIP(%q) = %v, esperava %v", entrada, got, esperado)
		}
	}
	if normalizarEmail("  Ana@X.com ") != "ana@x.com" {
		t.Error("e-mail deveria ser normalizado")
	}
	if intervalo(15*time.Minute) != "900 seconds" {
		t.Errorf("intervalo inesperado: %s", intervalo(15*time.Minute))
	}
}
