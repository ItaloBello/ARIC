import AsyncStorage from '@react-native-async-storage/async-storage';
import React, { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import { baseForumPosts } from '../data';
import { apiConfigured } from '../services/api';
import {
  createForumPost,
  fetchForumPosts,
  fetchModerationReports,
  fetchMyForumPosts,
  removeForumPostRemote,
  reportForumPostRemote,
  resolveReportRemote,
  toggleForumLikeRemote,
} from '../services/community';
import type { AbuseReport, CommunityRestriction, ForumPost, ForumSubmission, MaterialCategory, ModerationDecision } from '../types';
import { useAuth } from './AuthContext';

const DEMO_KEY = '@aric/community/v3';

type SubmitPayload = {
  author: string;
  neighborhood: string;
  category: MaterialCategory;
  title: string;
  body: string;
  imageUrl?: string;
  aiReviewConsent?: boolean;
  communityRulesAccepted: boolean;
  communityRulesVersion?: string;
  isTip?: boolean;
};
type CommunityResult = { ok: true; message?: string } | { ok: false; message: string };
type DemoStore = { posts: ForumPost[]; submissions: ForumSubmission[]; reports: AbuseReport[]; restrictions: CommunityRestriction[] };
const initialDemo: DemoStore = { posts: baseForumPosts, submissions: [], reports: [], restrictions: [] };

type CommunityContextValue = {
  mode: 'api' | 'demo';
  loading: boolean;
  posts: ForumPost[];
  mySubmissions: ForumSubmission[];
  moderationQueue: ForumSubmission[];
  reports: AbuseReport[];
  restrictions: CommunityRestriction[];
  submitPost: (payload: SubmitPayload) => Promise<CommunityResult>;
  reviewSubmission: (id: string, decision: ModerationDecision, note?: string) => Promise<CommunityResult>;
  rerunAiReview: (id: string) => Promise<CommunityResult>;
  toggleLike: (postId: string, currentlyLiked?: boolean) => Promise<CommunityResult>;
  reportPost: (post: ForumPost, reason: string, details?: string) => Promise<CommunityResult>;
  removePost: (postId: string, note?: string) => Promise<CommunityResult>;
  resolveReport: (reportId: string, action?: 'remover' | 'manter', reason?: string) => Promise<CommunityResult>;
  suspendUser: (targetUid: string, durationHours: 24 | 168 | 720 | -1, reason: string) => Promise<CommunityResult>;
  restoreUser: (targetUid: string) => Promise<CommunityResult>;
  refreshCommunity: () => Promise<void>;
};

const CommunityContext = createContext<CommunityContextValue | null>(null);

export function CommunityProvider({ children }: { children: React.ReactNode }) {
  const { uid, isAdmin, session } = useAuth();
  const useApi = apiConfigured && session?.provider === 'api' && Boolean(uid);
  const [loading, setLoading] = useState(true);
  const [posts, setPosts] = useState<ForumPost[]>(baseForumPosts);
  const [mySubmissions, setMySubmissions] = useState<ForumSubmission[]>([]);
  const [moderationQueue, setModerationQueue] = useState<ForumSubmission[]>([]);
  const [reports, setReports] = useState<AbuseReport[]>([]);
  const [restrictions, setRestrictions] = useState<CommunityRestriction[]>([]);
  const [demo, setDemo] = useState<DemoStore>(initialDemo);

  const refreshCommunity = useCallback(async () => {
    if (!useApi || !uid) return;
    setLoading(true);
    try {
      const [nextPosts, nextMine, nextReports] = await Promise.all([
        fetchForumPosts(),
        fetchMyForumPosts(uid),
        isAdmin ? fetchModerationReports() : Promise.resolve([]),
      ]);
      setPosts(nextPosts);
      setMySubmissions(nextMine);
      setReports(nextReports);
      setModerationQueue([]);
      setRestrictions([]);
    } catch (error) {
      setPosts([]);
      setMySubmissions([]);
      setReports([]);
      const message = error instanceof Error ? error.message : 'Falha ao carregar o fórum.';
      console.warn(`[ARIC] ${message}`);
    } finally {
      setLoading(false);
    }
  }, [isAdmin, uid, useApi]);

  useEffect(() => {
    if (useApi) {
      void refreshCommunity();
      return;
    }
    AsyncStorage.getItem(DEMO_KEY)
      .then((raw) => raw && setDemo({ ...initialDemo, ...(JSON.parse(raw) as Partial<DemoStore>) }))
      .catch(() => undefined)
      .finally(() => setLoading(false));
  }, [refreshCommunity, useApi]);

  useEffect(() => {
    if (useApi || loading) return;
    AsyncStorage.setItem(DEMO_KEY, JSON.stringify(demo)).catch(() => undefined);
    setPosts(demo.posts);
    setMySubmissions(uid ? demo.submissions.filter((item) => item.authorUid === uid) : []);
    setModerationQueue(isAdmin ? demo.submissions.filter((item) => item.status === 'pending_review') : []);
    setReports(isAdmin ? demo.reports.filter((item) => item.status === 'open') : []);
    setRestrictions(isAdmin ? demo.restrictions : []);
  }, [demo, isAdmin, loading, uid, useApi]);

  const submitPost = useCallback(async (payload: SubmitPayload): Promise<CommunityResult> => {
    if (!uid || uid === 'guest') return { ok: false, message: 'Entre em uma conta para publicar.' };
    if (!payload.communityRulesAccepted) return { ok: false, message: 'Aceite as regras da comunidade antes de publicar.' };
    try {
      if (useApi) {
        await createForumPost({ title: payload.title, body: payload.body, category: payload.category, isTip: payload.isTip });
        await refreshCommunity();
        return { ok: true, message: 'Publicação criada e disponível no fórum.' };
      }
      const now = new Date().toISOString();
      const post: ForumPost = {
        id: `demo-post-${Date.now()}`,
        author: payload.author,
        authorUid: uid,
        neighborhood: payload.neighborhood,
        category: payload.category,
        title: payload.title,
        body: payload.body,
        imageUrl: payload.imageUrl,
        status: 'Em aberto',
        time: 'agora',
        likes: 0,
        answers: 0,
        isTip: payload.isTip,
        publishedAt: now,
      };
      setDemo((prev) => ({ ...prev, posts: [post, ...prev.posts] }));
      return { ok: true, message: 'Publicação adicionada ao fórum de demonstração.' };
    } catch (error) {
      return { ok: false, message: error instanceof Error ? error.message : 'Não foi possível publicar.' };
    }
  }, [refreshCommunity, uid, useApi]);

  const reviewSubmission = useCallback(async (_id: string, _decision: ModerationDecision, _note = ''): Promise<CommunityResult> => ({
    ok: false,
    message: 'O backend Go publica posts imediatamente e não possui fila de aprovação prévia.',
  }), []);

  const rerunAiReview = useCallback(async (_id: string): Promise<CommunityResult> => ({
    ok: false,
    message: 'A API Go ainda não oferece reanálise automática por IA.',
  }), []);

  const toggleLike = useCallback(async (postId: string, currentlyLiked = false): Promise<CommunityResult> => {
    if (!uid || uid === 'guest') return { ok: false, message: 'Entre na sua conta para curtir.' };
    try {
      if (useApi) {
        await toggleForumLikeRemote(postId, currentlyLiked);
        await refreshCommunity();
      }
      return { ok: true };
    } catch (error) {
      return { ok: false, message: error instanceof Error ? error.message : 'Não foi possível curtir.' };
    }
  }, [refreshCommunity, uid, useApi]);

  const reportPost = useCallback(async (post: ForumPost, reason: string, details = ''): Promise<CommunityResult> => {
    if (!uid || uid === 'guest') return { ok: false, message: 'Entre na sua conta para denunciar.' };
    try {
      if (useApi) {
        await reportForumPostRemote(post, reason, details);
      } else {
        setDemo((prev) => ({ ...prev, reports: [{ id: `report-${Date.now()}`, postId: post.id, postTitle: post.title, reporterUid: uid, reason, details, status: 'open', createdAt: new Date().toISOString() }, ...prev.reports] }));
      }
      return { ok: true, message: 'Denúncia registrada.' };
    } catch (error) {
      return { ok: false, message: error instanceof Error ? error.message : 'Não foi possível registrar a denúncia.' };
    }
  }, [uid, useApi]);

  const removePost = useCallback(async (postId: string, _note = ''): Promise<CommunityResult> => {
    if (!isAdmin) return { ok: false, message: 'Apenas moderadores podem remover publicações de terceiros.' };
    try {
      if (useApi) {
        await removeForumPostRemote(postId);
        await refreshCommunity();
      } else {
        setDemo((prev) => ({ ...prev, posts: prev.posts.filter((post) => post.id !== postId) }));
      }
      return { ok: true };
    } catch (error) {
      return { ok: false, message: error instanceof Error ? error.message : 'Não foi possível remover.' };
    }
  }, [isAdmin, refreshCommunity, useApi]);

  const resolveReport = useCallback(async (reportId: string, action: 'remover' | 'manter' = 'manter', reason = ''): Promise<CommunityResult> => {
    if (!isAdmin) return { ok: false, message: 'Apenas moderadores podem resolver denúncias.' };
    try {
      if (useApi) {
        await resolveReportRemote(reportId, action, reason);
        await refreshCommunity();
      } else {
        setDemo((prev) => ({ ...prev, reports: prev.reports.map((report) => report.id === reportId ? { ...report, status: 'resolved' } : report) }));
      }
      return { ok: true };
    } catch (error) {
      return { ok: false, message: error instanceof Error ? error.message : 'Não foi possível resolver a denúncia.' };
    }
  }, [isAdmin, refreshCommunity, useApi]);

  const suspendUser = useCallback(async (_targetUid: string, _durationHours: 24 | 168 | 720 | -1, _reason: string): Promise<CommunityResult> => ({
    ok: false,
    message: 'O backend Go atual não possui recurso de suspensão de usuários.',
  }), []);
  const restoreUser = useCallback(async (_targetUid: string): Promise<CommunityResult> => ({
    ok: false,
    message: 'O backend Go atual não possui recurso de restauração de suspensão.',
  }), []);

  const value = useMemo<CommunityContextValue>(() => ({
    mode: useApi ? 'api' : 'demo', loading, posts, mySubmissions, moderationQueue, reports, restrictions,
    submitPost, reviewSubmission, rerunAiReview, toggleLike, reportPost, removePost, resolveReport,
    suspendUser, restoreUser, refreshCommunity,
  }), [loading, moderationQueue, mySubmissions, posts, refreshCommunity, removePost, reportPost, reports, rerunAiReview, resolveReport, restoreUser, reviewSubmission, restrictions, submitPost, suspendUser, toggleLike, useApi]);

  return <CommunityContext.Provider value={value}>{children}</CommunityContext.Provider>;
}

export function useCommunity() {
  const context = useContext(CommunityContext);
  if (!context) throw new Error('useCommunity must be used inside CommunityProvider');
  return context;
}
