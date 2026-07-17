import { create } from 'zustand';
import type { POI } from '../shared/types';

interface POIState {
  pois: POI[];
  searchKeyword: string;
  searchHistory: string[];
  isSearching: boolean;
  isLoading: boolean;

  setPois: (pois: POI[]) => void;
  setKeyword: (keyword: string) => void;
  setIsSearching: (searching: boolean) => void;
  setIsLoading: (loading: boolean) => void;
  addSearchHistory: (keyword: string) => void;
  removeSearchHistory: (keyword: string) => void;
  clearSearchHistory: () => void;
}

export const usePOIStore = create<POIState>((set, get) => ({
  pois: [],
  searchKeyword: '',
  searchHistory: [],
  isSearching: false,
  isLoading: false,

  setPois: (pois) => set({ pois }),

  setKeyword: (keyword) => set({ searchKeyword: keyword }),

  setIsSearching: (searching) => set({ isSearching: searching }),

  setIsLoading: (loading) => set({ isLoading: loading }),

  addSearchHistory: (keyword) => {
    const history = get().searchHistory.filter((h) => h !== keyword);
    set({ searchHistory: [keyword, ...history].slice(0, 20) });
  },

  removeSearchHistory: (keyword) => {
    set({ searchHistory: get().searchHistory.filter((h) => h !== keyword) });
  },

  clearSearchHistory: () => set({ searchHistory: [] }),
}));
