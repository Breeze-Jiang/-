import React from 'react';
import {
  Bell, Car, ChevronLeft, ChevronRight, CircleUserRound,
  Compass, Crosshair, Heart, Map, MapPin, MessageCircle,
  Navigation, ParkingCircle, PencilLine, Plus, Search, ShieldCheck,
  Store, Toilet, Utensils, Zap,
} from 'lucide-react-native';

export type AppIconName =
  | 'map' | 'discover' | 'message' | 'favorite' | 'profile'
  | 'search' | 'locate' | 'pin' | 'navigation' | 'plus'
  | 'back' | 'forward' | 'marker' | 'trust' | 'parking'
  | 'charging' | 'toilet' | 'store' | 'food' | 'car' | 'notice';

const iconMap = {
  map: Map,
  discover: Compass,
  message: MessageCircle,
  favorite: Heart,
  profile: CircleUserRound,
  search: Search,
  locate: Crosshair,
  pin: MapPin,
  navigation: Navigation,
  plus: Plus,
  back: ChevronLeft,
  forward: ChevronRight,
  marker: PencilLine,
  trust: ShieldCheck,
  parking: ParkingCircle,
  charging: Zap,
  toilet: Toilet,
  store: Store,
  food: Utensils,
  car: Car,
  notice: Bell,
} as const;

type Props = {
  name: AppIconName;
  color: string;
  size?: number;
  strokeWidth?: number;
  label?: string;
};

export function AppIcon({ name, color, size = 20, strokeWidth = 1.9, label }: Props) {
  const Icon = iconMap[name];
  return <Icon color={color} size={size} strokeWidth={strokeWidth} accessibilityLabel={label} />;
}
