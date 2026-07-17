import React from 'react';
import { StyleSheet, View } from 'react-native';
import Svg, { Circle, Path } from 'react-native-svg';
import { COLORS } from '../shared/utils/theme';

type MarkerState = 'default' | 'selected' | 'confirmed' | 'public';

export function MapMarker({ state = 'default', size = 48 }: { state?: MarkerState; size?: number }) {
  const fill = state === 'confirmed' ? COLORS.success : COLORS.primary;
  const stroke = state === 'selected' ? COLORS.primaryPressed : COLORS.white;
  const inner = state === 'public' ? COLORS.white : '#EAF2FF';
  return (
    <View accessible accessibilityLabel={state === 'public' ? '公开标注地点' : '地点标记'} style={[styles.wrap, { width: size, height: size }]}>
      <Svg width={size} height={size} viewBox="0 0 48 48">
        <Path d="M24 3C14.8 3 7.4 10.3 7.4 19.4c0 12.1 16.6 25.2 16.6 25.2s16.6-13.1 16.6-25.2C40.6 10.3 33.2 3 24 3Z" fill={fill} stroke={stroke} strokeWidth="2.5" />
        <Circle cx="24" cy="19" r="6" fill={inner} />
        {state === 'public' ? <Path d="M20.8 19.1l2.1 2.1 4.4-4.8" fill="none" stroke={fill} strokeWidth="2.1" strokeLinecap="round" strokeLinejoin="round" /> : null}
      </Svg>
    </View>
  );
}

const styles = StyleSheet.create({ wrap: { alignItems: 'center', justifyContent: 'center' } });
