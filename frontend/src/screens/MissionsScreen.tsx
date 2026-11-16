import React from 'react';
import { Alert, Pressable, StyleSheet, Text, View } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { AppImage } from '../components/AppImage';
import { BrandHeader } from '../components/BrandHeader';
import { ProgressBar } from '../components/ProgressBar';
import { ScreenShell } from '../components/ScreenShell';
import { useApp } from '../context/AppContext';
import { useAuth } from '../context/AuthContext';
import { missions } from '../data';
import { colors, radius, shadow, spacing } from '../theme';

export function MissionsScreen() {
  const { user, missionProgress, completedMissionIds, incrementMission, completeMission } = useApp();
  const { isGuest, signOut } = useAuth();
  const level = Math.max(1, Math.floor(user.points / 250) + 2);
  const currentBase = (level - 2) * 250;
  const nextAt = currentBase + 250;
  const progress = Math.max(0, Math.min(1, (user.points - currentBase) / 250));

  const interact = async (missionId: string, goal: number) => {
    if (isGuest) {
      Alert.alert('Entre para participar das missões', 'O progresso e os EcoPontos precisam ficar associados a uma conta.', [
        { text: 'Cancelar', style: 'cancel' }, { text: 'Entrar', onPress: () => signOut() },
      ]);
      return;
    }
    const current = missionProgress[missionId] ?? 0;
    if (completedMissionIds.includes(missionId)) return Alert.alert('Missão concluída', 'A recompensa desta missão já foi creditada.');
    const result = current + 1 >= goal ? await completeMission(missionId) : await incrementMission(missionId);
    if (!result.ok) return Alert.alert('Não foi possível registrar', result.message);
    if (result.completedNow) Alert.alert('Missão concluída! 🎉', `Você ganhou ${result.reward ?? 0} EcoPontos.`);
  };

  return (
    <ScreenShell>
      <BrandHeader />
      <View style={styles.levelCard}>
        <View style={styles.levelBadge}><Text style={styles.levelLabel}>NÍVEL</Text><Text style={styles.level}>{level}</Text></View>
        <View style={styles.levelCopy}>
          <Text style={styles.kicker}>SUA JORNADA CONTINUA</Text>
          <Text style={styles.guardian}>Guardião da Terra</Text>
          <Text style={styles.points}>{user.points} EcoPontos</Text>
          <View style={styles.progressWrap}><ProgressBar value={progress} height={10} accent="#BFE89A" /></View>
          <Text style={styles.next}>Faltam {Math.max(0, nextAt - user.points)} pts para o próximo nível</Text>
        </View>
        <Ionicons name="trophy" size={39} color="#FFD465" />
      </View>

      <Pressable accessibilityRole="button" style={styles.benefits} onPress={() => Alert.alert('Clube de Benefícios', 'Área planejada para parceiros locais e recompensas sustentáveis.') }>
        <View style={styles.gift}><Ionicons name="gift-outline" size={27} color={colors.primary} /></View>
        <View style={{ flex: 1 }}><Text style={styles.benefitsTitle}>Clube de Benefícios</Text><Text style={styles.benefitsText}>No futuro, EcoPontos poderão virar vantagens com parceiros.</Text></View>
        <Ionicons name="chevron-forward" size={19} color={colors.primary} />
      </Pressable>

      <View style={styles.sectionTitleRow}><Ionicons name="flag-outline" size={26} color={colors.primary} /><View><Text style={styles.sectionTitle}>Missões da Semana</Text><Text style={styles.sectionSub}>Complete ações reais e acompanhe seu progresso.</Text></View></View>

      {missions.map((mission) => {
        const current = missionProgress[mission.id] ?? 0;
        const done = completedMissionIds.includes(mission.id);
        const value = done ? 1 : current / mission.goal;
        return (
          <View key={mission.id} style={styles.card}>
            <View style={styles.image}><AppImage source={mission.image} height={126} borderRadius={radius.md} /></View>
            <View style={styles.cardBody}>
              <View style={styles.cardTop}><View style={styles.category}><Text style={styles.categoryText}>{mission.category}</Text></View><Text style={styles.reward}>+{mission.reward} pts</Text></View>
              <Text style={styles.missionTitle}>{mission.title}</Text>
              <Text style={styles.description}>{mission.description}</Text>
              <View style={styles.progressRow}><Text style={styles.progressLabel}>{done ? mission.goal : current} de {mission.goal} {mission.unit}</Text><Text style={styles.percent}>{Math.round(value * 100)}%</Text></View>
              <ProgressBar value={value} />
              <Pressable accessibilityRole="button" onPress={() => interact(mission.id, mission.goal)} style={[styles.action, done && styles.actionDone]}>
                <Ionicons name={done ? 'checkmark-circle' : 'add-circle-outline'} size={18} color={done ? colors.success : colors.primary} />
                <Text style={[styles.actionText, done && { color: colors.success }]}>{done ? 'Missão concluída' : 'Registrar progresso'}</Text>
              </Pressable>
            </View>
          </View>
        );
      })}
    </ScreenShell>
  );
}

