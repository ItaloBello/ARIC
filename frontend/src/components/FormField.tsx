import React, { useState } from 'react';
import { Pressable, StyleSheet, Text, TextInput, type TextInputProps, View } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { colors, radius } from '../theme';

type Props = TextInputProps & {
  label: string;
  icon?: keyof typeof Ionicons.glyphMap;
  error?: string;
  password?: boolean;
};

export function FormField({ label, icon, error, password, style, ...props }: Props) {
  const [hidden, setHidden] = useState(Boolean(password));
  return (
    <View style={styles.wrap}>
      <Text style={styles.label}>{label}</Text>
      <View style={[styles.field, error && styles.fieldError]}>
        {icon ? <Ionicons name={icon} size={19} color={colors.textMuted} /> : null}
        <TextInput
          {...props}
          style={[styles.input, style]}
          placeholderTextColor={colors.textSoft}
          secureTextEntry={password ? hidden : props.secureTextEntry}
          autoCapitalize={props.autoCapitalize ?? 'none'}
        />
        {password ? (
          <Pressable accessibilityRole="button" accessibilityLabel={hidden ? 'Mostrar senha' : 'Ocultar senha'} onPress={() => setHidden((value) => !value)} hitSlop={10}>
            <Ionicons name={hidden ? 'eye-outline' : 'eye-off-outline'} size={20} color={colors.textMuted} />
          </Pressable>
        ) : null}
      </View>
      {error ? <Text style={styles.error}>{error}</Text> : null}
    </View>
  );
}

const styles = StyleSheet.create({
  wrap: { gap: 7 },
  label: { color: colors.text, fontWeight: '800', fontSize: 13 },
  field: { minHeight: 54, backgroundColor: colors.surface, borderWidth: 1, borderColor: colors.border, borderRadius: radius.md, flexDirection: 'row', alignItems: 'center', paddingHorizontal: 14, gap: 9 },
  fieldError: { borderColor: colors.danger },
  input: { flex: 1, minHeight: 50, color: colors.text, fontSize: 15 },
  error: { color: colors.danger, fontSize: 12 },
});
