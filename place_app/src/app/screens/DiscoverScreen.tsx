import React, { useMemo, useState } from 'react';
import { Pressable, ScrollView, StyleSheet, Text, TextInput, View } from 'react-native';
import { communityPosts } from '../demoData';
import type { CommunityPost } from '../types';
import { AppHeader, Page, PageScroll, PrimaryButton, SectionTitle, StatusPill } from '../../shared/components/AppUI';
import { COLORS, FONTS, RADIUS, SPACING } from '../../shared/utils/theme';
import { AppIcon } from '../../ui/AppIcon';
import { TrustBadge } from '../../ui/TrustBadge';
import { SafeAreaView } from 'react-native-safe-area-context';

const channels = ['热点旅游资源', '热门标注分享', '旅游计划探讨'];

export default function DiscoverScreen({ navigation }: any) {
  const [top, setTop] = useState('推荐');
  const [channel, setChannel] = useState(channels[0]);
  const posts = useMemo(() => {
    if (top === '推荐') return communityPosts;
    if (channel === '热门标注分享') return communityPosts.filter((post) => post.kind === 'marker');
    if (channel === '旅游计划探讨') return communityPosts.filter((post) => post.kind === 'plan' || post.kind === 'discussion');
    return communityPosts.filter((post) => post.kind === 'resource' || post.kind === 'discussion');
  }, [channel, top]);

  return <Page><SafeAreaView edges={['top']} style={styles.safeArea}>
    <View style={styles.topArea}>
      <View style={styles.searchRow}>
        <View style={styles.brandMark}><Text style={styles.brandText}>沿途</Text></View>
        <Pressable accessibilityRole="button" accessibilityLabel="搜索景点、旅行问题或资源" onPress={() => navigation.navigate('CommunitySearch')} style={styles.searchBox}>
          <AppIcon name="search" color={COLORS.primary} size={18} label="搜索" />
          <Text style={styles.searchPlaceholder}>搜索景点、旅行问题或资源</Text>
        </Pressable>
        {top === '公开' ? <Pressable accessibilityRole="button" accessibilityLabel="发布旅行内容" onPress={() => navigation.navigate('CreatePost')}><Text style={styles.publish}>＋发布</Text></Pressable> : null}
      </View>
      <View style={styles.rootTabs}>
        {['推荐', '公开'].map((item) => <Pressable key={item} accessibilityRole="tab" accessibilityState={{ selected: top === item }} onPress={() => setTop(item)} style={[styles.rootTab, top === item && styles.rootTabActive]}><Text style={[styles.rootText, top === item && styles.rootTextActive]}>{item}</Text></Pressable>)}
      </View>
    </View>
    {top === '公开' ? <ScrollView horizontal style={styles.channelScroller} showsHorizontalScrollIndicator={false} contentContainerStyle={styles.channelBar}>{channels.map((item) => <Pressable key={item} accessibilityRole="tab" accessibilityState={{ selected: channel === item }} onPress={() => setChannel(item)} style={styles.channel}><Text style={[styles.channelText, channel === item && styles.channelTextActive]}>{item}</Text></Pressable>)}</ScrollView> : null}
    <ScrollView style={styles.feed} contentContainerStyle={styles.feedContent} showsVerticalScrollIndicator={false}>
      <TravelNotice top={top} />
      {posts.map((post) => <PostCard key={post.id} post={post} onPress={() => navigation.navigate('PostDetail', { postId: post.id })} />)}
    </ScrollView>
  </SafeAreaView></Page>;
}

function TravelNotice({ top }: { top: string }) {
  return <View style={styles.notice}><AppIcon name="notice" color={COLORS.warning} size={16} label="旅行提示" /><View style={{ flex: 1 }}><Text style={styles.noticeTitle}>{top === '推荐' ? '出发前看看 · 广州周末自驾' : '广州 · 自驾周末旅行资源'}</Text><Text style={styles.noticeText}>{top === '推荐' ? '因为你最近搜索过停车和江边散步，可随时调整推荐范围。' : '停车、补给与雨天备选资源正在持续讨论。'}</Text></View><AppIcon name="forward" color={COLORS.textTertiary} size={17} label="调整筛选" /></View>;
}

