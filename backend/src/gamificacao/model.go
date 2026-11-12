package gamificacao

import (
	"database/sql"
	"time"

	"ARIC/src/missao"
)

const AcaoRegistroManual = "registro_manual"

// ---------- perfil ----------

type Nivel struct {
	Numero   int    `json:"numero"`
	Titulo   string `json:"titulo"`
	MinimoXP int    `json:"minimoXp"`
}

type LigaResumo struct {
	ID       int     `json:"id"`
	Nome     string  `json:"nome"`
	Ordem    int     `json:"ordem"`
	Icone    *string `json:"icone"`
	Posicao  *int    `json:"posicao"` // posição na liga nesta semana
	XPSemana int     `json:"xpSemana"`
}

type Impacto struct {
	ResiduosKg         float64 `json:"residuosKg"`         // "kg de resíduos reutilizados"
	ProjetosConcluidos int     `json:"projetosConcluidos"` // projetos que o usuário fez
	ProjetosPublicados int     `json:"projetosPublicados"`
	VidasInspiradas    int     `json:"vidasInspiradas"` // pessoas que fizeram projetos dele
}

// Perfil alimenta as telas Perfil e Missões (card de nível) e o perfil público.
type Perfil struct {
	ID          int       `json:"id"`
	Nome        string    `json:"nome"`
	FotoURL     *string   `json:"fotoUrl"`
	Bio         *string   `json:"bio"`
	Bairro      string    `json:"bairro"`
	Cidade      string    `json:"cidade"`
	UF          string    `json:"uf"`
	MembroDesde time.Time `json:"membroDesde"`

	XPTotal        int         `json:"xpTotal"` // "EcoPontos" no topo das telas
	Nivel          *Nivel      `json:"nivel"`
	ProximoNivel   *Nivel      `json:"proximoNivel"`   // null no nível máximo
	XPParaProximo  int         `json:"xpParaProximo"`  // quanto falta
	ProgressoNivel int         `json:"progressoNivel"` // 0 a 100, para a barra
	SequenciaDias  int         `json:"sequenciaDias"`  // já zerada se quebrou
	Liga           *LigaResumo `json:"liga"`
	Impacto        Impacto     `json:"impacto"`
}

type NivelCatalogo struct {
	ID       int    `db:"id" json:"id"`
	Numero   int    `db:"numero" json:"numero"`
	Titulo   string `db:"titulo" json:"titulo"`
	MinimoXP int    `db:"minimo_xp" json:"minimoXp"`
}

type EventoXP struct {
	ID           int64     `json:"id"`
	Quantidade   int       `json:"quantidade"`
	Origem       string    `json:"origem"`
	ReferenciaID *int64    `json:"referenciaId"`
	Descricao    *string   `json:"descricao"`
	CriadoEm     time.Time `json:"criadoEm"`
}

// ---------- ligas ----------

type Liga struct {
	ID       int     `db:"id" json:"id"`
	Ordem    int     `db:"ordem" json:"ordem"`
	Nome     string  `db:"nome" json:"nome"`
	Icone    *string `db:"icone" json:"icone"`
	QtdSobe  int     `db:"qtd_sobe" json:"qtdSobe"`
	QtdDesce int     `db:"qtd_desce" json:"qtdDesce"`
}

type LinhaRanking struct {
	Posicao   int     `db:"posicao" json:"posicao"`
	UsuarioID int     `db:"usuario_id" json:"usuarioId"`
	Nome      string  `db:"nome" json:"nome"`
	FotoURL   *string `db:"foto_url" json:"fotoUrl"`
	Bairro    string  `db:"bairro" json:"bairro"`
	XPSemana  int     `db:"xp_semana" json:"xpSemana"`
	Tendencia int     `db:"tendencia" json:"tendencia"` // + subiu, - caiu (desde ontem)
	Zona      string  `db:"zona" json:"zona"`           // sobe | desce | mantem
	Eu        bool    `db:"-" json:"eu"`
}

type ResultadoSemana struct {
	SemanaInicio time.Time `db:"semana_inicio" json:"semanaInicio"`
	XPSemana     int       `db:"xp_semana" json:"xpSemana"`
	Posicao      int       `db:"posicao" json:"posicao"`
	Resultado    string    `db:"resultado" json:"resultado"` // subiu | desceu | manteve
	LigaAnterior string    `db:"liga_anterior" json:"ligaAnterior"`
	LigaNova     string    `db:"liga_nova" json:"ligaNova"`
}

