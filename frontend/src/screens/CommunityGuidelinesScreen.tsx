import React from 'react';
import { Alert, Linking, Pressable, StyleSheet, Text, View } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { ScreenShell } from '../components/ScreenShell';
import { useApp } from '../context/AppContext';
import { appConfig } from '../config';
import { colors, radius, spacing } from '../theme';

export function CommunityGuidelinesScreen({ goBack }: { goBack: () => void }) {
  const { blockedAuthors, unblockAuthor } = useApp();

  const contactSupport = async () => {
    if (!appConfig.supportEmail) {
      Alert.alert('Canal de suporte não configurado', 'Antes da publicação, configure EXPO_PUBLIC_SUPPORT_EMAIL no ambiente de produção.');
      return;
    }
    const url = `mailto:${appConfig.supportEmail}?subject=Suporte%20ARIC`;
    const supported = await Linking.canOpenURL(url);
    if (supported) await Linking.openURL(url);
  };

  return (
    <ScreenShell>
      <View style={styles.topbar}>
        <Pressable accessibilityRole="button" accessibilityLabel="Voltar" onPress={goBack} style={styles.back}>
          <Ionicons name="arrow-back" size={22} color={colors.primary} />
        </Pressable>
        <Text style={styles.topTitle}>Regras e termos</Text>
        <View style={{ width: 42 }} />
      </View>

      <View style={styles.hero}>
        <View style={styles.heroIcon}><Ionicons name="shield-checkmark" size={30} color={colors.primary} /></View>
        <Text style={styles.title}>Regras da comunidade e termos de uso</Text>
        <Text style={styles.subtitle}>O ARIC existe para trocar ideias sustentáveis. Conteúdo abusivo, perigoso ou que exponha dados pessoais não pertence aqui.</Text>
      </View>

      <Rule icon="document-text-outline" title="Termos para publicar" text="Ao enviar conteúdo, você declara que tem direito de compartilhá-lo, que as informações são de boa-fé e que o material pode ser analisado, rejeitado, removido ou arquivado pela moderação quando violar estas regras ou a segurança da comunidade." />
      <Rule icon="heart-outline" title="Respeite as pessoas" text="Sem ataques, assédio, ameaças, discriminação ou discurso de ódio." />
      <Rule icon="leaf-outline" title="Mantenha o foco" text="Publique dúvidas, projetos e orientações relacionados a reutilização, reciclagem e sustentabilidade." />
      <Rule icon="person-outline" title="Proteja dados pessoais" text="Não compartilhe endereço completo, telefone, documentos, senhas ou dados sensíveis." />
      <Rule icon="warning-outline" title="Evite instruções perigosas" text="Não incentive práticas que exponham pessoas a químicos, eletricidade, fogo ou ferramentas sem orientação de segurança." />
      <Rule icon="shield-checkmark-outline" title="Publicação e moderação" text="As publicações aparecem imediatamente. A equipe de moderação pode analisar denúncias e remover conteúdos que violem as regras." />
      <Rule icon="flag-outline" title="Denuncie e bloqueie" text="Mesmo após a publicação, use o menu de cada publicação para denunciar conteúdo problemático e ocultar usuários abusivos." />

      <View style={styles.card}>
        <Text style={styles.cardTitle}>Usuários bloqueados</Text>
        {blockedAuthors.length === 0 ? <Text style={styles.muted}>Você não bloqueou ninguém.</Text> : blockedAuthors.map((author) => (
          <View key={author} style={styles.blockedRow}>
            <View style={styles.avatar}><Text style={styles.avatarText}>{author.charAt(0)}</Text></View>
            <Text style={styles.blockedName}>{author}</Text>
            <Pressable accessibilityRole="button" onPress={() => unblockAuthor(author)} style={styles.unblock}><Text style={styles.unblockText}>Desbloquear</Text></Pressable>
          </View>
        ))}
      </View>

      <Pressable accessibilityRole="button" onPress={contactSupport} style={styles.support}>
        <Ionicons name="mail-outline" size={20} color={colors.primary} />
        <View style={{ flex: 1 }}><Text style={styles.supportTitle}>Falar com o suporte</Text><Text style={styles.muted}>{appConfig.supportEmail || 'E-mail será configurado no build de produção'}</Text></View>
        <Ionicons name="chevron-forward" size={18} color={colors.textSoft} />
      </Pressable>

      <Text style={styles.footnote}>No backend Go, as publicações do fórum são imediatas. Denúncias são registradas no servidor e analisadas pela equipe de moderação. O bloqueio pessoal oculta autores apenas neste aparelho.</Text>
    </ScreenShell>
  );
}

