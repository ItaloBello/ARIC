import { apiRequest } from './api';
import type { UserProfile } from '../types';

export type GamificationProfile = {
  id: number;
  nome: string;
  fotoUrl?: string | null;
  bio?: string | null;
  bairro?: string | null;
  cidade?: string | null;
  xpTotal?: number;
  sequenciaDias?: number;
  nivel?: { titulo?: string; numero?: number } | null;
  impacto?: { residuosKg?: number; projetosConcluidos?: number; projetosPublicados?: number; vidasInspiradas?: number } | null;
  liga?: { nome?: string; posicao?: number } | null;
};

export type CloudProgressProfile = UserProfile & {
  completedProjectIds: string[];
  missionProgress: Record<string, number>;
  completedMissionIds: string[];
};

export type LeaderboardEntry = { uid: string; name: string; neighborhood: string; points: number; title?: string };

export function profileFromApi(data: GamificationProfile): CloudProgressProfile {
  const impact = data.impacto ?? {};
  return {
    name: data.nome ?? 'Usuário ARIC',
    city: [data.bairro, data.cidade].filter(Boolean).join(' • ') || 'Comunidade ARIC',
    bio: data.bio ?? '',
    title: data.nivel?.titulo ?? 'Criador Circular',
    points: Number(data.xpTotal ?? 0),
    streak: Number(data.sequenciaDias ?? 0),
    impactKg: Number(impact.residuosKg ?? 0),
    co2Kg: 0,
    projectsCompleted: Number(impact.projetosConcluidos ?? 0),
    peopleInspired: Number(impact.vidasInspiradas ?? 0),
    completedProjectIds: [],
    missionProgress: {},
    completedMissionIds: [],
  };
}

export async function fetchGamificationProfile() {
  const data = await apiRequest<GamificationProfile>('/gamificacao/perfil');
  return profileFromApi(data);
}

export async function updateMyProfileRemote(payload: { name: string; bio: string }) {
  return apiRequest('/auth/me', {
    method: 'PUT',
    body: JSON.stringify({ nome: payload.name, bio: payload.bio }),
  });
}

export async function completeProjectRemote(projectId: string) {
  return apiRequest<{ xpGanho?: number; xpTotal?: number; missoesConcluidas?: unknown[] }>(`/projetos/${encodeURIComponent(projectId)}/conclusao`, {
    method: 'POST',
    body: JSON.stringify({}),
  });
}

export async function progressMissionRemote(_missionId: string): Promise<{ completedNow?: boolean; reward?: number }> {
  throw new Error('O backend Go não possui incremento manual de missões. Missões automáticas avançam pelas ações do app; missões com foto usam evidências.');
}


export type NeighborhoodImpact = { kgHoje?: number; projetosHoje?: number; pessoasHoje?: number; kgSemana?: number; kgTotal?: number; bairro?: string };

export async function fetchNeighborhoodImpact() {
  return apiRequest<NeighborhoodImpact>('/gamificacao/impacto-bairro');
}

// O backend Go é REST e não mantém listeners em tempo real. Mantemos os adaptadores vazios
// para telas antigas que ainda não foram removidas da versão simplificada.
export function listenUserProgress(_uid: string, _onChange: (profile: CloudProgressProfile) => void, _onError?: (error: unknown) => void) {
  return () => undefined;
}

export function listenLeaderboard(_onChange: (entries: LeaderboardEntry[]) => void, _onError?: (error: unknown) => void) {
  return () => undefined;
}