export function PostCard({ post, onPress }: { post: CommunityPost; onPress?: () => void }) {
  const type = post.kind === 'marker' ? '公开标注' : post.kind === 'plan' ? '公开计划' : post.kind === 'resource' ? '旅游资源' : '旅行问答';
  const action = post.kind === 'plan' ? '复制计划' : post.kind === 'discussion' ? '阅读回答' : '查看地点并导航';
  return <Pressable accessibilityRole="button" accessibilityLabel={`查看${post.title}`} onPress={onPress} style={({ pressed }) => [styles.post, pressed && styles.pressed]}>
    <View style={styles.kickerRow}><Text style={[styles.kicker, post.kind === 'marker' && styles.markerKicker]}>{type}</Text>{post.kind === 'marker' ? <TrustBadge label="已审核" /> : null}</View>
    <Text style={styles.postTitle}>{post.title}</Text>
    <View style={styles.author}><View style={[styles.avatar, { backgroundColor: post.imageTone }]}><Text style={styles.avatarText}>{post.author.slice(0, 1)}</Text></View><Text style={styles.authorText}>{post.author} · {post.city} · 2 小时前</Text></View>
    <Text numberOfLines={3} style={styles.postSummary}>{post.summary}</Text>
    {post.place ? <Pressable accessibilityRole="button" accessibilityLabel={`${action}：${post.place}`} onPress={(event) => { event.stopPropagation(); onPress?.(); }} style={styles.placeAction}><AppIcon name="pin" color={COLORS.success} size={16} label="关联地点" /><Text numberOfLines={1} style={styles.placeText}>{post.place}</Text><Text style={styles.placeGo}>{action} ›</Text></Pressable> : null}
    <View style={styles.metrics}><Text style={styles.metric}>△ {post.likes} 有用</Text><Text style={styles.metric}>☆ {post.saved} 收藏</Text><Text style={styles.metric}>◌ {post.comments} 评论</Text></View>
  </Pressable>;
}

export function PostDetailScreen({ navigation, route }: any) {
  const post = communityPosts.find((item) => item.id === route.params?.postId) || communityPosts[0];
  return <Page><AppHeader title="旅行讨论" onBack={() => navigation.goBack()} /><PageScroll><PostCard post={post} /><SectionTitle title="旅行者回答" /><Answer author="阿帆自驾" text="停车后沿江边走会更顺，不用回到主路。傍晚人多时建议预留半小时找位。" place={post.place} /><Answer author="江边散步的人" text="如果下雨，附近商场的充电宝和便利店可以先补给，等雨小一些再继续走。" /></PageScroll></Page>;
}

function Answer({ author, text, place }: { author: string; text: string; place?: string }) {
  return <View style={styles.answer}><Text style={styles.answerAuthor}>{author} · 旅行 46 城</Text><Text style={styles.answerText}>{text}</Text>{place ? <View style={styles.answerPlace}><AppIcon name="pin" color={COLORS.success} size={15} label="地点" /><Text style={styles.answerPlaceText}>{place} · 查看地点并导航</Text></View> : null}<View style={styles.answerMetrics}><Text>△ 286 有用</Text><Text>☆ 收藏</Text><Text>◌ 回复</Text></View></View>;
}

export function CommunitySearchScreen({ navigation }: any) {
  const [query, setQuery] = useState('');
  return <Page><AppHeader title="搜索旅行资源" onBack={() => navigation.goBack()} /><PageScroll><View style={styles.searchInput}><AppIcon name="search" color={COLORS.textTertiary} label="搜索" /><TextInput value={query} onChangeText={setQuery} placeholder="景点、城市、标签或作者" placeholderTextColor={COLORS.textTertiary} style={styles.input} accessibilityLabel="搜索旅行资源" /></View><SectionTitle title={query ? '综合结果' : '热门搜索'} />{(query ? communityPosts : communityPosts.slice(0, 3)).map((post) => <PostCard key={post.id} post={post} onPress={() => navigation.navigate('PostDetail', { postId: post.id })} />)}</PageScroll></Page>;
}

