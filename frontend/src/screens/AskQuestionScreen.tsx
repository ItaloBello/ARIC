import React, { useState } from 'react';
import { Alert, Pressable, StyleSheet, Text, TextInput, View } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { Chip } from '../components/Chip';
import { PrimaryButton } from '../components/PrimaryButton';
import { ScreenShell } from '../components/ScreenShell';
import { useApp } from '../context/AppContext';
import { useCommunity } from '../context/CommunityContext';
import type { MaterialCategory } from '../types';
import { colors, radius, spacing } from '../theme';

const categories: MaterialCategory[] = ['Plástico', 'Papelão', 'Vidro', 'Metal', 'Madeira', 'Horta'];

export function AskQuestionScreen({ goBack, openGuidelines }: { goBack: () => void; openGuidelines: () => void }) {
  const { user } = useApp();
  const { submitPost } = useCommunity();
  const [title, setTitle] = useState('');
  const [body, setBody] = useState('');
  const [category, setCategory] = useState<MaterialCategory>('Plástico');
  const [sending, setSending] = useState(false);
  const [rulesAccepted, setRulesAccepted] = useState(false);

  const submit = async () => {
    if (title.trim().length < 8 || body.trim().length < 15) return Alert.alert('Conte um pouco mais', 'Escreva um título claro e explique melhor sua dúvida.');
    if (!rulesAccepted) return Alert.alert('Aceite as regras', 'Para enviar conteúdo à comunidade, leia e aceite as Regras da Comunidade e os Termos de uso.');
    setSending(true);
    const result = await submitPost({
      author: user.name,
      neighborhood: user.city,
      category,
      title: title.trim(),
      body: body.trim(),
      communityRulesAccepted: true,
      communityRulesVersion: '2026-11-15',
    });
    setSending(false);
    if (!result.ok) return Alert.alert('Não foi possível enviar', result.message);
    Alert.alert('Publicação enviada', result.message ?? 'Sua publicação já está disponível no fórum. Ela poderá ser denunciada e removida se violar as regras.', [
      { text: 'OK', onPress: goBack },
    ]);
  };

  return (
    <ScreenShell>
      <View style={styles.topbar}>
        <Pressable accessibilityRole="button" accessibilityLabel="Voltar" onPress={goBack} style={styles.back}><Ionicons name="arrow-back" size={22} color={colors.primary} /></Pressable>
        <Text style={styles.topTitle}>Nova publicação</Text>
        <View style={{ width: 42 }} />
      </View>

      <View style={styles.reviewBanner}>
        <View style={styles.reviewIcon}><Ionicons name="shield-checkmark" size={25} color={colors.primary} /></View>
        <View style={{ flex: 1 }}>
          <Text style={styles.reviewTitle}>Publicação na comunidade</Text>
          <Text style={styles.reviewText}>Sua publicação aparece imediatamente. Denúncias podem ser analisadas pela equipe de moderação.</Text>
        </View>
      </View>

      <Text style={styles.title}>Compartilhe sua dúvida</Text>
      <Text style={styles.subtitle}>Dê contexto suficiente para que a comunidade consiga ajudar, sem expor dados pessoais.</Text>

      <Text style={styles.label}>Material</Text>
      <View style={styles.chips}>{categories.map((item) => <Chip key={item} label={item} selected={category === item} onPress={() => setCategory(item)} />)}</View>

      <Text style={styles.label}>Título</Text>
      <TextInput value={title} onChangeText={setTitle} style={styles.input} placeholder="Ex.: O que posso fazer com caixas de ovos?" placeholderTextColor={colors.textSoft} maxLength={90} />
      <Text style={styles.counter}>{title.length}/90</Text>

      <Text style={styles.label}>Descrição</Text>
      <TextInput value={body} onChangeText={setBody} style={[styles.input, styles.textarea]} placeholder="Explique o que você tem, quantidade e o que gostaria de fazer..." placeholderTextColor={colors.textSoft} multiline maxLength={600} textAlignVertical="top" />
      <Text style={styles.counter}>{body.length}/600</Text>

      <View style={styles.tip}><Ionicons name="bulb-outline" size={19} color={colors.orange} /><Text style={styles.tipText}>Não publique telefone, endereço completo, documentos, conteúdo ofensivo, sexual, ameaças ou material sem relação com sustentabilidade.</Text></View>

      <Pressable accessibilityRole="checkbox" accessibilityState={{ checked: rulesAccepted }} onPress={() => setRulesAccepted((value) => !value)} style={styles.consent}>
        <View style={[styles.checkbox, rulesAccepted && styles.checkboxOn]}>{rulesAccepted ? <Ionicons name="checkmark" size={15} color={colors.white} /> : null}</View>
        <View style={{ flex: 1 }}>
          <Text style={styles.consentText}>Li e concordo com as Regras da Comunidade e os Termos de uso para envio de conteúdo.</Text>
          <Pressable accessibilityRole="button" onPress={openGuidelines} hitSlop={8}><Text style={styles.rulesLink}>Ler regras e termos</Text></Pressable>
        </View>
      </Pressable>

      <PrimaryButton label={sending ? 'Publicando...' : 'Publicar no fórum'} icon="paper-plane-outline" onPress={sending ? () => undefined : submit} />
    </ScreenShell>
  );
}

