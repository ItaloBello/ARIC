import React from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { ScreenShell } from '../components/ScreenShell';
import { useCommunity } from '../context/CommunityContext';
import { colors, radius, spacing } from '../theme';

export function MySubmissionsScreen({ goBack }: { goBack: () => void }) {
  const { mySubmissions, mode, loading } = useCommunity();
  return (
    <ScreenShell>
      <View style={styles.topbar}>
        <Pressable onPress={goBack} style={styles.back}><Ionicons name="arrow-back" size={22} color={colors.primary} /></Pressable>
        <Text style={styles.topTitle}>Minhas publicações</Text>
        <View style={{ width: 42 }} />
      </View>
      <Text style={styles.title}>Sua participação</Text>
      <Text style={styles.subtitle}>Posts e dúvidas publicados por você. No backend atual, eles ficam visíveis imediatamente. Fonte: {mode === 'api' ? 'API Go' : 'demonstração'}.</Text>
      {mySubmissions.map((item) => (
        <View key={item.id} style={styles.card}>
          <View style={styles.status}><Ionicons name="checkmark-circle" size={15} color={colors.success} /><Text style={styles.statusText}>Publicado</Text></View>
          <Text style={styles.cardTitle}>{item.title}</Text>
          <Text style={styles.body} numberOfLines={5}>{item.body}</Text>
          <Text style={styles.date}>{item.createdAt}</Text>
        </View>
      ))}
      {loading ? <Text style={styles.emptyText}>Carregando publicações...</Text> : null}
      {!loading && mySubmissions.length === 0 ? (
        <View style={styles.empty}><Ionicons name="document-text-outline" size={34} color={colors.primary} /><Text style={styles.emptyTitle}>Você ainda não publicou no fórum</Text><Text style={styles.emptyText}>Quando publicar uma dúvida ou dica, ela aparecerá aqui.</Text></View>
      ) : null}
    </ScreenShell>
  );
}

const styles = StyleSheet.create({
  topbar: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' },
  back: { width: 42, height: 42, borderRadius: 21, backgroundColor: colors.surface, borderWidth: 1, borderColor: colors.border, alignItems: 'center', justifyContent: 'center' },
  topTitle: { color: colors.text, fontWeight: '900', fontSize: 18 },
  title: { color: colors.text, fontSize: 27, lineHeight: 33, fontWeight: '900', marginTop: spacing.lg },
  subtitle: { color: colors.textMuted, lineHeight: 19, marginTop: 5, marginBottom: spacing.md },
  card: { backgroundColor: colors.surface, borderWidth: 1, borderColor: colors.border, borderRadius: radius.lg, padding: 15, marginBottom: 12 },
  status: { alignSelf: 'flex-start', flexDirection: 'row', alignItems: 'center', gap: 5, borderRadius: radius.pill, backgroundColor: colors.mint, paddingHorizontal: 9, paddingVertical: 6 },
  statusText: { color: colors.success, fontWeight: '900', fontSize: 10.5 },
  cardTitle: { color: colors.text, fontWeight: '900', fontSize: 17, marginTop: 10 },
  body: { color: colors.textMuted, lineHeight: 19, fontSize: 12.5, marginTop: 5 },
  date: { color: colors.textSoft, fontSize: 10.5, marginTop: 10 },
  empty: { alignItems: 'center', paddingVertical: 46, paddingHorizontal: 26 },
  emptyTitle: { color: colors.text, fontWeight: '900', fontSize: 16, marginTop: 10 },
  emptyText: { color: colors.textMuted, textAlign: 'center', lineHeight: 18, marginTop: 5, fontSize: 12 },
});
