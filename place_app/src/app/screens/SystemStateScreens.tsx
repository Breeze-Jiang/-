import React, { useState } from 'react';
import { Linking, StyleSheet, Text, TextInput, View } from 'react-native';
import { AppHeader, EmptyState, Page } from '../../shared/components/AppUI';
import { useLocationPermissionStore } from '../../store/useLocationPermissionStore';
import { COLORS, FONTS, RADIUS } from '../../shared/utils/theme';

export function LocationPermissionScreen({ navigation }: any) {
  const [manualPlace, setManualPlace] = useState('广州市天河区');
  const useManualLocation = () => navigation.navigate('PlaceResults', { label: manualPlace.trim() || '输入地点', manualLocation: true });
  return <Page><AppHeader title="需要位置权限" onBack={() => navigation.goBack()} /><EmptyState title="允许位置后可查找附近地点" description="位置仅用于附近搜索和导航。你仍可搜索城市、景点或输入地点继续使用。" /><View style={styles.manual}><Text style={styles.manualTitle}>手动输入地点</Text><TextInput value={manualPlace} onChangeText={setManualPlace} style={styles.input} placeholder="城市、景点或地址" placeholderTextColor={COLORS.textTertiary} accessibilityLabel="手动输入地点" /><Text onPress={useManualLocation} accessibilityRole="button" style={styles.link}>按此地点搜索 ›</Text></View><EmptyState title="已在系统中拒绝权限？" description="可前往系统设置重新开启位置权限。" action="打开系统设置" onAction={() => Linking.openSettings()} /></Page>;
}

export function NetworkErrorScreen({ navigation }: any) {
  return <Page><AppHeader title="暂时无法加载" onBack={() => navigation.goBack()} /><EmptyState title="网络连接不可用" description="请检查网络后重试；已编辑的草稿会在重新联网后继续保存。" action="重试" onAction={() => navigation.goBack()} /></Page>;
}

export function ContentUnavailableScreen({ navigation }: any) {
  return <Page><AppHeader title="内容不可用" onBack={() => navigation.goBack()} /><EmptyState title="这条内容已不可查看" description="可能已被作者撤回、删除、审核未通过，或你没有访问权限。" action="返回发现" onAction={() => navigation.navigate('DiscoverTab')} /></Page>;
}

const styles = StyleSheet.create({ manual: { margin: 16, padding: 16, backgroundColor: COLORS.white, borderWidth: 1, borderColor: COLORS.border, borderRadius: RADIUS.lg }, manualTitle: { ...FONTS.body, color: COLORS.text, fontWeight: '700' }, input: { minHeight: 48, marginTop: 10, paddingHorizontal: 12, borderRadius: RADIUS.md, backgroundColor: COLORS.background, color: COLORS.text, ...FONTS.body }, link: { ...FONTS.caption, color: COLORS.primary, fontWeight: '700', marginTop: 12 } });
