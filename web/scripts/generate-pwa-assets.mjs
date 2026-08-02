// Generate PWA icons from the existing logo.png (180x180, dark background).
// Outputs:
//   public/pwa-192x192.png          — standard icon (any purpose)
//   public/pwa-512x512.png          — standard icon (any purpose)
//   public/pwa-maskable-512x512.png — maskable icon (content inside 80% safe zone)
import sharp from 'sharp'

const SRC = 'public/logo.png'
const OUT = 'public/'

// 1) Standard icons: fill the whole square (keeps the dark background)
for (const size of [192, 512]) {
  await sharp(SRC)
    .resize(size, size)
    .png()
    .toFile(`${OUT}pwa-${size}x${size}.png`)
}

// 2) Maskable icon: 512 canvas with dark background, logo scaled to ~66% and
//    centered so the ring stays inside the 80% safe zone used by platforms
//    that crop icons into circles / rounded squares.
const MASK = 512
const content = Math.round(MASK * 0.66)
const offset = Math.round((MASK - content) / 2)
await sharp({
  create: {
    width: MASK,
    height: MASK,
    channels: 4,
    background: { r: 0, g: 0, b: 0, alpha: 1 },
  },
})
  .composite([
    {
      input: await sharp(SRC).resize(content, content).png().toBuffer(),
      left: offset,
      top: offset,
    },
  ])
  .png()
  .toFile(`${OUT}pwa-maskable-512x512.png`)

console.log('PWA icons generated (192, 512, maskable-512)')
