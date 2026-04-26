import { Apple, Check, Download, HelpCircle, Monitor, Smartphone } from 'lucide-react'
import { useEffect, useState } from 'react'

import { Button } from '@/shared/ui/button'

import {
  APP_VERSION,
  DOWNLOADS,
  detectDevice,
  platformLabel,
  type DetectedDevice,
  type Platform,
} from './platform'

export function DownloadsPage() {
  const [device, setDevice] = useState<DetectedDevice>({
    platform: 'unknown',
    arch: 'unknown',
  })

  useEffect(() => {
    setDevice(
      detectDevice(typeof navigator === 'undefined' ? undefined : navigator.userAgent),
    )
  }, [])

  const recommended: 'windows' | 'macos' =
    device.platform === 'macos' ? 'macos' : 'windows'

  return (
    <div className="-m-6 p-6">
      <PageHead />
      <DetectionBanner device={device} recommended={recommended} />

      <div className="mb-3 mt-6 text-[14px] font-semibold">Desktop</div>
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <PlatformCard
          platform="windows"
          title="Windows"
          requirement="Windows 10 or later · x64"
          file={DOWNLOADS.windows.file}
          size={DOWNLOADS.windows.size}
          href={DOWNLOADS.windows.href}
          isRecommended={recommended === 'windows'}
        />
        <PlatformCard
          platform="macos"
          title="macOS"
          requirement="macOS 12+ · Apple Silicon"
          file={DOWNLOADS.macos.file}
          size={DOWNLOADS.macos.size}
          href={DOWNLOADS.macos.href}
          isRecommended={recommended === 'macos'}
        />
      </div>

      <div className="mb-3 mt-8 text-[14px] font-semibold">Mobile</div>
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <ComingSoonCard platform="android" title="Android" eta="Q3 2026" />
        <ComingSoonCard platform="ios" title="iOS & iPadOS" eta="Q4 2026" />
      </div>

      <HelpRow />
    </div>
  )
}

function HelpRow() {
  return (
    <div className="mt-8 grid grid-cols-1 gap-4 rounded-lg border bg-[color:var(--color-card)] p-5 md:grid-cols-3">
      <HelpItem
        title="Sign in once"
        body="Use your same GenPos account on the desktop app — your registers, products, and reports sync automatically."
      />
      <HelpItem
        title="Hardware ready"
        body="Star and Epson receipt printers, USB and Bluetooth scanners, cash drawers, and Stripe terminals work out of the box."
      />
      <HelpItem
        title="Works offline"
        body="Keep ringing up sales when the internet drops. Everything syncs the moment you're back online."
      />
    </div>
  )
}

function HelpItem({ title, body }: { title: string; body: string }) {
  return (
    <div>
      <div className="mb-1 text-[13px] font-semibold">{title}</div>
      <p className="m-0 text-[12.5px] leading-[1.5] text-[color:var(--color-muted-foreground)]">
        {body}
      </p>
    </div>
  )
}

function PageHead() {
  return (
    <div className="mb-5 flex items-end justify-between gap-4">
      <div>
        <h1 className="m-0 text-[22px] font-bold tracking-[-0.01em]">
          Get the GenPos app
        </h1>
        <p className="mt-1 text-[13px] text-[color:var(--color-muted-foreground)]">
          Native desktop apps for the counter and back office. Version{' '}
          <span className="font-mono">{APP_VERSION}</span>.
        </p>
      </div>
      <div className="flex gap-2">
        <Button variant="outline" size="sm" asChild>
          <a
            href={`https://github.com/tranhoangvuit/genpos-desk/releases/tag/v${APP_VERSION}`}
            target="_blank"
            rel="noreferrer"
          >
            Release notes
          </a>
        </Button>
        <Button variant="outline" size="sm">
          <HelpCircle className="mr-1.5 h-4 w-4" />
          Setup help
        </Button>
      </div>
    </div>
  )
}

