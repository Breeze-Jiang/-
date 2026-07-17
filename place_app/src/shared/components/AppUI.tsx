import React from 'react';
import {
  Pressable, ScrollView, StyleProp, StyleSheet, Text, TextStyle, View, ViewStyle,
} from 'react-native';
import type { Place } from '../../app/types';
import { COLORS, FONTS, RADIUS, SHADOWS, SPACING } from '../utils/theme';

type IconProps = { symbol: string; label: string; color?: string; size?: number };
export function Glyph({ symbol, label, color = COLORS.text, size = 20 }: IconProps) {
  return <Text accessibilityLabel={label} style={{ color, fontSize: size, fontWeight: '600', lineHeight: size + 4 }}>{symbol}</Text>;
}

export function AppHeader({ title, subtitle, onBack, right, compact }: { title: string; subtitle?: string; onBack?: () => void; right?: React.ReactNode; compact?: boolean }) {
  return (
    <View style={[styles.header, compact && styles.headerCompact]}>
      <View style={styles.headerSide}>
        {onBack ? <Pressable accessibilityRole="button" accessibilityLabel="返回" hitSlop={10} onPress={onBack} style={styles.headerButton}><Glyph symbol="‹" label="返回" size={32} /></Pressable> : null}
      </View>
      <View style={styles.headerCenter}>
        <Text numberOfLines={1} style={styles.headerTitle}>{title}</Text>
        {subtitle ? <Text numberOfLines={1} style={styles.headerSubtitle}>{subtitle}</Text> : null}
      </View>
      <View style={[styles.headerSide, styles.headerRight]}>{right}</View>
    </View>
  );
}

export function PrimaryButton({ label, onPress, variant = 'primary', disabled, style }: { label: string; onPress?: () => void; variant?: 'primary' | 'secondary' | 'quiet' | 'danger'; disabled?: boolean; style?: StyleProp<ViewStyle> }) {
  const tone = variant === 'primary' ? styles.primary : variant === 'danger' ? styles.danger : variant === 'secondary' ? styles.secondary : styles.quiet;
  const textTone = variant === 'primary' ? styles.primaryText : variant === 'danger' ? styles.dangerText : styles.secondaryText;
  return (
    <Pressable accessibilityRole="button" accessibilityLabel={label} disabled={disabled} onPress={onPress} style={({ pressed }) => [styles.button, tone, disabled && styles.disabled, pressed && !disabled && styles.pressed, style]}>
      <Text style={[styles.buttonText, textTone]}>{label}</Text>
    </Pressable>
  );
}

export function StatusPill({ label, tone = 'neutral' }: { label: string; tone?: 'success' | 'warning' | 'primary' | 'neutral' | 'danger' }) {
  const map = { success: [styles.successPill, styles.successText], warning: [styles.warningPill, styles.warningText], primary: [styles.primaryPill, styles.primaryPillText], danger: [styles.dangerPill, styles.dangerPillText], neutral: [styles.neutralPill, styles.neutralText] } as const;
  return <View style={[styles.pill, map[tone][0]]}><Text style={[styles.pillText, map[tone][1]]}>{label}</Text></View>;
}

export function Segment({ options, value, onChange }: { options: string[]; value: string; onChange: (value: string) => void }) {
  return <View style={styles.segment}>{options.map((item) => <Pressable key={item} accessibilityRole="tab" accessibilityState={{ selected: value === item }} onPress={() => onChange(item)} style={[styles.segmentItem, value === item && styles.segmentActive]}><Text style={[styles.segmentText, value === item && styles.segmentTextActive]}>{item}</Text></Pressable>)}</View>;
}

export function Chip({ label, selected, onPress }: { label: string; selected?: boolean; onPress?: () => void }) {
  return <Pressable accessibilityRole="button" accessibilityLabel={label} onPress={onPress} style={({ pressed }) => [styles.chip, selected && styles.chipSelected, pressed && styles.pressed]}><Text style={[styles.chipText, selected && styles.chipTextSelected]}>{label}</Text></Pressable>;
}

