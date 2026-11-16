import { projects as localProjects } from '../data';
import type { MaterialCategory, Project } from '../types';
import { apiMediaUrl, apiRequest } from './api';

type ApiProject = {
  id: number;
  titulo?: string;
  subtitulo?: string;
  descricao?: string;
  status?: string;
  xp?: number;
  tempoEstimado?: number;
  residuosKg?: number;
  dificuldade?: { id?: number; descricao?: string } | null;
  capaUrl?: string | null;
  imagens?: string[];
  autor?: { id?: number; nome?: string; fotoUrl?: string | null; bairro?: string } | null;
  categorias?: Array<{ id?: number; nome?: string; cor?: string }>;
  materiais?: Array<{ nome?: string; quantidade?: string; imagemUrl?: string }>;
  passos?: Array<{ titulo?: string; descricao?: string; imagemUrl?: string }>;
  curtidas?: number;
  conclusoes?: number;
  notaMedia?: number;
  curtido?: boolean;
  salvo?: boolean;
  concluido?: boolean;
};
type ApiPage<T> = { itens?: T[]; total?: number } | T[];

function items<T>(result: ApiPage<T>): T[] {
  return Array.isArray(result) ? result : result.itens ?? [];
}

function materialCategory(name?: string): MaterialCategory {
  const normalized = String(name ?? '').normalize('NFD').replace(/[\u0300-\u036f]/g, '').toLocaleLowerCase('pt-BR');
  if (normalized.includes('plastic')) return 'Plástico';
  if (normalized.includes('papel') || normalized.includes('carton')) return 'Papelão';
  if (normalized.includes('vidro')) return 'Vidro';
  if (normalized.includes('metal') || normalized.includes('alumin')) return 'Metal';
  if (normalized.includes('madeira') || normalized.includes('pallet')) return 'Madeira';
  if (normalized.includes('horta') || normalized.includes('organ') || normalized.includes('compost')) return 'Horta';
  return 'Todos';
}

function localImage(category?: string) {
  const found = localProjects.find((item) => item.category === materialCategory(category));
  return found?.image ?? localProjects[0].image;
}

export function projectFromApi(item: ApiProject): Project {
  const category = materialCategory(item.categorias?.[0]?.nome);
  const remoteImage = apiMediaUrl(item.capaUrl ?? item.imagens?.[0]);
  const rawDifficulty = String(item.dificuldade?.descricao ?? '').normalize('NFD').replace(/[\u0300-\u036f]/g, '').toLocaleLowerCase('pt-BR');
  const difficulty = rawDifficulty.includes('avanc') ? 'Avançado' : rawDifficulty.includes('inter') || rawDifficulty.includes('medio') ? 'Intermediário' : 'Iniciante';
  return {
    id: String(item.id),
    title: item.titulo ?? 'Projeto ARIC',
    subtitle: item.subtitulo ?? '',
    category,
    secondaryCategory: item.categorias?.[1]?.nome ? materialCategory(item.categorias[1].nome) : undefined,
    time: item.tempoEstimado ? `${item.tempoEstimado} min` : 'Tempo variável',
    difficulty,
    points: Number(item.xp ?? 0),
    impactKg: Number(item.residuosKg ?? 0),
    co2Kg: 0,
    image: remoteImage ? { uri: remoteImage } : localImage(item.categorias?.[0]?.nome),
    creator: item.autor?.nome ?? 'ARIC',
    neighborhood: item.autor?.bairro ?? 'Comunidade ARIC',
    likes: Number(item.curtidas ?? 0),
    saved: item.salvo === true,
    completed: item.concluido === true,
    description: item.descricao ?? item.subtitulo ?? '',
    materials: item.materiais?.map((material) => [material.nome, material.quantidade].filter(Boolean).join(' — ')) ?? [],
    steps: item.passos?.map((step) => ({ title: step.titulo ?? 'Etapa', description: step.descricao ?? '' })) ?? [],
  };
}

export async function fetchProjects(search = '') {
  const query = search ? `&busca=${encodeURIComponent(search)}` : '';
  const result = await apiRequest<ApiPage<ApiProject>>(`/projetos?ordem=recentes&pagina=1&limite=50${query}`);
  return items(result).map(projectFromApi);
}

export async function fetchProjectDetail(projectId: string) {
  const result = await apiRequest<ApiProject>(`/projetos/${encodeURIComponent(projectId)}`);
  return projectFromApi(result);
}

export async function fetchSavedProjectIds() {
  const result = await apiRequest<ApiPage<ApiProject>>('/projetos/salvos?pagina=1&limite=50');
  return items(result).map((item) => String(item.id));
}

export async function toggleSavedProjectRemote(projectId: string, currentlySaved: boolean) {
  return apiRequest<void>(`/projetos/${encodeURIComponent(projectId)}/salvo`, { method: currentlySaved ? 'DELETE' : 'PUT' });
}