function DetectionBanner({
  device,
  recommended,
}: {
  device: DetectedDevice
  recommended: 'windows' | 'macos'
}) {
  const dl = DOWNLOADS[recommended]
  const detected =
    device.platform === 'unknown' ? 'Detecting your device…' : platformLabel(device.platform)
  const Icon = recommended === 'macos' ? Apple : Monitor

  return (
    <div className="flex flex-col items-start gap-3 rounded-lg border-l-4 border-l-[color:var(--color-foreground)] border-y border-r bg-[color:var(--color-card)] px-4 py-3 sm:flex-row sm:items-center">
      <div className="grid h-10 w-10 flex-shrink-0 place-items-center rounded-md bg-[color:var(--color-muted)] text-[color:var(--color-foreground)]">
        <Icon className="h-5 w-5" />
      </div>
      <div className="min-w-0 flex-1">
        <div className="text-[11px] uppercase tracking-wider text-[color:var(--color-muted-foreground)]">
          Detected on this device
        </div>
        <div className="text-[14px] font-medium">{detected}</div>
      </div>
      <Button asChild size="sm">
        <a href={dl.href}>
          <Download className="mr-1.5 h-4 w-4" />
          {dl.label}
        </a>
      </Button>
    </div>
  )
}

function PlatformCard({
  platform,
  title,
  requirement,
  file,
  size,
  href,
  isRecommended,
}: {
  platform: Platform
  title: string
  requirement: string
  file: string
  size: string
  href: string
  isRecommended: boolean
}) {
  return (
    <article className="flex flex-col rounded-lg border bg-[color:var(--color-card)] p-4">
      <div className="mb-3 flex items-center gap-3">
        <PlatformIcon platform={platform} />
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <span className="text-[15px] font-semibold">{title}</span>
            {isRecommended && (
              <span className="inline-flex items-center gap-1 rounded-full bg-[color:var(--color-secondary)] px-2 py-0.5 text-[10.5px] font-medium text-[color:var(--color-secondary-foreground)]">
                <Check className="h-3 w-3" />
                Detected
              </span>
            )}
          </div>
          <div className="truncate text-[12px] text-[color:var(--color-muted-foreground)]">
            {requirement}
          </div>
        </div>
      </div>
      <div className="mb-3 flex items-baseline justify-between gap-2 border-t pt-2.5 text-[11.5px] text-[color:var(--color-muted-foreground)]">
        <span className="truncate font-mono">{file}</span>
        <span className="flex-shrink-0 font-mono">{size}</span>
      </div>
      <Button
        asChild
        size="sm"
        variant={isRecommended ? 'default' : 'outline'}
        className="w-full"
      >
        <a href={href}>
          <Download className="mr-1.5 h-4 w-4" />
          Download
        </a>
      </Button>
    </article>
  )
}

function ComingSoonCard({
  platform,
  title,
  eta,
}: {
  platform: 'android' | 'ios'
  title: string
  eta: string
}) {
  return (
    <article className="flex items-center gap-3 rounded-lg border border-dashed bg-[color:var(--color-card)] p-5">
      <PlatformIcon platform={platform} muted />
      <div className="flex-1">
        <div className="text-[14px] font-medium">{title}</div>
        <div className="text-[12px] text-[color:var(--color-muted-foreground)]">
          Coming {eta}
        </div>
      </div>
      <span className="rounded-md border px-2 py-1 text-[11px] font-medium text-[color:var(--color-muted-foreground)]">
        Soon
      </span>
    </article>
  )
}

function PlatformIcon({
  platform,
  muted,
}: {
  platform: Platform
  muted?: boolean
}) {
  const cls = muted
    ? 'grid h-9 w-9 place-items-center rounded-md bg-[color:var(--color-muted)] text-[color:var(--color-muted-foreground)]'
    : 'grid h-9 w-9 place-items-center rounded-md bg-[color:var(--color-muted)] text-[color:var(--color-foreground)]'
  if (platform === 'windows') {
    return (
      <span className={cls}>
        <Monitor className="h-5 w-5" strokeWidth={1.8} />
      </span>
    )
  }
  if (platform === 'android') {
    return (
      <span className={cls}>
        <Smartphone className="h-5 w-5" strokeWidth={1.8} />
      </span>
    )
  }
  return (
    <span className={cls}>
      <Apple className="h-5 w-5" />
    </span>
  )
}
