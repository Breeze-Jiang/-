import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import type { BottomTabScreenProps } from '@react-navigation/bottom-tabs';
import type { CompositeScreenProps } from '@react-navigation/native';

// Auth Stack
export type AuthStackParamList = {
  PhoneEntry: undefined;
  Login: { phone: string };
  Register: { phone: string };
  EditProfile: undefined;
  Preferences: undefined;
};

export type PhoneEntryScreenProps = NativeStackScreenProps<AuthStackParamList, 'PhoneEntry'>;
export type LoginScreenProps = NativeStackScreenProps<AuthStackParamList, 'Login'>;
export type RegisterScreenProps = NativeStackScreenProps<AuthStackParamList, 'Register'>;
export type EditProfileScreenProps = NativeStackScreenProps<AuthStackParamList, 'EditProfile'>;
export type PreferencesScreenProps = NativeStackScreenProps<AuthStackParamList, 'Preferences'>;

// Main Tab Navigator
export type MainTabParamList = {
  MapTab: undefined;
  FavoritesTab: undefined;
  HistoryTab: undefined;
  ProfileTab: undefined;
};

// Map Stack (within MapTab)
export type MapStackParamList = {
  Map: undefined;
  POIDetail: { poiId: string };
  Identifier: undefined;
};

// Profile Stack (within ProfileTab)
export type ProfileStackParamList = {
  Profile: undefined;
  EditProfile: undefined;
  Preferences: undefined;
  ShareSetup: undefined;
  SharedMap: { sessionId: string };
  FriendsList: undefined;
  MyMarkers: undefined;
};

// Root Navigator
export type RootStackParamList = {
  Auth: undefined;
  Main: undefined;
};

// Screen prop types
export type MapScreenProps = CompositeScreenProps<
  NativeStackScreenProps<MapStackParamList, 'Map'>,
  BottomTabScreenProps<MainTabParamList>
>;
