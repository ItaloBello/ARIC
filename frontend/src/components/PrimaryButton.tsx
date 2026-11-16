import React from 'react';
import { ActivityIndicator, Pressable, StyleSheet, Text, View } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { colors, radius } from '../theme';

type Props = {
  label: string;
  onPress: () => void;
  icon?: keyof typeof Ionicons.glyphMap;
  loading?: boolean;
  disabled?: boolean;
  variant?: 'primary' | 'secondary' | 'danger';
};

export function PrimaryButton({ label, onPress, icon, loading, disabled, variant = 'primary' }: Props) {
  const inactive = Boolean(disabled || loading);
  const palette = variant === 'danger'
    ? { bg: colors.dangerSoft, fg: colors.danger, border: '#F2C8C4' }
    : variant === 'secondary'
      ? { bg: colors.surface, fg: colors.primary, border: colors.borderStrong }
      : { bg: colors.primary, fg: colors.white, border: colors.primary };

  return (
    <Pressable
      accessibilityRole="button"
      accessibilityLabel={label}
      disabled={inactive}
      onPress={onPress}
      style={({ pressed }) => [styles.button, { backgroundColor: palette.bg, borderColor: palette.border }, inactive && styles.disabled, pressed && !inactive && styles.pressed]}
    >
      {loading ? <ActivityIndicator color={palette.fg} /> : (
        <View style={styles.content}>
          {icon ? <Ionicons name={icon} size={19} color={palette.fg} /> : null}
          <Text style={[styles.label, { color: palette.fg }]}>{label}</Text>
        </View>
      )}
    </Pressable>
  );
}

const styles = StyleSheet.create({
  button: { minHeight: 54, borderRadius: radius.md, borderWidth: 1, alignItems: 'center', justifyContent: 'center', paddingHorizontal: 18 },
  content: { flexDirection: 'row', alignItems: 'center', justifyContent: 'center', gap: 8 },
  label: { fontSize: 15, fontWeight: '900' },
  disabled: { opacity: 0.55 },
  pressed: { transform: [{ scale: 0.99 }], opacity: 0.92 },
});