const styles = StyleSheet.create({
  levelCard: { ...shadow, marginTop: spacing.lg, borderRadius: radius.xl, backgroundColor: colors.primaryDeep, padding: spacing.md, flexDirection: 'row', alignItems: 'center', gap: 12, overflow: 'hidden' },
  levelBadge: { width: 72, height: 90, borderRadius: 24, borderWidth: 3, borderColor: '#BCE6A4', alignItems: 'center', justifyContent: 'center', backgroundColor: 'rgba(5,101,63,0.7)' },
  levelLabel: { color: '#D7F5C6', fontWeight: '900', fontSize: 10 },
  level: { color: colors.white, fontWeight: '900', fontSize: 38, lineHeight: 42 },
  levelCopy: { flex: 1 },
  kicker: { color: '#BFE8CE', fontWeight: '800', fontSize: 9.5, letterSpacing: 0.7 },
  guardian: { color: colors.white, fontWeight: '900', fontSize: 20, marginTop: 2 },
  points: { color: '#DDF5E5', fontWeight: '800', fontSize: 13, marginTop: 2 },
  progressWrap: { marginTop: 10 },
  next: { color: 'rgba(255,255,255,0.76)', fontSize: 9.5, marginTop: 5 },
  benefits: { backgroundColor: colors.mint, borderRadius: radius.lg, borderWidth: 1, borderColor: '#D5ECDE', padding: 13, flexDirection: 'row', alignItems: 'center', gap: 11, marginTop: 13 },
  gift: { width: 48, height: 48, borderRadius: 17, backgroundColor: 'rgba(255,255,255,0.72)', alignItems: 'center', justifyContent: 'center' },
  benefitsTitle: { color: colors.text, fontWeight: '900', fontSize: 15.5 },
  benefitsText: { color: colors.textMuted, fontSize: 11.5, lineHeight: 16, marginTop: 2 },
  sectionTitleRow: { flexDirection: 'row', gap: 10, alignItems: 'flex-start', marginTop: spacing.lg, marginBottom: 12 },
  sectionTitle: { color: colors.text, fontWeight: '900', fontSize: 22 },
  sectionSub: { color: colors.textMuted, fontSize: 12, marginTop: 2 },
  card: { ...shadow, backgroundColor: colors.surface, borderRadius: radius.lg, borderWidth: 1, borderColor: colors.border, padding: 10, marginBottom: 12, flexDirection: 'row', gap: 12 },
  image: { width: 118, overflow: 'hidden', borderRadius: radius.md },
  cardBody: { flex: 1, minWidth: 0 },
  cardTop: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', gap: 7 },
  category: { backgroundColor: colors.mint, paddingHorizontal: 8, paddingVertical: 5, borderRadius: radius.pill },
  categoryText: { color: colors.primaryDark, fontWeight: '900', fontSize: 9.5 },
  reward: { color: colors.orange, fontWeight: '900', fontSize: 11.5 },
  missionTitle: { color: colors.text, fontWeight: '900', fontSize: 15, lineHeight: 19, marginTop: 7 },
  description: { color: colors.textMuted, fontSize: 11.5, lineHeight: 16, marginTop: 3 },
  progressRow: { flexDirection: 'row', justifyContent: 'space-between', marginTop: 9, marginBottom: 5 },
  progressLabel: { color: colors.textMuted, fontSize: 10 },
  percent: { color: colors.primary, fontWeight: '900', fontSize: 10.5 },
  action: { marginTop: 9, minHeight: 38, borderRadius: radius.sm, backgroundColor: colors.mintSoft, flexDirection: 'row', alignItems: 'center', justifyContent: 'center', gap: 5 },
  actionDone: { backgroundColor: '#EAF8EF' },
  actionText: { color: colors.primary, fontWeight: '900', fontSize: 11 },
});
