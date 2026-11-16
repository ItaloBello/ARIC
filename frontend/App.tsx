import React, { useCallback, useMemo, useState } from 'react';
import { BackHandler, StyleSheet, View } from 'react-native';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import { StatusBar } from 'expo-status-bar';
import { AppProvider } from './src/context/AppContext';
import { AuthProvider, useAuth } from './src/context/AuthContext';
import { CommunityProvider } from './src/context/CommunityContext';
import { BottomNav } from './src/components/BottomNav';
import { ErrorBoundary } from './src/components/ErrorBoundary';
import { LoadingScreen } from './src/components/LoadingScreen';
import { HomeScreen } from './src/screens/HomeScreen';
import { ProjectsScreen } from './src/screens/ProjectsScreen';
import { ProjectDetailScreen } from './src/screens/ProjectDetailScreen';
import { ForumScreen } from './src/screens/ForumScreen';
import { AskQuestionScreen } from './src/screens/AskQuestionScreen';
import { MySubmissionsScreen } from './src/screens/MySubmissionsScreen';
import { AdminModerationScreen } from './src/screens/AdminModerationScreen';
import { ProfileScreen } from './src/screens/ProfileScreen';
import { EditProfileScreen } from './src/screens/EditProfileScreen';
import { LoginScreen } from './src/screens/LoginScreen';
import { RegisterScreen } from './src/screens/RegisterScreen';
import { ForgotPasswordScreen } from './src/screens/ForgotPasswordScreen';
import { PrivacyScreen } from './src/screens/PrivacyScreen';
import { CommunityGuidelinesScreen } from './src/screens/CommunityGuidelinesScreen';
import type { AuthRoute, Route, TabKey } from './src/navigation';
import { colors } from './src/theme';

const ROOT_ROUTES: Record<TabKey, Route> = {
  home: { name: 'home' },
  forum: { name: 'forum' },
  profile: { name: 'profile' },
};

function activeTab(route: Route): TabKey {
  switch (route.name) {
    case 'projects':
    case 'projectDetail': return 'home';
    case 'askQuestion': return 'forum';
    case 'mySubmissions':
    case 'adminModeration':
    case 'editProfile':
    case 'privacy':
    case 'communityGuidelines': return 'profile';
    case 'home': return 'home';
    case 'forum': return 'forum';
    default: return 'profile';
  }
}

function AuthRoot() {
  const [stack, setStack] = useState<AuthRoute[]>(['login']);
  const route = stack[stack.length - 1] ?? 'login';
  const navigate = useCallback((next: AuthRoute) => setStack((prev) => [...prev, next]), []);
  const goBack = useCallback(() => setStack((prev) => (prev.length > 1 ? prev.slice(0, -1) : ['login'])), []);

  switch (route) {
    case 'register': return <RegisterScreen navigate={navigate} />;
    case 'forgot': return <ForgotPasswordScreen navigate={navigate} />;
    case 'privacy': return <PrivacyScreen goBack={goBack} />;
    case 'communityGuidelines': return <CommunityGuidelinesScreen goBack={goBack} />;
    default: return <LoginScreen navigate={navigate} />;
  }
}

function AppRoot() {
  const [stack, setStack] = useState<Route[]>([{ name: 'home' }]);
  const route: Route = stack[stack.length - 1] ?? { name: 'home' };

  const navigate = useCallback((next: Route) => setStack((prev) => [...prev, next]), []);
  const replaceRoot = useCallback((tab: TabKey) => setStack([ROOT_ROUTES[tab]]), []);
  const goBack = useCallback(() => setStack((prev) => (prev.length > 1 ? prev.slice(0, -1) : prev)), []);

  React.useEffect(() => {
    const sub = BackHandler.addEventListener('hardwareBackPress', () => {
      if (stack.length > 1) {
        goBack();
        return true;
      }
      return false;
    });
    return () => sub.remove();
  }, [goBack, stack.length]);

  const screen = useMemo(() => {
    switch (route.name) {
      case 'home': return <HomeScreen navigate={navigate} />;
      case 'projects': return <ProjectsScreen navigate={navigate} initialQuery={route.query} initialCategory={route.category} />;
      case 'projectDetail': return <ProjectDetailScreen projectId={route.projectId} goBack={goBack} />;
      case 'forum': return <ForumScreen navigate={navigate} />;
      case 'askQuestion': return <AskQuestionScreen goBack={goBack} openGuidelines={() => navigate({ name: 'communityGuidelines' })} />;
      case 'mySubmissions': return <MySubmissionsScreen goBack={goBack} />;
      case 'adminModeration': return <AdminModerationScreen goBack={goBack} />;
      case 'profile': return <ProfileScreen navigate={navigate} />;
      case 'editProfile': return <EditProfileScreen goBack={goBack} />;
      case 'privacy': return <PrivacyScreen goBack={goBack} />;
      case 'communityGuidelines': return <CommunityGuidelinesScreen goBack={goBack} />;
    }
  }, [goBack, navigate, route]);

  return (
    <View style={styles.app}>
      <View style={styles.content}>{screen}</View>
      <BottomNav active={activeTab(route)} onTab={replaceRoot} />
    </View>
  );
}

function RootGate() {
  const { status } = useAuth();
  if (status === 'loading') return <LoadingScreen />;
  if (status === 'signedOut') return <AuthRoot />;
  return <AppRoot />;
}

export default function App() {
  return (
    <SafeAreaProvider>
      <StatusBar style="dark" />
      <ErrorBoundary>
        <AuthProvider>
          <AppProvider>
            <CommunityProvider>
              <RootGate />
            </CommunityProvider>
          </AppProvider>
        </AuthProvider>
      </ErrorBoundary>
    </SafeAreaProvider>
  );
}

const styles = StyleSheet.create({
  app: { flex: 1, backgroundColor: colors.background },
  content: { flex: 1 },
});
