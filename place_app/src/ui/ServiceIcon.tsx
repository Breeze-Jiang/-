import React from 'react';
import { View, StyleSheet } from 'react-native';
import { AppIcon, AppIconName } from './AppIcon';
import { COLORS } from '../shared/utils/theme';

const serviceIconMap: Record<string, AppIconName> = {
  toilet: 'toilet', parking: 'parking', charging: 'charging', gas: 'car',
  convenience: 'store', pharmacy: 'store', hospital: 'trust', police: 'trust', food: 'food',
};

export function ServiceIcon({ category, size = 40 }: { category: string; size?: number }) {
  const name = serviceIconMap[category] || 'pin';
  return <View style={[styles.wrap, { width: size, height: size, borderRadius: Math.round(size * .3) }]}><AppIcon name={name} color={COLORS.primary} size={Math.round(size * .48)} label={category} /></View>;
}

const styles = StyleSheet.create({ wrap: { alignItems: 'center', justifyContent: 'center', backgroundColor: COLORS.primaryLight } });
