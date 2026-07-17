import React, { useState } from 'react';
import { StyleSheet, Text, TextInput, View } from 'react-native';
import { EmptyState, Glyph, Page, PrimaryButton } from '../../shared/components/AppUI';
import { P0ApiError } from '../../p0/client';
import { useP0SessionStore } from '../../p0/session';
import { COLORS, FONTS, RADIUS } from '../../shared/utils/theme';

function messageFor(error: unknown) {
  return error instanceof P0ApiError ? error.message : '暂时无法完成登录，请稍后重试。';
}

export function AuthenticationScreen({ navigation, route }: any) {
  const sendCode = useP0SessionStore((state) => state.sendCode);
  const verifyCode = useP0SessionStore((state) => state.verifyCode);
  const action = route.params?.action || '继续此操作';
  const [phone, setPhone] = useState('');
  const [code, setCode] = useState('');
  const [challenge, setChallenge] = useState<{ challengeId: string; expiresIn: string }>();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string>();
  const normalizedPhone = phone.replace(/\s/g, '').replace(/^\+86/, '');

  const send = async () => {
    if (!/^1\d{10}$/.test(normalizedPhone)) {
      setError('请输入 11 位中国大陆手机号。');
      return;
    }
    setBusy(true);
    setError(undefined);
    try {
      setChallenge(await sendCode(normalizedPhone));
    } catch (cause) {
      setError(messageFor(cause));
    } finally {
      setBusy(false);
    }
  };

  const verify = async () => {
    if (!challenge) return;
    if (!/^\d{6}$/.test(code)) {
      setError('请输入 6 位验证码。');
      return;
    }
    setBusy(true);
    setError(undefined);
    try {
      await verifyCode(challenge.challengeId, code);
      if (route.params?.resumeRoute) {
        navigation.replace(route.params.resumeRoute, route.params.resumeParams);
      } else {
        navigation.goBack();
      }
    } catch (cause) {
      setError(messageFor(cause));
    } finally {
      setBusy(false);
    }
  };

  return <Page><View style={styles.login}>
    <View style={styles.mark}><Glyph symbol="◉" label="沿途" color={COLORS.primary} size={38} /></View>
    <Text style={styles.title}>登录后可{action}</Text>
    <Text style={styles.description}>游客可继续浏览、搜索地点和使用导航。手机号仅用于创建最小会话与防止滥用，不会显示给其他用户。</Text>
    {!challenge ? <>
      <TextInput value={phone} onChangeText={setPhone} style={styles.input} placeholder="手机号" placeholderTextColor={COLORS.textTertiary} keyboardType="phone-pad" textContentType="telephoneNumber" autoComplete="tel" maxLength={14} accessibilityLabel="手机号" />
      <PrimaryButton label={busy ? '正在发送…' : '发送验证码'} disabled={busy} onPress={() => void send()} style={styles.primary} />
    </> : <>
      <Text style={styles.challenge}>验证码已发送，有效期 {challenge.expiresIn}</Text>
      <TextInput value={code} onChangeText={(value) => setCode(value.replace(/\D/g, ''))} style={styles.input} placeholder="6 位验证码" placeholderTextColor={COLORS.textTertiary} keyboardType="number-pad" textContentType="oneTimeCode" autoComplete="sms-otp" maxLength={6} accessibilityLabel="验证码" />
      <PrimaryButton label={busy ? '正在验证…' : '验证并登录'} disabled={busy} onPress={() => void verify()} style={styles.primary} />
      <PrimaryButton label="重新发送" variant="secondary" disabled={busy} onPress={() => void send()} style={styles.secondary} />
    </>}
    {error ? <Text accessibilityRole="alert" style={styles.error}>{error}</Text> : null}
    <Text style={styles.disclaimer}>验证码、访问令牌和刷新令牌不会写入应用日志；刷新令牌仅保存在设备的系统安全存储中。</Text>
  </View></Page>;
}

export function LockedFeatureTab({ navigation, title, description, action }: { navigation: any; title: string; description: string; action: string }) {
  return <Page><EmptyState title={title} description={description} action="登录后继续" onAction={() => navigation.navigate('LoginGate', { action })} /></Page>;
}

export function UnavailableFeatureTab({ title, description }: { title: string; description: string }) {
  return <Page><EmptyState title={title} description={description} action="返回地图" onAction={() => undefined} /></Page>;
}

const styles = StyleSheet.create({
  login: { flex: 1, paddingHorizontal: 28, justifyContent: 'center', alignItems: 'center' }, mark: { width: 82, height: 82, borderRadius: 26, backgroundColor: COLORS.primaryLight, alignItems: 'center', justifyContent: 'center' }, title: { ...FONTS.titleMedium, color: COLORS.text, marginTop: 22 }, description: { ...FONTS.body, color: COLORS.textSecondary, lineHeight: 24, textAlign: 'center', marginTop: 8 }, input: { alignSelf: 'stretch', minHeight: 52, borderRadius: RADIUS.md, borderWidth: 1, borderColor: COLORS.border, paddingHorizontal: 14, color: COLORS.text, ...FONTS.body, marginTop: 22 }, primary: { alignSelf: 'stretch', marginTop: 10 }, secondary: { alignSelf: 'stretch', marginTop: 10 }, challenge: { ...FONTS.caption, color: COLORS.textSecondary, alignSelf: 'stretch', marginTop: 22 }, error: { ...FONTS.caption, color: COLORS.danger, alignSelf: 'stretch', textAlign: 'center', marginTop: 12 }, disclaimer: { ...FONTS.caption, color: COLORS.textTertiary, lineHeight: 18, textAlign: 'center', marginTop: 22 },
});
