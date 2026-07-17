import React from 'react';
import { StyleSheet, View } from 'react-native';
import type { Coordinate } from '../../p0/contracts';

// Keep the AMap package behind an untyped boundary: its published TypeScript
// sources target an older React type model, while the native component is used
// normally by Metro/Android at runtime.
// eslint-disable-next-line @typescript-eslint/no-var-requires
const AMap = require('react-native-amap3d');
const NativeMapView = AMap.MapView as React.ComponentType<any>;

function coordinateFromLocationEvent(event: any): Coordinate | undefined {
  const coords = event?.nativeEvent?.coords;
  const latitude = Number(coords?.latitude);
  const longitude = Number(coords?.longitude);
  if (!Number.isFinite(latitude) || !Number.isFinite(longitude) || latitude < -90 || latitude > 90 || longitude < -180 || longitude > 180) {
    return undefined;
  }
  // AMap native location delivers mainland China map coordinates. The explicit
  // declaration prevents them from being sent to the API as WGS84.
  return { latitude, longitude, coordinateSystem: 'gcj02' };
}

export function TravelMapSurface({ locationEnabled = false, onMapPress, onLocation }: { locationEnabled?: boolean; onMapPress?: () => void; onLocation?: (coordinate: Coordinate) => void }) {
  return <View style={StyleSheet.absoluteFill}><NativeMapView style={StyleSheet.absoluteFill} initialCameraPosition={{ target: { latitude: 23.1291, longitude: 113.2644 }, zoom: 13 }} myLocationEnabled={locationEnabled} compassEnabled zoomControlsEnabled={false} scaleControlsEnabled={false} onPress={onMapPress} onLocation={(event: any) => { const coordinate = coordinateFromLocationEvent(event); if (coordinate) onLocation?.(coordinate); }} /></View>;
}
