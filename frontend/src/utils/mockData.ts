// Mock data for missing API endpoints
// This file contains placeholder data for features not yet implemented in the backend

export const mockCredits = {
  balance: 250,
  currency: 'credits'
};

export const mockWorldCharacteristics = [
  { name: 'Technology Level', value: 85, color: '#2094f3' },
  { name: 'Magic Presence', value: 30, color: '#a855f7' },
  { name: 'Social Structure', value: 70, color: '#10b981' },
  { name: 'Geographic Diversity', value: 90, color: '#f59e0b' }
];

export const mockWorldHistory = [
  {
    era: 'Foundation Era',
    year: '2024',
    event: 'World created by AI generation system',
    description: 'Initial world parameters and base characters established'
  },
  {
    era: 'Early Development',
    year: '2024',
    event: 'First character interactions',
    description: 'AI characters began forming relationships and conflicts'
  },
  {
    era: 'Current Era',
    year: '2024',
    event: 'Active community phase',
    description: 'Multiple storylines and character arcs in progress'
  }
];

export const mockCharacterTraits = [
  'Analytical',
  'Empathetic', 
  'Strategic',
  'Creative',
  'Logical',
  'Intuitive'
];

export const mockCharacterSpecializations = [
  'Technology',
  'Diplomacy',
  'Research',
  'Leadership',
  'Innovation',
  'Communication'
];

export const mockUserSettings = {
  account: {
    username: 'user123',
    email: 'user@example.com',
    joinDate: '2024-01-15'
  },
  notifications: {
    newPosts: true,
    mentions: true,
    worldUpdates: false,
    weeklyDigest: true
  },
  privacy: {
    profileVisibility: 'public',
    showActivity: true,
    allowMessages: true
  }
};

export const mockCreditPackages = [
  {
    id: 'starter',
    name: 'Starter Pack',
    credits: 100,
    price: 9.99,
    popular: false
  },
  {
    id: 'creator',
    name: 'Creator Pack',
    credits: 500,
    price: 39.99,
    popular: true
  },
  {
    id: 'explorer',
    name: 'Explorer Pack',
    credits: 1000,
    price: 69.99,
    popular: false
  }
];

export const mockTransactionHistory = [
  {
    id: '1',
    type: 'purchase',
    amount: 500,
    date: '2024-03-01',
    description: 'Creator Pack purchased'
  },
  {
    id: '2',
    type: 'spend',
    amount: -50,
    date: '2024-03-02',
    description: 'World creation: Cyberpunk City'
  },
  {
    id: '3',
    type: 'spend',
    amount: -25,
    date: '2024-03-03',
    description: 'Character generation boost'
  }
];

export const mockWorldStatistics = {
  totalCharacters: 45,
  totalPosts: 234,
  totalLikes: 1892,
  activeUsers: 12,
  worldRank: 7,
  categoryRank: 2
};