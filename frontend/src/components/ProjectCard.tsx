import React, { memo } from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import type { Project } from '../types';
import { colors, radius, shadow, spacing } from '../theme';
import { AppImage } from './AppImage';

function ProjectCardBase({ project, onPress, favorite, onFavorite }: { project: Project; onPress: () => void; favorite: boolean; onFavorite: () => void }) {
  return (
    <Pressable accessibilityRole="button" accessibilityLabel={`Abrir projeto ${project.title}`} onPress={onPress} style={({ pressed }) => [styles.card, pressed && styles.pressed]}>
      <View style={styles.imageWrap}>
        <AppImage source={project.image} height={176} />
        <Pressable accessibilityRole="button" accessibilityLabel={favorite ? 'Remover dos favoritos' : 'Salvar nos favoritos'} style={styles.favorite} onPress={(event) => { event.stopPropagation(); onFavorite(); }}>
          <Ionicons name={favorite ? 'heart' : 'heart-outline'} size={21} color={favorite ? colors.orange : colors.primaryDeep} />
        </Pressable>
        <View style={styles.category}><Text style={styles.categoryText}>{project.category}</Text></View>
      </View>
      <View style={styles.body}>
        <Text style={styles.title}>{project.title}</Text>
        <Text style={styles.subtitle} numberOfLines={2}>{project.subtitle}</Text>
        <View style={styles.meta}>
          <Meta icon="time-outline" text={project.time} />
          <Meta icon="stats-chart-outline" text={project.difficulty} />
          <View style={styles.pointsWrap}><Ionicons name="ribbon-outline" size={15} color={colors.orange} /><Text style={styles.points}>+{project.points} pts</Text></View>
        </View>
      </View>
    </Pressable>
  );
}

function Meta({ icon, text }: { icon: keyof typeof Ionicons.glyphMap; text: string }) {
  return <View style={styles.metaItem}><Ionicons name={icon} size={15} color={colors.textMuted} /><Text style={styles.metaText}>{text}</Text></View>;
}

export const ProjectCard = memo(ProjectCardBase);

const styles = StyleSheet.create({
  card: { ...shadow, backgroundColor: colors.surface, borderRadius: radius.lg, overflow: 'hidden', borderWidth: 1, borderColor: colors.border, marginBottom: spacing.md },
  pressed: { opacity: 0.94, transform: [{ scale: 0.995 }] },
  imageWrap: { position: 'relative', overflow: 'hidden' },
  favorite: { position: 'absolute', top: 12, right: 12, width: 39, height: 39, borderRadius: 20, backgroundColor: 'rgba(255,255,255,0.94)', alignItems: 'center', justifyContent: 'center' },
  category: { position: 'absolute', bottom: 11, left: 11, backgroundColor: 'rgba(232,247,237,0.96)', paddingHorizontal: 12, paddingVertical: 6, borderRadius: radius.pill },
  categoryText: { color: colors.primaryDark, fontWeight: '900', fontSize: 12 },
  body: { padding: spacing.md },
  title: { fontSize: 18, lineHeight: 23, fontWeight: '900', color: colors.text },
  subtitle: { marginTop: 4, color: colors.textMuted, lineHeight: 19, fontSize: 13 },
  meta: { flexDirection: 'row', alignItems: 'center', gap: 11, marginTop: 13, flexWrap: 'wrap' },
  metaItem: { flexDirection: 'row', alignItems: 'center', gap: 4 },
  metaText: { color: colors.textMuted, fontSize: 11.5, fontWeight: '700' },
  pointsWrap: { flexDirection: 'row', alignItems: 'center', gap: 4, marginLeft: 'auto' },
  points: { color: colors.orange, fontWeight: '900', fontSize: 12 },
});
