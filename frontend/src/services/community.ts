import { apiRequest, apiMediaUrl } from './api';
import type { AbuseReport, ForumPost, ForumSubmission, MaterialCategory } from '../types';

type ApiCategory = { id: number; nome: string; cor?: string };
type ApiForumPost = {
  id: number;
  tipo?: 'duvida' | 'dica';
  titulo?: string;
  resumo?: string;
  conteudo?: string;
  imagemUrl?: string | null;
  resolvido?: boolean;
  autor?: { id?: number; nome?: string; bairro?: string; fotoUrl?: string | null } | null;
  categoria?: { id?: number; nome?: string; cor?: string } | null;
  curtidas?: number;
  respostas?: number;
  curtido?: boolean;
  criadoEm?: string;
  atualizadoEm?: string;
};
type ApiPage<T> = { itens?: T[]; total?: number; pagina?: number; limite?: number } | T[];
type ApiReport = {
  tipo?: string;
  alvoTipo?: string;
  id?: number;
  alvoId?: number;
  titulo?: string;
  trecho?: string;
  autor?: string | { nome?: string };
  quantidadeDenuncias?: number;
  motivos?: string[];
  descricoes?: string[];
  noAr?: boolean;
};

let categoryCache: ApiCategory[] = [];

function categoryName(value?: string | null): MaterialCategory {
  const normalized = String(value ?? '').normalize('NFD').replace(/[\u0300-\u036f]/g, '').toLocaleLowerCase('pt-BR');
  if (normalized.includes('plastic')) return 'Plástico';
  if (normalized.includes('papel') || normalized.includes('carton')) return 'Papelão';
  if (normalized.includes('vidro')) return 'Vidro';
  if (normalized.includes('metal') || normalized.includes('alumin')) return 'Metal';
  if (normalized.includes('madeira') || normalized.includes('pallet')) return 'Madeira';
  if (normalized.includes('horta') || normalized.includes('organ') || normalized.includes('compost')) return 'Horta';
  return 'Todos';
}

function pageItems<T>(value: ApiPage<T>): T[] {
  return Array.isArray(value) ? value : value.itens ?? [];
}

function dateLabel(value?: string) {
  if (!value) return 'agora';
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : date.toLocaleDateString('pt-BR');
}

export function forumPostFromApi(item: ApiForumPost): ForumPost {
  return {
    id: String(item.id),
    author: item.autor?.nome ?? 'Usuário ARIC',
    authorUid: item.autor?.id != null ? String(item.autor.id) : undefined,
    neighborhood: item.autor?.bairro ?? 'Comunidade ARIC',
    time: dateLabel(item.criadoEm),
    category: categoryName(item.categoria?.nome),
    status: item.resolvido ? 'Resolvido' : 'Em aberto',
    title: item.titulo ?? 'Publicação da comunidade',
    body: item.conteudo ?? item.resumo ?? '',
    likes: Number(item.curtidas ?? 0),
    answers: Number(item.respostas ?? 0),
    imageUrl: apiMediaUrl(item.imagemUrl),
    isTip: item.tipo === 'dica',
    likedByMe: item.curtido === true,
    publishedAt: item.criadoEm,
  };
}

export async function fetchForumPosts(query = '') {
  const suffix = query ? `&${query}` : '';
  const result = await apiRequest<ApiPage<ApiForumPost>>(`/forum/posts?ordem=recentes&pagina=1&limite=50${suffix}`);
  return pageItems(result).map(forumPostFromApi);
}

export async function fetchMyForumPosts(userId: string): Promise<ForumSubmission[]> {
  const result = await apiRequest<ApiPage<ApiForumPost>>(`/forum/posts?autor=${encodeURIComponent(userId)}&ordem=recentes&pagina=1&limite=50`);
  return pageItems(result).map((item) => {
    const post = forumPostFromApi(item);
    return {
      id: post.id,
      author: post.author,
      authorUid: post.authorUid ?? userId,
      neighborhood: post.neighborhood,
      category: post.category,
      title: post.title,
      body: post.body,
      status: 'approved',
      createdAt: item.criadoEm ?? new Date().toISOString(),
      imageUrl: post.imageUrl,
      aiReview: { status: 'not_run', flagged: false, categories: [], recommendation: 'review' },
    };
  });
}

