const fs = require('fs');
const path = require('path');

const root = path.resolve(__dirname, '..');
const read = (relativePath) => fs.readFileSync(path.join(root, relativePath), 'utf8');
const expect = (source, text, message) => {
  if (!source.includes(text)) throw new Error(message);
};
const expectNot = (source, text, message) => {
  if (source.includes(text)) throw new Error(message);
};

const client = read('src/p0/client.ts');
const session = read('src/p0/session.ts');
const auth = read('src/app/screens/AuthenticationScreen.tsx');
const detail = read('src/app/screens/PlaceDetailScreen.tsx');
const navigator = read('src/navigation/RootNavigator.tsx');
const profile = read('src/app/screens/AccountScreens.tsx');
const compose = read('../backend/docker-compose.frontend-test.yml');

expect(client, '/api/v1', 'P0 client must use the versioned API prefix');
expectNot(client, '/api/pois/', 'P0 client must not call removed FastAPI routes');
expect(client, 'P0RuntimeConfig', 'API origin must come from native build configuration');
expect(client, 'api_not_configured', 'Release without an API origin must fail clearly');
expect(session, 'react-native-keychain', 'Refresh tokens must use system secure storage');
expect(session, 'accessTokenForWrite', 'Authenticated writes must rotate expired access tokens');
expect(session, 'newIdempotencyKey', 'Place confirmation retries need one idempotency key');
expect(session, 'p0Api.logout', 'Logout must revoke the device session');
expect(auth, 'sendCode', 'Login must create an SMS challenge');
expect(auth, 'verifyCode', 'Login must verify the SMS challenge');
expectNot(auth, '微信登录', 'P0 must not claim an unimplemented WeChat login');
expect(detail, 'p0Api.confirm', 'Place detail must submit an authenticated confirmation');
expect(detail, 'confirmationKey', 'Confirmation retries must reuse the idempotency key');
expect(profile, 'useP0SessionStore', 'Profile must read the real P0 session');
expect(profile, 'signOut', 'Profile logout must revoke and clear the P0 session');
expect(navigator, 'UnavailableFeatureTab', 'P1–P3 local demos must stay unavailable in P0');
expectNot(navigator, 'useP0SessionStore', 'P0 session must not unlock P1–P3 demos');
expect(compose, 'env_file: !reset []', 'Frontend E2E must not inherit backend/.env credentials');
console.log('P0 contract verification passed.');
