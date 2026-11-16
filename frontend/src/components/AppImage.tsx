import React, { memo, useState } from 'react';
import { Image, type ImageSourcePropType, StyleSheet, Text, View } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { colors } from '../theme';

function AppImageBase({ source, height = 190, borderRadius = 0 }: { source: ImageSourcePropType; height?: number; borderRadius?: number }) {
  const [failed, setFailed] = useState(false);
  if (failed) {
    return (
      <View style={[styles.fallback, { height, borderRadius }]}>
        <Ionicons name="leaf-outline" size={36} color={colors.primary} />
        <Text style={styles.fallbackText}>Imagem indisponível</Text>
      </View>
    );
  }
  return <Image source={source} style={[styles.image, { height, borderRadius }]} resizeMode="cover" onError={() => setFailed(true)} />;
}

export const AppImage = memo(AppImageBase);

const styles = StyleSheet.create({
  image: { width: '100%', backgroundColor: colors.mint },
  fallback: { width: '100%', backgroundColor: colors.mint, alignItems: 'center', justifyContent: 'center', gap: 6, overflow: 'hidden' },
  fallbackText: { color: colors.primary, fontWeight: '800' },
});
