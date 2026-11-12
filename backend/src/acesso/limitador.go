// Package acesso limita tentativas de login, de confirmação de senha e de cadastro.
//
// As tentativas ficam em aric.tentativas_acesso (e não na memória), então o limite
// sobrevive a reinícios e vale para todas as instâncias da API.
package acesso

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"net"
	"strings"
	"time"

	"ARIC/utils/apperror"

	"github.com/jmoiron/sqlx"
)

// Tipos de tentativa (coluna tipo)
const (
	Login    = "login"    // POST /auth/login
	Senha    = "senha"    // senha digitada para trocar a senha ou excluir a conta
	Cadastro = "cadastro" // POST /auth/registro
)

// Regras. Login e confirmação de senha contam juntos por e-mail.
const (
	janelaSenha       = 15 * time.Minute
	maxFalhasEmail    = 5  // por conta: protege contra adivinhar a senha de alguém
	maxFalhasIP       = 50 // por IP: alto porque escolas e laboratórios saem pelo mesmo IP
	janelaCadastro    = time.Hour
	maxCadastrosIP    = 20
	retencaoRegistros = 30 * 24 * time.Hour
)

// Liberação do e-mail: quando a falha mais antiga entre as últimas N sai da janela.
// Falhas anteriores ao último acerto não contam (acertar a senha zera o contador).
const qLiberacaoEmail = `
SELECT min(criado_em) + $2::interval
FROM (
    SELECT criado_em FROM aric.tentativas_acesso
    WHERE email = $1 AND NOT sucesso AND tipo IN ('login', 'senha')
      AND criado_em > now() - $2::interval
      AND criado_em > COALESCE((SELECT max(criado_em) FROM aric.tentativas_acesso
                                WHERE email = $1 AND sucesso AND tipo IN ('login', 'senha')),
                               '-infinity')
    ORDER BY criado_em DESC
    LIMIT $3
) f
HAVING count(*) >= $3`

const qLiberacaoIP = `
SELECT min(criado_em) + $2::interval
FROM (
    SELECT criado_em FROM aric.tentativas_acesso
    WHERE ip = $1::inet AND NOT sucesso AND tipo = 'login'
      AND criado_em > now() - $2::interval
    ORDER BY criado_em DESC
    LIMIT $3
) f
HAVING count(*) >= $3`

const qLiberacaoCadastro = `
SELECT min(criado_em) + $2::interval
FROM (
    SELECT criado_em FROM aric.tentativas_acesso
    WHERE ip = $1::inet AND tipo = 'cadastro' AND sucesso
      AND criado_em > now() - $2::interval
    ORDER BY criado_em DESC
    LIMIT $3
) f
HAVING count(*) >= $3`

const qRegistrar = `
INSERT INTO aric.tentativas_acesso (tipo, email, ip, sucesso)
VALUES ($1, NULLIF($2, ''), $3::inet, $4)`

const qLimpar = `DELETE FROM aric.tentativas_acesso WHERE criado_em < now() - $1::interval`

type Limitador struct {
	db *sqlx.DB
}

func NewLimitador(db *sqlx.DB) *Limitador {
	return &Limitador{db: db}
}

// VerificarSenha deve ser chamado ANTES de conferir a senha (login ou confirmação).
// Devolve 429 se o e-mail ou o IP estiverem bloqueados. Com ip vazio, só olha o e-mail.
func (l *Limitador) VerificarSenha(ctx context.Context, email, ip string) error {
	email = normalizarEmail(email)
	liberaEmail, err := l.liberacao(ctx, qLiberacaoEmail, email, janelaSenha, maxFalhasEmail)
	if err != nil {
		return err
	}
	var liberaIP sql.NullTime
	if ipValido := limparIP(ip); ipValido != nil {
		if liberaIP, err = l.liberacao(ctx, qLiberacaoIP, ipValido, janelaSenha, maxFalhasIP); err != nil {
			return err
		}
	}
	return bloqueio(liberaEmail, liberaIP)
}

// RegistrarSenha anota o resultado de uma conferência de senha.
// Acertar zera o contador do e-mail.
func (l *Limitador) RegistrarSenha(ctx context.Context, tipo, email, ip string, sucesso bool) {
	l.registrar(ctx, tipo, normalizarEmail(email), ip, sucesso)
}

// VerificarCadastro limita cadastros por IP (sem IP válido, não limita).
func (l *Limitador) VerificarCadastro(ctx context.Context, ip string) error {
	ipValido := limparIP(ip)
	if ipValido == nil {
		return nil
	}
	libera, err := l.liberacao(ctx, qLiberacaoCadastro, ipValido, janelaCadastro, maxCadastrosIP)
	if err != nil {
		return err
	}
	return bloqueio(libera, sql.NullTime{})
}

func (l *Limitador) RegistrarCadastro(ctx context.Context, email, ip string) {
	l.registrar(ctx, Cadastro, normalizarEmail(email), ip, true)
}

func (l *Limitador) liberacao(ctx context.Context, query string, chave any, janela time.Duration, max int) (sql.NullTime, error) {
	var libera sql.NullTime
	err := l.db.GetContext(ctx, &libera, query, chave, intervalo(janela), max)
	if errors.Is(err, sql.ErrNoRows) {
		return sql.NullTime{}, nil
	}
	if err != nil {
		log.Println("[acesso] erro ao verificar tentativas:", err)
		return sql.NullTime{}, apperror.Internal("erro ao verificar tentativas")
	}
	return libera, nil
}

// registrar não falha a requisição: se o banco recusar, só fica no log.
func (l *Limitador) registrar(ctx context.Context, tipo, email, ip string, sucesso bool) {
	if _, err := l.db.ExecContext(ctx, qRegistrar, tipo, email, limparIP(ip), sucesso); err != nil {
		log.Println("[acesso] erro ao registrar tentativa:", err)
	}
}

// IniciarLimpeza apaga, de hora em hora, registros com mais de 30 dias.
func (l *Limitador) IniciarLimpeza(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for {
			if _, err := l.db.ExecContext(ctx, qLimpar, intervalo(retencaoRegistros)); err != nil {
				log.Println("[acesso] erro ao limpar tentativas antigas:", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

// ---------- helpers ----------

// bloqueio devolve 429 com o tempo de espera, usando a liberação mais distante.
func bloqueio(liberacoes ...sql.NullTime) error {
	var ate time.Time
	for _, l := range liberacoes {
		if l.Valid && l.Time.After(ate) {
			ate = l.Time
		}
	}
	espera := time.Until(ate)
	if espera <= 0 {
		return nil
	}
	minutos := int(math.Ceil(espera.Minutes()))
	msg := "muitas tentativas, tente de novo em 1 minuto"
	if minutos > 1 {
		msg = fmt.Sprintf("muitas tentativas, tente de novo em %d minutos", minutos)
	}
	return apperror.TooManyRequests(msg, int(math.Ceil(espera.Seconds())))
}

func intervalo(d time.Duration) string {
	return fmt.Sprintf("%d seconds", int(d.Seconds()))
}

func normalizarEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// limparIP devolve o IP em texto, ou nil se não for um IP válido.
func limparIP(ip string) any {
	parsed := net.ParseIP(strings.TrimSpace(ip))
	if parsed == nil {
		return nil
	}
	return parsed.String()
}
