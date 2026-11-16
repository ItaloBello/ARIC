import React, { useState } from 'react';
import { Pressable, StyleSheet, Text } from 'react-native';
import { AuthShell } from '../components/AuthShell';
import { FormField } from '../components/FormField';
import { PrimaryButton } from '../components/PrimaryButton';
import { useAuth } from '../context/AuthContext';
import type { AuthRoute } from '../navigation';
import { colors, radius } from '../theme';

export function ForgotPasswordScreen({ navigate }: { navigate: (route: AuthRoute) => void }) {
  const { resetPassword } = useAuth();
  const [email, setEmail] = useState('');
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const submit = async () => {
    setError('');
    setMessage('');
    setLoading(true);
    const result = await resetPassword(email);
    setLoading(false);
    if (!result.ok) return setError(result.message);
    setMessage('Se o e-mail estiver cadastrado, você receberá as instruções de recuperação.');
  };

  return (
    <AuthShell title="Recuperar acesso" subtitle="A recuperação por e-mail ainda não foi implementada no backend Go. Entre em contato com o suporte para redefinir seu acesso.">
      <FormField label="E-mail" icon="mail-outline" value={email} onChangeText={setEmail} keyboardType="email-address" autoComplete="email" placeholder="voce@email.com" />
      {error ? <Text style={styles.error}>{error}</Text> : null}
      {message ? <Text style={styles.success}>{message}</Text> : null}
      <PrimaryButton label="Verificar disponibilidade" icon="paper-plane-outline" onPress={submit} loading={loading} />
      <Pressable onPress={() => navigate('login')} style={styles.back}><Text style={styles.link}>Voltar para o login</Text></Pressable>
    </AuthShell>
  );
}

const styles = StyleSheet.create({
  error: { color: colors.danger, backgroundColor: colors.dangerSoft, padding: 10, borderRadius: radius.sm, fontSize: 12.5 },
  success: { color: colors.success, backgroundColor: colors.mintSoft, padding: 10, borderRadius: radius.sm, fontSize: 12.5, lineHeight: 18 },
  back: { alignItems: 'center', paddingVertical: 5 },
  link: { color: colors.primary, fontWeight: '900' },
});