// Ranking é a tela Comunidade.
type Ranking struct {
	Liga         Liga             `json:"liga"`
	Primeira     bool             `json:"primeira"`  // ninguém desce
	Ultima       bool             `json:"ultima"`    // ninguém sobe
	FimSemana    time.Time        `json:"fimSemana"` // para o contador "termina em 2d 4h"
	Itens        []LinhaRanking   `json:"itens"`
	TotalNaLiga  int              `json:"totalNaLiga"`
	MinhaPosicao *LinhaRanking    `json:"minhaPosicao"`
	UltimaSemana *ResultadoSemana `json:"ultimaSemana"` // null se nunca fechou uma semana
}

// ---------- missões ----------

type Categoria struct {
	ID   int     `json:"id"`
	Nome string  `json:"nome"`
	Cor  *string `json:"cor"`
}

type MissaoSemana struct {
	ID          int        `json:"id"`
	Titulo      string     `json:"titulo"`
	Descricao   string     `json:"descricao"`
	ImagemURL   *string    `json:"imagemUrl"`
	Categoria   *Categoria `json:"categoria"`
	Meta        int        `json:"meta"`
	Unidade     string     `json:"unidade"`
	XP          int        `json:"xp"`
	Progresso   int        `json:"progresso"`  // limitado à meta
	Percentual  int        `json:"percentual"` // 0 a 100
	Concluida   bool       `json:"concluida"`
	ConcluidaEm *time.Time `json:"concluidaEm"`
	// true: o usuário precisa enviar foto (missão de registro manual).
	// false: avança sozinha com ações no app.
	EnviaFoto      bool `json:"enviaFoto"`
	FotosEmAnalise int  `json:"fotosEmAnalise"`
}

type MissoesResponse struct {
	FimSemana time.Time      `json:"fimSemana"`
	Missoes   []MissaoSemana `json:"missoes"`
}

type EvidenciaRequest struct {
	FotoURL    string `json:"fotoUrl"`
	Quantidade int    `json:"quantidade"`
}

type Evidencia struct {
	ID             int        `json:"id"`
	MissaoID       int        `json:"missaoId"`
	MissaoTitulo   string     `json:"missaoTitulo"`
	FotoURL        string     `json:"fotoUrl"`
	Quantidade     int        `json:"quantidade"`
	Status         string     `json:"status"` // pendente | aprovada | rejeitada
	MotivoRejeicao *string    `json:"motivoRejeicao"`
	SemanaInicio   *time.Time `json:"semanaInicio,omitempty"`
	CriadoEm       time.Time  `json:"criadoEm"`
	AvaliadoEm     *time.Time `json:"avaliadoEm"`
}

type EvidenciaPendente struct {
	ID           int       `json:"id"`
	FotoURL      string    `json:"fotoUrl"`
	Quantidade   int       `json:"quantidade"`
	CriadoEm     time.Time `json:"criadoEm"`
	SemanaInicio time.Time `json:"semanaInicio"`
	Progresso    int       `json:"progresso"`
	Missao       struct {
		ID      int    `json:"id"`
		Titulo  string `json:"titulo"`
		Meta    int    `json:"meta"`
		Unidade string `json:"unidade"`
	} `json:"missao"`
	Usuario struct {
		ID      int     `json:"id"`
		Nome    string  `json:"nome"`
		FotoURL *string `json:"fotoUrl"`
	} `json:"usuario"`
}

// AprovacaoResponse diz se a aprovação concluiu a missão (e quanto XP o usuário ganhou).
type AprovacaoResponse struct {
	Progresso       int               `json:"progresso"`
	MissaoConcluida *missao.Concluida `json:"missaoConcluida"`
}

// ---------- linhas do banco ----------

type perfilRow struct {
	ID                 int            `db:"id"`
	Nome               string         `db:"nome"`
	FotoURL            sql.NullString `db:"foto_url"`
	Bio                sql.NullString `db:"bio"`
	Bairro             string         `db:"bairro"`
	Cidade             string         `db:"cidade"`
	UF                 string         `db:"uf"`
	XPTotal            int            `db:"xp_total"`
	CriadoEm           time.Time      `db:"criado_em"`
	SequenciaAtual     int            `db:"sequencia_atual"`
	NivelNumero        sql.NullInt64  `db:"nivel_numero"`
	NivelTitulo        sql.NullString `db:"nivel_titulo"`
	NivelMinimoXP      sql.NullInt64  `db:"nivel_minimo_xp"`
	ProxNumero         sql.NullInt64  `db:"prox_numero"`
	ProxTitulo         sql.NullString `db:"prox_titulo"`
	ProxMinimoXP       sql.NullInt64  `db:"prox_minimo_xp"`
	LigaID             sql.NullInt64  `db:"liga_id"`
	LigaNome           sql.NullString `db:"liga_nome"`
	LigaOrdem          sql.NullInt64  `db:"liga_ordem"`
	LigaIcone          sql.NullString `db:"liga_icone"`
	ResiduosKg         float64        `db:"residuos_kg"`
	ProjetosConcluidos int            `db:"projetos_concluidos"`
	ProjetosPublicados int            `db:"projetos_publicados"`
	VidasInspiradas    int            `db:"vidas_inspiradas"`
}