export function SectionTitle({ title, action, onAction }: { title: string; action?: string; onAction?: () => void }) {
  return <View style={styles.sectionHeader}><Text style={styles.sectionTitle}>{title}</Text>{action ? <Pressable accessibilityRole="button" accessibilityLabel={action} onPress={onAction}><Text style={styles.sectionAction}>{action} ›</Text></Pressable> : null}</View>;
}

export function EmptyState({ title, description, action, onAction }: { title: string; description: string; action?: string; onAction?: () => void }) {
  return <View style={styles.empty}><View style={styles.emptyMark}><Glyph symbol="◇" label="空状态" size={30} color={COLORS.primary} /></View><Text style={styles.emptyTitle}>{title}</Text><Text style={styles.emptyDescription}>{description}</Text>{action ? <PrimaryButton label={action} onPress={onAction} variant="secondary" style={styles.emptyButton} /> : null}</View>;
}

export function Page({ children, style }: { children: React.ReactNode; style?: StyleProp<ViewStyle> }) {
  return <View style={[styles.page, style]}>{children}</View>;
}

export function PageScroll({ children, style }: { children: React.ReactNode; style?: StyleProp<ViewStyle> }) {
  return <ScrollView contentContainerStyle={[styles.scrollContent, style]} showsVerticalScrollIndicator={false}>{children}</ScrollView>;
}

export function PlaceCard({ place, onPress, onNavigate }: { place: Place; onPress?: () => void; onNavigate?: () => void }) {
  const trustTone = place.existence === '地点存在已确认' || place.existence === '已收录' ? 'success' : 'warning';
  return <Pressable accessibilityRole="button" accessibilityLabel={`查看${place.name}`} onPress={onPress} style={({ pressed }) => [styles.placeCard, pressed && styles.pressed]}>
    <View style={styles.placeTop}><View style={styles.placeIcon}><Glyph symbol="⌖" label="地点" color={COLORS.primary} /></View><View style={styles.placeMain}><Text numberOfLines={1} style={styles.placeName}>{place.name}</Text><Text numberOfLines={1} style={styles.placeAddress}>{place.distance} · {place.address}</Text></View><StatusPill label={place.existence} tone={trustTone} /></View>
    <View style={styles.placeMeta}><Text style={styles.placeMetaText}>营业：{place.hours}</Text><Text style={styles.placeMetaText}>{place.confirmedAt}</Text></View>
    <View style={styles.placeBottom}>{place.tags.map((tag) => <StatusPill key={tag} label={tag} />)}<View style={{ flex: 1 }} />{onNavigate ? <Pressable accessibilityRole="button" accessibilityLabel={`导航至${place.name}`} onPress={(event) => { event.stopPropagation(); onNavigate(); }} style={styles.inlineRoute}><Text style={styles.inlineRouteText}>导航</Text><Glyph symbol="›" label="导航" color={COLORS.primary} /></Pressable> : null}</View>
  </Pressable>;
}

