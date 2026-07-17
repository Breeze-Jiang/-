import React, { useEffect, useMemo, useState } from 'react';
import { ActivityIndicator, FlatList, ScrollView, StyleSheet, Text, View } from 'react-native';
import { AppHeader, Chip, EmptyState, Page, PlaceCard } from '../../shared/components/AppUI';
import { COLORS, FONTS, SPACING } from '../../shared/utils/theme';
import { p0Api, P0ApiError } from '../../p0/client';
import type { Coordinate, P0Place, P0PlaceCategory, PlacePage } from '../../p0/contracts';
import { presentP0Place } from '../../p0/presentation';

function userMessage(error: unknown) {
  if (error instanceof P0ApiError) return error.message;
  return '暂时无法加载地点，请检查网络后重试。';
}

export default function PlaceResultsScreen({ navigation, route }: any) {
  const [filter, setFilter] = useState<'distance' | 'confirmed'>('distance');
  const [page, setPage] = useState<PlacePage>();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>();
  const title = route.params?.label || '地点搜索';
  const category = route.params?.category as P0PlaceCategory | undefined;
  const center = route.params?.center as Coordinate | undefined;
  const query = route.params?.query as string | undefined;

  const load = async () => {
    setError(undefined);
    if (!category && !query) {
      setPage(undefined);
      return;
    }
    setLoading(true);
    try {
      const nextPage = center
        ? await p0Api.nearby({ center, category, radiusMeters: 5000, limit: 20 })
        : await p0Api.search({ keyword: query || title, category, limit: 20 });
      setPage(nextPage);
    } catch (cause) {
      setPage(undefined);
      setError(userMessage(cause));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { void load(); }, [category, center?.coordinateSystem, center?.latitude, center?.longitude, query]);

  const places = useMemo(() => {
    const items = page?.items || [];
    if (filter === 'distance') return items;
    return [...items].sort((left, right) => {
      const leftTime = left.trust.exists.confirmedAt ? Date.parse(left.trust.exists.confirmedAt) : 0;
      const rightTime = right.trust.exists.confirmedAt ? Date.parse(right.trust.exists.confirmedAt) : 0;
      return rightTime - leftTime;
    });
  }, [filter, page?.items]);

  const openPlace = (place: P0Place) => navigation.navigate('PlaceDetail', { placeId: place.id, origin: center });

  return <Page>
    <AppHeader title={title} subtitle={center ? '按距离排序' : '搜索结果'} onBack={() => navigation.goBack()} />
    <View style={styles.filters}>
      <ScrollView horizontal showsHorizontalScrollIndicator={false}>
        <Chip label="距离" selected={filter === 'distance'} onPress={() => setFilter('distance')} />
        <Chip label="最近确认" selected={filter === 'confirmed'} onPress={() => setFilter('confirmed')} />
      </ScrollView>
    </View>
    {loading ? <View style={styles.center}><ActivityIndicator color={COLORS.primary} /><Text style={styles.status}>正在查询地点…</Text></View> : null}
    {!loading && error ? <EmptyState title="地点加载失败" description={error} action="重试" onAction={() => void load()} /> : null}
    {!loading && !error && !category && !query ? <EmptyState title="请选择一项服务" description="为了避免把不同类别的地点混在一起，请从地图中的服务入口选择要查找的类别。" action="返回地图" onAction={() => navigation.goBack()} /> : null}
    {!loading && !error && (category || query) && !places.length ? <EmptyState title={page?.coverage === 'not_covered' ? '该服务暂未覆盖' : '暂未找到匹配地点'} description={page?.coverage === 'not_covered' ? '此类别尚未经过供应商类型核验，请选择其他已覆盖服务或稍后再试。' : '可修改关键词、选择其他服务类别，或在获得定位后查看附近地点。'} action="返回地图" onAction={() => navigation.goBack()} /> : null}
    {!loading && !error && places.length ? <FlatList data={places} keyExtractor={(item) => item.id} contentContainerStyle={styles.list} ListHeaderComponent={page ? <Text style={styles.freshness}>{page.dataFreshness === 'live' ? '实时查询结果' : page.dataFreshness === 'cached' ? '缓存结果' : '降级结果；可能不是实时数据'}</Text> : null} renderItem={({ item }) => <PlaceCard place={presentP0Place(item)} onPress={() => openPlace(item)} onNavigate={() => navigation.navigate('Navigation', { placeId: item.id, origin: center })} />} /> : null}
  </Page>;
}

const styles = StyleSheet.create({
  filters: { backgroundColor: COLORS.white, paddingVertical: 12, paddingLeft: SPACING.lg, borderBottomWidth: StyleSheet.hairlineWidth, borderColor: COLORS.border },
  list: { padding: SPACING.lg, paddingBottom: 40 },
  center: { alignItems: 'center', padding: 42 },
  status: { ...FONTS.caption, color: COLORS.textSecondary, marginTop: 10 },
  freshness: { ...FONTS.caption, color: COLORS.textTertiary, marginBottom: 10 },
});
