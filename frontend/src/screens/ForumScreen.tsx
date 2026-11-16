import React, { useMemo, useState } from 'react';
import { Alert, Pressable, StyleSheet, Text, TextInput, View } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { AppImage } from '../components/AppImage';
import { BrandHeader } from '../components/BrandHeader';
import { Chip } from '../components/Chip';
import { ScreenShell } from '../components/ScreenShell';
import { useApp } from '../context/AppContext';
import { useAuth } from '../context/AuthContext';
import { useCommunity } from '../context/CommunityContext';
import type { ForumPost } from '../types';
import type { Route } from '../navigation';
import { colors, radius, shadow, spacing, typography } from '../theme';

export function ForumScreen({ navigate }: { navigate: (route: Route) => void }) {
  const { likedPostIds, reportedPostIds, blockedAuthors, toggleLikePost, reportPost: rememberReport, blockAuthor } = useApp();
  const { posts, reportPost, toggleLike, mode } = useCommunity();
  const { isGuest, signOut } = useAuth();
  const [query, setQuery] = useState('');
  const [filter, setFilter] = useState<'Mais votadas' | 'Dúvidas abertas' | 'Resolvidas'>('Mais votadas');

  const list = useMemo(() => {
    const q = query.trim().toLowerCase();
    let filtered = posts.filter(
      (post) => !blockedAuthors.includes(post.author) && (!q || `${post.title} ${post.body} ${post.category}`.toLowerCase().includes(q)),
    );
    if (filter === 'Dúvidas abertas') filtered = filtered.filter((post) => post.status === 'Em aberto');
    if (filter === 'Resolvidas') filtered = filtered.filter((post) => post.status === 'Resolvido');
    if (filter === 'Mais votadas') filtered = [...filtered].sort((a, b) => b.likes - a.likes);
    return filtered;
  }, [blockedAuthors, filter, posts, query]);

  const ask = () => {
    if (!isGuest) return navigate({ name: 'askQuestion' });
    Alert.alert('Entre para participar', 'O fórum pode ser lido como visitante. Para enviar uma publicação, entre ou crie uma conta.', [
      { text: 'Agora não', style: 'cancel' },
      { text: 'Entrar', onPress: () => signOut() },
    ]);
  };

  const sendReport = async (post: ForumPost, reason: string) => {
    const result = await reportPost(post, reason);
    if (!result.ok) return Alert.alert('Não foi possível denunciar', result.message);
    rememberReport(post.id);
    Alert.alert('Denúncia enviada', result.message ?? 'A equipe de moderação analisará a publicação.');
  };


  const likePost = async (post: ForumPost) => {
    if (isGuest) {
      Alert.alert('Entre para curtir', 'As curtidas online precisam ficar associadas a uma conta.');
      return;
    }
    const currentlyLiked = post.likedByMe ?? likedPostIds.includes(post.id);
    const result = await toggleLike(post.id, currentlyLiked);
    if (!result.ok) return Alert.alert('Não foi possível curtir', result.message);
    toggleLikePost(post.id);
  };

  const moderationMenu = (post: ForumPost) => {
    const alreadyReported = reportedPostIds.includes(post.id);
    Alert.alert('Segurança da comunidade', 'Escolha uma ação para esta publicação.', [
      { text: 'Cancelar', style: 'cancel' },
      {
        text: alreadyReported ? 'Já denunciado' : 'Denunciar: conteúdo impróprio',
        onPress: () => { if (!alreadyReported) sendReport(post, 'Conteúdo impróprio ou fora do propósito do ARIC'); },
      },
      {
        text: 'Denunciar: spam ou fraude',
        onPress: () => { if (!alreadyReported) sendReport(post, 'Spam, fraude ou propaganda indevida'); },
      },
      {
        text: `Bloquear ${post.author}`,
        style: 'destructive',
        onPress: () => {
          blockAuthor(post.author);
          Alert.alert('Usuário bloqueado', `As publicações de ${post.author} foram ocultadas neste dispositivo.`);
        },
      },
    ]);
  };

  return (
    <ScreenShell>
      <BrandHeader onProfile={() => navigate({ name: 'profile' })} />
      <View style={styles.hero}>
        <View style={styles.heroIcon}><Ionicons name="people" size={29} color={colors.primary} /></View>
        <View style={styles.heroCopy}>
          <Text style={styles.title}>Chat da Comunidade</Text>
          <Text style={styles.subtitle}>Espaço público de perguntas e dicas — não é conversa privada em tempo real. Publicações aparecem imediatamente. Fonte: {mode === 'api' ? 'comunidade online' : 'modo demonstração'}.</Text>
        </View>
      </View>

      <View style={styles.safetyRow}>
        <Ionicons name="shield-checkmark-outline" size={16} color={colors.primary} />
        <Text style={styles.safetyText}>Publicação imediata + denúncias + moderação.</Text>
        <Pressable accessibilityRole="button" onPress={() => navigate({ name: 'communityGuidelines' })}>
          <Text style={styles.safetyLink}>Ver regras</Text>
        </Pressable>
      </View>

      <View style={styles.search}>
        <Ionicons name="search" size={21} color={colors.textMuted} />
        <TextInput value={query} onChangeText={setQuery} placeholder="Buscar no fórum..." placeholderTextColor={colors.textSoft} style={styles.input} />
      </View>
      <Pressable accessibilityRole="button" onPress={ask} style={styles.askButton}><Ionicons name="add" size={22} color={colors.white} /><Text style={styles.askText}>Publicar uma pergunta</Text></Pressable>

      <View style={styles.filters}>{(['Mais votadas', 'Dúvidas abertas', 'Resolvidas'] as const).map((item) => <Chip key={item} label={item} selected={filter === item} onPress={() => setFilter(item)} />)}</View>

      {blockedAuthors.length > 0 ? (
        <View style={styles.blockedNote}>
          <Ionicons name="eye-off-outline" size={16} color={colors.textMuted} />
          <Text style={styles.blockedText}>{blockedAuthors.length} usuário(s) bloqueado(s) neste aparelho.</Text>
        </View>
      ) : null}

      {list.map((post) => {
        const liked = post.likedByMe ?? likedPostIds.includes(post.id);
        const reported = reportedPostIds.includes(post.id);
        const visibleLikes = mode === 'api' ? post.likes : post.likes + (liked ? 1 : 0);
        return (
          <View key={post.id} style={styles.card}>
            <View style={styles.headerRow}>
              <View style={styles.avatar}><Text style={styles.avatarText}>{post.author.charAt(0)}</Text></View>
              <View style={styles.authorCopy}><Text style={styles.author}>{post.author}</Text><Text style={styles.meta}>{post.neighborhood} • {post.time}</Text></View>
              <View style={[styles.status, post.status === 'Resolvido' ? styles.statusSolved : styles.statusOpen]}><Text style={[styles.statusText, post.status === 'Resolvido' ? styles.statusTextSolved : styles.statusTextOpen]}>{post.status}</Text></View>
              <Pressable accessibilityRole="button" accessibilityLabel="Opções de segurança" hitSlop={8} onPress={() => moderationMenu(post)} style={styles.more}>
                <Ionicons name="ellipsis-horizontal" size={20} color={colors.textMuted} />
              </Pressable>
            </View>
            {reported ? <View style={styles.reported}><Ionicons name="flag" size={13} color={colors.orange} /><Text style={styles.reportedText}>Denúncia registrada</Text></View> : null}
            <Text style={styles.postTitle}>{post.title}</Text>
            <Text style={styles.body} numberOfLines={4}>{post.body}</Text>
            {post.image ? <View style={styles.imageWrap}><AppImage source={post.image} height={170} borderRadius={radius.md} /></View> : null}
            {post.imageUrl ? <View style={styles.imageWrap}><AppImage source={{ uri: post.imageUrl }} height={170} borderRadius={radius.md} /></View> : null}
            <View style={styles.bottomRow}>
              <View style={styles.category}><Ionicons name="leaf-outline" size={14} color={colors.primary} /><Text style={styles.categoryText}>{post.category}</Text></View>
              <Pressable accessibilityRole="button" accessibilityLabel={liked ? 'Descurtir' : 'Curtir'} onPress={() => likePost(post)} style={styles.action}><Ionicons name={liked ? 'heart' : 'heart-outline'} size={20} color={liked ? colors.orange : colors.textMuted} /><Text style={styles.actionText}>{visibleLikes}</Text></Pressable>
              <View style={styles.action}><Ionicons name="chatbubble-outline" size={18} color={colors.textMuted} /><Text style={styles.actionText}>{post.answers}</Text></View>
            </View>
          </View>
        );
      })}

      {list.length === 0 ? (
        <View style={styles.empty}>
          <Ionicons name="search-outline" size={30} color={colors.primary} />
          <Text style={styles.emptyTitle}>Nenhuma publicação encontrada</Text>
          <Text style={styles.emptyText}>Tente outro termo ou mude o filtro.</Text>
        </View>
      ) : null}
    </ScreenShell>
  );
}

