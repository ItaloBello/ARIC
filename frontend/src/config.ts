export const appConfig = {
  name: 'ARIC',
  fullName: 'Aplicativo de Reciclagem Inteligente da Comunidade',
  version: '3.0.0',
  privacyPolicyUrl: process.env.EXPO_PUBLIC_PRIVACY_POLICY_URL?.trim() ?? '',
  accountDeletionUrl: process.env.EXPO_PUBLIC_ACCOUNT_DELETION_URL?.trim() ?? '',
  supportEmail: process.env.EXPO_PUBLIC_SUPPORT_EMAIL?.trim() ?? '',
};
