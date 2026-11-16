import type { ImageSourcePropType } from 'react-native';

export type MaterialCategory = 'Todos' | 'Plástico' | 'Papelão' | 'Vidro' | 'Metal' | 'Madeira' | 'Horta';

export type Project = {
  id: string;
  title: string;
  subtitle: string;
  category: MaterialCategory;
  secondaryCategory?: MaterialCategory;
  time: string;
  difficulty: 'Iniciante' | 'Intermediário' | 'Avançado';
  points: number;
  impactKg: number;
  co2Kg: number;
  image: ImageSourcePropType;
  creator: string;
  neighborhood: string;
  likes: number;
  saved?: boolean;
  completed?: boolean;
  description: string;
  materials: string[];
  steps: { title: string; description: string }[];
};

export type Mission = {
  id: string;
  title: string;
  description: string;
  category: MaterialCategory;
  reward: number;
  goal: number;
  unit: string;
  image: ImageSourcePropType;
};

export type ForumPost = {
  id: string;
  author: string;
  authorUid?: string;
  neighborhood: string;
  time: string;
  category: MaterialCategory;
  status: 'Resolvido' | 'Em aberto';
  title: string;
  body: string;
  likes: number;
  answers: number;
  image?: ImageSourcePropType;
  imageUrl?: string;
  isTip?: boolean;
  likedByMe?: boolean;
  publishedAt?: string;
};

export type ModerationDecision = 'approve' | 'reject';
export type ModerationStatus = 'pending_review' | 'approved' | 'rejected';
export type AiReviewStatus = 'not_run' | 'safe' | 'flagged' | 'error';

export type AiModerationSummary = {
  status: AiReviewStatus;
  flagged: boolean;
  categories: string[];
  recommendation: 'approve' | 'review' | 'reject';
  checkedAt?: string;
};

export type ForumSubmission = {
  id: string;
  author: string;
  authorUid: string;
  neighborhood: string;
  category: MaterialCategory;
  title: string;
  body: string;
  status: ModerationStatus;
  createdAt: string;
  imageUrl?: string;
  moderatorNote?: string;
  aiReview: AiModerationSummary;
};


export type CommunityRestriction = {
  uid: string;
  permanent: boolean;
  suspendedUntil?: string;
  reason: string;
  updatedAt: string;
};

export type AbuseReport = {
  id: string;
  postId: string;
  postTitle: string;
  postAuthorUid?: string;
  reporterUid: string;
  reason: string;
  details?: string;
  status: 'open' | 'resolved';
  createdAt: string;
};

export type UserProfile = {
  name: string;
  email?: string;
  city: string;
  bio: string;
  title: string;
  points: number;
  streak: number;
  impactKg: number;
  co2Kg: number;
  projectsCompleted: number;
  peopleInspired: number;
};

export type PersistedState = {
  user: UserProfile;
  favoriteProjectIds: string[];
  completedProjectIds: string[];
  missionProgress: Record<string, number>;
  completedMissionIds: string[];
  likedPostIds: string[];
  reportedPostIds: string[];
  blockedAuthors: string[];
  forumPosts: ForumPost[];
};
