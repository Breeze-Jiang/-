import { create } from 'zustand';
import type { POI, POIType } from '../shared/types';

interface MapState {
  currentLat: number | null;
  currentLng: number | null;
  zoom: number;
  selectedPOI: POI | null;
  activeFilters: POIType[];
  isLocating: boolean;

  setLocation: (lat: number, lng: number) => void;
  setZoom: (zoom: number) => void;
  selectPOI: (poi: POI | null) => void;
  toggleFilter: (type: POIType) => void;
  setFilters: (types: POIType[]) => void;
  clearFilters: () => void;
}

export const useMapStore = create<MapState>((set, get) => ({
  currentLat: null,
  currentLng: null,
  zoom: 15,
  selectedPOI: null,
  activeFilters: [],
  isLocating: true,

  setLocation: (lat, lng) => {
    set({ currentLat: lat, currentLng: lng, isLocating: false });
  },

  setZoom: (zoom) => set({ zoom }),

  selectPOI: (poi) => set({ selectedPOI: poi }),

  toggleFilter: (type) => {
    const filters = get().activeFilters;
    if (filters.includes(type)) {
      set({ activeFilters: filters.filter((f) => f !== type) });
    } else {
      set({ activeFilters: [...filters, type] });
    }
  },

  setFilters: (types) => set({ activeFilters: types }),

  clearFilters: () => set({ activeFilters: [] }),
}));
