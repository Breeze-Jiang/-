import React, { useState } from 'react';
import { PermissionsAndroid, Platform, Pressable, ScrollView, StyleSheet, Text, TextInput, View } from 'react-native';
import { alongtuServices, urgentServices } from '../demoData';
import { TravelMapSurface } from '../components/TravelMapSurface';
import { Chip, Glyph, StatusPill } from '../../shared/components/AppUI';
import { COLORS, FONTS, RADIUS, SHADOWS, SPACING } from '../../shared/utils/theme';
import { useLocationPermissionStore } from '../../store/useLocationPermissionStore';
import { AppIcon } from '../../ui/AppIcon';
import { ServiceIcon } from '../../ui/ServiceIcon';
import type { Place } from '../types';
import type { Coordinate } from '../../p0/contracts';
import { toP0Category } from '../../p0/presentation';

export default function MapHomeScreen({ navigation }: any) {
  const [serviceOpen, setServiceOpen] = useState(true);
  const [query, setQuery] = useState('');
  const [currentLocation, setCurrentLocation] = useState<Coordinate>();
  const [locationHint, setLocationHint] = useState('');
  const locationStatus = useLocationPermissionStore((state) => state.status);
  const openCategory = (category: Place['category'], label: string) => navigation.navigate('PlaceResults', {
    category: toP0Category(category),
    label,
    center: currentLocation,
    query: currentLocation ? undefined : label,
  });
  const submitSearch = () => {
    const keyword = query.trim();
    if (keyword) navigation.navigate('PlaceResults', { label: keyword, query: keyword });
  };
  const locate = async () => {
    if (locationStatus === 'granted') {
      setLocationHint(currentLocation ? '已获得当前位置，可选择服务类别查看附近地点。' : '正在获取当前位置，请稍候后选择服务类别。');
      return;
    }
    if (Platform.OS !== 'android') {
      navigation.navigate('LocationPermission');
      return;
    }
    try {
      const result = await PermissionsAndroid.request(PermissionsAndroid.PERMISSIONS.ACCESS_FINE_LOCATION, {
        title: '允许沿途使用位置',
        message: '仅用于查找附近的厕所、停车场、充电站等地点，以及导航。',
        buttonPositive: '允许',
        buttonNegative: '暂不允许',
      });
      if (result === PermissionsAndroid.RESULTS.GRANTED) {
        useLocationPermissionStore.getState().setStatus('granted');
        setLocationHint('正在获取当前位置，请稍候后选择服务类别。');
      } else {
        useLocationPermissionStore.getState().setStatus('denied');
        navigation.navigate('LocationPermission');
      }
    } catch {
      useLocationPermissionStore.getState().setStatus('denied');
      navigation.navigate('LocationPermission');
    }
  };

  return (
    <View style={styles.page}>
      <View style={styles.map}>
        <TravelMapSurface locationEnabled={locationStatus === 'granted'} onMapPress={() => setServiceOpen(false)} onLocation={(coordinate) => { setCurrentLocation(coordinate); setLocationHint('已获得当前位置，可按服务类别查看附近地点。'); }} />
      </View>

      <View style={styles.searchWrap}>
        <AppIcon name="search" label="搜索" color={COLORS.textSecondary} />
        <TextInput value={query} onChangeText={setQuery} placeholder="搜索地点、地址、景点" placeholderTextColor={COLORS.textTertiary} returnKeyType="search" onSubmitEditing={submitSearch} style={styles.search} accessibilityLabel="搜索地点、地址、景点" />
        <Pressable accessibilityRole="button" accessibilityLabel="定位当前位置" onPress={locate} style={styles.locate}>
          <AppIcon name="locate" label="定位当前位置" color={COLORS.primary} />
        </Pressable>
      </View>

      <Pressable accessibilityRole="button" accessibilityLabel="标注此处" onPress={() => navigation.navigate('CreateMarker')} style={styles.markerAction}>
        <AppIcon name="plus" label="新增标注" color={COLORS.white} size={18} strokeWidth={2.4} />
        <Text style={styles.markerActionText}>标注此处</Text>
      </Pressable>

      {serviceOpen ? (
        <View style={styles.sheet}>
          <View style={styles.handle} />
          <View style={styles.sheetTitleRow}>
            <View><Text style={styles.sheetTitle}>急需服务</Text><Text style={styles.sheetSubtitle}>优先显示附近可导航地点</Text></View>
            <StatusPill label="距离优先" tone="primary" />
          </View>
          <View style={styles.serviceGrid}>
            {urgentServices.map((item) => <Pressable key={item.key} accessibilityRole="button" accessibilityLabel={`查找附近${item.label}`} onPress={() => openCategory(item.key, item.label)} style={styles.serviceItem}><ServiceIcon category={item.key} /><Text style={styles.serviceLabel}>{item.label}</Text></Pressable>)}
          </View>
          {locationHint ? <Text style={styles.locationHint}>{locationHint}</Text> : null}
          <View style={styles.alongtuHeader}>
            <View><Text style={styles.alongtuTitle}>沿途服务</Text><Text style={styles.alongtuSubtitle}>把旅行中常漏掉的资源直接放在地图上</Text></View>
            <Pressable accessibilityRole="button" accessibilityLabel="查看全部沿途服务" onPress={() => navigation.navigate('PlaceResults', { label: '沿途服务' })}><Text style={styles.more}>全部 ›</Text></Pressable>
          </View>
          <ScrollView horizontal showsHorizontalScrollIndicator={false} contentContainerStyle={styles.chips}>
            {alongtuServices.map((item) => <Chip key={item.key} label={`${item.symbol}  ${item.label}`} onPress={() => openCategory(item.key, item.label)} />)}
          </ScrollView>
        </View>
      ) : <Pressable accessibilityRole="button" accessibilityLabel="展开沿途服务" onPress={() => setServiceOpen(true)} style={styles.reopen}><Glyph symbol="▤" label="沿途服务" color={COLORS.primary} /><Text style={styles.reopenText}>沿途服务</Text></Pressable>}
    </View>
  );
}

