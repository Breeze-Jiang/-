export type PlaceCategory =
  | 'toilet' | 'fuel' | 'charge' | 'market' | 'sight' | 'hotel'
  | 'parking' | 'powerbank' | 'pharmacy' | 'hospital' | 'convenience' | 'police' | 'food' | 'unclassified';

export type PlaceExistence = '地点存在已确认' | '地点信息待核实' | '地点存在待确认' | '已收录' | '待核实';

export type TravelMode = 'walk' | 'ride' | 'drive' | 'transit' | 'taxi';

export interface Place {
  id: string;
  name: string;
  category: PlaceCategory;
  distance: string;
  address: string;
  hours: string;
  confirmedAt: string;
  existence: PlaceExistence;
  tags: string[];
}

export type PostKind = 'resource' | 'marker' | 'plan' | 'discussion';

export interface CommunityPost {
  id: string;
  kind: PostKind;
  author: string;
  city: string;
  title: string;
  summary: string;
  place?: string;
  imageTone: string;
  likes: number;
  comments: number;
  saved: number;
}

export interface SavedFolder {
  id: string;
  name: string;
  subtitle: string;
  count: number;
  accent: string;
}

export interface TravelPlan {
  id: string;
  title: string;
  city: string;
  days: number;
  places: number;
  mode: TravelMode;
  visibility: '仅自己' | '好友可见' | '公开审核中' | '已公开';
}
