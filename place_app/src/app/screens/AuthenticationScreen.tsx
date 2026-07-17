import React from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import { EmptyState, Glyph, Page, PrimaryButton } from '../../shared/components/AppUI';
import { usePresentationSessionStore } from '../../store/usePresentationSessionStore';
import { COLORS, FONTS, RADIUS } from '../../shared/utils/theme';

export function AuthenticationScreen({ navigation, route }: any) {
  const complete = usePresentationSessionStore((state) => state.completeLocalSignIn);
  const action = route.params?.action || '继续此操作';
  const signIn = () => {
    complete();
    if (route.params?.resumeRoute) {
      navigation.replace(route.params.resumeRoute, route.params.resumeParams);
      return;
    }
    navigation.goBack();
  };
  return <Page><View style={styles.login}><View style={styles.mark}><Glyph symbol="◇" label="沿途" color={COLORS.primary} size={38} /></View><Text style={styles.title}>登录后可{action}</Text><Text style={styles.description}>浏览、搜索地点和开始导航始终无需登录。登录后才会保存你的旅行资产与关系数据。</Text><PrimaryButton label="手机号验证码登录" onPress={signIn} style={styles.primary} /><PrimaryButton label="微信登录" variant="secondary" onPress={signIn} style={styles.secondary} /><Pressable accessibilityRole="button" accessibilityLabel="使用密码登录" onPress={signIn}><Text style={styles.password}>使用密码登录</Text></Pressable><Text style={styles.disclaimer}>当前为前端演示认证流程；真实手机号、微信和密码认证会在后端契约接入后启用。</Text></View></Page>;
}

export function LockedFeatureTab({ navigation, title, description, action }: { navigation: any; title: string; description: string; action: string }) {
  return <Page><EmptyState title={title} description={description} action="登录后继续" onAction={() => navigation.navigate('LoginGate', { action })} /></Page>;
}

const styles = StyleSheet.create({ login: { flex: 1, paddingHorizontal: 28, justifyContent: 'center', alignItems: 'center' }, mark: { width: 82, height: 82, borderRadius: 26, backgroundColor: COLORS.primaryLight, alignItems: 'center', justifyContent: 'center' }, title: { ...FONTS.titleMedium, color: COLORS.text, marginTop: 22 }, description: { ...FONTS.body, color: COLORS.textSecondary, lineHeight: 24, textAlign: 'center', marginTop: 8 }, primary: { alignSelf: 'stretch', marginTop: 28 }, secondary: { alignSelf: 'stretch', marginTop: 10 }, password: { ...FONTS.caption, color: COLORS.primary, fontWeight: '700', marginTop: 18 }, disclaimer: { ...FONTS.caption, color: COLORS.textTertiary, lineHeight: 18, textAlign: 'center', marginTop: 22 } });
