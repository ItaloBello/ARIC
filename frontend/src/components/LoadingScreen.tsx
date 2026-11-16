import React from 'react';
import { ActivityIndicator, StyleSheet, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { AppLogo } from './AppLogo';
import { colors } from '../theme';

export function LoadingScreen() {
  return (
    <SafeAreaView style={styles.safe}>
      <View style={styles.center}>
        <AppLogo />
        <ActivityIndicator style={styles.loader} color={colors.primary} size="large" />
        <Text style={styles.text}>Preparando sua comunidade circular…</Text>
      </View>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.background },
  center: { flex: 1, alignItems: 'center', justifyContent: 'center', padding: 24 },
  loader: { marginTop: 26 },
  text: { color: colors.textMuted, marginTop: 12 },
});
