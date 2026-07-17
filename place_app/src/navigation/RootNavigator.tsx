import React, { useEffect } from 'react';
import { View } from 'react-native';
import { NavigationContainer } from '@react-navigation/native';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { COLORS } from '../shared/utils/theme';
import MapHomeScreen from '../app/screens/MapHomeScreen';
import PlaceResultsScreen from '../app/screens/PlaceResultsScreen';
import PlaceDetailScreen from '../app/screens/PlaceDetailScreen';
import { DrivingModeScreen } from '../app/screens/NavigationScreen';
import RouteModeScreen from '../app/screens/RouteModeScreen';
import ExternalNavigationScreen from '../app/screens/ExternalNavigationScreen';
import DiscoverScreen, { CommunitySearchScreen, CreatePostScreen, PostDetailScreen } from '../app/screens/DiscoverScreen';
import FavoritesScreenV2, { CreateFolderScreen, CreateMarkerScreen, FolderDetailScreen, ManageFavoritesScreen, MarkerDetailScreen, MarkerListScreen, PlanDetailScreen, PlanEditorScreen } from '../app/screens/AssetScreens';
import MessagesScreen, { ChatScreen, FriendRequestScreen, FriendsScreen, GroupManagementScreen } from '../app/screens/SocialScreens';
import ProfileScreenV2, { ModerationStatusScreen, NotificationsScreen, PrivacyScreen, SettingsScreen } from '../app/screens/AccountScreens';
import { ContentUnavailableScreen, LocationPermissionScreen, NetworkErrorScreen } from '../app/screens/SystemStateScreens';
import { CreateMarkerComposerScreen, CreatePostComposerScreen, PlanComposerScreen } from '../app/screens/ComposerScreens';
import ChatComposerScreen from '../app/screens/ChatComposerScreen';
import { AuthenticationScreen, UnavailableFeatureTab } from '../app/screens/AuthenticationScreen';
import { AppIcon, AppIconName } from '../ui/AppIcon';
import { useSafeAreaInsets } from 'react-native-safe-area-context';

const RootStack = createNativeStackNavigator();
const Tabs = createBottomTabNavigator();

const tabLabels: Record<string, string> = { MapTab: '地图', DiscoverTab: '发现', MessagesTab: '消息', FavoritesTab: '收藏', ProfileTab: '我的' };
const tabIcons: Record<string, AppIconName> = { MapTab: 'map', DiscoverTab: 'discover', MessagesTab: 'message', FavoritesTab: 'favorite', ProfileTab: 'profile' };

function unavailableFeature(action: string) {
  return function UnavailableScreen() {
    return <UnavailableFeatureTab title="功能尚未接入" description={`${action}将随 P1–P3 的服务端能力、数据模型和审核边界一起上线；当前 P0 仅支持地点查询、路线和地点确认。`} />;
  };
}

const ProtectedCreatePost = unavailableFeature('发布旅行内容');
const ProtectedMarkerList = unavailableFeature('管理个人标注');
const ProtectedCreateMarker = unavailableFeature('创建地点标注');
const ProtectedMarkerDetail = unavailableFeature('编辑个人标注');
const ProtectedPlanEditor = unavailableFeature('创建旅行计划');
const ProtectedPlanDetail = unavailableFeature('查看个人旅行计划');
const ProtectedFolderDetail = unavailableFeature('查看收藏夹');
const ProtectedCreateFolder = unavailableFeature('创建收藏夹');
const ProtectedManageFavorites = unavailableFeature('管理收藏地点');
const ProtectedFriends = unavailableFeature('管理好友');
const ProtectedFriendRequest = unavailableFeature('添加好友');
const ProtectedChat = unavailableFeature('使用聊天');
const ProtectedGroupManagement = unavailableFeature('管理私密群聊');
const ProtectedNotifications = unavailableFeature('查看通知');

function TabIcon({ routeName, color, focused }: { routeName: string; color: string; focused: boolean }) {
  return <AppIcon name={tabIcons[routeName]} color={color} size={22} strokeWidth={focused ? 2.3 : 1.8} label={tabLabels[routeName]} />;
}

function MainTabs() {
  const insets = useSafeAreaInsets();
  // A small optical inset sits above the icons; the physical inset below them
  // is calculated from the device. This mirrors mature mobile apps: controls
  // never crowd either the content or the Android gesture area.
  const tabBarBottomInset = Math.max(insets.bottom, 8);
  const tabBarHeight = 60 + tabBarBottomInset;
  return <Tabs.Navigator screenOptions={({ route }) => ({ headerShown: false, tabBarActiveTintColor: COLORS.primary, tabBarInactiveTintColor: COLORS.textTertiary, tabBarStyle: { height: tabBarHeight, paddingTop: 7, paddingBottom: tabBarBottomInset, backgroundColor: COLORS.white, borderTopColor: COLORS.border }, tabBarLabelStyle: { fontSize: 11, fontWeight: '600', marginTop: 1 }, tabBarIcon: ({ color, focused }) => <TabIcon routeName={route.name} color={color} focused={focused} /> })}>
    <Tabs.Screen name="MapTab" component={MapHomeScreen} options={{ tabBarLabel: '地图' }} />
    <Tabs.Screen name="DiscoverTab" component={DiscoverScreen} options={{ tabBarLabel: '发现' }} />
    <Tabs.Screen name="MessagesTab" component={MessagesTab} options={{ tabBarLabel: '消息' }} />
    <Tabs.Screen name="FavoritesTab" component={FavoritesTab} options={{ tabBarLabel: '收藏' }} />
    <Tabs.Screen name="ProfileTab" component={ProfileScreenV2} options={{ tabBarLabel: '我的' }} />
  </Tabs.Navigator>;
}

