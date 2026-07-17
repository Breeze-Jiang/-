import React, { useEffect, useState } from 'react';
import { ActivityIndicator, Alert, Linking, StyleSheet, Text, View } from 'react-native';
import { AppHeader, EmptyState, Page, PageScroll, PrimaryButton, StatusPill } from '../../shared/components/AppUI';
import { COLORS, FONTS, RADIUS } from '../../shared/utils/theme';
import { p0Api, P0ApiError } from '../../p0/client';
import type { Coordinate, ExternalLink, NavigationMode, P0Place } from '../../p0/contracts';

const providerName: Record<ExternalLink['provider'], string> = { amap: '高德地图', baidu: '百度地图', tencent: '腾讯地图', ride_hailing: '第三方叫车服务' };

function modeLabel(mode: NavigationMode) {
  return { walking: '步行', cycling: '骑行', driving: '驾车', transit: '公交地铁', taxi: '打车' }[mode];
}

function errorMessage(error: unknown) {
  return error instanceof P0ApiError ? error.message : '暂时无法获取外部导航入口，请稍后重试。';
}

export default function ExternalNavigationScreen({ navigation, route }: any) {
  const [place, setPlace] = useState<P0Place>();
  const [links, setLinks] = useState<ExternalLink[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>();
  const [opening, setOpening] = useState<string>();
  const placeId = route.params?.placeId as string | undefined;
  const origin = route.params?.origin as Coordinate | undefined;
  const mode = (route.params?.mode || 'driving') as NavigationMode;
  const kind = route.params?.type === 'taxi' ? '第三方叫车服务' : '其他地图';

  const load = async () => {
    if (!placeId || !origin) {
      setError('需要目的地和当前位置后才能生成外部导航入口。');
      setLoading(false);
      return;
    }
    setLoading(true);
    setError(undefined);
    try {
      const destination = await p0Api.place(placeId);
      const external = await p0Api.externalLinks(origin, destination.location, destination.name, mode);
      setPlace(destination);
      setLinks(external.links);
    } catch (cause) {
      setError(errorMessage(cause));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { void load(); }, [mode, origin?.coordinateSystem, origin?.latitude, origin?.longitude, placeId]);

  const open = async (link: ExternalLink) => {
    if (!link.available || !link.uri) return;
    setOpening(link.provider);
    try {
      await Linking.openURL(link.uri);
    } catch {
      Alert.alert('无法打开地图服务', '请检查对应应用是否已安装，或稍后重试。');
    } finally {
      setOpening(undefined);
    }
  };

  if (loading) return <Page><AppHeader title={kind} onBack={() => navigation.goBack()} /><View style={styles.loading}><ActivityIndicator color={COLORS.primary} /><Text style={styles.loadingText}>正在获取外部导航入口…</Text></View></Page>;
  if (error) return <Page><AppHeader title={kind} onBack={() => navigation.goBack()} /><EmptyState title="外部导航不可用" description={error} action="重试" onAction={() => void load()} /></Page>;

  return <Page><AppHeader title={kind} onBack={() => navigation.goBack()} /><PageScroll>
    <View style={styles.hero}><Text style={styles.title}>选择要打开的地图</Text><Text style={styles.description}>链接由服务端按坐标系和出行方式生成；客户端不自行拼接供应商 URI。</Text></View>
    <View style={styles.destination}><Text style={styles.destinationLabel}>目的地</Text><Text style={styles.destinationName}>{place?.name}</Text><Text style={styles.destinationText}>方式：{modeLabel(mode)}</Text></View>
    {links.map((link) => <View key={link.provider} style={styles.provider}><View style={{ flex: 1 }}><Text style={styles.providerName}>{providerName[link.provider]}</Text><Text style={styles.providerText}>{link.available && link.uri ? '可打开地图或网页导航。' : '当前不可用，请选择其他入口。'}</Text></View><PrimaryButton label={opening === link.provider ? '正在打开…' : link.available && link.uri ? '打开' : '不可用'} disabled={opening !== undefined || !link.available || !link.uri} onPress={() => void open(link)} style={styles.open} /></View>)}
    <View style={styles.notice}><StatusPill label="第三方服务" tone="neutral" /><Text style={styles.noticeText}>沿途不会创建订单、获取支付信息或伪造实时路线。返回应用后可以继续查看地点。</Text></View>
  </PageScroll></Page>;
}

const styles = StyleSheet.create({
  loading: { alignItems: 'center', padding: 44 }, loadingText: { ...FONTS.caption, color: COLORS.textSecondary, marginTop: 10 }, hero: { alignItems: 'center', backgroundColor: COLORS.primaryLight, borderRadius: RADIUS.lg, padding: 24 }, title: { ...FONTS.titleMedium, color: COLORS.text }, description: { ...FONTS.body, color: COLORS.textSecondary, lineHeight: 23, textAlign: 'center', marginTop: 7 },
  destination: { marginTop: 16, backgroundColor: COLORS.white, borderWidth: 1, borderColor: COLORS.border, borderRadius: RADIUS.lg, padding: 16 }, destinationLabel: { ...FONTS.caption, color: COLORS.textTertiary }, destinationName: { ...FONTS.body, color: COLORS.text, fontWeight: '700', marginTop: 5 }, destinationText: { ...FONTS.caption, color: COLORS.textSecondary, marginTop: 4 }, provider: { flexDirection: 'row', alignItems: 'center', backgroundColor: COLORS.white, borderWidth: 1, borderColor: COLORS.border, borderRadius: RADIUS.lg, padding: 16, marginTop: 12 }, providerName: { ...FONTS.body, color: COLORS.text, fontWeight: '700' }, providerText: { ...FONTS.caption, color: COLORS.textSecondary, marginTop: 4 }, open: { minWidth: 74, marginLeft: 12 }, notice: { marginTop: 16, padding: 14, borderRadius: RADIUS.md, backgroundColor: COLORS.background }, noticeText: { ...FONTS.caption, color: COLORS.textSecondary, lineHeight: 20, marginTop: 8 },
});
