import React, { useEffect } from 'react';
import { Alert, Pressable, StyleSheet, Text, View } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { ScreenShell } from '../components/ScreenShell';
import { useAuth } from '../context/AuthContext';
import { useCommunity } from '../context/CommunityContext';
import { colors, radius, spacing } from '../theme';

export function AdminModerationScreen({ goBack }: { goBack: () => void }) {
  const { isAdmin } = useAuth();
  const { reports, loading, resolveReport, refreshCommunity } = useCommunity();

  useEffect(() => { if (isAdmin) void refreshCommunity(); }, [isAdmin, refreshCommunity]);

  if (!isAdmin) {
    return <ScreenShell><Text style={styles.denied}>Acesso restrito à equipe de moderação.</Text></ScreenShell>;
  }

  const decide = (reportId: string, action: 'remover' | 'manter', title: string) => {
    Alert.alert(action === 'remover' ? 'Remover conteúdo?' : 'Manter conteúdo?', action === 'remover'
      ? 'O conteúdo será retirado do ar e as denúncias relacionadas serão marcadas como procedentes.'
      : 'O conteúdo continuará disponível e as denúncias relacionadas serão marcadas como improcedentes.', [
      { text: 'Cancelar', style: 'cancel' },
      { text: action === 'remover' ? 'Remover' : 'Manter', style: action === 'remover' ? 'destructive' : 'default', onPress: async () => {
        const result = await resolveReport(reportId, action, `Decisão da moderação: ${title}`);
        Alert.alert(result.ok ? 'Análise concluída' : 'Não foi possível concluir', result.ok ? 'A decisão foi registrada no servidor.' : result.message);
      } },
    ]);
  };

  return (
    <ScreenShell>
      <View style={styles.topbar}>
        <Pressable onPress={goBack} style={styles.back}><Ionicons name="arrow-back" size={22} color={colors.primary} /></Pressable>
        <Text style={styles.topTitle}>Painel de moderação</Text>
        <Pressable onPress={() => void refreshCommunity()} style={styles.back}><Ionicons name="refresh" size={20} color={colors.primary} /></Pressable>
      </View>
      <View style={styles.hero}>
        <View style={styles.heroIcon}><Ionicons name="shield-checkmark" size={27} color={colors.primary} /></View>
        <View style={{ flex: 1 }}>
          <Text style={styles.title}>Denúncias da comunidade</Text>
          <Text style={styles.subtitle}>O backend Go reúne denúncias por conteúdo. A equipe decide se o conteúdo deve ser removido ou mantido.</Text>
        </View>
      </View>

      {loading ? <Text style={styles.emptyText}>Carregando denúncias...</Text> : null}
      {reports.map((report) => (
        <View key={report.id} style={styles.card}>
          <View style={styles.metaRow}>
            <Text style={styles.author}>{report.reason}</Text>
            <Text style={styles.meta}>{report.createdAt}</Text>
          </View>
          <Text style={styles.cardTitle}>{report.postTitle}</Text>
          {report.details ? <Text style={styles.body}>{report.details}</Text> : null}
          <View style={styles.actions}>
            <Pressable onPress={() => decide(report.id, 'manter', report.postTitle)} style={[styles.actionButton, styles.keep]}>
              <Ionicons name="checkmark-circle-outline" size={17} color={colors.primary} /><Text style={styles.keepText}>Manter no ar</Text>
            </Pressable>
            <Pressable onPress={() => decide(report.id, 'remover', report.postTitle)} style={[styles.actionButton, styles.remove]}>
              <Ionicons name="trash-outline" size={17} color={colors.white} /><Text style={styles.removeText}>Remover</Text>
            </Pressable>
          </View>
        </View>
      ))}
      {!loading && reports.length === 0 ? (
        <View style={styles.empty}>
          <Ionicons name="shield-checkmark-outline" size={36} color={colors.primary} />
          <Text style={styles.emptyTitle}>Nenhuma denúncia pendente</Text>
          <Text style={styles.emptyText}>Quando uma publicação for denunciada, ela aparecerá aqui para análise.</Text>
        </View>
      ) : null}
    </ScreenShell>
  );
}

const styles = StyleSheet.create({
  denied: { color: colors.danger, fontWeight: '900', textAlign: 'center', marginTop: 80 },
  topbar: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' },
  back: { width: 42, height: 42, borderRadius: 21, backgroundColor: colors.surface, borderWidth: 1, borderColor: colors.border, alignItems: 'center', justifyContent: 'center' },
  topTitle: { color: colors.text, fontWeight: '900', fontSize: 18 },
  hero: { flexDirection: 'row', gap: 12, marginTop: spacing.lg, alignItems: 'center', marginBottom: spacing.md },
  heroIcon: { width: 56, height: 56, borderRadius: 19, backgroundColor: colors.mint, alignItems: 'center', justifyContent: 'center' },
  title: { color: colors.text, fontWeight: '900', fontSize: 20 },
  subtitle: { color: colors.textMuted, fontSize: 12, lineHeight: 18, marginTop: 3 },
  card: { backgroundColor: colors.surface, borderWidth: 1, borderColor: colors.border, borderRadius: radius.lg, padding: 15, marginBottom: 12 },
  metaRow: { flexDirection: 'row', justifyContent: 'space-between', gap: 8 },
  author: { color: colors.orange, fontWeight: '900', fontSize: 11.5, flex: 1 },
  meta: { color: colors.textSoft, fontSize: 10.5 },
  cardTitle: { color: colors.text, fontWeight: '900', fontSize: 17, marginTop: 10 },
  body: { color: colors.textMuted, lineHeight: 18, fontSize: 12.5, marginTop: 6 },
  actions: { flexDirection: 'row', gap: 9, marginTop: 14 },
  actionButton: { flex: 1, minHeight: 43, borderRadius: radius.md, flexDirection: 'row', alignItems: 'center', justifyContent: 'center', gap: 6 },
  keep: { backgroundColor: colors.mint },
  keepText: { color: colors.primary, fontWeight: '900', fontSize: 12 },
  remove: { backgroundColor: colors.danger },
  removeText: { color: colors.white, fontWeight: '900', fontSize: 12 },
  empty: { alignItems: 'center', paddingVertical: 42, paddingHorizontal: 25 },
  emptyTitle: { color: colors.text, fontWeight: '900', fontSize: 16, marginTop: 10 },
  emptyText: { color: colors.textMuted, textAlign: 'center', lineHeight: 18, fontSize: 12, marginTop: 5 },
});
