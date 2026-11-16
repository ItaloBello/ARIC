import React, { useEffect, useMemo, useState } from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { Chip } from '../components/Chip';
import { ScreenShell } from '../components/ScreenShell';
import { useApp } from '../context/AppContext';
import { useAuth } from '../context/AuthContext';
import { rankingUsers } from '../data';
import { listenLeaderboard, type LeaderboardEntry } from '../services/progress';
import { colors, radius, shadow, spacing } from '../theme';

type RankingFilter = 'Bairro' | 'Cidade' | 'Mensal';

type RankingRow = LeaderboardEntry & { current: boolean; position: number };

export function RankingScreen({ goBack }: { goBack: () => void }) {
  const { user, cloudProgressEnabled } = useApp();
  const { uid } = useAuth();
  const [filter, setFilter] = useState<RankingFilter>('Bairro');
  const [onlineRows, setOnlineRows] = useState<LeaderboardEntry[]>([]);

  useEffect(() => {
    if (!cloudProgressEnabled) {
      setOnlineRows([]);
      return;
    }
    return listenLeaderboard(setOnlineRows, () => undefined);
  }, [cloudProgressEnabled]);

  const ranking = useMemo<RankingRow[]>(() => {
    if (cloudProgressEnabled && onlineRows.length) {
      const rows = onlineRows.map((item) => ({ ...item, current: item.uid === uid }));
      if (!rows.some((item) => item.current) && uid) {
        rows.push({ uid, name: user.name, neighborhood: user.city, points: user.points, title: user.title, current: true });
      }
      return rows.sort((a, b) => b.points - a.points).map((item, index) => ({ ...item, position: index + 1 }));
    }

    const current: LeaderboardEntry & { current: true } = { uid: 'current', name: user.name, neighborhood: user.city, points: user.points, title: user.title, current: true };
    return [...rankingUsers.map((item, index) => ({ uid: `demo-${index}`, ...item, current: false })), current]
      .sort((a, b) => b.points - a.points)
      .map((item, index) => ({ ...item, position: index + 1 }));
  }, [cloudProgressEnabled, onlineRows, uid, user.city, user.name, user.points, user.title]);

  const me = ranking.find((item) => item.current) ?? ranking[0];
  const top3 = ranking.slice(0, 3);
  const rest = ranking.slice(3, 10);
  const tenth = ranking[9]?.points ?? 0;
  const pointsToTop10 = me && me.position > 10 ? Math.max(0, tenth - user.points + 1) : 0;

  return (
    <ScreenShell>
      <View style={styles.topbar}><Pressable accessibilityRole="button" accessibilityLabel="Voltar" onPress={goBack} style={styles.back}><Ionicons name="arrow-back" size={22} color={colors.primary} /></Pressable><Text style={styles.topTitle}>Ranking Circular</Text><View style={{ width: 42 }} /></View>
      <View style={styles.hero}><View style={styles.trophy}><Ionicons name="trophy" size={34} color={colors.primary} /></View><View style={{ flex: 1 }}><Text style={styles.title}>Ranking Circular</Text><Text style={styles.muted}>Reconhecendo quem transforma boas ideias em ações.</Text></View></View>
      <View style={styles.filters}>{(['Bairro', 'Cidade', 'Mensal'] as RankingFilter[]).map((item) => <Chip key={item} label={item} selected={filter === item} onPress={() => setFilter(item)} />)}</View>

      <View style={styles.podium}>
        {top3.map((item, index) => (
          <View key={`${item.uid}-${index}`} style={[styles.podiumCard, index === 0 && styles.firstCard]}>
            <Text style={styles.medal}>{index === 0 ? '👑' : index === 1 ? '🥈' : '🥉'}</Text>
            <View style={[styles.avatar, item.current && styles.currentAvatar]}><Text style={styles.avatarText}>{item.name.charAt(0)}</Text></View>
            <Text style={styles.podiumName} numberOfLines={1}>{item.name}</Text>
            <Text style={styles.podiumCity} numberOfLines={1}>{item.neighborhood}</Text>
            <View style={styles.pointsPill}><Ionicons name="leaf" size={13} color={colors.primary} /><Text style={styles.podiumPoints}>{item.points} pts</Text></View>
          </View>
        ))}
      </View>

      <View style={styles.table}>
        {rest.map((item) => (
          <View key={`${item.uid}-${item.position}`} style={[styles.row, item.current && styles.currentRow]}>
            <Text style={styles.position}>{item.position}</Text>
            <View style={styles.smallAvatar}><Text style={styles.smallAvatarText}>{item.name.charAt(0)}</Text></View>
            <View style={{ flex: 1 }}><Text style={styles.rowName}>{item.name}</Text><Text style={styles.rowCity}>{item.neighborhood}</Text></View>
            <Text style={styles.rowPoints}>{item.points} pts</Text>
          </View>
        ))}
      </View>

      {me ? (
        <View style={styles.meCard}>
          <View style={styles.avatar}><Text style={styles.avatarText}>{user.name.charAt(0)}</Text></View>
          <View style={{ flex: 1 }}><Text style={styles.meTitle}>Sua posição: #{me.position}</Text><Text style={styles.mePoints}>{user.points} EcoPontos</Text><Text style={styles.meHint}>{me.position <= 10 ? 'Você está no Top 10. Continue participando!' : `Faltam ${pointsToTop10} pts para entrar no Top 10.`}</Text></View>
        </View>
      ) : null}
      <Text style={styles.note}>{cloudProgressEnabled ? 'O ranking do backend Go usa as ligas semanais e os EcoPontos calculados no servidor.' : `Filtro selecionado: ${filter}. Modo demonstração com dados locais.`}</Text>
    </ScreenShell>
  );
}

