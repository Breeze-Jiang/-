import { POI_TYPE_CONFIG, POIType } from '../types/poi';

// 主题色
export const COLORS = {
  primary: '#246BDF',
  primaryPressed: '#1D56B3',
  primaryLight: '#EAF2FF',
  success: '#149B80',
  successLight: '#EAF8F4',
  warning: '#AA6919',
  warningLight: '#FFF3E5',
  danger: '#C94F58',
  dangerLight: '#FFF0F1',
  text: '#172521',
  textSecondary: '#5F6E69',
  textTertiary: '#72817C',
  border: '#E2EAEC',
  background: '#F5F8F9',
  white: '#FFFFFF',
  overlay: 'rgba(23,37,33,0.48)',
};

// 字体
export const FONTS = {
  titleLarge: { fontSize: 26, lineHeight: 34, fontWeight: '700' as const },
  titleMedium: { fontSize: 19, lineHeight: 26, fontWeight: '700' as const },
  body: { fontSize: 16, fontWeight: '400' as const },
  caption: { fontSize: 13, lineHeight: 19, fontWeight: '400' as const },
};

// 间距
export const SPACING = {
  xs: 4,
  sm: 8,
  md: 12,
  lg: 16,
  xl: 24,
  xxl: 32,
};

// 圆角
export const RADIUS = {
  sm: 8,
  md: 12,
  lg: 16,
  xl: 24,
  full: 999,
};

// 阴影
export const SHADOWS = {
  card: {
    shadowColor: '#172521',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.08,
    shadowRadius: 12,
    elevation: 3,
  },
  button: {
    shadowColor: '#246BDF',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.3,
    shadowRadius: 8,
    elevation: 4,
  },
};

// POI 类型颜色映射
export const getPOIColor = (type: POIType): string => {
  return POI_TYPE_CONFIG[type]?.color ?? COLORS.primary;
};

// 高德 Web JS API Key（WebView 加载 AMap JS API v2.0，必须用 Web端 key）
export const AMAP_KEY = '630f6157213401f2c332141aea316cfa';

// API Base URL（Android 模拟器用 10.0.2.2 访问宿主机）
export const API_BASE_URL = 'http://10.0.2.2:8080';