const styles = StyleSheet.create({
  page: { flex: 1, backgroundColor: COLORS.background },
  map: { ...StyleSheet.absoluteFillObject, backgroundColor: '#E8F0EC', overflow: 'hidden' },
  searchWrap: { position: 'absolute', top: 14, left: 16, right: 16, minHeight: 54, borderRadius: RADIUS.lg, backgroundColor: COLORS.white, flexDirection: 'row', alignItems: 'center', paddingLeft: 16, ...SHADOWS.card },
  search: { flex: 1, minHeight: 52, color: COLORS.text, ...FONTS.body, paddingHorizontal: 10 },
  locate: { height: 44, width: 44, justifyContent: 'center', alignItems: 'center', borderLeftWidth: StyleSheet.hairlineWidth, borderColor: COLORS.border },
  markerAction: { position: 'absolute', right: 16, bottom: 90, minHeight: 44, paddingHorizontal: 14, borderRadius: RADIUS.full, backgroundColor: COLORS.primary, flexDirection: 'row', alignItems: 'center', gap: 6, ...SHADOWS.button },
  markerActionText: { ...FONTS.caption, color: COLORS.white, fontWeight: '700' },
  sheet: { position: 'absolute', left: 0, right: 0, bottom: 0, maxHeight: '72%', backgroundColor: COLORS.white, borderTopLeftRadius: RADIUS.xl, borderTopRightRadius: RADIUS.xl, paddingHorizontal: SPACING.lg, paddingBottom: 14, ...SHADOWS.card },
  handle: { height: 4, width: 38, alignSelf: 'center', borderRadius: 3, backgroundColor: COLORS.border, marginVertical: 10 },
  sheetTitleRow: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' },
  sheetTitle: { ...FONTS.titleMedium, color: COLORS.text },
  sheetSubtitle: { ...FONTS.caption, color: COLORS.textTertiary, marginTop: 2 },
  serviceGrid: { flexDirection: 'row', flexWrap: 'wrap', marginTop: 14 },
  serviceItem: { width: '16.66%', alignItems: 'center', marginBottom: 12, minHeight: 68 },
  serviceLabel: { ...FONTS.caption, color: COLORS.text, marginTop: 5, fontWeight: '600' },
  locationHint: { ...FONTS.caption, color: COLORS.textSecondary, marginTop: -3, marginBottom: 10 },
  alongtuHeader: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', borderTopWidth: StyleSheet.hairlineWidth, borderTopColor: COLORS.border, paddingTop: 14 },
  alongtuTitle: { ...FONTS.body, color: COLORS.text, fontWeight: '700' },
  alongtuSubtitle: { ...FONTS.caption, color: COLORS.textTertiary, marginTop: 2 },
  more: { ...FONTS.caption, color: COLORS.primary, fontWeight: '700' },
  chips: { paddingVertical: 12 },
  reopen: { position: 'absolute', bottom: 24, left: 16, height: 48, borderRadius: RADIUS.full, paddingHorizontal: 16, backgroundColor: COLORS.white, ...SHADOWS.card, flexDirection: 'row', alignItems: 'center' },
  reopenText: { ...FONTS.caption, color: COLORS.primary, fontWeight: '700', marginLeft: 6 },
});