const styles = StyleSheet.create({
  page: { flex: 1, backgroundColor: COLORS.background }, scrollContent: { padding: SPACING.lg, paddingBottom: 40 },
  header: { minHeight: 62, backgroundColor: COLORS.white, borderBottomWidth: StyleSheet.hairlineWidth, borderBottomColor: COLORS.border, flexDirection: 'row', alignItems: 'center', paddingHorizontal: 12 }, headerCompact: { minHeight: 52 }, headerSide: { width: 54, minHeight: 44, justifyContent: 'center' }, headerRight: { alignItems: 'flex-end' }, headerCenter: { flex: 1, alignItems: 'center' }, headerButton: { width: 44, height: 44, alignItems: 'flex-start', justifyContent: 'center' }, headerTitle: { ...FONTS.titleMedium, color: COLORS.text }, headerSubtitle: { ...FONTS.caption, color: COLORS.textTertiary, marginTop: 1 },
  button: { minHeight: 48, borderRadius: RADIUS.md, alignItems: 'center', justifyContent: 'center', paddingHorizontal: SPACING.lg }, buttonText: { ...FONTS.body, fontWeight: '700' }, primary: { backgroundColor: COLORS.primary, ...SHADOWS.button }, primaryText: { color: COLORS.white }, secondary: { backgroundColor: COLORS.white, borderWidth: 1, borderColor: COLORS.border }, secondaryText: { color: COLORS.primary }, quiet: { backgroundColor: 'transparent' }, danger: { backgroundColor: COLORS.dangerLight }, dangerText: { color: COLORS.danger }, disabled: { opacity: 0.45 }, pressed: { opacity: 0.78 },
  pill: { paddingHorizontal: 8, minHeight: 24, borderRadius: RADIUS.full, justifyContent: 'center' }, pillText: { ...FONTS.caption, fontSize: 11, fontWeight: '600' }, successPill: { backgroundColor: COLORS.successLight }, successText: { color: COLORS.success }, warningPill: { backgroundColor: COLORS.warningLight }, warningText: { color: COLORS.warning }, primaryPill: { backgroundColor: COLORS.primaryLight }, primaryPillText: { color: COLORS.primary }, neutralPill: { backgroundColor: '#EEF2F3' }, neutralText: { color: COLORS.textSecondary }, dangerPill: { backgroundColor: COLORS.dangerLight }, dangerPillText: { color: COLORS.danger },
  segment: { flexDirection: 'row', backgroundColor: '#EDF2F3', padding: 3, borderRadius: RADIUS.md }, segmentItem: { flex: 1, minHeight: 40, alignItems: 'center', justifyContent: 'center', borderRadius: 9 }, segmentActive: { backgroundColor: COLORS.white, ...SHADOWS.card }, segmentText: { ...FONTS.caption, color: COLORS.textSecondary, fontWeight: '600' }, segmentTextActive: { color: COLORS.text },
  chip: { minHeight: 34, justifyContent: 'center', paddingHorizontal: 12, borderRadius: RADIUS.full, backgroundColor: COLORS.white, borderWidth: 1, borderColor: COLORS.border, marginRight: 8 }, chipSelected: { backgroundColor: COLORS.primaryLight, borderColor: COLORS.primary }, chipText: { ...FONTS.caption, color: COLORS.textSecondary, fontWeight: '600' }, chipTextSelected: { color: COLORS.primary },
  sectionHeader: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', marginTop: 24, marginBottom: 12 }, sectionTitle: { ...FONTS.titleMedium, color: COLORS.text }, sectionAction: { ...FONTS.caption, color: COLORS.primary, fontWeight: '700' },
  empty: { alignItems: 'center', paddingHorizontal: 30, paddingVertical: 48 }, emptyMark: { width: 68, height: 68, borderRadius: 34, backgroundColor: COLORS.primaryLight, alignItems: 'center', justifyContent: 'center', marginBottom: 18 }, emptyTitle: { ...FONTS.titleMedium, color: COLORS.text }, emptyDescription: { ...FONTS.body, color: COLORS.textSecondary, textAlign: 'center', lineHeight: 24, marginTop: 8 }, emptyButton: { marginTop: 22, alignSelf: 'stretch' },
  placeCard: { backgroundColor: COLORS.white, borderRadius: RADIUS.lg, padding: 14, marginBottom: 12, borderWidth: 1, borderColor: COLORS.border }, placeTop: { flexDirection: 'row', alignItems: 'center' }, placeIcon: { width: 40, height: 40, borderRadius: 12, backgroundColor: COLORS.primaryLight, alignItems: 'center', justifyContent: 'center', marginRight: 10 }, placeMain: { flex: 1, paddingRight: 8 }, placeName: { ...FONTS.body, color: COLORS.text, fontWeight: '700' }, placeAddress: { ...FONTS.caption, color: COLORS.textSecondary, marginTop: 3 }, placeMeta: { marginTop: 12, paddingTop: 10, borderTopWidth: StyleSheet.hairlineWidth, borderTopColor: COLORS.border }, placeMetaText: { ...FONTS.caption, color: COLORS.textSecondary, marginBottom: 3 }, placeBottom: { flexDirection: 'row', alignItems: 'center', marginTop: 10 }, inlineRoute: { minHeight: 36, paddingLeft: 8, flexDirection: 'row', alignItems: 'center' }, inlineRouteText: { ...FONTS.caption, color: COLORS.primary, fontWeight: '700' },
});

export const uiStyles = styles;
