import React from 'react';
import { Alert, Pressable, StyleSheet, Text, View } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { AppImage } from '../components/AppImage';
import { PrimaryButton } from '../components/PrimaryButton';
import { ScreenShell } from '../components/ScreenShell';
import { useApp } from '../context/AppContext';
import { useAuth } from '../context/AuthContext';
import { useProjectDetail } from '../hooks/useProjects';
import { colors, radius, shadow, spacing } from '../theme';

export function ProjectDetailScreen({ projectId, goBack }: { projectId: string; goBack: () => void }) {
  const { favoriteProjectIds, toggleFavorite, completedProjectIds, completeProject } = useApp();
  const { isGuest } = useAuth();
  const { project, loading, error } = useProjectDetail(projectId);
  if (!project) return <ScreenShell><Text style={styles.title}>{loading ? 'Carregando projeto...' : 'Projeto indisponível'}</Text><Text style={styles.subtitle}>{error || 'Não foi possível localizar este projeto.'}</Text></ScreenShell>;
  const favorite = favoriteProjectIds.includes(project.id) || project.saved === true;
  const completed = completedProjectIds.includes(project.id) || project.completed === true;

  const finish = async () => {
    if (isGuest) {
      Alert.alert('Entre para registrar seu impacto', 'Você pode explorar o tutorial como visitante, mas EcoPontos e progresso ficam vinculados a uma conta.');
      return;
    }
    if (completed) return Alert.alert('Projeto já concluído', 'Os EcoPontos deste projeto já foram creditados.');
    const result = await completeProject(project.id);
    if (!result.ok) return Alert.alert('Não foi possível registrar', result.message);
    if (!result.awarded) return Alert.alert('Projeto já concluído', 'Os EcoPontos deste projeto já foram creditados.');
    Alert.alert('Projeto concluído! 🌱', `Você ganhou ${result.reward ?? project.points} EcoPontos e registrou ${project.impactKg.toFixed(1)} kg de material reaproveitado.`);
  };

  return (
    <ScreenShell padded={false}>
      <View style={styles.topbar}>
        <Pressable accessibilityRole="button" accessibilityLabel="Voltar" onPress={goBack} style={styles.circle}><Ionicons name="arrow-back" size={22} color={colors.primary} /></Pressable>
        <Text style={styles.topTitle}>Projeto</Text>
        <Pressable accessibilityRole="button" accessibilityLabel={favorite ? 'Remover favorito' : 'Salvar favorito'} onPress={() => { void toggleFavorite(project.id); }} style={styles.circle}><Ionicons name={favorite ? 'bookmark' : 'bookmark-outline'} size={22} color={colors.primary} /></Pressable>
      </View>

      <AppImage source={project.image} height={300} />
      <View style={styles.content}>
        <View style={styles.tags}><Tag icon="leaf-outline" label={project.category} /><Tag icon="sparkles-outline" label={project.secondaryCategory ?? 'Upcycling'} /></View>
        <Text style={styles.title}>{project.title}</Text>
        <Text style={styles.subtitle}>{project.subtitle}</Text>
        <View style={styles.metrics}>
          <Metric icon="time-outline" label={project.time} />
          <Metric icon="stats-chart-outline" label={project.difficulty} />
          <Metric icon="ribbon-outline" label={`+${project.points} pts`} accent />
        </View>

        <Text style={styles.description}>{project.description}</Text>

        <View style={styles.materialCard}>
          <View style={styles.cardTitleRow}><Ionicons name="basket-outline" size={23} color={colors.primary} /><Text style={styles.cardTitle}>Materiais necessários</Text></View>
          {project.materials.map((material, index) => <View style={styles.material} key={material}><View style={styles.number}><Text style={styles.numberText}>{index + 1}</Text></View><Text style={styles.materialText}>{material}</Text></View>)}
        </View>

        <View style={styles.stepsHeader}><Text style={styles.cardTitle}>Passo a passo</Text><Text style={styles.stepsCount}>{project.steps.length} etapas</Text></View>
        {project.steps.map((step, index) => (
          <View key={step.title} style={styles.step}>
            <View style={styles.stepNumber}><Text style={styles.stepNumberText}>{index + 1}</Text></View>
            <View style={styles.stepCopy}><Text style={styles.stepTitle}>{step.title}</Text><Text style={styles.stepText}>{step.description}</Text></View>
          </View>
        ))}

        <PrimaryButton label={completed ? 'Projeto concluído' : 'Concluir projeto'} icon={completed ? 'checkmark-circle' : 'play-circle-outline'} onPress={finish} variant={completed ? 'secondary' : 'primary'} />
        <View style={styles.social}><Ionicons name="people-outline" size={19} color={colors.primary} /><Text style={styles.socialText}>{project.likes + 957} pessoas já fizeram ou salvaram este projeto</Text></View>
      </View>
    </ScreenShell>
  );
}

