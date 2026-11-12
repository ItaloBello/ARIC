package gamificacao

import (
	"context"
	"log"
	"time"
)

// IniciarFechamentoLigas roda em segundo plano: ao subir o servidor e depois a cada hora,
// fecha a semana anterior das ligas se ela ainda não foi fechada.
//
// Rodar de hora em hora (em vez de agendar para segunda 00:00) faz o fechamento acontecer
// mesmo se o servidor estiver fora do ar na virada: ele fecha assim que voltar.
// Se dois servidores rodarem juntos, a função do banco usa um lock e recusa fechar
// a mesma semana duas vezes.
func IniciarFechamentoLigas(ctx context.Context, repo *Repository) {
	go func() {
		fecharSeNecessario(ctx, repo)

		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				fecharSeNecessario(ctx, repo)
			}
		}
	}()
}

func fecharSeNecessario(ctx context.Context, repo *Repository) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	semana, fechada, err := repo.SemanaParaFechar(ctx)
	if err != nil {
		log.Println("[ligas] erro ao verificar semana:", err)
		return
	}
	if fechada {
		return
	}

	n, err := repo.FecharSemana(ctx, semana)
	if err != nil {
		log.Printf("[ligas] erro ao fechar a semana de %s: %v", semana.Format("02/01/2006"), err)
		return
	}
	if n > 0 {
		log.Printf("[ligas] semana de %s fechada: %d usuários processados", semana.Format("02/01/2006"), n)
	}
}
