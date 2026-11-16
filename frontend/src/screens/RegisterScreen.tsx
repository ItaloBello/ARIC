import React, { useState } from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import { AuthShell } from '../components/AuthShell';
import { FormField } from '../components/FormField';
import { PrimaryButton } from '../components/PrimaryButton';
import { useAuth } from '../context/AuthContext';
import type { AuthRoute } from '../navigation';
import { colors, radius } from '../theme';

export function RegisterScreen({ navigate }: { navigate: (route: AuthRoute) => void }) {
  const { register } = useAuth();
  const [name, setName] = useState('');
  const [cep, setCep] = useState('');
  const [numeroEndereco, setNumeroEndereco] = useState('');
  const [bairro, setBairro] = useState('');
  const [logradouro, setLogradouro] = useState('');
  const [complementoEndereco, setComplementoEndereco] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [accepted, setAccepted] = useState(false);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const submit = async () => {
    setError('');
    if (password !== confirm) return setError('As senhas não coincidem.');
    if (!accepted) return setError('Leia e aceite a Política de Privacidade e as Regras da Comunidade para continuar.');
    setLoading(true);
    const result = await register(name, email, password, cep, numeroEndereco, bairro, logradouro, complementoEndereco);
    setLoading(false);
    if (!result.ok) return setError(result.message);
  };

  return (
    <AuthShell title="Crie sua conta" subtitle="Informe seu endereço para localizar sua comunidade e calcular o impacto do bairro.">
      <FormField label="Nome" icon="person-outline" value={name} onChangeText={setName} autoCapitalize="words" placeholder="Seu nome" />
      <FormField label="CEP" icon="location-outline" value={cep} onChangeText={setCep} keyboardType="number-pad" placeholder="00000-000" />
      <FormField label="Número do endereço" icon="home-outline" value={numeroEndereco} onChangeText={setNumeroEndereco} keyboardType="default" placeholder="Número da residência" />
      <FormField label="Bairro (se o CEP não identificar)" icon="map-outline" value={bairro} onChangeText={setBairro} autoCapitalize="words" placeholder="Opcional" />
      <FormField label="Logradouro (se o CEP não identificar)" icon="navigate-outline" value={logradouro} onChangeText={setLogradouro} autoCapitalize="words" placeholder="Opcional" />
      <FormField label="Complemento (opcional)" icon="business-outline" value={complementoEndereco} onChangeText={setComplementoEndereco} placeholder="Apartamento, bloco..." />
      <FormField label="E-mail" icon="mail-outline" value={email} onChangeText={setEmail} keyboardType="email-address" autoComplete="email" placeholder="voce@email.com" />
      <FormField label="Senha" icon="lock-closed-outline" value={password} onChangeText={setPassword} password autoComplete="new-password" placeholder="8+ caracteres, com maiúscula, minúscula, número e símbolo" />
      <FormField label="Confirmar senha" icon="lock-closed-outline" value={confirm} onChangeText={setConfirm} password placeholder="Repita a senha" />

      <Pressable accessibilityRole="checkbox" accessibilityState={{ checked: accepted }} onPress={() => setAccepted((value) => !value)} style={styles.consent}>
        <View style={[styles.checkbox, accepted && styles.checkboxOn]}>{accepted ? <Text style={styles.check}>✓</Text> : null}</View>
        <Text style={styles.consentText}>Li e concordo com a <Text style={styles.link} onPress={() => navigate('privacy')}>Política de Privacidade</Text> e com as <Text style={styles.link} onPress={() => navigate('communityGuidelines')}>Regras da Comunidade</Text>.</Text>
      </Pressable>

      {error ? <Text style={styles.error}>{error}</Text> : null}
      <PrimaryButton label="Criar conta" icon="person-add-outline" onPress={submit} loading={loading} />
      <Pressable accessibilityRole="button" onPress={() => navigate('login')} style={styles.back}><Text style={styles.backText}>Já tenho conta. <Text style={styles.link}>Entrar</Text></Text></Pressable>
    </AuthShell>
  );
}

const styles = StyleSheet.create({
  consent: { flexDirection: 'row', gap: 10, alignItems: 'flex-start' },
  checkbox: { width: 22, height: 22, borderRadius: 7, borderWidth: 1.5, borderColor: colors.borderStrong, backgroundColor: colors.surface, alignItems: 'center', justifyContent: 'center', marginTop: 1 },
  checkboxOn: { backgroundColor: colors.primary, borderColor: colors.primary },
  check: { color: colors.white, fontWeight: '900' },
  consentText: { flex: 1, color: colors.textMuted, lineHeight: 20, fontSize: 13 },
  link: { color: colors.primary, fontWeight: '900' },
  error: { color: colors.danger, backgroundColor: colors.dangerSoft, padding: 10, borderRadius: radius.sm, fontSize: 12.5 },
  back: { alignItems: 'center', paddingVertical: 5 },
  backText: { color: colors.textMuted, fontSize: 13.5 },
});