function MessagesTab() {
  return <UnavailableFeatureTab title="消息功能尚未接入" description="好友、群聊、地点定向分享和实时位置共享将在后续服务端阶段接入；P0 登录不会解锁本地演示内容。" />;
}

function FavoritesTab() {
  return <UnavailableFeatureTab title="收藏功能尚未接入" description="收藏、标注和旅行计划将在后续服务端阶段接入；P0 登录不会创建或展示任何本地演示资产。" />;
}

function tabRedirect(target: string) {
  return function TabRedirect({ navigation }: any) {
    useEffect(() => { navigation.navigate('Main', { screen: target }); }, [navigation]);
    return <View style={{ flex: 1 }} />;
  };
}

const DiscoverEntry = tabRedirect('DiscoverTab');
const FavoritesEntry = tabRedirect('FavoritesTab');
const ProfileEntry = tabRedirect('ProfileTab');
const MapEntry = tabRedirect('MapTab');

export default function RootNavigator() {
  return <NavigationContainer>
    <RootStack.Navigator screenOptions={{ headerShown: false, animation: 'slide_from_right' }}>
      <RootStack.Screen name="Main" component={MainTabs} />
      <RootStack.Screen name="DiscoverTab" component={DiscoverEntry} />
      <RootStack.Screen name="FavoritesTab" component={FavoritesEntry} />
      <RootStack.Screen name="ProfileTab" component={ProfileEntry} />
      <RootStack.Screen name="MapTab" component={MapEntry} />
      <RootStack.Group screenOptions={{ presentation: 'card' }}>
        <RootStack.Screen name="PlaceResults" component={PlaceResultsScreen} />
        <RootStack.Screen name="PlaceDetail" component={PlaceDetailScreen} />
        <RootStack.Screen name="Navigation" component={RouteModeScreen} />
        <RootStack.Screen name="DrivingMode" component={DrivingModeScreen} />
        <RootStack.Screen name="ExternalMap" component={ExternalNavigationScreen} />
        <RootStack.Screen name="CommunitySearch" component={CommunitySearchScreen} />
        <RootStack.Screen name="PostDetail" component={PostDetailScreen} />
        <RootStack.Screen name="CreatePost" component={ProtectedCreatePost} />
        <RootStack.Screen name="MarkerList" component={ProtectedMarkerList} />
        <RootStack.Screen name="CreateMarker" component={ProtectedCreateMarker} />
        <RootStack.Screen name="MarkerDetail" component={ProtectedMarkerDetail} />
        <RootStack.Screen name="PlanEditor" component={ProtectedPlanEditor} />
        <RootStack.Screen name="PlanDetail" component={ProtectedPlanDetail} />
        <RootStack.Screen name="FolderDetail" component={ProtectedFolderDetail} />
        <RootStack.Screen name="CreateFolder" component={ProtectedCreateFolder} />
        <RootStack.Screen name="ManageFavorites" component={ProtectedManageFavorites} />
        <RootStack.Screen name="Friends" component={ProtectedFriends} />
        <RootStack.Screen name="FriendRequest" component={ProtectedFriendRequest} />
        <RootStack.Screen name="Chat" component={ProtectedChat} />
        <RootStack.Screen name="GroupManagement" component={ProtectedGroupManagement} />
        <RootStack.Screen name="Notifications" component={ProtectedNotifications} />
        <RootStack.Screen name="Privacy" component={PrivacyScreen} />
        <RootStack.Screen name="Settings" component={SettingsScreen} />
        <RootStack.Screen name="ModerationStatus" component={ModerationStatusScreen} />
        <RootStack.Screen name="LocationPermission" component={LocationPermissionScreen} />
        <RootStack.Screen name="NetworkError" component={NetworkErrorScreen} />
        <RootStack.Screen name="ContentUnavailable" component={ContentUnavailableScreen} />
      </RootStack.Group>
      <RootStack.Group screenOptions={{ presentation: 'modal', animation: 'slide_from_bottom' }}>
        <RootStack.Screen name="LoginGate" component={AuthenticationScreen} />
      </RootStack.Group>
    </RootStack.Navigator>
  </NavigationContainer>;
}
