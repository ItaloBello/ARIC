import type { ViewStyle } from 'react-native';

export const colors = {
  background: '#F8F7F2',
  backgroundWarm: '#F3F1E9',
  surface: '#FFFFFF',
  surfaceMuted: '#F1F5F1',
  primary: '#087548',
  primaryDark: '#075334',
  primaryDeep: '#053A26',
  mint: '#E4F4EA',
  mintStrong: '#BFE5CD',
  mintSoft: '#F0F9F3',
  orange: '#D8662F',
  orangeSoft: '#FFF0E7',
  gold: '#E7A52E',
  goldSoft: '#FFF7DE',
  text: '#16231C',
  textMuted: '#647068',
  textSoft: '#889189',
  border: '#E0E7E1',
  borderStrong: '#CDD9D0',
  danger: '#C9483C',
  dangerSoft: '#FFF0EF',
  success: '#1A7A4A',
  white: '#FFFFFF',
  black: '#000000',
};

export const spacing = { xs: 6, sm: 10, md: 16, lg: 22, xl: 28, xxl: 36 };
export const radius = { sm: 12, md: 18, lg: 24, xl: 30, pill: 999 };

export const shadow: ViewStyle = {
  shadowColor: '#0B2D1E',
  shadowOffset: { width: 0, height: 8 },
  shadowOpacity: 0.08,
  shadowRadius: 18,
  elevation: 3,
};

export const typography = {
  title: { fontSize: 30, lineHeight: 36, fontWeight: '900' as const, color: colors.text },
  h2: { fontSize: 22, lineHeight: 28, fontWeight: '900' as const, color: colors.text },
  body: { fontSize: 15, lineHeight: 22, color: colors.text },
  muted: { fontSize: 13, lineHeight: 19, color: colors.textMuted },
  label: { fontSize: 13, lineHeight: 18, fontWeight: '800' as const, color: colors.text },
};
