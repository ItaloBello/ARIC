import React, { useState } from 'react';
import { Alert, Modal, Pressable, StyleSheet, Text, TextInput, View } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { AppImage } from '../components/AppImage';
import { BrandHeader } from '../components/BrandHeader';
import { PrimaryButton } from '../components/PrimaryButton';
import { ScreenShell } from '../components/ScreenShell';
import { useApp } from '../context/AppContext';
import { useAuth } from '../context/AuthContext';
import { useProjects } from '../hooks/useProjects';
import type { Route } from '../navigation';
import { colors, radius, shadow, spacing } from '../theme';

export function ProfileScreen({ navigate }: { navigate: (route: Route) => void }) {
  const { user, favoriteProjectIds, resetDemo } = useApp();
  const { projects } = useProjects();
  const { session, isGuest, isAdmin, signOut, deleteAccount } = useAuth();
  const [deleteModal, setDeleteModal] = useState(false);
  const [deletePassword, setDeletePassword] = useState('');
  const [deleteContent, setDeleteContent] = useState(false);

  const confirmDelete = () => {
    if (isGuest) return Alert.alert('Perfil de visitante', 'Nenhuma conta online foi criada neste modo.');
    Alert.alert('Excluir conta e dados?', 'A API vai anonimizar sua conta. Você pode escolher se os conteúdos publicados continuam atribuídos a “Usuário removido” ou se também serão apagados.', [
      { text: 'Cancelar', style: 'cancel' },
      { text: 'Continuar', style: 'destructive', onPress: () => setDeleteModal(true) },
    ]);
  };

  return (
    <ScreenShell>
      <BrandHeader />
      <View style={styles.profileCard}>
        <View style={styles.avatar}><Text style={styles.avatarText}>{(isGuest ? 'V' : user.name.charAt(0)).toUpperCase()}</Text><View style={styles.online} /></View>
        <View style={styles.identity}>
          <Text style={styles.name}>{isGuest ? 'Visitante' : user.name}</Text>
          <Text style={styles.city}><Ionicons name="location-outline" size={13} /> {isGuest ? 'Modo de exploração' : user.city}</Text>
          <View style={styles.badge}><Ionicons name="leaf" size={13} color={colors.orange} /><Text style={styles.badgeText}>{isGuest ? 'Explorador ARIC' : user.title}</Text></View>
        </View>
        {!isGuest ? <Pressable accessibilityRole="button" onPress={() => navigate({ name: 'editProfile' })} style={styles.edit}><Ionicons name="pencil" size={16} color={colors.primary} /><Text style={styles.editText}>Editar</Text></Pressable> : null}
        <Text style={styles.bio}>{isGuest ? 'Explore projetos, ideias e ações sustentáveis. Entre em uma conta para registrar seu impacto e participar da comunidade.' : user.bio}</Text>
        {isGuest ? <PrimaryButton label="Entrar ou criar conta" icon="log-in-outline" onPress={() => signOut()} /> : null}
      </View>

      {!isGuest ? (
        <>
          <View style={styles.miniGrid}>
            <View style={styles.miniCard}>
              <View style={[styles.miniIcon, { backgroundColor: colors.orangeSoft }]}><Ionicons name="flame-outline" size={23} color={colors.orange} /></View>
              <View><Text style={styles.miniLabel}>Sequência ativa</Text><Text style={styles.miniValue}>{user.streak} dias</Text></View>
            </View>
          </View>

          <View style={styles.sectionRow}><View><Text style={styles.title}>Seu impacto circular</Text><Text style={styles.muted}>Dados sincronizados com a comunidade ARIC.</Text></View><View style={styles.year}><Text style={styles.yearText}>2026</Text></View></View>
          <View style={styles.statsGrid}>
            <Stat icon="refresh-circle" value={`${user.impactKg.toFixed(1)} kg`} label="resíduos reutilizados" />
            <Stat icon="construct" value={`${user.projectsCompleted}`} label="projetos concluídos" accent />
            <Stat icon="people" value={`${user.peopleInspired}`} label="vidas inspiradas" />
          </View>

          <View style={styles.summaryCard}>
            <Text style={styles.summaryTitle}>Resumo da sua atividade</Text>
            <Summary icon="heart-outline" text={`${favoriteProjectIds.length} projeto(s) salvo(s)`} />
            <Summary icon="checkmark-circle-outline" text={`${user.projectsCompleted} projeto(s) concluído(s)`} />
            <Summary icon="leaf-outline" text={`${user.points} EcoPontos acumulados`} />
          </View>

          <View style={styles.sectionRow}><Text style={styles.title}>Projetos recentes</Text></View>
          <View style={styles.recentGrid}>
            {projects.slice(0, 2).map((project) => (
              <Pressable key={project.id} style={styles.recentCard} onPress={() => navigate({ name: 'projectDetail', projectId: project.id })}>
                <AppImage source={project.image} height={112} borderRadius={radius.md} />
                <Text style={styles.recentTitle} numberOfLines={2}>{project.title}</Text>
                <Text style={styles.muted}>{project.category} • {project.likes} curtidas</Text>
              </Pressable>
            ))}
          </View>
        </>
      ) : null}

      <View style={styles.settings}>
        <Text style={styles.settingsTitle}>Conta e privacidade</Text>
        {!isGuest ? <Setting icon="documents-outline" label="Minhas publicações" onPress={() => navigate({ name: 'mySubmissions' })} /> : null}
        {isAdmin ? <Setting icon="shield-half-outline" label="Painel de moderação" onPress={() => navigate({ name: 'adminModeration' })} /> : null}
        <Setting icon="shield-checkmark-outline" label="Privacidade e uso de dados" onPress={() => navigate({ name: 'privacy' })} />
        <Setting icon="people-outline" label="Regras e segurança da comunidade" onPress={() => navigate({ name: 'communityGuidelines' })} />
        {!isGuest ? <Setting icon="log-out-outline" label="Sair da conta" onPress={() => signOut()} /> : null}
        {!isGuest ? <Setting icon="trash-outline" label="Excluir conta" danger onPress={confirmDelete} /> : null}
        <Text style={styles.accountInfo}>{isGuest ? 'Sessão de visitante' : `Conta: ${session?.email ?? user.email ?? 'e-mail não informado'}`}</Text>
      </View>
      <Modal visible={deleteModal} transparent animationType="fade" onRequestClose={() => setDeleteModal(false)}>
        <View style={styles.modalBackdrop}>
          <View style={styles.deleteCard}>
            <Text style={styles.deleteTitle}>Confirmar exclusão</Text>
            <Text style={styles.deleteDescription}>Digite sua senha para confirmar. Essa ação não pode ser desfeita.</Text>
            <TextInput value={deletePassword} onChangeText={setDeletePassword} secureTextEntry placeholder="Sua senha atual" placeholderTextColor={colors.textSoft} style={styles.deleteInput} />
            <Pressable accessibilityRole="checkbox" accessibilityState={{ checked: deleteContent }} onPress={() => setDeleteContent((value) => !value)} style={styles.deleteOption}>
              <View style={[styles.deleteCheckbox, deleteContent && styles.deleteCheckboxOn]}>{deleteContent ? <Text style={styles.deleteCheck}>✓</Text> : null}</View>
              <Text style={styles.deleteOptionText}>Apagar também meus projetos, posts, respostas e imagens publicados.</Text>
            </Pressable>
            <View style={styles.deleteActions}>
              <Pressable onPress={() => { setDeleteModal(false); setDeletePassword(''); }} style={styles.deleteCancel}><Text style={styles.deleteCancelText}>Cancelar</Text></Pressable>
              <Pressable onPress={async () => {
                const result = await deleteAccount(deletePassword, deleteContent);
                if (!result.ok) return Alert.alert('Não foi possível excluir', result.message);
                setDeleteModal(false);
                setDeletePassword('');
                await resetDemo();
              }} style={styles.deleteConfirm}><Text style={styles.deleteConfirmText}>Excluir conta</Text></Pressable>
            </View>
          </View>
        </View>
      </Modal>
    </ScreenShell>
  );
}