function Rule({ icon, title, text }: { icon: keyof typeof Ionicons.glyphMap; title: string; text: string }) {
  return (
    <View style={styles.rule}>
      <View style={styles.ruleIcon}><Ionicons name={icon} size={20} color={colors.primary} /></View>
      <View style={{ flex: 1 }}><Text style={styles.ruleTitle}>{title}</Text><Text style={styles.ruleText}>{text}</Text></View>
    </View>
  );
}

const styles = StyleSheet.create({
  topbar: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' },
  back: { width: 42, height: 42, borderRadius: 21, backgroundColor: colors.surface, borderWidth: 1, borderColor: colors.border, alignItems: 'center', justifyContent: 'center' },
  topTitle: { color: colors.text, fontWeight: '900', fontSize: 17 },
  hero: { alignItems: 'center', marginTop: spacing.lg, marginBottom: spacing.md, backgroundColor: colors.mintSoft, borderRadius: radius.xl, padding: spacing.lg, borderWidth: 1, borderColor: '#DCEDE2' },
  heroIcon: { width: 62, height: 62, borderRadius: 22, backgroundColor: colors.white, alignItems: 'center', justifyContent: 'center' },
  title: { color: colors.text, fontSize: 25, lineHeight: 31, fontWeight: '900', textAlign: 'center', marginTop: 12 },
  subtitle: { color: colors.textMuted, lineHeight: 20, fontSize: 13, textAlign: 'center', marginTop: 6 },
  rule: { flexDirection: 'row', gap: 12, backgroundColor: colors.surface, borderRadius: radius.lg, borderWidth: 1, borderColor: colors.border, padding: 14, marginBottom: 9 },
  ruleIcon: { width: 42, height: 42, borderRadius: 15, backgroundColor: colors.mint, alignItems: 'center', justifyContent: 'center' },
  ruleTitle: { color: colors.text, fontWeight: '900', fontSize: 14 },
  ruleText: { color: colors.textMuted, lineHeight: 18, fontSize: 12, marginTop: 3 },
  card: { backgroundColor: colors.surface, borderRadius: radius.lg, borderWidth: 1, borderColor: colors.border, padding: 14, marginTop: spacing.sm },
  cardTitle: { color: colors.text, fontWeight: '900', fontSize: 16, marginBottom: 8 },
  muted: { color: colors.textMuted, fontSize: 11.5, lineHeight: 16 },
  blockedRow: { minHeight: 52, flexDirection: 'row', alignItems: 'center', gap: 9, borderTopWidth: 1, borderTopColor: colors.border },
  avatar: { width: 34, height: 34, borderRadius: 17, backgroundColor: colors.mint, alignItems: 'center', justifyContent: 'center' },
  avatarText: { color: colors.primaryDark, fontWeight: '900' },
  blockedName: { flex: 1, color: colors.text, fontWeight: '800' },
  unblock: { paddingHorizontal: 10, paddingVertical: 7, borderRadius: radius.pill, backgroundColor: colors.mint },
  unblockText: { color: colors.primary, fontWeight: '900', fontSize: 10.5 },
  support: { flexDirection: 'row', alignItems: 'center', gap: 10, backgroundColor: colors.surface, borderRadius: radius.lg, borderWidth: 1, borderColor: colors.border, padding: 14, marginTop: 12 },
  supportTitle: { color: colors.text, fontWeight: '900', fontSize: 13.5 },
  footnote: { color: colors.textSoft, fontSize: 10.5, lineHeight: 15, marginTop: 12, marginBottom: 8 },
});
