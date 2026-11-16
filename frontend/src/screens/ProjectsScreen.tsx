import React, { useMemo, useState } from 'react';
import { ScrollView, StyleSheet, Text, TextInput, View } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { BrandHeader } from '../components/BrandHeader';
import { Chip } from '../components/Chip';
import { ProjectCard } from '../components/ProjectCard';
import { ScreenShell } from '../components/ScreenShell';
import { useApp } from '../context/AppContext';
import { useProjects } from '../hooks/useProjects';
import type { Route } from '../navigation';
import type { MaterialCategory } from '../types';
import { colors, radius, spacing, typography } from '../theme';

const categories: MaterialCategory[] = ['Todos', 'Plástico', 'Papelão', 'Vidro', 'Metal', 'Madeira', 'Horta'];

export function ProjectsScreen({ navigate, initialQuery, initialCategory }: { navigate: (route: Route) => void; initialQuery?: string; initialCategory?: MaterialCategory }) {
  const { favoriteProjectIds, toggleFavorite } = useApp();
  const { projects, loading, error } = useProjects();
  const [query, setQuery] = useState(initialQuery ?? '');
  const [category, setCategory] = useState<MaterialCategory>(initialCategory ?? 'Todos');

  const list = useMemo(() => projects.filter((project) => {
    const q = query.toLowerCase().trim();
    const text = `${project.title} ${project.description} ${project.category} ${project.secondaryCategory ?? ''} ${project.materials.join(' ')}`.toLowerCase();
    return (category === 'Todos' || project.category === category || project.secondaryCategory === category) && (!q || text.includes(q));
  }), [category, query]);

  return (
    <ScreenShell>
      <BrandHeader onProfile={() => navigate({ name: 'profile' })} />
      <Text style={styles.title}>Projetos sustentáveis</Text>
      <Text style={styles.subtitle}>Escolha um material e descubra formas seguras e criativas de reaproveitá-lo.</Text>

      <View style={styles.search}>
        <Ionicons name="search" size={21} color={colors.textMuted} />
        <TextInput value={query} onChangeText={setQuery} placeholder="Ex.: garrafa PET, papelão, lata..." placeholderTextColor={colors.textSoft} style={styles.input} returnKeyType="search" />
      </View>
      <ScrollView horizontal showsHorizontalScrollIndicator={false} contentContainerStyle={styles.chips}>
        {categories.map((item) => <Chip key={item} label={item} selected={item === category} onPress={() => setCategory(item)} />)}
      </ScrollView>

      {error ? <Text style={styles.subtitle}>{error}</Text> : null}
      {loading ? <Text style={styles.count}>Carregando projetos do servidor...</Text> : null}
      <Text style={styles.count}>{list.length} projeto{list.length === 1 ? '' : 's'} encontrado{list.length === 1 ? '' : 's'}</Text>
      {list.map((project) => (
        <ProjectCard
          key={project.id}
          project={project}
          favorite={favoriteProjectIds.includes(project.id)}
          onFavorite={() => toggleFavorite(project.id)}
          onPress={() => navigate({ name: 'projectDetail', projectId: project.id })}
        />
      ))}
      {list.length === 0 ? (
        <View style={styles.empty}>
          <Ionicons name="search-outline" size={34} color={colors.primary} />
          <Text style={styles.emptyTitle}>Nenhum projeto encontrado</Text>
          <Text style={styles.emptyText}>Tente outro material ou uma palavra mais simples.</Text>
        </View>
      ) : null}
    </ScreenShell>
  );
}

const styles = StyleSheet.create({
  title: { ...typography.title, marginTop: spacing.lg },
  subtitle: { ...typography.muted, marginTop: 5 },
  search: { marginTop: spacing.md, backgroundColor: colors.surface, borderRadius: radius.md, borderWidth: 1, borderColor: colors.border, flexDirection: 'row', alignItems: 'center', paddingHorizontal: 14 },
  input: { flex: 1, minHeight: 52, marginLeft: 8, color: colors.text },
  chips: { gap: 8, paddingVertical: spacing.md },
  count: { color: colors.textMuted, fontWeight: '800', marginBottom: 12, fontSize: 12.5 },
  empty: { alignItems: 'center', padding: 30, backgroundColor: colors.surface, borderRadius: radius.lg, borderWidth: 1, borderColor: colors.border },
  emptyTitle: { color: colors.text, fontSize: 17, fontWeight: '900', marginTop: 8 },
  emptyText: { color: colors.textMuted, marginTop: 4, textAlign: 'center' },
});