const styles = StyleSheet.create({
  hero: { flexDirection: 'row', alignItems: 'center', gap: 12, marginTop: spacing.lg },
  heroIcon: { width: 58, height: 58, borderRadius: 20, backgroundColor: colors.mint, alignItems: 'center', justifyContent: 'center' },
  heroCopy: { flex: 1 },
  title: { ...typography.h2, fontSize: 26, lineHeight: 31 },
  subtitle: { ...typography.muted, marginTop: 4 },
  safetyRow: { marginTop: 12, minHeight: 42, borderRadius: radius.md, backgroundColor: colors.mintSoft, borderWidth: 1, borderColor: '#DCEDE2', flexDirection: 'row', alignItems: 'center', gap: 7, paddingHorizontal: 12 },
  safetyText: { flex: 1, color: colors.textMuted, fontSize: 11.5 },
  safetyLink: { color: colors.primary, fontSize: 11.5, fontWeight: '900' },
  search: { marginTop: spacing.md, minHeight: 52, backgroundColor: colors.surface, borderRadius: radius.md, borderWidth: 1, borderColor: colors.border, flexDirection: 'row', alignItems: 'center', paddingHorizontal: 14 },
  input: { flex: 1, minHeight: 50, marginLeft: 8, color: colors.text },
  askButton: { minHeight: 52, backgroundColor: colors.primary, borderRadius: radius.md, flexDirection: 'row', alignItems: 'center', justifyContent: 'center', gap: 7, marginTop: 10 },
  askText: { color: colors.white, fontWeight: '900', fontSize: 14.5 },
  filters: { flexDirection: 'row', gap: 7, flexWrap: 'wrap', marginVertical: spacing.md },
  blockedNote: { flexDirection: 'row', alignItems: 'center', gap: 7, backgroundColor: colors.surfaceMuted, borderRadius: radius.md, padding: 10, marginBottom: 10 },
  blockedText: { flex: 1, color: colors.textMuted, fontSize: 11.5, lineHeight: 16 },
  card: { ...shadow, backgroundColor: colors.surface, borderWidth: 1, borderColor: colors.border, borderRadius: radius.lg, padding: spacing.md, marginBottom: 13 },
  headerRow: { flexDirection: 'row', alignItems: 'center', gap: 9 },
  avatar: { width: 42, height: 42, borderRadius: 21, backgroundColor: colors.mintStrong, alignItems: 'center', justifyContent: 'center' },
  avatarText: { color: colors.primaryDark, fontWeight: '900', fontSize: 16 },
  authorCopy: { flex: 1 },
  author: { color: colors.text, fontWeight: '900', fontSize: 14 },
  meta: { color: colors.textMuted, fontSize: 10.5, marginTop: 1 },
  status: { borderRadius: radius.pill, paddingHorizontal: 9, paddingVertical: 6 },
  statusSolved: { backgroundColor: colors.mint },
  statusOpen: { backgroundColor: colors.orangeSoft },
  statusText: { fontSize: 10.5, fontWeight: '900' },
  statusTextSolved: { color: colors.success },
  statusTextOpen: { color: colors.orange },
  more: { width: 30, height: 30, borderRadius: 15, alignItems: 'center', justifyContent: 'center' },
  reported: { marginTop: 10, alignSelf: 'flex-start', flexDirection: 'row', alignItems: 'center', gap: 5, backgroundColor: colors.orangeSoft, borderRadius: radius.pill, paddingHorizontal: 8, paddingVertical: 5 },
  reportedText: { color: colors.orange, fontSize: 10, fontWeight: '900' },
  postTitle: { color: colors.text, fontSize: 19, lineHeight: 24, fontWeight: '900', marginTop: 13 },
  body: { color: colors.textMuted, lineHeight: 20, fontSize: 13.5, marginTop: 5 },
  imageWrap: { marginTop: 12, overflow: 'hidden', borderRadius: radius.md },
  bottomRow: { flexDirection: 'row', alignItems: 'center', gap: 14, marginTop: 12 },
  category: { flexDirection: 'row', alignItems: 'center', gap: 4, backgroundColor: colors.mint, borderRadius: radius.pill, paddingHorizontal: 9, paddingVertical: 6 },
  categoryText: { color: colors.primaryDark, fontWeight: '800', fontSize: 10.5 },
  action: { flexDirection: 'row', alignItems: 'center', gap: 4 },
  actionText: { color: colors.textMuted, fontSize: 11.5, fontWeight: '800' },
  empty: { alignItems: 'center', paddingVertical: 32, paddingHorizontal: 24 },
  emptyTitle: { color: colors.text, fontWeight: '900', fontSize: 16, marginTop: 8 },
  emptyText: { color: colors.textMuted, textAlign: 'center', lineHeight: 18, marginTop: 4, fontSize: 12 },
});
