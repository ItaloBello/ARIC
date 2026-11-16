import React, { memo } from 'react';
import { ImageBackground, Pressable, StyleSheet, Text, View } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import type { Project } from '../types';
import { colors, radius, shadow } from '../theme';

function FeaturedProjectCardBase({ project, onPress }: { project: Project; onPress: () => void }) {
  return (
    <Pressable accessibilityRole="button" accessibilityLabel={`Ver tutorial ${project.title}`} onPress={onPress} style={({ pressed }) => [styles.card, pressed && styles.pressed]}>
      <ImageBackground source={project.image} style={styles.image} imageStyle={styles.imageRadius} resizeMode="cover">
        <View style={styles.scrim} />
        <View style={styles.topRow}>
          <View style={styles.weekBadge}><Ionicons name="sparkles" size={14} color={colors.white} /><Text style={styles.weekText}>DESTAQUE DA SEMANA</Text></View>
          <View style={styles.impactBadge}><Ionicons name="refresh" size={14} color={colors.primary} /><Text style={styles.impactText}>-{project.impactKg.toFixed(1)} kg</Text></View>
        </View>
        <View style={styles.copy}>
          <Text style={styles.eyebrow}>{project.secondaryCategory ?? project.category}</Text>
          <Text style={styles.title}>{project.title}</Text>
          <Text style={styles.subtitle}>{project.subtitle}</Text>
        </View>
      </ImageBackground>
      <View style={styles.bottom}>
        <View style={styles.meta}><Ionicons name="time-outline" size={18} color={colors.primary} /><Text style={styles.metaText}>{project.time}</Text></View>
        <View style={styles.meta}><Ionicons name="stats-chart-outline" size={18} color={colors.primary} /><Text style={styles.metaText}>{project.difficulty}</Text></View>
        <View style={styles.points}><Ionicons name="ribbon-outline" size={18} color={colors.orange} /><Text style={styles.pointsText}>+{project.points}</Text></View>
        <View style={styles.cta}><Text style={styles.ctaText}>Ver tutorial</Text><Ionicons name="arrow-forward" size={18} color={colors.white} /></View>
      </View>
    </Pressable>
  );
}

export const FeaturedProjectCard = memo(FeaturedProjectCardBase);

const styles = StyleSheet.create({
  card: { ...shadow, backgroundColor: colors.surface, borderRadius: radius.lg, overflow: 'hidden', borderWidth: 1, borderColor: colors.border },
  pressed: { opacity: 0.96 },
  image: { height: 300, padding: 16, justifyContent: 'space-between' },
  imageRadius: { borderTopLeftRadius: radius.lg, borderTopRightRadius: radius.lg },
  scrim: { position: 'absolute', top: 0, right: 0, bottom: 0, left: 0, backgroundColor: 'rgba(4,35,23,0.38)' },
  topRow: { flexDirection: 'row', justifyContent: 'space-between', gap: 8 },
  weekBadge: { flexDirection: 'row', alignItems: 'center', gap: 5, backgroundColor: colors.orange, borderRadius: radius.pill, paddingHorizontal: 11, paddingVertical: 7 },
  weekText: { color: colors.white, fontWeight: '900', fontSize: 10.5 },
  impactBadge: { flexDirection: 'row', alignItems: 'center', gap: 4, backgroundColor: 'rgba(255,255,255,0.94)', borderRadius: radius.pill, paddingHorizontal: 10, paddingVertical: 7 },
  impactText: { color: colors.primaryDark, fontWeight: '900', fontSize: 11 },
  copy: { zIndex: 1 },
  eyebrow: { color: '#BFF0D0', fontSize: 12, fontWeight: '900', letterSpacing: 0.7, textTransform: 'uppercase' },
  title: { color: colors.white, fontWeight: '900', fontSize: 27, lineHeight: 32, marginTop: 5 },
  subtitle: { color: 'rgba(255,255,255,0.93)', fontSize: 14, lineHeight: 20, marginTop: 5 },
  bottom: { flexDirection: 'row', alignItems: 'center', gap: 10, padding: 13, flexWrap: 'wrap' },
  meta: { flexDirection: 'row', alignItems: 'center', gap: 4 },
  metaText: { color: colors.text, fontSize: 11.5, fontWeight: '700' },
  points: { flexDirection: 'row', alignItems: 'center', gap: 3 },
  pointsText: { color: colors.orange, fontWeight: '900', fontSize: 12 },
  cta: { marginLeft: 'auto', backgroundColor: colors.primary, borderRadius: radius.md, paddingHorizontal: 14, minHeight: 43, flexDirection: 'row', alignItems: 'center', gap: 7 },
  ctaText: { color: colors.white, fontWeight: '900', fontSize: 12.5 },
});
