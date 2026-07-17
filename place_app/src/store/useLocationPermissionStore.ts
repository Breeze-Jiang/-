import { create } from 'zustand';

export type LocationPermissionState = 'checking' | 'granted' | 'denied';

interface LocationPermissionStore {
  status: LocationPermissionState;
  setStatus: (status: LocationPermissionState) => void;
}

export const useLocationPermissionStore = create<LocationPermissionStore>((set) => ({
  status: 'checking',
  setStatus: (status) => set({ status }),
}));
