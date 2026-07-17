import { create } from 'zustand';

export type MarkerVisibility = '仅自己' | '指定好友' | '全部好友' | '公开到论坛';
export type MarkerReviewState = '私密' | '审核中' | '已公开';

export type TravelMarker = {
  id: string;
  name: string;
  category: string;
  note: string;
  visibility: MarkerVisibility;
  reviewState: MarkerReviewState;
  placeName: string;
  city: string;
  updatedAt: string;
};

type NewMarker = Omit<TravelMarker, 'id' | 'reviewState' | 'updatedAt'>;

type MarkerState = {
  markers: TravelMarker[];
  addMarker: (marker: NewMarker) => string;
};

export const useMarkerStore = create<MarkerState>((set, get) => ({
  markers: [
    { id: 'marker-1', name: '北山街湖边小亭', category: 'sight', note: '入口在树荫后的步道，傍晚适合短暂停留。', visibility: '仅自己', reviewState: '私密', placeName: '北山街湖边小亭', city: '杭州 · 西湖区', updatedAt: '今天 14:26' },
    { id: 'marker-2', name: '雨天停车入口', category: 'parking', note: '从辅路进入，雨天也不需要绕到主路。', visibility: '公开到论坛', reviewState: '审核中', placeName: '苏堤南口停车场', city: '杭州 · 西湖区', updatedAt: '昨天' },
  ],
  addMarker: (marker) => {
    const id = `marker-${Date.now()}`;
    const newMarker: TravelMarker = {
      ...marker,
      id,
      reviewState: marker.visibility === '公开到论坛' ? '审核中' : '私密',
      updatedAt: '刚刚保存',
    };
    set({ markers: [newMarker, ...get().markers] });
    return id;
  },
}));
