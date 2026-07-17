const { getDefaultConfig, mergeConfig } = require('@react-native/metro-config');

/**
 * Metro configuration
 * https://reactnative.dev/docs/metro
 *
 * @type {import('@react-native/metro-config').MetroConfig}
 */
// Android release bundling runs on machines that may already have Metro
// workers from other local projects. Keep this project to one transformer
// worker so the Windows Node process stays within its native resource limit.
const config = {
  maxWorkers: 1,
};

module.exports = mergeConfig(getDefaultConfig(__dirname), config);
