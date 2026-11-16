import React from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Ionicons } from '@expo/vector-icons';
import { colors, radius, spacing } from '../theme';

export class ErrorBoundary extends React.Component<{ children: React.ReactNode }, { hasError: boolean }> {
  state = { hasError: false };

  static getDerivedStateFromError() {
    return { hasError: true };
  }

  render() {
    if (!this.state.hasError) return this.props.children;
    return (
      <SafeAreaView style={styles.safe}>
        <View style={styles.card}>
          <Ionicons name="warning-outline" size={42} color={colors.orange} />
          <Text style={styles.title}>Algo não saiu como esperado</Text>
          <Text style={styles.text}>Feche e abra o ARIC novamente. Se o problema continuar, registre a tela e envie para a equipe do projeto.</Text>
        </View>
      </SafeAreaView>
    );
  }
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.background, padding: spacing.lg, justifyContent: 'center' },
  card: { backgroundColor: colors.surface, borderRadius: radius.lg, borderWidth: 1, borderColor: colors.border, padding: spacing.xl, alignItems: 'center' },
  title: { color: colors.text, fontSize: 22, fontWeight: '900', textAlign: 'center', marginTop: 12 },
  text: { color: colors.textMuted, lineHeight: 21, textAlign: 'center', marginTop: 8 },
});
