import React, { useState } from 'react';
import { Alert, Pressable, StyleSheet, Text, TextInput, View } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { FormField } from '../components/FormField';
import { PrimaryButton } from '../components/PrimaryButton';
import { ScreenShell } from '../components/ScreenShell';
import { useApp } from '../context/AppContext';
import { colors, radius, spacing } from '../theme';

export function EditProfileScreen({ goBack }: { goBack: () => void }) {
  const { user, updateProfile } = useApp();
  const [name, setName] = useState(user.name);
  const [bio, setBio] = useState(user.bio);

  const save = async () => {
    if (!name.trim()) return Alert.alert('Campo obrigatório', 'Informe seu nome.');
    const result = await updateProfile({ name: name.trim(), bio: bio.trim() });
    if (!result.ok) return Alert.alert('Perfil salvo parcialmente', result.message);
    Alert.alert('Perfil atualizado', 'As alterações foram sincronizadas.', [{ text: 'OK', onPress: goBack }]);
  };

  return (
    <ScreenShell>
      <View style={styles.topbar}>
        <Pressable accessibilityRole="button" accessibilityLabel="Voltar" onPress={goBack} style={styles.back}><Ionicons name="arrow-back" size={22} color={colors.primary} /></Pressable>
        <Text style={styles.topTitle}>Editar perfil</Text>
        <View style={{ width: 42 }} />
      </View>
      <Text style={styles.title}>Seu perfil no ARIC</Text>
      <Text style={styles.subtitle}>Atualize seu nome e sua apresentação. O e-mail e o endereço são protegidos e não podem ser alterados nesta tela.</Text>

      <View style={styles.form}>
        <FormField label="Nome" icon="person-outline" value={name} onChangeText={setName} autoCapitalize="words" />
        <View style={styles.bioWrap}>
          <Text style={styles.label}>Bio sustentável</Text>
          <TextInput value={bio} onChangeText={setBio} multiline style={styles.textarea} placeholder="Conte um pouco sobre seus interesses sustentáveis." placeholderTextColor={colors.textSoft} textAlignVertical="top" maxLength={500} />
          <Text style={styles.counter}>{bio.length}/500</Text>
        </View>
        <PrimaryButton label="Salvar alterações" icon="save-outline" onPress={save} />
      </View>
    </ScreenShell>
  );
}

const styles = StyleSheet.create({
  topbar: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between' },
  back: { width: 42, height: 42, borderRadius: 21, backgroundColor: colors.surface, borderWidth: 1, borderColor: colors.border, alignItems: 'center', justifyContent: 'center' },
  topTitle: { color: colors.text, fontWeight: '900', fontSize: 18 },
  title: { color: colors.text, fontSize: 28, fontWeight: '900', marginTop: spacing.lg },
  subtitle: { color: colors.textMuted, lineHeight: 20, marginTop: 5 },
  form: { gap: 15, marginTop: spacing.lg },
  bioWrap: { gap: 7 },
  label: { color: colors.text, fontWeight: '800', fontSize: 13 },
  textarea: { minHeight: 138, borderRadius: radius.md, borderWidth: 1, borderColor: colors.border, backgroundColor: colors.surface, padding: 14, color: colors.text, fontSize: 14 },
  counter: { textAlign: 'right', color: colors.textSoft, fontSize: 10.5 },
});
