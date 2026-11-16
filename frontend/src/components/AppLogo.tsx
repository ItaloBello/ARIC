import React from 'react';
import { Image, StyleSheet, Text, View } from 'react-native';
import { colors } from '../theme';

export function AppLogo({ compact = false }: { compact?: boolean }) {
  return (
    <View style={styles.row} accessibilityRole="header">
      <Image source={require('../../assets/branding/icon.png')} style={[styles.image, compact && styles.imageCompact]} resizeMode="contain" />
      {!compact ? (
        <View style={styles.copy}>
          <Text style={styles.name}>ARIC</Text>
          <Text style={styles.subtitle}>Reciclagem inteligente da comunidade</Text>
        </View>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', alignItems: 'center' },
  image: { width: 62, height: 62, borderRadius: 18 },
  imageCompact: { width: 44, height: 44, borderRadius: 13 },
  copy: { marginLeft: 9, flexShrink: 1 },
  name: { fontSize: 25, lineHeight: 28, fontWeight: '900', color: colors.primary },
  subtitle: { marginTop: 1, fontSize: 10.5, lineHeight: 14, color: colors.textMuted, maxWidth: 172 },
});