const styles = StyleSheet.create({
  topbar: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' },
  back: { width: 42, height: 42, borderRadius: 21, backgroundColor: colors.surface, borderWidth: 1, borderColor: colors.border, alignItems: 'center', justifyContent: 'center' },
  topTitle: { color: colors.text, fontWeight: '900', fontSize: 18 },
  hero: { marginTop: spacing.lg, flexDirection: 'row', gap: 12, alignItems: 'center' },
  trophy: { width: 60, height: 60, borderRadius: 20, backgroundColor: colors.mint, alignItems: 'center', justifyContent: 'center' },
  title: { color: colors.text, fontWeight: '900', fontSize: 27 },
  muted: { color: colors.textMuted, fontSize: 12, lineHeight: 17 },
  filters: { flexDirection: 'row', gap: 8, marginTop: spacing.md, flexWrap: 'wrap' },
  podium: { flexDirection: 'row', gap: 7, alignItems: 'flex-end', marginTop: spacing.lg },
  podiumCard: { ...shadow, flex: 1, backgroundColor: colors.surface, borderRadius: radius.lg, borderWidth: 1, borderColor: colors.border, padding: 9, alignItems: 'center', minHeight: 184 },
  firstCard: { backgroundColor: colors.mintSoft, borderColor: '#CFE8D9', minHeight: 202 },
  medal: { fontSize: 23 },
  avatar: { width: 54, height: 54, borderRadius: 27, backgroundColor: colors.primary, alignItems: 'center', justifyContent: 'center', marginTop: 3 },
  currentAvatar: { backgroundColor: colors.orange },
  avatarText: { color: colors.white, fontWeight: '900', fontSize: 21 },
  podiumName: { color: colors.text, fontWeight: '900', marginTop: 8, fontSize: 12.5 },
  podiumCity: { color: colors.textMuted, fontSize: 9.5, marginTop: 2, textAlign: 'center' },
  pointsPill: { flexDirection: 'row', alignItems: 'center', gap: 3, backgroundColor: colors.white, paddingVertical: 5, paddingHorizontal: 7, borderRadius: radius.pill, marginTop: 8 },
  podiumPoints: { color: colors.primary, fontWeight: '900', fontSize: 11 },
  table: { backgroundColor: colors.surface, borderRadius: radius.lg, borderWidth: 1, borderColor: colors.border, marginTop: spacing.md, overflow: 'hidden' },
  row: { flexDirection: 'row', alignItems: 'center', gap: 9, paddingHorizontal: 12, paddingVertical: 10, borderBottomWidth: 1, borderBottomColor: colors.border },
  currentRow: { backgroundColor: colors.mint },
  position: { width: 22, color: colors.textMuted, fontWeight: '900' },
  smallAvatar: { width: 34, height: 34, borderRadius: 17, backgroundColor: colors.mintStrong, alignItems: 'center', justifyContent: 'center' },
  smallAvatarText: { color: colors.primaryDark, fontWeight: '900' },
  rowName: { color: colors.text, fontWeight: '900', fontSize: 12.5 },
  rowCity: { color: colors.textMuted, fontSize: 9.5 },
  rowPoints: { color: colors.primary, fontWeight: '900', fontSize: 11.5 },
  meCard: { marginTop: spacing.md, backgroundColor: colors.mint, borderRadius: radius.lg, padding: spacing.md, flexDirection: 'row', gap: 12, alignItems: 'center', borderWidth: 1, borderColor: '#CDE8D7' },
  meTitle: { color: colors.primaryDark, fontWeight: '900', fontSize: 19 },
  mePoints: { color: colors.primaryDark, fontWeight: '800', marginTop: 2 },
  meHint: { color: colors.primaryDark, fontWeight: '700', marginTop: 5, fontSize: 11.5 },
  note: { color: colors.textSoft, fontSize: 10.5, lineHeight: 15, marginTop: 9 },
});
