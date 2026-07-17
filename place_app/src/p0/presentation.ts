import type { Place } from '../app/types';
import { formatDistance, formatRelativeTime } from '../shared/utils/format';
import type { P0Place, P0PlaceCategory } from './contracts';

const categoryMap: Record<P0PlaceCategory, Place['category']> = {
  unclassified: 'unclassified',
  toilet: 'toilet',
  fuel: 'fuel',
  charging: 'charge',
  supermarket: 'market',
  attraction: 'sight',
  hotel: 'hotel',
  parking: 'parking',
  power_bank: 'powerbank',
  pharmacy: 'pharmacy',
  hospital: 'hospital',
  convenience_store: 'convenience',
  police: 'police',
  food: 'food',
};

export function toP0Category(category: Place['category']): P0PlaceCategory {
  const map: Record<Place['category'], P0PlaceCategory> = {
    unclassified: 'unclassified',
    toilet: 'toilet',
    fuel: 'fuel',
    charge: 'charging',
    market: 'supermarket',
    sight: 'attraction',
    hotel: 'hotel',
    parking: 'parking',
    powerbank: 'power_bank',
    pharmacy: 'pharmacy',
    hospital: 'hospital',
    convenience: 'convenience_store',
    police: 'police',
    food: 'food',
  };
  return map[category];
}

function existenceLabel(place: P0Place): Place['existence'] {
  if (place.trust.exists.status === 'confirmed') return '地点存在已确认';
  if (place.trust.exists.status === 'incorrect') return '地点信息待核实';
  return '地点存在待确认';
}

function confirmationLabel(place: P0Place) {
  const confirmedAt = place.trust.exists.confirmedAt;
  return confirmedAt ? `最近确认：${formatRelativeTime(confirmedAt)}` : '暂无用户确认';
}

export function presentP0Place(place: P0Place): Place {
  return {
    id: place.id,
    name: place.name,
    category: categoryMap[place.category],
    distance: place.distanceMeters === undefined ? '距离未知' : formatDistance(place.distanceMeters),
    address: place.address || '地址待补充',
    hours: '未知/待确认',
    confirmedAt: confirmationLabel(place),
    existence: existenceLabel(place),
    tags: place.trust.openingHours.status === 'unknown' ? ['营业时间待确认'] : ['营业信息已确认'],
  };
}
