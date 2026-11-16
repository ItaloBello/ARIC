import React, { memo } from 'react';
import { StyleSheet, View } from 'react-native';
import { colors, radius } from '../theme';

function ProgressBarBase({ value, height = 9, accent = colors.primary }: { value: number; height?: number; accent?: string }) {
  const safe = Math.max(0, Math.min(1, value));
  return (
    <View accessibilityRole="progressbar" accessibilityValue={{ min: 0, max: 100, now: Math.round(safe * 100) }} style={[styles.track, { height }]}>
      <View style={[styles.fill, { width: `${safe * 100}%`, backgroundColor: accent }]} />
    </View>
  );
}

export const ProgressBar = memo(ProgressBarBase);

const styles = StyleSheet.create({
  track: { borderRadius: radius.pill, backgroundColor: '#DCE5DF', overflow: 'hidden' },
  fill: { height: '100%', borderRadius: radius.pill },
});
