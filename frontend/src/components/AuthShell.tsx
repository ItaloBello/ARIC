import React from 'react';
import { KeyboardAvoidingView, Platform, ScrollView, StyleSheet, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { AppLogo } from './AppLogo';
import { colors, radius, shadow, spacing } from '../theme';

export function AuthShell({ title, subtitle, children, footer }: { title: string; subtitle: string; children: React.ReactNode; footer?: React.ReactNode }) {
  return (
    <SafeAreaView style={styles.safe}>
      <KeyboardAvoidingView style={styles.flex} behavior={Platform.OS === 'ios' ? 'padding' : undefined}>
        <ScrollView contentContainerStyle={styles.scroll} keyboardShouldPersistTaps="handled" showsVerticalScrollIndicator={false}>
          <View style={styles.brandArea}>
            <View style={styles.orbA} />
            <View style={styles.orbB} />
            <AppLogo />
            <Text style={styles.kicker}>COMUNIDADE • REUSO • IMPACTO</Text>
            <Text style={styles.title}>{title}</Text>
            <Text style={styles.subtitle}>{subtitle}</Text>
          </View>
          <View style={styles.card}>{children}</View>
          {footer ? <View style={styles.footer}>{footer}</View> : null}
        </ScrollView>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.background },
  flex: { flex: 1 },
  scroll: { flexGrow: 1, paddingHorizontal: spacing.md, paddingTop: spacing.md, paddingBottom: spacing.xl },
  brandArea: { backgroundColor: colors.primaryDeep, borderRadius: radius.xl, padding: spacing.lg, overflow: 'hidden', minHeight: 285, justifyContent: 'flex-end' },
  orbA: { position: 'absolute', width: 210, height: 210, borderRadius: 105, backgroundColor: 'rgba(123,218,157,0.14)', right: -64, top: -64 },
  orbB: { position: 'absolute', width: 115, height: 115, borderRadius: 58, backgroundColor: 'rgba(224,139,79,0.18)', left: -40, bottom: -30 },
  kicker: { color: '#BFE8CE', fontSize: 10.5, fontWeight: '900', letterSpacing: 1.2, marginTop: 22 },
  title: { color: colors.white, fontSize: 32, lineHeight: 38, fontWeight: '900', marginTop: 8 },
  subtitle: { color: 'rgba(255,255,255,0.78)', fontSize: 14, lineHeight: 21, marginTop: 7, maxWidth: 315 },
  card: { ...shadow, backgroundColor: colors.surface, borderRadius: radius.xl, padding: spacing.lg, marginTop: -18, borderWidth: 1, borderColor: colors.border, gap: 15 },
  footer: { paddingTop: spacing.lg, alignItems: 'center' },
});
