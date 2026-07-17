import React, { useEffect, useState } from 'react';
import { ActivityIndicator, ScrollView, StyleSheet, Text, View } from 'react-native';
import type { TravelMode } from '../types';
import { AppHeader, Chip, EmptyState, Page, PrimaryButton, StatusPill } from '../../shared/components/AppUI';
import { COLORS, FONTS, RADIUS, SPACING } from '../../shared/utils/theme';
import { p0Api, P0ApiError } from '../../p0/client';
import type { Coordinate, NavigationMode, P0Place, Route } from '../../p0/contracts';
import { formatDistance } from '../../shared/utils/format';

const modes: { id: TravelMode; apiMode: NavigationMode; label: string; symbol: string }[] = [
  { id: 'walk', apiMode: 'walking', label: '步行', symbol: '◌' },
  { id: 'ride', apiMode: 'cycling', label: '骑行', symbol: '◈' },
  { id: 'drive', apiMode: 'driving', label: '驾车', symbol: '▰' },
  { id: 'transit', apiMode: 'transit', label: '公交地铁', symbol: '▣' },
  { id: 'taxi', apiMode: 'taxi', label: '打车', symbol: '◉' },
];

function durationLabel(seconds: number) {
  const minutes = Math.max(1, Math.round(seconds / 60));
  return minutes >= 60 ? `${Math.floor(minutes / 60)} 小时 ${minutes % 60} 分钟` : `${minutes} 分钟`;
}

function errorMessage(error: unknown) {
  return error instanceof P0ApiError ? error.message : '暂时无法计算路线，请使用外部地图继续导航。';
}

export default function RouteModeScreen({ navigation, route }: any) {
  const [mode, setMode] = useState<TravelMode>('walk');
  const [place, setPlace] = useState<P0Place>();
  const [routeData, setRouteData] = useState<Route>();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>();
  const placeId = route.params?.placeId as string | undefined;
  const origin = route.params?.origin as Coordinate | undefined;
  const selected = modes.find((item) => item.id === mode)!;

  useEffect(() => {
    let active = true;
    const load = async () => {
      if (!placeId || !origin) {
        if (active) { setError('需要当前位置后才能计算路线。'); setLoading(false); }
        return;
      }
      if (selected.apiMode === 'transit') {
        if (active) { setRouteData(undefined); setError('公交地铁需要出发城市上下文；当前定位尚未取得该信息，请使用外部地图导航。'); setLoading(false); }
        return;
      }
      setLoading(true);
      setError(undefined);
      try {
        const destination = await p0Api.place(placeId);
        const nextRoute = await p0Api.route(origin, destination.location, selected.apiMode, { destination: destination.cityCode });
        if (active) { setPlace(destination); setRouteData(nextRoute); }
      } catch (cause) {
        if (active) { setRouteData(undefined); setError(errorMessage(cause)); }
      } finally {
        if (active) setLoading(false);
      }
    };
    void load();
    return () => { active = false; };
  }, [origin?.coordinateSystem, origin?.latitude, origin?.longitude, placeId, selected.apiMode]);

  const openExternal = () => navigation.navigate('ExternalMap', { placeId, origin, mode: selected.apiMode, type: mode === 'taxi' ? 'taxi' : 'map' });
  if (!origin) return <Page><AppHeader title="选择出行方式" onBack={() => navigation.goBack()} /><EmptyState title="需要当前位置" description="路线和外部导航都需要真实起点坐标。请回到地图获得定位后再试。" action="返回地图" onAction={() => navigation.navigate('MapTab')} /></Page>;

  return <Page>
    <AppHeader title="选择出行方式" onBack={() => navigation.goBack()} />
    <ScrollView style={styles.panel} contentContainerStyle={styles.panelContent}>
      <Text style={styles.destination}>{place ? `前往 ${place.name}` : '正在获取目的地信息'}</Text>
      {place?.address ? <Text style={styles.address}>{place.address}</Text> : null}
      <View style={styles.modes}>{modes.map((item) => <Chip key={item.id} label={`${item.symbol} ${item.label}`} selected={mode === item.id} onPress={() => setMode(item.id)} />)}</View>
      {loading ? <View style={styles.loading}><ActivityIndicator color={COLORS.primary} /><Text style={styles.loadingText}>正在计算路线…</Text></View> : null}
      {!loading && routeData ? <RouteSummary route={routeData} /> : null}
      {!loading && error ? <View style={styles.failure}><StatusPill label="路线暂不可用" tone="warning" /><Text style={styles.failureText}>{error}</Text></View> : null}
      <Text style={styles.notice}>路线摘要来自服务端；首版不提供语音、偏航重算或逐向导航。需要完整导航时，可继续使用外部地图。</Text>
      <PrimaryButton label="使用外部地图" onPress={openExternal} />
    </ScrollView>
  </Page>;
}

function RouteSummary({ route }: { route: Route }) {
  if (route.dataFreshness === 'external_only') {
    return <View style={styles.routeCard}><Text style={styles.eta}>应用内不提供该路线</Text><Text style={styles.routeMeta}>此方式仅支持跳转至外部服务，不显示虚构距离或预计时间。</Text><StatusPill label="仅外部导航" tone="primary" /></View>;
  }
  return <View style={styles.routeCard}>
    <View><Text style={styles.eta}>约 {durationLabel(route.durationSeconds)}</Text><Text style={styles.routeMeta}>{formatDistance(route.distanceMeters)} · {route.dataFreshness === 'live' ? '实时路线' : '缓存路线'}</Text></View>
    <StatusPill label="路线摘要" tone="primary" />
    {route.steps.slice(0, 3).map((step, index) => <Text key={`${step.instruction}-${index}`} style={styles.step}>{index + 1}. {step.instruction} · {formatDistance(step.distanceMeters)}</Text>)}
  </View>;
}

const styles = StyleSheet.create({
  panel: { flex: 1, backgroundColor: COLORS.white }, panelContent: { padding: SPACING.lg, paddingBottom: 32 }, destination: { ...FONTS.body, color: COLORS.text, fontWeight: '700' }, address: { ...FONTS.caption, color: COLORS.textSecondary, marginTop: 4 }, modes: { flexDirection: 'row', flexWrap: 'wrap', marginTop: 16, rowGap: 8 },
  loading: { alignItems: 'center', padding: 30 }, loadingText: { ...FONTS.caption, color: COLORS.textSecondary, marginTop: 10 },
  routeCard: { marginVertical: 16, padding: SPACING.lg, borderRadius: RADIUS.lg, backgroundColor: COLORS.primaryLight }, eta: { ...FONTS.titleMedium, color: COLORS.text }, routeMeta: { ...FONTS.caption, color: COLORS.textSecondary, marginTop: 4, marginBottom: 12 }, step: { ...FONTS.caption, color: COLORS.textSecondary, marginTop: 7, lineHeight: 19 },
  failure: { marginVertical: 16, padding: SPACING.lg, borderRadius: RADIUS.lg, backgroundColor: COLORS.warningLight }, failureText: { ...FONTS.caption, color: COLORS.textSecondary, marginTop: 8, lineHeight: 20 }, notice: { ...FONTS.caption, color: COLORS.textSecondary, lineHeight: 20, marginBottom: 16 },
});
