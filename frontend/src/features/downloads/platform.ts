export type Platform = 'windows' | 'macos' | 'android' | 'ios' | 'linux' | 'unknown'

export type Arch = 'x64' | 'arm64' | 'unknown'

export type DetectedDevice = {
  platform: Platform
  arch: Arch
}

export const APP_VERSION = '0.1.3'

export const DOWNLOADS = {
  windows: {
    label: 'Download for Windows',
    href: `https://github.com/tranhoangvuit/genpos-desk/releases/download/v${APP_VERSION}/genpos_0.1.2_x64-setup.exe`,
    file: `genpos_0.1.2_x64-setup.exe`,
    size: '84 MB',
  },
  macos: {
    label: 'Download for macOS',
    href: `https://github.com/tranhoangvuit/genpos-desk/releases/download/v${APP_VERSION}/genpos_0.1.2_aarch64.dmg`,
    file: `genpos_0.1.2_aarch64.dmg`,
    size: '96 MB',
  },
} as const

export function detectDevice(userAgent: string | undefined): DetectedDevice {
  if (!userAgent) return { platform: 'unknown', arch: 'unknown' }
  const ua = userAgent.toLowerCase()

  let platform: Platform = 'unknown'
  if (/iphone|ipad|ipod/.test(ua)) platform = 'ios'
  else if (/android/.test(ua)) platform = 'android'
  else if (/mac os x|macintosh/.test(ua)) platform = 'macos'
  else if (/windows/.test(ua)) platform = 'windows'
  else if (/linux/.test(ua)) platform = 'linux'

  let arch: Arch = 'unknown'
  if (/arm64|aarch64|apple silicon/.test(ua)) arch = 'arm64'
  else if (/x86_64|win64|wow64|x64/.test(ua)) arch = 'x64'

  // macOS reports Intel UA on Apple Silicon — treat as arm64 by default for the dmg.
  if (platform === 'macos' && arch === 'unknown') arch = 'arm64'

  return { platform, arch }
}

export function platformLabel(p: Platform): string {
  switch (p) {
    case 'windows': return 'Windows · 64-bit'
    case 'macos': return 'macOS · Apple Silicon'
    case 'android': return 'Android'
    case 'ios': return 'iOS'
    case 'linux': return 'Linux'
    default: return 'Your device'
  }
}