export function CreatePostScreen({ navigation }: any) {
  const [kind, setKind] = useState('地点标注');
  return <Page><AppHeader title="发布" onBack={() => navigation.goBack()} /><PageScroll><View style={styles.kindRow}>{['地点标注', '旅行讨论', '旅行计划'].map((item) => <Pressable key={item} onPress={() => setKind(item)} style={[styles.kind, kind === item && styles.kindActive]}><Text style={[styles.kindText, kind === item && styles.kindTextActive]}>{item}</Text></Pressable>)}</View><View style={styles.createGuide}><Text style={styles.createTitle}>{kind === '地点标注' ? '分享一个值得去或实用的地点' : kind === '旅行计划' ? '将计划发布到旅行计划探讨' : '和旅行者讨论一个具体问题'}</Text><Text style={styles.createText}>{kind === '地点标注' ? '公开标注需要地点、文字说明和照片，并会进入内容与地点审核。' : '请不要公开同行人的实时位置、私人行程或敏感地点。'}</Text></View><FormField label="标题" value={kind === '旅行讨论' ? '写下你想讨论的旅行问题' : '给内容起个清楚的标题'} /><FormField label="正文" value="补充路线、时间、注意事项和你的真实体验" tall />{kind !== '旅行讨论' ? <FormField label="关联地点" value="选择地点或在地图上标注" /> : null}<PrimaryButton label="提交审核" onPress={() => navigation.navigate('ModerationStatus', { type: kind })} /></PageScroll></Page>;
}

function FormField({ label, value, tall }: { label: string; value: string; tall?: boolean }) { return <View style={[styles.field, tall && styles.fieldTall]}><Text style={styles.fieldLabel}>{label}</Text><Text style={styles.fieldValue}>{value}</Text></View>; }

