import React from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { AppIcon } from './AppIcon';
import { COLORS, FONTS } from '../shared/utils/theme';

export function TrustBadge({ label = '已确认' }: { label?: string }) {
  return <View style={styles.badge}><AppIcon name="trust" color={COLORS.success} size={14} label={label} /><Text style={styles.text}>{label}</Text></View>;
}

const styles = StyleSheet.create({ badge: { flexDirection: 'row', alignItems: 'center', minHeight: 24, paddingHorizontal: 7, gap: 3, borderRadius: 999, backgroundColor: COLORS.successLight }, text: { ...FONTS.caption, fontSize: 11, color: COLORS.success, fontWeight: '700' } });
