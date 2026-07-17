const fs = require('fs');
const path = require('path');

const projectRoot = path.resolve(__dirname, '..');
const read = (relativePath) => fs.readFileSync(path.join(projectRoot, relativePath), 'utf8');
const expect = (source, expected, message) => {
  if (!source.includes(expected)) {
    throw new Error(message);
  }
};
const expectNot = (source, unexpected, message) => {
  if (source.includes(unexpected)) {
    throw new Error(message);
  }
};

const navigator = read('src/navigation/RootNavigator.tsx');
const appConfig = JSON.parse(read('app.json'));
const packageJson = JSON.parse(read('package.json'));
const mainActivity = read('android/app/src/main/java/com/placenative/MainActivity.kt');
expect(mainActivity, `getMainComponentName(): String = "${appConfig.name}"`, 'Android main component name must match app.json registration name');
for (const tab of ['MapTab', 'DiscoverTab', 'MessagesTab', 'FavoritesTab', 'ProfileTab']) {
  expect(navigator, `name="${tab}"`, `Missing required root tab: ${tab}`);
}
for (const screen of ['PlaceResults', 'PlaceDetail', 'Navigation', 'DrivingMode', 'CreatePost', 'CreateMarker', 'PlanEditor', 'Friends', 'Chat', 'LoginGate', 'ModerationStatus', 'LocationPermission', 'NetworkError', 'ContentUnavailable']) {
  expect(navigator, `name="${screen}"`, `Missing required user path: ${screen}`);
}

const mapHome = read('src/app/screens/MapHomeScreen.tsx');
expect(mapHome, '沿途服务', 'Map screen must expose Alongtu services');
expect(mapHome, 'onMapPress={() => setServiceOpen(false)}', 'A map tap must collapse the services panel through the native map event');
expectNot(mapHome, 'mapDecorations', 'Native map must not render legacy demo decorations above the map');
expectNot(mapHome, 'mapTapShield', 'Native map must not use a full-screen touch shield that blocks gestures');
expect(read('src/app/components/TravelMapSurface.tsx'), 'onPress={onMapPress}', 'Native AMap press event must be forwarded to the map screen');
for (const category of ['toilet', 'fuel', 'charge', 'market', 'sight', 'hotel', 'parking', 'powerbank', 'pharmacy', 'hospital', 'convenience', 'police', 'food']) {
  expect(read('src/app/demoData.ts'), `category: '${category}'`, `Missing demo place for service category: ${category}`);
}
expect(read('src/app/screens/DiscoverScreen.tsx'), '热门标注分享', 'Discover must expose public community channels');
expect(read('src/app/screens/DiscoverScreen.tsx'), '出发前看看', 'Discover must render a travel-context notice above the reading feed');
expect(read('src/app/screens/DiscoverScreen.tsx'), 'style={styles.channelScroller}', 'Public channel scroller must reserve its own fixed height');
expect(read('src/app/screens/DiscoverScreen.tsx'), '<SafeAreaView', 'Discover header must respect the Android status-bar safe area');
expect(read('src/app/screens/AssetScreens.tsx'), '公开标注需补充审核材料', 'Marker creation must describe review requirements');
expect(read('src/app/screens/AssetScreens.tsx'), 'useMarkerStore', 'Marker creation and management must share local marker state');
expect(read('src/app/screens/MapHomeScreen.tsx'), '标注此处', 'Map home must provide a direct marker creation entry');
expect(read('src/navigation/RootNavigator.tsx'), 'tabIcons', 'Bottom navigation must use the unified icon mapping');
for (const component of ['src/ui/AppIcon.tsx', 'src/ui/MapMarker.tsx', 'src/ui/ServiceIcon.tsx', 'src/ui/TrustBadge.tsx']) {
  if (!fs.existsSync(path.join(projectRoot, component))) throw new Error(`Missing shared UI component: ${component}`);
}
for (const dependency of ['lucide-react-native', 'react-native-svg']) {
  if (!packageJson.dependencies[dependency]) throw new Error(`Missing UI dependency: ${dependency}`);
}
expect(read('src/app/screens/SocialScreens.tsx'), '发起者退出后，本次共享立即结束', 'Location sharing exit rule is missing');
expect(read('src/app/screens/AccountScreens.tsx'), '游客浏览不受影响', 'Guest-first login behavior is missing');

console.log('UI route coverage verification passed.');
