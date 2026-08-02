// Generate the production service worker with workbox-build.
// Precaches every built static asset so the dashboard works offline;
// API calls (/api, /pg, /mj) are never cached and always hit the network.
import { generateSW } from 'workbox-build'

const result = await generateSW({
  globDirectory: 'dist',
  globPatterns: [
    '**/*.html',
    '**/*.js',
    '**/*.css',
    '**/*.png',
    '**/*.svg',
    '**/*.ico',
    '**/*.woff',
    '**/*.woff2',
    '**/*.ttf',
    '**/*.eot',
  ],
  globIgnores: ['sw.js', 'workbox-*.js', '**/*.LICENSE.txt'],
  swDest: 'dist/sw.js',
  navigateFallback: '/index.html',
  navigateFallbackDenylist: [/^\/api\//, /^\/pg\//, /^\/mj\//],
  skipWaiting: true,
  clientsClaim: true,
  cleanupOutdatedCaches: true,
  // Largest built chunk is ~6.7 MB (route-level code splitting); raise the
  // workbox default (2 MB) so every route is available offline.
  maximumFileSizeToCacheInBytes: 8 * 1024 * 1024,
})

if (result.warnings.length > 0) {
  console.warn('workbox warnings:\n' + result.warnings.join('\n'))
}
console.log(
  `Generated dist/sw.js — precaching ${result.count} entries (${(result.size / 1024 / 1024).toFixed(2)} MiB)`
)