export async function fetchCategories() {
  const result = await apiRequest<ApiPage<ApiCategory>>('/categorias');
  categoryCache = pageItems(result);
  return categoryCache;
}

export async function createForumPost(payload: { title: string; body: string; category: MaterialCategory; isTip?: boolean }) {
  if (!categoryCache.length) await fetchCategories();
  const category = categoryCache.find((item) => categoryName(item.nome) === payload.category);
  return apiRequest('/forum/posts', {
    method: 'POST',
    body: JSON.stringify({
      tipo: payload.isTip ? 'dica' : 'duvida',
      titulo: payload.title,
      conteudo: payload.body,
      categoriaId: category?.id ?? null,
    }),
  });
}

export async function toggleForumLikeRemote(postId: string, currentlyLiked: boolean) {
  const method = currentlyLiked ? 'DELETE' : 'PUT';
  return apiRequest<{ curtido?: boolean; curtidas?: number }>(`/forum/posts/${encodeURIComponent(postId)}/curtida`, { method });
}

function reportReason(reason: string): 'spam' | 'ofensivo' | 'perigoso' | 'falso' | 'outro' {
  const text = reason.toLocaleLowerCase('pt-BR');
  if (text.includes('spam') || text.includes('fraude') || text.includes('propaganda')) return 'spam';
  if (text.includes('ofens') || text.includes('assédio') || text.includes('ódio')) return 'ofensivo';
  if (text.includes('perig') || text.includes('segur')) return 'perigoso';
  if (text.includes('falso') || text.includes('desinform')) return 'falso';
  return 'outro';
}

export async function reportForumPostRemote(post: ForumPost, reason: string, details = '') {
  return apiRequest('/denuncias', {
    method: 'POST',
    body: JSON.stringify({
      alvoTipo: 'forum_post',
      alvoId: Number(post.id),
      motivo: reportReason(reason),
      descricao: [reason, details].filter(Boolean).join(' — ').slice(0, 500),
    }),
  });
}

export async function removeForumPostRemote(postId: string) {
  return apiRequest<void>(`/forum/posts/${encodeURIComponent(postId)}`, { method: 'DELETE' });
}

export async function fetchModerationReports(): Promise<AbuseReport[]> {
  const result = await apiRequest<ApiPage<ApiReport>>('/moderacao/denuncias?pagina=1&limite=50');
  return pageItems(result).map((item) => {
    const targetType = item.alvoTipo ?? item.tipo ?? 'forum_post';
    const targetId = item.alvoId ?? item.id ?? 0;
    const author = typeof item.autor === 'string' ? item.autor : item.autor?.nome ?? 'Autor não identificado';
    const reasons = Array.isArray(item.motivos) ? item.motivos.join(', ') : 'Denúncia recebida';
    return {
      id: `${targetType}:${targetId}`,
      postId: String(targetId),
      postTitle: item.titulo ?? item.trecho ?? 'Conteúdo denunciado',
      postAuthorUid: undefined,
      reporterUid: '',
      reason: reasons,
      details: [author, ...(item.descricoes ?? [])].filter(Boolean).join(' • '),
      status: 'open' as const,
      createdAt: item.noAr ? 'Conteúdo já removido pelo autor' : 'Aguardando análise',
    };
  });
}

export async function resolveReportRemote(reportId: string, action: 'remover' | 'manter', reason = '') {
  const [alvoTipo, rawId] = reportId.split(':');
  const alvoId = Number(rawId);
  return apiRequest<void>('/moderacao/denuncias/resolver', {
    method: 'POST',
    body: JSON.stringify({ alvoTipo, alvoId, acao: action, motivo: reason }),
  });
}
