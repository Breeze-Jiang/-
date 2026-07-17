import React, { useEffect } from 'react';
import { PermissionsAndroid, Platform, StatusBar } from 'react-native';
import RootNavigator from './src/navigation/RootNavigator';
import { useLocationPermissionStore } from './src/store/useLocationPermissionStore';

export default function App() {
  useEffect(() => {
    const requestLocation = async () => {
      if (Platform.OS !== 'android') {
        useLocationPermissionStore.getState().setStatus('granted');
        return;
      }
      try {
        const granted = await PermissionsAndroid.check(PermissionsAndroid.PERMISSIONS.ACCESS_FINE_LOCATION);
        useLocationPermissionStore.getState().setStatus(granted ? 'granted' : 'denied');
      } catch {
        useLocationPermissionStore.getState().setStatus('denied');
      }
    };
    void requestLocation();
  }, []);

  return (
    <>
      <StatusBar barStyle="dark-content" backgroundColor="#FFFFFF" />
      <RootNavigator />
    </>
  );
}
