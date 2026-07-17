// POI 类型定义
export enum POIType {
  PARKING = 'parking',
  CHARGING = 'charging',
  TOILET = 'toilet',
  POWERBANK = 'powerbank',
  CONVENIENCE = 'convenience',
}

// POI 类型显示信息
export const POI_TYPE_CONFIG: Record<POIType, { label: string; icon: string; color: string }> = {
  [POIType.PARKING]:     { label: '停车场',   icon: '🅿️', color: '#1677FF' },
  [POIType.CHARGING]:    { label: '充电桩',   icon: '🔋', color: '#52C41A' },
  [POIType.TOILET]:      { label: '公共厕所', icon: '🚻', color: '#FA8C16' },
  [POIType.POWERBANK]:   { label: '充电宝',   icon: '🔌', color: '#FAAD14' },
  [POIType.CONVENIENCE]: { label: '便利店',   icon: '🏪', color: '#722ED1' },
};

// POI 数据模型
export interface POI {
  id: string;
  name: string;
  type: POIType;
  lat: number;
  lng: number;
  address: string;
  distance?: number;
  tel?: string;
  rating?: number;
}

// POI 搜索参数
export interface POISearchParams {
  keyword: string;
  lat: number;
  lng: number;
  radius?: number;
  types?: POIType[];
}

// 高德 POI 类型对应的搜索编码
export const AMAP_POI_CODES: Record<POIType, string> = {
  [POIType.PARKING]:     '150100|150200',
  [POIType.CHARGING]:    '011200',
  [POIType.TOILET]:      '200300',
  [POIType.POWERBANK]:   '060000',
  [POIType.CONVENIENCE]: '060200',
};
