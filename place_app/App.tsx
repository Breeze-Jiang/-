import React, { useEffect } from 'react';
import { PermissionsAndroid, Platform, StatusBar } from 'react-native';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import RootNavigator from './src/navigation/RootNavigator';
import { useLocationPermissionStore } from './src/store/useLocationPermissionStore';
import { useP0SessionStore } from './src/p0/session';

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

  useEffect(() => {
    void useP0SessionStore.getState().bootstrap();
  }, []);

  return (
    <>
      <StatusBar barStyle="dark-content" backgroundColor="#FFFFFF" />
      <SafeAreaProvider><RootNavigator /></SafeAreaProvider>
    </>
  );
}