type posicaoRow struct {
	Posicao   int    `db:"posicao"`
	XPSemana  int    `db:"xp_semana"`
	Tendencia int    `db:"tendencia"`
	Zona      string `db:"zona"`
}

type ligaUsuarioRow struct {
	Liga
	Ultima    bool      `db:"ultima"`
	Primeira  bool      `db:"primeira"`
	FimSemana time.Time `db:"fim_semana"`
}

type eventoRow struct {
	ID           int64          `db:"id"`
	Quantidade   int            `db:"quantidade"`
	Origem       string         `db:"origem"`
	ReferenciaID sql.NullInt64  `db:"referencia_id"`
	Descricao    sql.NullString `db:"descricao"`
	CriadoEm     time.Time      `db:"criado_em"`
	Total        int            `db:"total"`
}

type missaoSemanaRow struct {
	ID            int            `db:"id"`
	Titulo        string         `db:"titulo"`
	Descricao     string         `db:"descricao"`
	ImagemURL     sql.NullString `db:"imagem_url"`
	Acao          string         `db:"acao"`
	Meta          int            `db:"meta"`
	Unidade       string         `db:"unidade"`
	XP            int            `db:"xp"`
	CategoriaID   sql.NullInt64  `db:"categoria_id"`
	CategoriaNome sql.NullString `db:"categoria_nome"`
	CategoriaCor  sql.NullString `db:"categoria_cor"`
	Progresso     int            `db:"progresso"`
	ConcluidaEm   sql.NullTime   `db:"concluida_em"`
	Pendentes     int            `db:"pendentes"`
	FimSemana     time.Time      `db:"fim_semana"`
}

type missaoRow struct {
	ID     int    `db:"id"`
	Titulo string `db:"titulo"`
	Acao   string `db:"acao"`
	Meta   int    `db:"meta"`
	XP     int    `db:"xp"`
	Ativa  bool   `db:"ativa"`
}

type progressoRow struct {
	ID          int          `db:"id"`
	Progresso   int          `db:"progresso"`
	ConcluidaEm sql.NullTime `db:"concluida_em"`
	Pendentes   int          `db:"pendentes"`
}

type evidenciaCriadaRow struct {
	ID       int       `db:"id"`
	Status   string    `db:"status"`
	CriadoEm time.Time `db:"criado_em"`
}

type minhaEvidenciaRow struct {
	ID             int            `db:"id"`
	FotoURL        string         `db:"foto_url"`
	Quantidade     int            `db:"quantidade"`
	Status         string         `db:"status"`
	MotivoRejeicao sql.NullString `db:"motivo_rejeicao"`
	CriadoEm       time.Time      `db:"criado_em"`
	AvaliadoEm     sql.NullTime   `db:"avaliado_em"`
	MissaoID       int            `db:"missao_id"`
	MissaoTitulo   string         `db:"missao_titulo"`
	SemanaInicio   time.Time      `db:"semana_inicio"`
	Total          int            `db:"total"`
}

type pendenteRow struct {
	ID             int            `db:"id"`
	FotoURL        string         `db:"foto_url"`
	Quantidade     int            `db:"quantidade"`
	CriadoEm       time.Time      `db:"criado_em"`
	SemanaInicio   time.Time      `db:"semana_inicio"`
	Progresso      int            `db:"progresso"`
	MissaoID       int            `db:"missao_id"`
	MissaoTitulo   string         `db:"missao_titulo"`
	Meta           int            `db:"meta"`
	Unidade        string         `db:"unidade"`
	UsuarioID      int            `db:"usuario_id"`
	UsuarioNome    string         `db:"usuario_nome"`
	UsuarioFotoURL sql.NullString `db:"usuario_foto_url"`
	Total          int            `db:"total"`
}

type evidenciaAvaliarRow struct {
	ID              int    `db:"id"`
	Status          string `db:"status"`
	Quantidade      int    `db:"quantidade"`
	UsuarioMissaoID int    `db:"usuario_missao_id"`
	UsuarioID       int    `db:"usuario_id"`
}

type somaProgressoRow struct {
	ID          int          `db:"id"`
	UsuarioID   int          `db:"usuario_id"`
	Progresso   int          `db:"progresso"`
	ConcluidaEm sql.NullTime `db:"concluida_em"`
	Meta        int          `db:"meta"`
	XP          int          `db:"xp"`
	Titulo      string       `db:"titulo"`
	MissaoID    int          `db:"missao_id"`
}