function Stat({ icon, value, label, accent }: { icon: keyof typeof Ionicons.glyphMap; value: string; label: string; accent?: boolean }) {
  return (
    <View style={[styles.statCard, accent && { backgroundColor: colors.orangeSoft, borderColor: '#F5D8C8' }]}>
      <View style={[styles.statIcon, accent && { backgroundColor: '#FFF8F4' }]}><Ionicons name={icon} size={23} color={accent ? colors.orange : colors.primary} /></View>
      <Text style={styles.statValue}>{value}</Text>
      <Text style={[styles.statLabel, accent && { color: '#8E4827' }]}>{label}</Text>
    </View>
  );
}
function Summary({ icon, text }: { icon: keyof typeof Ionicons.glyphMap; text: string }) {
  return <View style={styles.summaryLine}><Ionicons name={icon} size={18} color={colors.primary} /><Text style={styles.summaryText}>{text}</Text></View>;
}
function Setting({ icon, label, danger, onPress }: { icon: keyof typeof Ionicons.glyphMap; label: string; danger?: boolean; onPress: () => void }) {
  return (
    <Pressable accessibilityRole="button" onPress={onPress} style={styles.settingRow}>
      <Ionicons name={icon} size={20} color={danger ? colors.danger : colors.primary} />
      <Text style={[styles.settingLabel, danger && { color: colors.danger }]}>{label}</Text>
      <Ionicons name="chevron-forward" size={18} color={colors.textSoft} />
    </Pressable>
  );
}

