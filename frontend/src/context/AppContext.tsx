import AsyncStorage from '@react-native-async-storage/async-storage';
import React, { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import { baseForumPosts, defaultUser, missions, projects } from '../data';
import { apiConfigured } from '../services/api';
import { fetchSavedProjectIds, toggleSavedProjectRemote } from '../services/projects';
import {
  completeProjectRemote,
  fetchGamificationProfile,
  progressMissionRemote,
  updateMyProfileRemote,
} from '../services/progress';
import type { ForumPost, PersistedState, UserProfile } from '../types';
import { useAuth } from './AuthContext';

const STORAGE_KEY = '@aric/state/v2.1';
const LEGACY_STORAGE_KEY = '@aric/state/v2';

type ActionResult = { ok: true; awarded?: boolean; completedNow?: boolean; reward?: number } | { ok: false; message: string };

const initialState: PersistedState = {
  user: defaultUser,
  favoriteProjectIds: [],
  completedProjectIds: [],
  missionProgress: { 'pet-10': 7, 'papelao-1': 0, 'compostagem-3': 2 },
  completedMissionIds: [],
  likedPostIds: [],
  reportedPostIds: [],
  blockedAuthors: [],
  forumPosts: baseForumPosts,
};

type AppContextValue = PersistedState & {
  hydrated: boolean;
  cloudProgressEnabled: boolean;
  toggleFavorite: (projectId: string) => Promise<ActionResult>;
  completeProject: (projectId: string) => Promise<ActionResult>;
  incrementMission: (missionId: string) => Promise<ActionResult>;
  completeMission: (missionId: string) => Promise<ActionResult>;
  toggleLikePost: (postId: string) => void;
  reportPost: (postId: string) => void;
  blockAuthor: (author: string) => void;
  unblockAuthor: (author: string) => void;
  addForumPost: (post: Omit<ForumPost, 'id' | 'time' | 'likes' | 'answers'>) => void;
  updateProfile: (profile: Pick<UserProfile, 'name' | 'bio'>) => Promise<ActionResult>;
  resetDemo: () => Promise<void>;
};

const AppContext = createContext<AppContextValue | null>(null);

function actionError(error: unknown) {
  const text = error instanceof Error ? error.message : String(error ?? '');
  if (text.includes('unauthenticated')) return 'Entre na sua conta para registrar progresso.';
  if (text.includes('permission-denied')) return 'Você não tem permissão para esta ação.';
  if (text.includes('failed-precondition')) return 'Seu perfil ainda está sincronizando. Tente novamente em alguns segundos.';
  return 'Não foi possível sincronizar com o servidor. Verifique sua internet e tente novamente.';
}

export function AppProvider({ children }: { children: React.ReactNode }) {
  const { session, uid } = useAuth();
  const cloudProgressEnabled = apiConfigured && session?.provider === 'api' && Boolean(uid);
  const [state, setState] = useState<PersistedState>(initialState);
  const [hydrated, setHydrated] = useState(false);

  useEffect(() => {
    Promise.all([AsyncStorage.getItem(STORAGE_KEY), AsyncStorage.getItem(LEGACY_STORAGE_KEY)])
      .then(([v21, v2]) => {
        const raw = v21 ?? v2;
        if (!raw) return;
        const parsed = JSON.parse(raw) as Partial<PersistedState>;
        setState((prev) => ({
          ...prev,
          ...parsed,
          user: { ...prev.user, ...(parsed.user ?? {}) },
          forumPosts: parsed.forumPosts?.length ? parsed.forumPosts : prev.forumPosts,
        }));
      })
      .catch(() => undefined)
      .finally(() => setHydrated(true));
  }, []);

  useEffect(() => {
    if (!hydrated) return;
    AsyncStorage.setItem(STORAGE_KEY, JSON.stringify(state)).catch(() => undefined);
  }, [hydrated, state]);

  useEffect(() => {
    if (!cloudProgressEnabled || !uid) return;
    let active = true;
    fetchGamificationProfile()
      .then((profile) => {
        if (!active) return;
        setState((prev) => ({ ...prev, user: { ...prev.user, ...profile, email: session?.email ?? prev.user.email } }));
      })
      .catch((error) => console.warn('[ARIC] Não foi possível carregar o perfil:', error));
    return () => { active = false; };
  }, [cloudProgressEnabled, session?.email, uid]);

  useEffect(() => {
    if (!cloudProgressEnabled) return;
    let active = true;
    fetchSavedProjectIds()
      .then((ids) => { if (active) setState((prev) => ({ ...prev, favoriteProjectIds: ids })); })
      .catch((error) => console.warn('[ARIC] Não foi possível carregar projetos salvos:', error));
    return () => { active = false; };
  }, [cloudProgressEnabled]);

  const toggleFavorite = useCallback(async (projectId: string): Promise<ActionResult> => {
    const wasSaved = state.favoriteProjectIds.includes(projectId);
    if (cloudProgressEnabled && /^\d+$/.test(projectId)) {
      try {
        await toggleSavedProjectRemote(projectId, wasSaved);
      } catch (error) {
        return { ok: false, message: actionError(error) };
      }
    }
    setState((prev) => ({
      ...prev,
      favoriteProjectIds: wasSaved
        ? prev.favoriteProjectIds.filter((id) => id !== projectId)
        : [...prev.favoriteProjectIds, projectId],
    }));
    return { ok: true };
  }, [cloudProgressEnabled, state.favoriteProjectIds]);

  const completeProject = useCallback(async (projectId: string): Promise<ActionResult> => {
    const project = projects.find((item) => item.id === projectId);

    if (cloudProgressEnabled && !/^\d+$/.test(projectId)) return { ok: false, message: 'Este projeto ainda não está vinculado a um registro do catálogo no servidor.' };

    if (cloudProgressEnabled && /^\d+$/.test(projectId)) {
      try {
        const result = await completeProjectRemote(projectId);
        const refreshedProfile = await fetchGamificationProfile().catch(() => null);
        setState((prev) => ({ ...prev, completedProjectIds: prev.completedProjectIds.includes(projectId) ? prev.completedProjectIds : [...prev.completedProjectIds, projectId], user: refreshedProfile ? { ...prev.user, ...refreshedProfile } : { ...prev.user, points: Number(result.xpTotal ?? prev.user.points + Number(result.xpGanho ?? 0)) } }));
        return { ok: true, awarded: Number(result.xpGanho ?? 0) > 0, reward: Number(result.xpGanho ?? 0) };
      } catch (error) {
        return { ok: false, message: actionError(error) };
      }
    }

    if (!project) return { ok: false, message: 'Projeto não encontrado no catálogo local. Atualize a lista de projetos.' };

    let awarded = false;
    setState((prev) => {
      if (prev.completedProjectIds.includes(projectId)) return prev;
      awarded = true;
      return {
        ...prev,
        completedProjectIds: [...prev.completedProjectIds, projectId],
        user: {
          ...prev.user,
          points: prev.user.points + project.points,
          impactKg: Number((prev.user.impactKg + project.impactKg).toFixed(1)),
          co2Kg: Number((prev.user.co2Kg + project.co2Kg).toFixed(1)),
          projectsCompleted: prev.user.projectsCompleted + 1,
          peopleInspired: prev.user.peopleInspired + 2,
        },
      };
    });
    return { ok: true, awarded, reward: awarded ? project.points : 0 };
  }, [cloudProgressEnabled]);

  const registerMissionStep = useCallback(async (missionId: string): Promise<ActionResult> => {
    const mission = missions.find((item) => item.id === missionId);
    if (!mission) return { ok: false, message: 'Missão não encontrada.' };

    if (cloudProgressEnabled) {
      try {
        const result = await progressMissionRemote(missionId);
        return { ok: true, completedNow: result.completedNow, reward: result.reward };
      } catch (error) {
        return { ok: false, message: actionError(error) };
      }
    }

    let completedNow = false;
    setState((prev) => {
      if (prev.completedMissionIds.includes(missionId)) return prev;
      const next = Math.min(mission.goal, (prev.missionProgress[missionId] ?? 0) + 1);
      if (next < mission.goal) {
        return { ...prev, missionProgress: { ...prev.missionProgress, [missionId]: next } };
      }
      completedNow = true;
      return {
        ...prev,
        completedMissionIds: [...prev.completedMissionIds, missionId],
        missionProgress: { ...prev.missionProgress, [missionId]: mission.goal },
        user: { ...prev.user, points: prev.user.points + mission.reward, streak: prev.user.streak + 1 },
      };
    });
    return { ok: true, completedNow, reward: completedNow ? mission.reward : 0 };
  }, [cloudProgressEnabled]);

  const incrementMission = useCallback((missionId: string) => registerMissionStep(missionId), [registerMissionStep]);
  const completeMission = useCallback((missionId: string) => registerMissionStep(missionId), [registerMissionStep]);

  const toggleLikePost = useCallback((postId: string) => {
    setState((prev) => {
      const liked = prev.likedPostIds.includes(postId);
      return {
        ...prev,
        likedPostIds: liked ? prev.likedPostIds.filter((id) => id !== postId) : [...prev.likedPostIds, postId],
        forumPosts: prev.forumPosts.map((post) =>
          post.id === postId ? { ...post, likes: Math.max(0, post.likes + (liked ? -1 : 1)) } : post,
        ),
      };
    });
  }, []);

  const reportPost = useCallback((postId: string) => {
    setState((prev) => ({
      ...prev,
      reportedPostIds: prev.reportedPostIds.includes(postId) ? prev.reportedPostIds : [...prev.reportedPostIds, postId],
    }));
  }, []);

  const blockAuthor = useCallback((author: string) => {
    setState((prev) => ({
      ...prev,
      blockedAuthors: prev.blockedAuthors.includes(author) ? prev.blockedAuthors : [...prev.blockedAuthors, author],
    }));
  }, []);

  const unblockAuthor = useCallback((author: string) => {
    setState((prev) => ({ ...prev, blockedAuthors: prev.blockedAuthors.filter((item) => item !== author) }));
  }, []);

  const addForumPost = useCallback<AppContextValue['addForumPost']>((post) => {
    setState((prev) => ({
      ...prev,
      forumPosts: [{ ...post, id: `user-${Date.now()}`, time: 'agora', likes: 0, answers: 0 }, ...prev.forumPosts],
    }));
  }, []);

  const updateProfile = useCallback<AppContextValue['updateProfile']>(async (profile) => {
    const name = profile.name.trim();
    const bio = profile.bio.trim();
    if (cloudProgressEnabled) {
      try {
        await updateMyProfileRemote({ name, bio });
        setState((prev) => ({ ...prev, user: { ...prev.user, name, bio } }));
        return { ok: true };
      } catch (error) {
        return { ok: false, message: actionError(error) };
      }
    }
    setState((prev) => ({ ...prev, user: { ...prev.user, name, bio } }));
    return { ok: true };
  }, [cloudProgressEnabled]);

  const resetDemo = useCallback(async () => {
    await AsyncStorage.multiRemove([STORAGE_KEY, LEGACY_STORAGE_KEY]);
    setState(initialState);
  }, []);

  const value = useMemo<AppContextValue>(() => ({
    ...state,
    hydrated,
    cloudProgressEnabled,
    toggleFavorite,
    completeProject,
    incrementMission,
    completeMission,
    toggleLikePost,
    reportPost,
    blockAuthor,
    unblockAuthor,
    addForumPost,
    updateProfile,
    resetDemo,
  }), [
    addForumPost,
    blockAuthor,
    cloudProgressEnabled,
    completeMission,
    completeProject,
    hydrated,
    incrementMission,
    reportPost,
    resetDemo,
    state,
    toggleFavorite,
    toggleLikePost,
    unblockAuthor,
    updateProfile,
  ]);

  return <AppContext.Provider value={value}>{children}</AppContext.Provider>;
}

export function useApp() {
  const context = useContext(AppContext);
  if (!context) throw new Error('useApp must be used inside AppProvider');
  return context;
}
