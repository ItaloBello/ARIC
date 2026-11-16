import { useCallback, useEffect, useState } from 'react';
import { projects as localProjects } from '../data';
import { apiConfigured } from '../services/api';
import { fetchProjectDetail, fetchProjects } from '../services/projects';
import { useAuth } from '../context/AuthContext';
import type { Project } from '../types';

export function useProjects() {
  const { session } = useAuth();
  const useApi = apiConfigured && session?.provider === 'api';
  const [items, setItems] = useState<Project[]>(localProjects);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const refresh = useCallback(async () => {
    if (!useApi) {
      setItems(localProjects);
      setError('');
      return;
    }
    setLoading(true);
    try {
      const next = await fetchProjects();
      setItems(next);
      setError('');
    } catch (err) {
      setItems([]);
      setError(err instanceof Error ? err.message : 'Não foi possível carregar os projetos do servidor.');
    } finally {
      setLoading(false);
    }
  }, [useApi]);

  useEffect(() => { void refresh(); }, [refresh]);
  return { projects: items, loading, error, refresh, useApi };
}

export function useProjectDetail(projectId: string) {
  const { session } = useAuth();
  const useApi = apiConfigured && session?.provider === 'api' && /^\d+$/.test(projectId);
  const [project, setProject] = useState<Project | null>(localProjects.find((item) => item.id === projectId) ?? null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    let active = true;
    if (!useApi) {
      setProject(localProjects.find((item) => item.id === projectId) ?? localProjects[0]);
      return () => { active = false; };
    }
    setLoading(true);
    fetchProjectDetail(projectId)
      .then((next) => { if (active) { setProject(next); setError(''); } })
      .catch((err) => { if (active) setError(err instanceof Error ? err.message : 'Não foi possível carregar este projeto.'); })
      .finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, [projectId, useApi]);
  return { project, loading, error, useApi };
}