const styles = StyleSheet.create({
  modalBackdrop: { flex: 1, backgroundColor: 'rgba(0,0,0,0.45)', justifyContent: 'center', padding: 22 },
  deleteCard: { backgroundColor: colors.surface, borderRadius: radius.xl, padding: 20, gap: 12 },
  deleteTitle: { color: colors.text, fontSize: 20, fontWeight: '900' },
  deleteDescription: { color: colors.textMuted, fontSize: 13, lineHeight: 19 },
  deleteInput: { minHeight: 50, borderWidth: 1, borderColor: colors.border, borderRadius: radius.md, paddingHorizontal: 12, color: colors.text },
  deleteOption: { flexDirection: 'row', alignItems: 'flex-start', gap: 10 },
  deleteCheckbox: { width: 22, height: 22, borderWidth: 1.5, borderColor: colors.borderStrong, borderRadius: 6, alignItems: 'center', justifyContent: 'center' },
  deleteCheckboxOn: { backgroundColor: colors.primary, borderColor: colors.primary },
  deleteCheck: { color: colors.white, fontWeight: '900' },
  deleteOptionText: { flex: 1, color: colors.textMuted, fontSize: 12.5, lineHeight: 18 },
  deleteActions: { flexDirection: 'row', justifyContent: 'flex-end', gap: 10, marginTop: 4 },
  deleteCancel: { paddingHorizontal: 14, paddingVertical: 11 },
  deleteCancelText: { color: colors.textMuted, fontWeight: '800' },
  deleteConfirm: { backgroundColor: colors.danger, borderRadius: radius.md, paddingHorizontal: 14, paddingVertical: 11 },
  deleteConfirmText: { color: colors.white, fontWeight: '900' },
  profileCard: { ...shadow, marginTop: spacing.lg, backgroundColor: colors.surface, borderRadius: radius.xl, borderWidth: 1, borderColor: colors.border, padding: spacing.md, flexDirection: 'row', flexWrap: 'wrap', alignItems: 'center', gap: 12 },
  avatar: { width: 76, height: 76, borderRadius: 38, backgroundColor: colors.primary, alignItems: 'center', justifyContent: 'center', position: 'relative', borderWidth: 4, borderColor: colors.mint },
  avatarText: { color: colors.white, fontSize: 30, fontWeight: '900' },
  online: { position: 'absolute', width: 16, height: 16, borderRadius: 8, backgroundColor: colors.success, right: -2, bottom: 7, borderWidth: 3, borderColor: colors.white },
  identity: { flex: 1, minWidth: 150 },
  name: { color: colors.text, fontSize: 24, fontWeight: '900' },
  city: { color: colors.textMuted, marginTop: 2, fontSize: 12.5 },
  badge: { alignSelf: 'flex-start', flexDirection: 'row', alignItems: 'center', gap: 4, backgroundColor: colors.orangeSoft, paddingHorizontal: 10, paddingVertical: 6, borderRadius: radius.pill, marginTop: 7 },
  badgeText: { color: '#934721', fontWeight: '900', fontSize: 10.5 },
  edit: { flexDirection: 'row', alignItems: 'center', gap: 5, backgroundColor: colors.mint, paddingVertical: 9, paddingHorizontal: 11, borderRadius: radius.pill },
  editText: { color: colors.primary, fontWeight: '900', fontSize: 12 },
  bio: { width: '100%', color: colors.textMuted, lineHeight: 20, fontSize: 13.5 },
  miniGrid: { flexDirection: 'row', gap: 10, marginTop: 12 },
  miniCard: { flex: 1, minHeight: 92, backgroundColor: colors.surface, borderRadius: radius.lg, borderWidth: 1, borderColor: colors.border, padding: 11, flexDirection: 'row', alignItems: 'center', gap: 8 },
  miniIcon: { width: 42, height: 42, borderRadius: 15, backgroundColor: colors.mint, alignItems: 'center', justifyContent: 'center' },
  miniLabel: { color: colors.textMuted, fontSize: 10.5 },
  miniValue: { color: colors.text, fontWeight: '900', marginTop: 2, fontSize: 12.5 },
  sectionRow: { marginTop: spacing.lg, marginBottom: 12, flexDirection: 'row', justifyContent: 'space-between', alignItems: 'flex-end', gap: 10 },
  title: { color: colors.text, fontSize: 22, lineHeight: 28, fontWeight: '900' },
  muted: { color: colors.textMuted, fontSize: 11.5, lineHeight: 16 },
  year: { backgroundColor: colors.mint, borderRadius: radius.pill, paddingHorizontal: 10, paddingVertical: 6 },
  yearText: { color: colors.primaryDark, fontWeight: '900', fontSize: 11 },
  statsGrid: { flexDirection: 'row', flexWrap: 'wrap', gap: 10 },
  statCard: { width: '48%', backgroundColor: colors.mintSoft, borderRadius: radius.lg, borderWidth: 1, borderColor: '#DDEEE3', padding: 13, minHeight: 137 },
  statIcon: { width: 42, height: 42, borderRadius: 15, backgroundColor: colors.white, alignItems: 'center', justifyContent: 'center' },
  statValue: { color: colors.text, fontSize: 22, fontWeight: '900', marginTop: 9 },
  statLabel: { color: colors.primaryDark, fontWeight: '800', marginTop: 2, fontSize: 11.5 },
  summaryCard: { backgroundColor: colors.surface, borderRadius: radius.lg, padding: spacing.md, borderWidth: 1, borderColor: colors.border, gap: 9, marginTop: spacing.md },
  summaryTitle: { color: colors.text, fontWeight: '900', fontSize: 16 },
  summaryLine: { flexDirection: 'row', alignItems: 'center', gap: 8 },
  summaryText: { color: colors.textMuted, lineHeight: 19, fontSize: 12.5 },
  recentGrid: { flexDirection: 'row', gap: 10 },
  recentCard: { flex: 1, backgroundColor: colors.surface, borderRadius: radius.lg, borderWidth: 1, borderColor: colors.border, padding: 7 },
  recentTitle: { color: colors.text, fontWeight: '900', fontSize: 12.5, lineHeight: 17, marginTop: 7, paddingHorizontal: 3 },
  settings: { backgroundColor: colors.surface, borderRadius: radius.lg, borderWidth: 1, borderColor: colors.border, marginTop: spacing.lg, overflow: 'hidden' },
  settingsTitle: { color: colors.text, fontWeight: '900', fontSize: 16, padding: 14, borderBottomWidth: 1, borderBottomColor: colors.border },
  settingRow: { minHeight: 52, flexDirection: 'row', alignItems: 'center', gap: 10, paddingHorizontal: 14, borderBottomWidth: 1, borderBottomColor: colors.border },
  settingLabel: { flex: 1, color: colors.text, fontWeight: '700', fontSize: 13 },
  accountInfo: { color: colors.textSoft, fontSize: 10.5, padding: 12 },
});
