import type { MaterialCategory } from './types';

export type TabKey = 'home' | 'forum' | 'profile';

export type Route =
  | { name: 'home' }
  | { name: 'projects'; query?: string; category?: MaterialCategory }
  | { name: 'projectDetail'; projectId: string }
  | { name: 'forum' }
  | { name: 'askQuestion' }
  | { name: 'mySubmissions' }
  | { name: 'adminModeration' }
  | { name: 'profile' }
  | { name: 'editProfile' }
  | { name: 'privacy' }
  | { name: 'communityGuidelines' };

export type AuthRoute = 'login' | 'register' | 'forgot' | 'privacy' | 'communityGuidelines';
