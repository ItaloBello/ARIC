import React from 'react';
import { Linking, Pressable, StyleSheet, Text, View } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { ScreenShell } from '../components/ScreenShell';
import { PrimaryButton } from '../components/PrimaryButton';
import { appConfig } from '../config';
import { colors, radius, spacing } from '../theme';

export function PrivacyScreen({ goBack }: { goBack: () => void }) {
  const openPolicy = () => {
    if (appConfig.privacyPolicyUrl) Linking.openURL(appConfig.privacyPolicyUrl).catch(() => undefined);
  };
  const openDeletion = () => {
    if (appConfig.accountDeletionUrl) Linking.openURL(appConfig.accountDeletionUrl).catch(() => undefined);
  };

  return (
    <ScreenShell>
      <View style={styles.topbar}>
        <Pressable accessibilityRole="button" accessibilityLabel="Voltar" onPress={goBack} style={styles.back}><Ionicons name="arrow-back" size={22} color={colors.primary} /></Pressable>
        <Text style={styles.topTitle}>Privacidade</Text>
        <View style={{ width: 42 }} />
      </View>

      <View style={styles.hero}>
        <View style={styles.heroIcon}><Ionicons name="shield-checkmark" size={30} color={colors.primary} /></View>
        <Text style={styles.title}>Seus dados, com transparência</Text>
        <Text style={styles.subtitle}>Resumo da arquitetura de dados da versão 3.0. Antes de publicar nas lojas, disponibilize a política completa em uma URL pública.</Text>
      </View>

      <Policy title="Dados de conta" text="A API ARIC (Go + PostgreSQL) utiliza nome, e-mail, identificador da conta e dados de perfil para autenticação, suporte e identificação das publicações. No modo visitante, a sessão não é criada no servidor." />
      <Policy title="Comunidade e moderação" text="As publicações do fórum ficam visíveis imediatamente. Denúncias podem registrar o conteúdo, o motivo e a descrição para análise por administradores responsáveis pela moderação." />
      <Policy title="Moderação" text="O backend Go atual não executa revisão prévia por IA. A equipe responsável pode analisar denúncias e decidir pela remoção ou manutenção do conteúdo." />
      <Policy title="Denúncias" text="Denúncias registram a publicação, o motivo, o identificador do denunciante e o horário para permitir análise da equipe. Esses dados não são públicos." />
      <Policy title="Dados de uso" text="Com a API ativa, EcoPontos, perfil, projetos, curtidas, publicações e dados de impacto do bairro são consultados ou registrados no backend. Bloqueios pessoais e algumas preferências permanecem localmente no aparelho." />
      <Policy title="Exclusão de conta" text="Ao excluir a conta, a API Go solicita a senha e permite escolher se os conteúdos publicados serão apagados ou mantidos anonimizados. Dados pessoais e sessões são removidos ou invalidados conforme as regras do backend. Registros estritamente necessários por obrigação legal ou segurança devem ser tratados conforme a política pública final." />

      {appConfig.privacyPolicyUrl ? <PrimaryButton label="Abrir política completa" icon="open-outline" variant="secondary" onPress={openPolicy} /> : <Text style={styles.pending}>Defina EXPO_PUBLIC_PRIVACY_POLICY_URL antes de enviar para as lojas.</Text>}
      {appConfig.accountDeletionUrl ? <PrimaryButton label="Abrir página de exclusão" icon="trash-outline" variant="secondary" onPress={openDeletion} /> : <Text style={styles.pending}>Defina EXPO_PUBLIC_ACCOUNT_DELETION_URL para o caminho externo de exclusão.</Text>}
    </ScreenShell>
  );
}

function Policy({ title, text }: { title: string; text: string }) {
  return (
    <View style={styles.card}>
      <Text style={styles.cardTitle}>{title}</Text>
      <Text style={styles.cardText}>{text}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  topbar: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' },
  back: { width: 42, height: 42, borderRadius: 21, backgroundColor: colors.surface, borderWidth: 1, borderColor: colors.border, alignItems: 'center', justifyContent: 'center' },
  topTitle: { color: colors.text, fontWeight: '900', fontSize: 18 },
  hero: { marginTop: spacing.lg, marginBottom: spacing.md },
  heroIcon: { width: 58, height: 58, borderRadius: 19, backgroundColor: colors.mint, alignItems: 'center', justifyContent: 'center' },
  title: { color: colors.text, fontSize: 28, lineHeight: 34, fontWeight: '900', marginTop: 13 },
  subtitle: { color: colors.textMuted, lineHeight: 21, marginTop: 7 },
  card: { backgroundColor: colors.surface, borderRadius: radius.lg, borderWidth: 1, borderColor: colors.border, padding: spacing.md, marginBottom: 12 },
  cardTitle: { color: colors.text, fontWeight: '900', fontSize: 16 },
  cardText: { color: colors.textMuted, lineHeight: 20, marginTop: 6, fontSize: 13.5 },
  pending: { color: colors.orange, backgroundColor: colors.orangeSoft, borderRadius: radius.md, padding: 12, lineHeight: 18, marginBottom: 10, fontSize: 12.5, fontWeight: '700' },
});
