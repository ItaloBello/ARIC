package denuncia

import (
	"context"
	"errors"
	"strings"
	"testing"

	"ARIC/utils/apperror"
)

// As validações acontecem antes de qualquer acesso ao banco, então dá para testar sem banco.
func TestDenunciaInvalida(t *testing.T) {
	u := &UseCases{}
	casos := map[string]DenunciaRequest{
		"tipo inválido":       {AlvoTipo: "usuario", AlvoID: 1, Motivo: "spam"},
		"sem id":              {AlvoTipo: "projeto", Motivo: "spam"},
		"motivo inválido":     {AlvoTipo: "projeto", AlvoID: 1, Motivo: "chato"},
		"outro sem descrição": {AlvoTipo: "forum_post", AlvoID: 1, Motivo: "outro", Descricao: "   "},
		"descrição longa":     {AlvoTipo: "forum_post", AlvoID: 1, Motivo: "spam", Descricao: strings.Repeat("a", 501)},
	}
	for nome, req := range casos {
		_, err := u.Denunciar(context.Background(), 1, req)
		var appErr *apperror.AppError
		if !errors.As(err, &appErr) || appErr.StatusCode != 400 {
			t.Errorf("%s: esperava 400, veio %v", nome, err)
		}
	}
}

func TestResolverInvalido(t *testing.T) {
	u := &UseCases{}
	casos := map[string]ResolverRequest{
		"sem alvo":      {Acao: "remover"},
		"ação inválida": {AlvoTipo: "projeto", AlvoID: 1, Acao: "apagar"},
		"motivo longo":  {AlvoTipo: "projeto", AlvoID: 1, Acao: "remover", Motivo: strings.Repeat("a", 501)},
	}
	for nome, req := range casos {
		err := u.Resolver(context.Background(), 1, req)
		var appErr *apperror.AppError
		if !errors.As(err, &appErr) || appErr.StatusCode != 400 {
			t.Errorf("%s: esperava 400, veio %v", nome, err)
		}
	}
}
