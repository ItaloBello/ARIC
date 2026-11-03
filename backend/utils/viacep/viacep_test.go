package viacep

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func servidorFalso(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/18035000/json/":
			w.Write([]byte(`{"cep":"18035-000","logradouro":"Rua X","bairro":"Centro","localidade":"Sorocaba","uf":"SP"}`))
		case "/99999999/json/":
			w.Write([]byte(`{"erro": true}`))
		case "/88888888/json/":
			w.Write([]byte(`{"erro": "true"}`))
		case "/77777777/json/":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestBuscar(t *testing.T) {
	c := NewClient(servidorFalso(t).URL)
	ctx := context.Background()

	end, err := c.Buscar(ctx, "18035-000")
	if err != nil {
		t.Fatal(err)
	}
	if end.Cep != "18035000" || end.Localidade != "Sorocaba" || end.UF != "SP" || end.Bairro != "Centro" {
		t.Fatalf("endereço inesperado: %+v", end)
	}

	if _, err := c.Buscar(ctx, "123"); !errors.Is(err, ErrCepInvalido) {
		t.Errorf("CEP curto: esperava ErrCepInvalido, veio %v", err)
	}
	if _, err := c.Buscar(ctx, "99999999"); !errors.Is(err, ErrCepNaoEncontrado) {
		t.Errorf("erro bool: esperava ErrCepNaoEncontrado, veio %v", err)
	}
	if _, err := c.Buscar(ctx, "88888888"); !errors.Is(err, ErrCepNaoEncontrado) {
		t.Errorf("erro string: esperava ErrCepNaoEncontrado, veio %v", err)
	}
	if _, err := c.Buscar(ctx, "77777777"); err == nil || errors.Is(err, ErrCepNaoEncontrado) {
		t.Errorf("status 500: esperava erro de serviço, veio %v", err)
	}
}
