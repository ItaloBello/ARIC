import React, { memo } from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Ionicons } from '@expo/vector-icons';
import type { TabKey } from '../navigation';
import { colors, shadow } from '../theme';

const items: { key: TabKey; label: string; icon: keyof typeof Ionicons.glyphMap; activeIcon: keyof typeof Ionicons.glyphMap }[] = [
  { key: 'home', label: 'Início', icon: 'home-outline', activeIcon: 'home' },
  { key: 'forum', label: 'Chat', icon: 'chatbubbles-outline', activeIcon: 'chatbubbles' },
  { key: 'profile', label: 'Perfil', icon: 'person-outline', activeIcon: 'person' },
];

function BottomNavBase({ active, onTab }: { active: TabKey; onTab: (tab: TabKey) => void }) {
  return (
    <SafeAreaView edges={['bottom']} style={styles.safe}>
      <View style={styles.bar}>
        {items.map((item) => {
          const selected = active === item.key;
          return (
            <Pressable
              accessibilityRole="tab"
              accessibilityState={{ selected }}
              accessibilityLabel={item.label}
              key={item.key}
              style={styles.item}
              onPress={() => onTab(item.key)}
            >
              <View style={[styles.iconWrap, selected && styles.iconWrapActive]}>
                <Ionicons name={selected ? item.activeIcon : item.icon} size={22} color={selected ? colors.primary : colors.textMuted} />
              </View>
              <Text style={[styles.label, selected && styles.labelActive]}>{item.label}</Text>
            </Pressable>
          );
        })}
      </View>
    </SafeAreaView>
  );
}

export const BottomNav = memo(BottomNavBase);

const styles = StyleSheet.create({
  safe: { backgroundColor: colors.surface },
  bar: { ...shadow, flexDirection: 'row', backgroundColor: colors.surface, borderTopWidth: 1, borderTopColor: colors.border, paddingTop: 7, paddingBottom: 7 },
  item: { flex: 1, alignItems: 'center', gap: 1, minHeight: 48 },
  iconWrap: { width: 50, height: 30, alignItems: 'center', justifyContent: 'center', borderRadius: 16 },
  iconWrapActive: { backgroundColor: colors.mint },
  label: { fontSize: 10.5, color: colors.textMuted, fontWeight: '700' },
  labelActive: { color: colors.primary, fontWeight: '900' },
});