function Tag({ icon, label }: { icon: keyof typeof Ionicons.glyphMap; label: string }) {
  return <View style={styles.tag}><Ionicons name={icon} size={14} color={colors.primary} /><Text style={styles.tagText}>{label}</Text></View>;
}
function Metric({ icon, label, accent }: { icon: keyof typeof Ionicons.glyphMap; label: string; accent?: boolean }) {
  return <View style={styles.metric}><Ionicons name={icon} size={18} color={accent ? colors.orange : colors.primary} /><Text style={[styles.metricText, accent && { color: colors.orange }]}>{label}</Text></View>;
}

const styles = StyleSheet.create({
  topbar: { paddingHorizontal: spacing.md, paddingVertical: 8, flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' },
  circle: { width: 42, height: 42, borderRadius: 21, backgroundColor: colors.surface, borderWidth: 1, borderColor: colors.border, alignItems: 'center', justifyContent: 'center' },
  topTitle: { color: colors.text, fontWeight: '900', fontSize: 18 },
  content: { padding: spacing.md, paddingBottom: spacing.xl },
  tags: { flexDirection: 'row', gap: 8, flexWrap: 'wrap' },
  tag: { backgroundColor: colors.mint, borderRadius: radius.pill, paddingHorizontal: 11, paddingVertical: 7, flexDirection: 'row', alignItems: 'center', gap: 5 },
  tagText: { color: colors.primaryDark, fontWeight: '900', fontSize: 11.5 },
  title: { color: colors.text, fontWeight: '900', fontSize: 30, lineHeight: 35, marginTop: 12 },
  subtitle: { color: colors.textMuted, fontSize: 15, lineHeight: 21, marginTop: 5 },
  metrics: { flexDirection: 'row', flexWrap: 'wrap', gap: 15, paddingVertical: spacing.md, borderBottomWidth: 1, borderBottomColor: colors.border },
  metric: { flexDirection: 'row', alignItems: 'center', gap: 5 },
  metricText: { color: colors.text, fontWeight: '800', fontSize: 12.5 },
  description: { color: colors.textMuted, fontSize: 14.5, lineHeight: 23, marginVertical: spacing.md },
  materialCard: { ...shadow, backgroundColor: colors.mintSoft, borderRadius: radius.lg, borderWidth: 1, borderColor: '#D7EEDF', padding: spacing.md, gap: 10 },
  cardTitleRow: { flexDirection: 'row', alignItems: 'center', gap: 8, marginBottom: 2 },
  cardTitle: { color: colors.text, fontWeight: '900', fontSize: 20 },
  material: { backgroundColor: 'rgba(255,255,255,0.72)', borderRadius: radius.md, padding: 10, flexDirection: 'row', alignItems: 'center', gap: 10 },
  number: { width: 28, height: 28, borderRadius: 14, backgroundColor: colors.primary, alignItems: 'center', justifyContent: 'center' },
  numberText: { color: colors.white, fontWeight: '900', fontSize: 12 },
  materialText: { flex: 1, color: colors.text, fontWeight: '700', fontSize: 13.5 },
  stepsHeader: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', marginTop: spacing.lg, marginBottom: 10 },
  stepsCount: { color: colors.primary, fontWeight: '900' },
  step: { backgroundColor: colors.surface, borderWidth: 1, borderColor: colors.border, borderRadius: radius.md, padding: 13, flexDirection: 'row', gap: 12, marginBottom: 9 },
  stepNumber: { width: 36, height: 36, borderRadius: 18, backgroundColor: colors.mint, alignItems: 'center', justifyContent: 'center' },
  stepNumberText: { color: colors.primary, fontWeight: '900' },
  stepCopy: { flex: 1 },
  stepTitle: { color: colors.text, fontWeight: '900', fontSize: 14.5 },
  stepText: { color: colors.textMuted, fontSize: 12.5, lineHeight: 18, marginTop: 2 },
  social: { flexDirection: 'row', alignItems: 'center', gap: 7, justifyContent: 'center', marginTop: 14 },
  socialText: { color: colors.textMuted, fontSize: 11.5, flexShrink: 1 },
});
