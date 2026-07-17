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
import { AuthenticationScreen, LockedFeatureTab } from '../app/screens/AuthenticationScreen';
import { usePresentationSessionStore } from '../store/usePresentationSessionStore';
import { AppIcon, AppIconName } from '../ui/AppIcon';

const RootStack = createNativeStackNavigator();
const Tabs = createBottomTabNavigator();

const tabLabels: Record<string, string> = { MapTab: '地图', DiscoverTab: '发现', MessagesTab: '消息', FavoritesTab: '收藏', ProfileTab: '我的' };
const tabIcons: Record<string, AppIconName> = { MapTab: 'map', DiscoverTab: 'discover', MessagesTab: 'message', FavoritesTab: 'favorite', ProfileTab: 'profile' };

function withAuthentication(Component: React.ComponentType<any>, action: string) {
  return function AuthenticatedScreen(props: any) {
    const signedIn = usePresentationSessionStore((state) => state.isSignedIn);
    useEffect(() => {
      if (!signedIn) {
        props.navigation.replace('LoginGate', { action, resumeRoute: props.route.name, resumeParams: props.route.params });
      }
    }, [action, props.navigation, props.route.name, props.route.params, signedIn]);
    return signedIn ? <Component {...props} /> : <View style={{ flex: 1 }} />;
  };
}

const ProtectedCreatePost = withAuthentication(CreatePostComposerScreen, '发布旅行内容');
const ProtectedMarkerList = withAuthentication(MarkerListScreen, '管理个人标注');
const ProtectedCreateMarker = withAuthentication(CreateMarkerComposerScreen, '创建地点标注');
const ProtectedMarkerDetail = withAuthentication(MarkerDetailScreen, '编辑个人标注');
const ProtectedPlanEditor = withAuthentication(PlanComposerScreen, '创建旅行计划');
const ProtectedPlanDetail = withAuthentication(PlanDetailScreen, '查看个人旅行计划');
const ProtectedFolderDetail = withAuthentication(FolderDetailScreen, '查看收藏夹');
const ProtectedCreateFolder = withAuthentication(CreateFolderScreen, '创建收藏夹');
const ProtectedManageFavorites = withAuthentication(ManageFavoritesScreen, '管理收藏地点');
const ProtectedFriends = withAuthentication(FriendsScreen, '管理好友');
const ProtectedFriendRequest = withAuthentication(FriendRequestScreen, '添加好友');
const ProtectedChat = withAuthentication(ChatComposerScreen, '使用聊天');
const ProtectedGroupManagement = withAuthentication(GroupManagementScreen, '管理私密群聊');
const ProtectedNotifications = withAuthentication(NotificationsScreen, '查看通知');

function TabIcon({ routeName, color, focused }: { routeName: string; color: string; focused: boolean }) {
  return <AppIcon name={tabIcons[routeName]} color={color} size={22} strokeWidth={focused ? 2.3 : 1.8} label={tabLabels[routeName]} />;
}

function MainTabs() {
  return <Tabs.Navigator screenOptions={({ route }) => ({ headerShown: false, tabBarActiveTintColor: COLORS.primary, tabBarInactiveTintColor: COLORS.textTertiary, tabBarStyle: { height: 64, paddingTop: 5, paddingBottom: 7, backgroundColor: COLORS.white, borderTopColor: COLORS.border }, tabBarLabelStyle: { fontSize: 11, fontWeight: '600' }, tabBarIcon: ({ color, focused }) => <TabIcon routeName={route.name} color={color} focused={focused} /> })}>
    <Tabs.Screen name="MapTab" component={MapHomeScreen} options={{ tabBarLabel: '地图' }} />
    <Tabs.Screen name="DiscoverTab" component={DiscoverScreen} options={{ tabBarLabel: '发现' }} />
    <Tabs.Screen name="MessagesTab" component={MessagesTab} options={{ tabBarLabel: '消息' }} />
    <Tabs.Screen name="FavoritesTab" component={FavoritesTab} options={{ tabBarLabel: '收藏' }} />
    <Tabs.Screen name="ProfileTab" component={ProfileScreenV2} options={{ tabBarLabel: '我的' }} />
  </Tabs.Navigator>;
}

function MessagesTab({ navigation }: any) {
  const signedIn = usePresentationSessionStore((state) => state.isSignedIn);
  return signedIn ? <MessagesScreen navigation={navigation} /> : <LockedFeatureTab navigation={navigation} title="登录后查看消息" description="好友、群聊、地点定向分享和实时位置共享都需要先登录。" action="查看旅行消息" />;
}

function FavoritesTab({ navigation }: any) {
  const signedIn = usePresentationSessionStore((state) => state.isSignedIn);
  return signedIn ? <FavoritesScreenV2 navigation={navigation} /> : <LockedFeatureTab navigation={navigation} title="登录后管理收藏" description="收藏地点、建立自定义收藏夹和保存旅行计划需要登录。" action="管理旅行收藏" />;
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
