import type { CommunityPost, Place, SavedFolder, TravelPlan } from './types';

export const urgentServices = [
  { key: 'toilet', label: '厕所', symbol: '▣' },
  { key: 'fuel', label: '加油站', symbol: '◒' },
  { key: 'charge', label: '充电站', symbol: 'ϟ' },
  { key: 'market', label: '超市', symbol: '□' },
  { key: 'sight', label: '景点', symbol: '◇' },
  { key: 'hotel', label: '酒店', symbol: '⌂' },
] as const;

export const alongtuServices = [
  { key: 'parking', label: '停车场', symbol: 'P' },
  { key: 'powerbank', label: '充电宝', symbol: 'ϟ' },
  { key: 'pharmacy', label: '药店', symbol: '+' },
  { key: 'hospital', label: '医院', symbol: '✚' },
  { key: 'convenience', label: '便利店', symbol: '□' },
  { key: 'police', label: '警务点', symbol: '⌘' },
  { key: 'food', label: '美食', symbol: '○' },
] as const;

export const places: Place[] = [
  { id: 'toilet-1', name: '天河公园东门公共卫生间', category: 'toilet', distance: '210 m', address: '广州市天河区天府路 11 号东门旁', hours: '06:00–22:00', confirmedAt: '18 分钟前由 2 位用户确认', existence: '已收录', tags: ['无障碍', '母婴台'] },
  { id: 'toilet-2', name: '正佳广场 B1 卫生间', category: 'toilet', distance: '620 m', address: '广州市天河区天河路 228 号 B1', hours: '商场营业时间内', confirmedAt: '今天 10:26 由用户确认', existence: '已收录', tags: ['室内', '免费'] },
  { id: 'charge-1', name: '特来电汽车充电站', category: 'charge', distance: '850 m', address: '广州市天河区体育东路停车场', hours: '24 小时', confirmedAt: '昨天 21:08 由站点信息更新', existence: '已收录', tags: ['直流快充', '停车收费'] },
  { id: 'fuel-1', name: '中国石化体育东路加油站', category: 'fuel', distance: '1.1 km', address: '广州市天河区体育东路 18 号', hours: '24 小时', confirmedAt: '今天 09:12 由站点信息更新', existence: '已收录', tags: ['92/95 汽油', '洗手间'] },
  { id: 'market-1', name: '永旺超市天河城店', category: 'market', distance: '720 m', address: '广州市天河区天河路 208 号 B1', hours: '10:00–22:00', confirmedAt: '今天 11:35 由用户确认', existence: '已收录', tags: ['生鲜', '室内'] },
  { id: 'sight-1', name: '天河公园东门', category: 'sight', distance: '1.4 km', address: '广州市天河区天府路 11 号', hours: '06:00–22:00', confirmedAt: '昨天 17:42 由用户确认', existence: '已收录', tags: ['免费', '适合散步'] },
  { id: 'hotel-1', name: '天河希尔顿酒店', category: 'hotel', distance: '980 m', address: '广州市天河区林和西横路 215 号', hours: '全天', confirmedAt: '今天 08:20 由用户确认', existence: '已收录', tags: ['前台', '停车场'] },
  { id: 'parking-1', name: '正佳广场地下停车场', category: 'parking', distance: '530 m', address: '广州市天河区天河路 228 号', hours: '07:00–23:00', confirmedAt: '今天 12:06 由用户确认', existence: '已收录', tags: ['室内', '充电桩'] },
  { id: 'powerbank-1', name: '天河城一层共享充电宝', category: 'powerbank', distance: '690 m', address: '广州市天河区天河路 208 号北门', hours: '商场营业时间内', confirmedAt: '2 小时前由用户确认', existence: '已收录', tags: ['归还点', '多品牌'] },
  { id: 'pharmacy-1', name: '大参林药房体育西店', category: 'pharmacy', distance: '460 m', address: '广州市天河区体育西路 101 号', hours: '08:00–23:00', confirmedAt: '今天 10:41 由用户确认', existence: '已收录', tags: ['常用药', '可咨询'] },
  { id: 'hospital-1', name: '中山三院急诊入口', category: 'hospital', distance: '1.8 km', address: '广州市天河区天河路 600 号', hours: '24 小时', confirmedAt: '昨天 20:16 由机构信息更新', existence: '已收录', tags: ['急诊', '停车'] },
  { id: 'convenience-1', name: '全家便利店体育西店', category: 'convenience', distance: '330 m', address: '广州市天河区体育西路 87 号', hours: '24 小时', confirmedAt: '36 分钟前由用户确认', existence: '已收录', tags: ['热食', '饮水'] },
  { id: 'police-1', name: '天河路警务点', category: 'police', distance: '1.2 km', address: '广州市天河区天河路 518 号附近', hours: '以现场信息为准', confirmedAt: '3 天前由用户确认', existence: '待核实', tags: ['求助', '咨询'] },
  { id: 'food-1', name: '体育西路本地小吃街', category: 'food', distance: '400 m', address: '广州市天河区体育西横街', hours: '11:00–22:30', confirmedAt: '今天 12:10 由用户确认', existence: '已收录', tags: ['本地口味', '步行可达'] },
];

export const communityPosts: CommunityPost[] = [
  { id: 'post-1', kind: 'marker', author: '林岸', city: '广州 · 沙面', title: '傍晚最适合散步的江边入口', summary: '从东侧小门进，不用绕远路；日落后沿江灯光很舒服，附近也有公共卫生间。', place: '沙面公园东侧入口', imageTone: '#BFD8EE', likes: 482, comments: 39, saved: 126 },
  { id: 'post-2', kind: 'resource', author: '阿帆自驾', city: '清远', title: '连山公路补给点：加油、充电和便利店都在同一侧', summary: '长途自驾路过这里可以一次完成补给，停车入口在辅路，晚上照明正常。', place: 'G107 连山服务区', imageTone: '#CBDCC7', likes: 316, comments: 28, saved: 94 },
  { id: 'post-3', kind: 'plan', author: 'Momo 在路上', city: '成都', title: '三天两夜慢游：从宽窄巷子到都江堰', summary: '以地铁和步行为主，保留午后自由时间；已经标出每段换乘和适合休息的点。', imageTone: '#F0D6B3', likes: 721, comments: 64, saved: 203 },
  { id: 'post-4', kind: 'discussion', author: '晴天旅记', city: '杭州', title: '西湖骑行路线，雨天还有必要去吗？', summary: '想听听最近去过的朋友建议：雨天视野、停车和租车点实际情况如何？', imageTone: '#D9D5E9', likes: 88, comments: 51, saved: 17 },
];

export const savedFolders: SavedFolder[] = [
  { id: 'city', name: '广州 · 天河区', subtitle: '按距离查看', count: 18, accent: '#EAF2FF' },
  { id: 'drive', name: '自驾补给', subtitle: '停车 / 加油 / 充电', count: 12, accent: '#EAF8F4' },
  { id: 'weekend', name: '周末想去', subtitle: '自定义收藏夹', count: 26, accent: '#FFF3E5' },
];

export const travelPlans: TravelPlan[] = [
  { id: 'plan-1', title: '成都慢游 3 天游', city: '成都', days: 3, places: 11, mode: 'transit', visibility: '仅自己' },
  { id: 'plan-2', title: '粤北自驾补给路线', city: '清远', days: 2, places: 8, mode: 'drive', visibility: '好友可见' },
];
