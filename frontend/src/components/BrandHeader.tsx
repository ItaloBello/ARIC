import React from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { AppLogo } from './AppLogo';
import { colors, radius } from '../theme';
import { useApp } from '../context/AppContext';

export function BrandHeader({ onProfile }: { onProfile?: () => void }) {
  const { user } = useApp();
  return (
    <View style={styles.row}>
      <AppLogo compact />
      <View style={styles.wordmark}>
        <Text style={styles.name}>ARIC</Text>
        <Text style={styles.subtitle}>Comunidade circular</Text>
      </View>
      <View style={styles.spacer} />
      <View style={styles.points}>
        <Ionicons name="leaf" size={17} color={colors.primary} />
        <Text style={styles.pointsText}>{user.points}</Text>
        <Text style={styles.pts}>pts</Text>
      </View>
      <Pressable accessibilityRole="button" accessibilityLabel="Abrir perfil" onPress={onProfile} style={styles.avatar}>
        <Text style={styles.avatarText}>{user.name.trim().charAt(0).toUpperCase() || 'A'}</Text>
      </Pressable>
    </View>
  );
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', alignItems: 'center', minHeight: 52 },
  wordmark: { marginLeft: 7 },
  name: { color: colors.primary, fontSize: 20, lineHeight: 22, fontWeight: '900' },
  subtitle: { color: colors.textMuted, fontSize: 9.5, marginTop: 1 },
  spacer: { flex: 1 },
  points: { flexDirection: 'row', alignItems: 'baseline', gap: 3, backgroundColor: colors.mint, paddingHorizontal: 11, paddingVertical: 8, borderRadius: radius.pill },
  pointsText: { color: colors.primaryDark, fontWeight: '900', fontSize: 15 },
  pts: { color: colors.primaryDark, fontWeight: '700', fontSize: 11 },
  avatar: { width: 40, height: 40, borderRadius: 20, backgroundColor: colors.primary, alignItems: 'center', justifyContent: 'center', marginLeft: 8, borderWidth: 3, borderColor: colors.mintStrong },
  avatarText: { color: colors.white, fontWeight: '900', fontSize: 16 },
});
