import { create } from 'zustand';

interface PresentationSessionState {
  isSignedIn: boolean;
  displayName: string;
  completeLocalSignIn: () => void;
  signOut: () => void;
}

/**
 * UI-only session state for non-P0 demonstration screens. It must not be
 * treated as a credential source by the P0 API client.
 */
export const usePresentationSessionStore = create<PresentationSessionState>((set) => ({
  isSignedIn: false,
  displayName: '沿途旅者',
  completeLocalSignIn: () => set({ isSignedIn: true }),
  signOut: () => set({ isSignedIn: false }),
}));
