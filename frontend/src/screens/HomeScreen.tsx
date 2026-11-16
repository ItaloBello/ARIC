import React, { useMemo, useState } from 'react';
import { Pressable, ScrollView, StyleSheet, Text, TextInput, View } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { BrandHeader } from '../components/BrandHeader';
import { Chip } from '../components/Chip';
import { FeaturedProjectCard } from '../components/FeaturedProjectCard';
import { ProjectCard } from '../components/ProjectCard';
import { ScreenShell } from '../components/ScreenShell';
import { SectionHeader } from '../components/SectionHeader';
import { useApp } from '../context/AppContext';
import { useProjects } from '../hooks/useProjects';
import { fetchNeighborhoodImpact, type NeighborhoodImpact } from '../services/progress';
import { useAuth } from '../context/AuthContext';
import type { Route } from '../navigation';
import type { MaterialCategory } from '../types';
import { colors, radius, spacing } from '../theme';

const categories: MaterialCategory[] = ['Todos', 'Plástico', 'Papelão', 'Vidro', 'Metal', 'Madeira', 'Horta'];

export function HomeScreen({ navigate }: { navigate: (route: Route) => void }) {
  const { favoriteProjectIds, toggleFavorite } = useApp();
  const { projects, loading, error, useApi } = useProjects();
  const { session } = useAuth();
  const [impact, setImpact] = useState<NeighborhoodImpact | null>(null);
  React.useEffect(() => {
    let active = true;
    if (useApi && session?.provider === 'api') fetchNeighborhoodImpact().then((value) => { if (active) setImpact(value); }).catch(() => undefined);
    return () => { active = false; };
  }, [session?.provider, useApi]);
  const [query, setQuery] = useState('');
  const [category, setCategory] = useState<MaterialCategory>('Todos');

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    return projects.filter((project) => {
      const matchCategory = category === 'Todos' || project.category === category || project.secondaryCategory === category;
      const haystack = `${project.title} ${project.subtitle} ${project.category} ${project.description} ${project.materials.join(' ')}`.toLowerCase();
      return matchCategory && (!q || haystack.includes(q));
    });
  }, [category, query]);

  const featured = filtered[0] ?? projects[0];
  const nearby = featured ? filtered.filter((project) => project.id !== featured.id).slice(0, 2) : [];

  const openSearch = () => navigate({ name: 'projects', query: query || undefined, category: category !== 'Todos' ? category : undefined });

  return (
    <ScreenShell>
      <BrandHeader onProfile={() => navigate({ name: 'profile' })} />

      <Pressable accessibilityRole="button" onPress={() => navigate({ name: 'profile' })} style={styles.impact}>
        <View style={styles.impactIcon}><Ionicons name="leaf" size={26} color={colors.white} /></View>
        <View style={styles.impactCopy}>
          <Text style={styles.eyebrow}>IMPACTO DO BAIRRO</Text>
          <Text style={styles.impactTitle}>{impact ? `${Number(impact.kgHoje ?? 0).toLocaleString('pt-BR')} kg reaproveitados hoje` : 'Acompanhe o impacto da sua comunidade'}</Text>
          <Text style={styles.muted}>{impact?.bairro ? `Impacto de ${impact.bairro}` : 'Pequenas ações, grandes mudanças.'}</Text>
        </View>
        <View style={styles.arrow}><Ionicons name="chevron-forward" size={18} color={colors.primary} /></View>
      </Pressable>

      <View style={styles.search}>
        <Ionicons name="search" size={21} color={colors.textMuted} />
        <TextInput
          value={query}
          onChangeText={setQuery}
          onSubmitEditing={openSearch}
          returnKeyType="search"
          placeholder="Buscar material ou projeto"
          placeholderTextColor={colors.textSoft}
          style={styles.searchInput}
        />
        {query ? <Pressable accessibilityLabel="Limpar busca" onPress={() => setQuery('')}><Ionicons name="close-circle" size={20} color={colors.textMuted} /></Pressable> : <Pressable accessibilityLabel="Abrir projetos" onPress={openSearch}><Ionicons name="options-outline" size={20} color={colors.primary} /></Pressable>}
      </View>
      <Text style={styles.helper}>Tenho algo para descartar. O que posso fazer com isso?</Text>

      <ScrollView horizontal showsHorizontalScrollIndicator={false} contentContainerStyle={styles.chips}>
        {categories.map((item) => <Chip key={item} label={item} selected={category === item} onPress={() => setCategory(item)} />)}
      </ScrollView>

      <SectionHeader title="Destaque da semana" subtitle={loading ? 'Carregando catálogo...' : 'Uma ideia simples para começar hoje.'} actionLabel="Ver todos" onAction={() => navigate({ name: 'projects' })} />
      {error ? <Text style={styles.muted}>{error}</Text> : null}
      {featured ? <View style={styles.featuredWrap}><FeaturedProjectCard project={featured} onPress={() => navigate({ name: 'projectDetail', projectId: featured.id })} /></View> : <Text style={styles.muted}>Nenhum projeto publicado foi encontrado.</Text>}

      <View style={styles.nearbyHeader}>
        <SectionHeader title="Criações da vizinhança" subtitle="Projetos práticos compartilhados pela comunidade." />
      </View>
      {(nearby.length ? nearby : projects.filter((project) => project.id !== featured?.id).slice(0, 2)).map((project) => (
        <ProjectCard
          key={project.id}
          project={project}
          favorite={favoriteProjectIds.includes(project.id)}
          onFavorite={() => toggleFavorite(project.id)}
          onPress={() => navigate({ name: 'projectDetail', projectId: project.id })}
        />
      ))}
    </ScreenShell>
  );
}

const styles = StyleSheet.create({
  impact: { marginTop: spacing.lg, padding: spacing.md, borderRadius: radius.lg, backgroundColor: colors.mint, flexDirection: 'row', alignItems: 'center', gap: 12, borderWidth: 1, borderColor: '#D1EBDD' },
  impactIcon: { width: 54, height: 54, borderRadius: 18, backgroundColor: colors.primary, alignItems: 'center', justifyContent: 'center' },
  impactCopy: { flex: 1 },
  eyebrow: { color: colors.primary, fontSize: 10.5, fontWeight: '900', letterSpacing: 0.8 },
  impactTitle: { color: colors.text, fontWeight: '900', fontSize: 16.5, lineHeight: 21, marginTop: 2 },
  muted: { color: colors.textMuted, fontSize: 12.5, lineHeight: 18 },
  arrow: { width: 34, height: 34, borderRadius: 17, backgroundColor: 'rgba(255,255,255,0.74)', alignItems: 'center', justifyContent: 'center' },
  search: { marginTop: spacing.md, minHeight: 54, backgroundColor: colors.surface, borderRadius: radius.md, borderWidth: 1, borderColor: colors.border, flexDirection: 'row', alignItems: 'center', paddingHorizontal: 14 },
  searchInput: { flex: 1, minHeight: 50, marginLeft: 8, color: colors.text, fontSize: 15 },
  helper: { color: colors.textMuted, fontSize: 11.5, marginTop: 7 },
  chips: { gap: 8, paddingVertical: spacing.md },
  featuredWrap: { marginTop: 12, marginBottom: spacing.lg },
  nearbyHeader: { marginBottom: 12 },
});