const styles = StyleSheet.create({
  safeArea: { flex: 1, backgroundColor: COLORS.white },
  topArea: { backgroundColor: COLORS.white, paddingTop: 12, paddingHorizontal: SPACING.lg, borderBottomWidth: StyleSheet.hairlineWidth, borderBottomColor: COLORS.border },
  searchRow: { flexDirection: 'row', alignItems: 'center', gap: 8 }, brandMark: { width: 34, height: 34, borderRadius: 11, backgroundColor: COLORS.primary, alignItems: 'center', justifyContent: 'center' }, brandText: { color: COLORS.white, fontSize: 9, fontWeight: '700', letterSpacing: -1 }, searchBox: { flex: 1, minHeight: 38, borderWidth: 1.5, borderColor: '#BDD6FA', borderRadius: 20, alignItems: 'center', flexDirection: 'row', paddingHorizontal: 11, gap: 6 }, searchPlaceholder: { ...FONTS.caption, color: COLORS.textTertiary, fontSize: 12 }, publish: { ...FONTS.caption, color: COLORS.primary, fontWeight: '700' },
  rootTabs: { flexDirection: 'row', gap: 24, marginTop: 7 }, rootTab: { minHeight: 32, paddingHorizontal: 2, justifyContent: 'flex-end', borderBottomWidth: 2, borderBottomColor: 'transparent' }, rootTabActive: { borderBottomColor: COLORS.primary }, rootText: { ...FONTS.body, color: COLORS.textTertiary, fontWeight: '600' }, rootTextActive: { color: COLORS.text, fontWeight: '700' },
  channelScroller: { flexGrow: 0, flexShrink: 0, height: 40, backgroundColor: COLORS.white, borderBottomWidth: StyleSheet.hairlineWidth, borderBottomColor: COLORS.border }, channelBar: { height: 40, flexDirection: 'row', alignItems: 'center', paddingHorizontal: SPACING.lg, gap: 19 }, channel: { height: 40, justifyContent: 'center' }, channelText: { ...FONTS.caption, color: COLORS.textTertiary, fontWeight: '600' }, channelTextActive: { color: COLORS.primary, fontWeight: '700' },
  feed: { flex: 1, backgroundColor: COLORS.white }, feedContent: { paddingBottom: 24 }, notice: { flexDirection: 'row', alignItems: 'center', gap: 8, margin: 12, padding: 10, borderWidth: 1, borderColor: COLORS.border, borderRadius: RADIUS.md, backgroundColor: '#FAFBFC' }, noticeTitle: { ...FONTS.caption, color: COLORS.text, fontWeight: '700' }, noticeText: { ...FONTS.caption, color: COLORS.textSecondary, fontSize: 11, marginTop: 2 },
  post: { padding: SPACING.lg, borderBottomWidth: StyleSheet.hairlineWidth, borderBottomColor: COLORS.border, backgroundColor: COLORS.white }, pressed: { opacity: .76 }, kickerRow: { flexDirection: 'row', alignItems: 'center', gap: 6, marginBottom: 7 }, kicker: { ...FONTS.caption, fontSize: 11, color: COLORS.primary, fontWeight: '700' }, markerKicker: { color: COLORS.success }, postTitle: { ...FONTS.titleMedium, color: COLORS.text, fontSize: 20, lineHeight: 28 }, author: { flexDirection: 'row', alignItems: 'center', gap: 7, marginTop: 9 }, avatar: { width: 26, height: 26, borderRadius: 13, alignItems: 'center', justifyContent: 'center' }, avatarText: { ...FONTS.caption, color: COLORS.text, fontWeight: '700' }, authorText: { ...FONTS.caption, color: COLORS.textSecondary }, postSummary: { ...FONTS.body, color: COLORS.textSecondary, lineHeight: 25, marginTop: 9 }, placeAction: { flexDirection: 'row', alignItems: 'center', gap: 6, marginTop: 10, padding: 8, borderRadius: 9, backgroundColor: COLORS.successLight }, placeText: { ...FONTS.caption, color: COLORS.success, flex: 1, fontWeight: '600' }, placeGo: { ...FONTS.caption, color: COLORS.success, fontWeight: '700' }, metrics: { flexDirection: 'row', gap: 17, marginTop: 12 }, metric: { ...FONTS.caption, color: COLORS.textTertiary },
  answer: { padding: 15, borderRadius: RADIUS.md, borderWidth: 1, borderColor: COLORS.border, backgroundColor: COLORS.white, marginBottom: 10 }, answerAuthor: { ...FONTS.caption, color: COLORS.text, fontWeight: '700' }, answerText: { ...FONTS.body, color: COLORS.textSecondary, lineHeight: 24, marginTop: 7 }, answerPlace: { flexDirection: 'row', alignItems: 'center', gap: 6, padding: 8, marginTop: 10, borderRadius: 8, backgroundColor: COLORS.successLight }, answerPlaceText: { ...FONTS.caption, color: COLORS.success, fontWeight: '700' }, answerMetrics: { flexDirection: 'row', gap: 18, marginTop: 11 },
  searchInput: { minHeight: 52, borderRadius: RADIUS.md, borderWidth: 1, borderColor: COLORS.border, backgroundColor: COLORS.white, flexDirection: 'row', alignItems: 'center', paddingHorizontal: 14, gap: 8 }, input: { flex: 1, ...FONTS.body, color: COLORS.text }, kindRow: { flexDirection: 'row', gap: 8, marginTop: 8 }, kind: { paddingHorizontal: 10, minHeight: 38, alignItems: 'center', justifyContent: 'center', borderRadius: 9, borderWidth: 1, borderColor: COLORS.border }, kindActive: { backgroundColor: COLORS.primaryLight, borderColor: COLORS.primary }, kindText: { ...FONTS.caption, color: COLORS.textSecondary }, kindTextActive: { color: COLORS.primary, fontWeight: '700' }, createGuide: { padding: 16, backgroundColor: COLORS.primaryLight, borderRadius: RADIUS.lg, marginTop: 18, marginBottom: 8 }, createTitle: { ...FONTS.body, color: COLORS.text, fontWeight: '700' }, createText: { ...FONTS.caption, color: COLORS.textSecondary, lineHeight: 19, marginTop: 5 }, field: { minHeight: 72, backgroundColor: COLORS.white, borderWidth: 1, borderColor: COLORS.border, borderRadius: RADIUS.md, padding: 12, marginTop: 12 }, fieldTall: { minHeight: 116 }, fieldLabel: { ...FONTS.caption, color: COLORS.textTertiary }, fieldValue: { ...FONTS.body, color: COLORS.text, marginTop: 7 },
});
