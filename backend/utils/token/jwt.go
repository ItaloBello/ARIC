package token

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const emissor = "aric-api"

type Claims struct {
	UsuarioID int `json:"uid"`
	jwt.RegisteredClaims
}

type Manager struct {
	segredo   []byte
	expiracao time.Duration
}

// NewManagerFromEnv lê JWT_SECRET (obrigatório, mínimo 32 caracteres)
// e JWT_EXPIRACAO_HORAS (opcional, padrão 24).
func NewManagerFromEnv() (*Manager, error) {
	segredo := os.Getenv("JWT_SECRET")
	if len(segredo) < 32 {
		return nil, errors.New("JWT_SECRET não definido ou com menos de 32 caracteres")
	}

	horas := 24
	if v := os.Getenv("JWT_EXPIRACAO_HORAS"); v != "" {
		h, err := strconv.Atoi(v)
		if err != nil || h <= 0 {
			return nil, fmt.Errorf("JWT_EXPIRACAO_HORAS inválido: %q", v)
		}
		horas = h
	}

	return &Manager{segredo: []byte(segredo), expiracao: time.Duration(horas) * time.Hour}, nil
}

func (m *Manager) Gerar(usuarioID int) (string, time.Time, error) {
	return m.GerarEm(usuarioID, time.Now())
}

// GerarEm emite o token com o instante de emissão informado. Usado na troca de senha:
// o mesmo instante vai para usuarios.sessoes_validas_desde, então o token novo vale
// e os anteriores não.
func (m *Manager) GerarEm(usuarioID int, agora time.Time) (string, time.Time, error) {
	expira := agora.Add(m.expiracao)

	claims := Claims{
		UsuarioID: usuarioID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    emissor,
			Subject:   strconv.Itoa(usuarioID),
			IssuedAt:  jwt.NewNumericDate(agora),
			ExpiresAt: jwt.NewNumericDate(expira),
		},
	}

	assinado, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.segredo)
	if err != nil {
		return "", time.Time{}, err
	}
	return assinado, expira, nil
}

func (m *Manager) Validar(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		return m.segredo, nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(emissor),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, err
	}
	if claims.UsuarioID <= 0 {
		return nil, errors.New("token sem usuário")
	}
	return claims, nil
}
