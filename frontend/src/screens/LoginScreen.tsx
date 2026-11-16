import React, { useState } from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { AuthShell } from '../components/AuthShell';
import { FormField } from '../components/FormField';
import { PrimaryButton } from '../components/PrimaryButton';
import { useAuth } from '../context/AuthContext';
import type { AuthRoute } from '../navigation';
import { colors, radius } from '../theme';

export function LoginScreen({ navigate }: { navigate: (route: AuthRoute) => void }) {
  const { login, continueAsGuest, apiConfigured } = useAuth();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const submit = async () => {
    setError('');
    setLoading(true);
    const result = await login(email, password);
    setLoading(false);
    if (!result.ok) setError(result.message);
  };

  return (
    <AuthShell
      title="Bem-vindo ao ARIC"
      subtitle="Descubra novos usos para materiais, participe da comunidade e acompanhe o impacto das suas ações."
      footer={<Text style={styles.footer}>ARIC • conectado ao backend Go + PostgreSQL</Text>}
    >
      <View style={styles.modeRow}>
        <Ionicons name={apiConfigured ? 'shield-checkmark' : 'flask-outline'} size={16} color={apiConfigured ? colors.success : colors.orange} />
        <Text style={[styles.mode, apiConfigured && { color: colors.success }]}>{apiConfigured ? 'API Go configurada' : 'API não configurada — modo visitante'}</Text>
      </View>

      <FormField label="E-mail" icon="mail-outline" value={email} onChangeText={setEmail} keyboardType="email-address" autoComplete="email" placeholder="voce@email.com" />
      <FormField label="Senha" icon="lock-closed-outline" value={password} onChangeText={setPassword} password autoComplete="password" placeholder="Sua senha" />
      {error ? <Text style={styles.error}>{error}</Text> : null}

      <Pressable accessibilityRole="button" onPress={() => navigate('forgot')} style={styles.forgot}>
        <Text style={styles.link}>Esqueci minha senha</Text>
      </Pressable>

      <PrimaryButton label="Entrar" icon="log-in-outline" onPress={submit} loading={loading} />
      <PrimaryButton label="Continuar como visitante" icon="eye-outline" variant="secondary" onPress={() => continueAsGuest()} />

      <View style={styles.divider}><View style={styles.line} /><Text style={styles.or}>ou</Text><View style={styles.line} /></View>
      <Pressable accessibilityRole="button" onPress={() => navigate('register')} style={styles.create}>
        <Text style={styles.createText}>Ainda não tem conta? <Text style={styles.link}>Criar conta</Text></Text>
      </Pressable>
      <Pressable accessibilityRole="link" onPress={() => navigate('privacy')} style={styles.privacy}>
        <Text style={styles.privacyText}>Privacidade e uso de dados</Text>
      </Pressable>
    </AuthShell>
  );
}

const styles = StyleSheet.create({
  modeRow: { alignSelf: 'flex-start', flexDirection: 'row', alignItems: 'center', gap: 6, backgroundColor: colors.orangeSoft, borderRadius: radius.pill, paddingHorizontal: 10, paddingVertical: 6 },
  mode: { color: colors.orange, fontWeight: '800', fontSize: 11.5 },
  error: { color: colors.danger, backgroundColor: colors.dangerSoft, padding: 10, borderRadius: radius.sm, fontSize: 12.5, lineHeight: 18 },
  forgot: { alignSelf: 'flex-end', paddingVertical: 2 },
  link: { color: colors.primary, fontWeight: '900' },
  divider: { flexDirection: 'row', alignItems: 'center', gap: 10, marginVertical: 2 },
  line: { flex: 1, height: 1, backgroundColor: colors.border },
  or: { color: colors.textSoft, fontSize: 12 },
  create: { alignItems: 'center', paddingVertical: 4 },
  createText: { color: colors.textMuted, fontSize: 13.5 },
  privacy: { alignItems: 'center', paddingTop: 2 },
  privacyText: { color: colors.textMuted, textDecorationLine: 'underline', fontSize: 12 },
  footer: { color: colors.textSoft, fontSize: 11.5 },
});