const styles = StyleSheet.create({
  topbar: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' },
  back: { width: 42, height: 42, borderRadius: 21, backgroundColor: colors.surface, borderWidth: 1, borderColor: colors.border, alignItems: 'center', justifyContent: 'center' },
  topTitle: { color: colors.text, fontWeight: '900', fontSize: 18 },
  reviewBanner: { marginTop: spacing.lg, flexDirection: 'row', gap: 12, padding: 14, borderRadius: radius.lg, backgroundColor: colors.mintSoft, borderWidth: 1, borderColor: '#DCEDE2' },
  reviewIcon: { width: 48, height: 48, borderRadius: 16, backgroundColor: colors.mint, alignItems: 'center', justifyContent: 'center' },
  reviewTitle: { color: colors.primaryDeep, fontWeight: '900', fontSize: 15 },
  reviewText: { color: colors.textMuted, fontSize: 12, lineHeight: 17, marginTop: 3 },
  verifyBanner: { marginTop: 10, flexDirection: 'row', alignItems: 'center', gap: 10, padding: 12, borderRadius: radius.md, backgroundColor: colors.orangeSoft, borderWidth: 1, borderColor: '#F4D3BF' },
  verifyTitle: { color: colors.orange, fontWeight: '900', fontSize: 12.5 },
  verifyText: { color: colors.textMuted, fontSize: 10.5, lineHeight: 15, marginTop: 2 },
  title: { color: colors.text, fontSize: 28, lineHeight: 34, fontWeight: '900', marginTop: spacing.lg },
  subtitle: { color: colors.textMuted, lineHeight: 20, marginTop: 5 },
  label: { color: colors.text, fontWeight: '900', marginTop: spacing.lg, marginBottom: 8 },
  chips: { flexDirection: 'row', gap: 7, flexWrap: 'wrap' },
  input: { minHeight: 54, borderRadius: radius.md, borderWidth: 1, borderColor: colors.border, backgroundColor: colors.surface, paddingHorizontal: 14, color: colors.text, fontSize: 14 },
  textarea: { minHeight: 155, paddingTop: 14 },
  counter: { color: colors.textSoft, textAlign: 'right', fontSize: 10.5, marginTop: 4 },
  tip: { flexDirection: 'row', alignItems: 'flex-start', gap: 8, backgroundColor: colors.orangeSoft, borderRadius: radius.md, padding: 12, marginVertical: spacing.md },
  tipText: { flex: 1, color: colors.textMuted, lineHeight: 18, fontSize: 12 },
  consent: { flexDirection: 'row', gap: 10, alignItems: 'flex-start', marginBottom: spacing.md },
  checkbox: { width: 22, height: 22, borderRadius: 7, borderWidth: 1.5, borderColor: colors.borderStrong, backgroundColor: colors.surface, alignItems: 'center', justifyContent: 'center', marginTop: 1 },
  checkboxOn: { backgroundColor: colors.primary, borderColor: colors.primary },
  consentText: { flex: 1, color: colors.textMuted, lineHeight: 18, fontSize: 11.5 },
  rulesLink: { color: colors.primary, fontWeight: '900', fontSize: 11.5, marginTop: 4, textDecorationLine: 'underline' },
});
