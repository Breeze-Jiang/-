import React, { useEffect, useRef, useState } from 'react';
import { ActivityIndicator, StyleSheet, Text, View } from 'react-native';
import { AppHeader, Chip, EmptyState, Page, PageScroll, PrimaryButton, SectionTitle, StatusPill } from '../../shared/components/AppUI';
import { COLORS, FONTS, RADIUS, SPACING } from '../../shared/utils/theme';
import { p0Api, P0ApiError } from '../../p0/client';
import type { Coordinate, P0Place } from '../../p0/contracts';
import { presentP0Place } from '../../p0/presentation';
import { newIdempotencyKey, useP0SessionStore } from '../../p0/session';

function errorMessage(error: unknown) {
  return error instanceof P0ApiError ? error.message : '暂时无法加载地点详情，请稍后重试。';
}

export default function PlaceDetailScreen({ navigation, route }: any) {
  const placeId = route.params?.placeId as string | undefined;
  const origin = route.params?.origin as Coordinate | undefined;
  const [place, setPlace] = useState<P0Place>();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>();
  const [confirming, setConfirming] = useState(false);
  const [confirmationError, setConfirmationError] = useState<string>();
  const confirmationKey = useRef(newIdempotencyKey());
  const sessionStatus = useP0SessionStore((state) => state.status);
  const accessTokenForWrite = useP0SessionStore((state) => state.accessTokenForWrite);
  const clearLocal = useP0SessionStore((state) => state.clearLocal);

  const load = async () => {
    if (!placeId) {
      setError('缺少地点标识，无法加载详情。');
      setLoading(false);
      return;
    }
    setLoading(true);
    setError(undefined);
    try {
      setPlace(await p0Api.place(placeId));
    } catch (cause) {
      setError(errorMessage(cause));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { void load(); }, [placeId]);

  const confirmExistence = async () => {
    if (!place) return;
    if (sessionStatus !== 'authenticated') {
      navigation.navigate('LoginGate', { action: '确认地点存在', resumeRoute: 'PlaceDetail', resumeParams: { placeId, origin } });
      return;
    }
    setConfirming(true);
    setConfirmationError(undefined);
    try {
      const token = await accessTokenForWrite();
      await p0Api.confirm(place.id, token, confirmationKey.current, 'exists', 'confirmed');
      await load();
    } catch (cause) {
      if (cause instanceof P0ApiError && cause.status === 401) {
        await clearLocal();
        navigation.navigate('LoginGate', { action: '确认地点存在', resumeRoute: 'PlaceDetail', resumeParams: { placeId, origin } });
        return;
      }
      setConfirmationError(errorMessage(cause));
    } finally {
      setConfirming(false);
    }
  };

  if (loading) return <Page><AppHeader title="地点详情" onBack={() => navigation.goBack()} /><View style={styles.center}><ActivityIndicator color={COLORS.primary} /><Text style={styles.loading}>正在加载可信地点信息…</Text></View></Page>;
  if (error || !place) return <Page><AppHeader title="地点详情" onBack={() => navigation.goBack()} /><EmptyState title="地点详情不可用" description={error || '暂时没有可展示的信息。'} action="重试" onAction={() => void load()} /></Page>;

  const view = presentP0Place(place);
  const existenceTone = place.trust.exists.status === 'confirmed' ? 'success' : place.trust.exists.status === 'incorrect' ? 'warning' : 'neutral';
  return <Page>
    <AppHeader title="地点详情" onBack={() => navigation.goBack()} />
    <PageScroll>
      <View style={styles.hero}>
        <View style={styles.placeMark}><Text style={styles.placeMarkText}>⌖</Text></View>
        <Text style={styles.name}>{place.name}</Text>
        <Text style={styles.address}>{place.address || '地址待补充'}</Text>
        <View style={styles.heroMeta}><StatusPill label={view.existence} tone={existenceTone} />{place.distanceMeters !== undefined ? <Text style={styles.distance}>距你 {view.distance}</Text> : null}</View>
      </View>
      <SectionTitle title="可信地点信息" />
      <View style={styles.facts}>
        <Fact label="地点存在" value={view.existence} />
        <Fact label="营业时间" value="未知/待确认" />
        <Fact label="最近用户确认" value={view.confirmedAt} />
        <Fact label="信息说明" value="P0 仅展示地点存在、营业时间和用户最近确认；不把缺失信息伪装成实时状态。" muted />
      </View>
      <SectionTitle title="地点分类" />
      <View style={styles.tags}><Chip label={place.category === 'unclassified' ? '未分类' : place.category} /></View>
      {confirmationError ? <Text accessibilityRole="alert" style={styles.confirmationError}>{confirmationError}</Text> : null}
    </PageScroll>
    <View style={styles.bottom}>
      <PrimaryButton label={confirming ? '正在确认…' : '确认地点存在'} variant="secondary" disabled={confirming} onPress={() => void confirmExistence()} style={styles.confirm} />
      {origin ? <PrimaryButton label="查看路线" onPress={() => navigation.navigate('Navigation', { placeId: place.id, origin })} style={styles.route} /> : <PrimaryButton label="定位后查看路线" onPress={() => navigation.navigate('MapTab')} style={styles.route} />}
    </View>
  </Page>;
}

function Fact({ label, value, muted }: { label: string; value: string; muted?: boolean }) {
  return <View style={styles.fact}><Text style={styles.factLabel}>{label}</Text><Text style={[styles.factValue, muted && styles.factMuted]}>{value}</Text></View>;
}

const styles = StyleSheet.create({
  center: { alignItems: 'center', padding: 44 }, loading: { ...FONTS.caption, color: COLORS.textSecondary, marginTop: 10 },
  hero: { backgroundColor: COLORS.white, margin: -SPACING.lg, marginBottom: 0, padding: SPACING.xl, borderBottomWidth: StyleSheet.hairlineWidth, borderColor: COLORS.border },
  placeMark: { width: 48, height: 48, borderRadius: 16, backgroundColor: COLORS.primary, justifyContent: 'center', alignItems: 'center', marginBottom: 14 }, placeMarkText: { fontSize: 28, color: COLORS.white },
  name: { ...FONTS.titleLarge, color: COLORS.text }, address: { ...FONTS.body, color: COLORS.textSecondary, marginTop: 8, lineHeight: 23 }, heroMeta: { flexDirection: 'row', alignItems: 'center', marginTop: 14 }, distance: { ...FONTS.caption, color: COLORS.textTertiary, marginLeft: 10 },
  facts: { backgroundColor: COLORS.white, borderRadius: RADIUS.lg, borderWidth: 1, borderColor: COLORS.border, paddingHorizontal: SPACING.lg }, fact: { paddingVertical: 13, borderBottomWidth: StyleSheet.hairlineWidth, borderColor: COLORS.border }, factLabel: { ...FONTS.caption, color: COLORS.textTertiary }, factValue: { ...FONTS.body, color: COLORS.text, marginTop: 4 }, factMuted: { color: COLORS.textSecondary },
  tags: { flexDirection: 'row', flexWrap: 'wrap' }, confirmationError: { ...FONTS.caption, color: COLORS.danger, marginTop: 16 }, bottom: { flexDirection: 'row', backgroundColor: COLORS.white, padding: SPACING.md, borderTopWidth: StyleSheet.hairlineWidth, borderColor: COLORS.border }, confirm: { flex: 1, marginRight: 10 }, route: { flex: 1.2 },
});
