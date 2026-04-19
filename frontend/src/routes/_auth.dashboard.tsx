import { createFileRoute } from '@tanstack/react-router'
import {
  AlertTriangle,
  ArrowUpRight,
  ChevronDown,
  ChevronRight,
  ClipboardCheck,
  DollarSign,
  Download,
  Package,
  RefreshCw,
  ShoppingCart,
  TrendingUp,
  Undo2,
} from 'lucide-react'
import { useMemo, useState } from 'react'

import { useAuthStore } from '@/shared/auth/store'

export const Route = createFileRoute('/_auth/dashboard')({
  component: DashboardPage,
})

type ChartView = 'day' | 'hour' | 'weekday'

const CHART_DATA: Record<
  ChartView,
  { labels: string[]; values: number[]; fmt: (v: number) => string; yMax: number; yStep: number }
> = {
  day: {
    labels: ['01','02','03','04','05','06','07','08','09','10','11','12','13','14','15','16','17','18','19'],
    values: [108, 95, 102, 88, 128, 145, 162, 155, 140, 198, 170, 158, 172, 190, 68, 80, 70, 92, 132],
    fmt: (v) => `$${v}k`,
    yMax: 300,
    yStep: 60,
  },
  hour: {
    labels: ['7a','8a','9a','10a','11a','12p','1p','2p','3p','4p','5p','6p'],
    values: [42, 68, 95, 122, 186, 248, 265, 204, 158, 142, 118, 76],
    fmt: (v) => `$${v}`,
    yMax: 300,
    yStep: 60,
  },
  weekday: {
    labels: ['Mon','Tue','Wed','Thu','Fri','Sat','Sun'],
    values: [3820, 4120, 3940, 4510, 5280, 6150, 4290],
    fmt: (v) => `$${(v / 1000).toFixed(1)}k`,
    yMax: 7000,
    yStep: 1400,
  },
}

const TOP_ITEMS = [
  { n: 'Cappuccino (Large · Oat)', v: 1530 },
  { n: 'Latte (Regular)', v: 829 },
  { n: 'Cold brew (16oz)', v: 789 },
  { n: 'Matcha latte', v: 758 },
  { n: 'Avocado toast', v: 651 },
  { n: 'Espresso (double)', v: 689 },
  { n: 'Croissant', v: 676 },
  { n: 'Blueberry muffin', v: 478 },
  { n: 'Chai latte', v: 451 },
  { n: 'Americano', v: 424 },
]

const TOP_CUST = [
  { n: 'Alex Nguyen', v: 3879 },
  { n: 'Maya Ramirez', v: 1588 },
  { n: 'Jordan Park', v: 689 },
  { n: 'Priya Shah', v: 668 },
  { n: 'Leo Tran', v: 647 },
  { n: 'Sam Patel', v: 601 },
  { n: 'Chris Diaz', v: 586 },
  { n: 'Taylor Kim', v: 540 },
  { n: 'Jamie Chen', v: 526 },
  { n: 'Dana Smith', v: 481 },
]

type Activity = {
  kind: 'sale' | 'stock' | 'check'
  who: string
  what: string
  amt: string
  t: string
}

const ACTIVITY: Activity[] = [
  { kind: 'sale', who: 'Quyen', what: 'placed an order', amt: '$268.00', t: '7 hours ago' },
  { kind: 'sale', who: 'Vu', what: 'placed an order', amt: '$352.00', t: '9 hours ago' },
  { kind: 'sale', who: 'Nhat Phat', what: 'placed an order', amt: '$7,643.00', t: '1 day ago' },
  { kind: 'check', who: 'Quyen', what: 'ran inventory check', amt: '—', t: '1 day ago' },
  { kind: 'sale', who: 'Quyen', what: 'placed an order', amt: '$926.00', t: '1 day ago' },
  { kind: 'sale', who: 'Nhat Phat', what: 'placed an order', amt: '$25,088.00', t: '1 day ago' },
  { kind: 'sale', who: 'Quyen', what: 'placed an order', amt: '$17,725.00', t: '1 day ago' },
  { kind: 'sale', who: 'Vu', what: 'placed an order', amt: '$4,130.00', t: '1 day ago' },
  { kind: 'stock', who: 'Nga', what: 'restocked products', amt: '$151,985.71', t: '1 day ago' },
  { kind: 'sale', who: 'Nhat Phat', what: 'placed an order', amt: '$665.00', t: '1 day ago' },
  { kind: 'stock', who: 'Nga', what: 'restocked products', amt: '$4,320.00', t: '1 day ago' },
  { kind: 'sale', who: 'Nhat Phat', what: 'placed an order', amt: '$1,200.00', t: '2 days ago' },
]

const BLUE = 'hsl(221 83% 53%)'
const BLUE_SOFT = 'hsl(214 100% 96.7%)'
const BLUE_INK = 'hsl(224 76% 48%)'
const LOW_SOFT = 'hsl(33.3 100% 96.5%)'
const LOW_INK = 'hsl(22.7 82.5% 45.1%)'
const LOW_BORDER = 'hsl(24 95% 53% / 0.25)'
const IDLE_SOFT = 'hsl(210 40% 96.1%)'
const IDLE_INK = 'hsl(215.3 25% 26.7%)'
const DONE_SOFT = 'hsl(138.5 76.5% 96.7%)'
const DONE_INK = 'hsl(142.1 70.6% 29.2%)'
const BORDER = 'hsl(214.3 31.8% 91.4%)'

function formatDate() {
  return new Date().toLocaleDateString('en-US', {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  })
}

function DashboardPage() {
  const user = useAuthStore((s) => s.user)
  if (!user) return null
  const displayName = user.name || user.email?.split('@')[0] || 'there'

  return (
    <div className="-m-6 p-6">
      <PageHead name={displayName} />
      <div className="grid grid-cols-1 items-start gap-4 xl:grid-cols-[minmax(0,1fr)_320px]">
        <div className="flex min-w-0 flex-col gap-4">
          <KpiCard />
          <ChartCard />
          <TopTenRow />
        </div>
        <aside className="flex flex-col gap-4 xl:sticky xl:top-4">
          <PromoCard
            tone="pay"
            title="QR Payments"
            sub="Accept instantly — 0% fees for 90 days"
          />
          <PromoCard
            tone="loan"
            title="Working capital"
            sub="Small loans, fast — get pre-qualified"
          />
          <AlertCard />
          <ActivityCard />
        </aside>
      </div>
    </div>
  )
}

function PageHead({ name }: { name: string }) {
  return (
    <div className="mb-5 flex items-end justify-between gap-4">
      <div>
        <h1 className="m-0 text-[22px] font-bold tracking-[-0.01em]">
          Good afternoon, {name}
        </h1>
        <div className="mt-0.5 text-[13px] text-[hsl(215.4_16.3%_46.9%)]">
          Here's what's happening today · {formatDate()}
        </div>
      </div>
      <div className="flex gap-2">
        <OutlineBtn icon={<Download className="h-3.5 w-3.5" />}>Export</OutlineBtn>
        <OutlineBtn icon={<RefreshCw className="h-3.5 w-3.5" />}>Refresh</OutlineBtn>
      </div>
    </div>
  )
}

function OutlineBtn({ children, icon }: { children: React.ReactNode; icon?: React.ReactNode }) {
  return (
    <button
      type="button"
      className="inline-flex h-9 items-center gap-1.5 rounded-md border border-[hsl(214.3_31.8%_91.4%)] bg-white px-3.5 text-[13px] font-medium text-[hsl(215.4_16.3%_46.9%)] transition hover:bg-[hsl(210_40%_96%)] hover:text-[hsl(222.2_84%_4.9%)]"
    >
      {icon}
      {children}
    </button>
  )
}

function SelectBtn({ children }: { children: React.ReactNode }) {
  return (
    <button
      type="button"
      className="inline-flex items-center gap-1.5 rounded-md border border-[hsl(214.3_31.8%_91.4%)] bg-white px-2.5 py-1.5 text-[12px] font-medium text-[hsl(222.2_84%_4.9%)] hover:bg-[hsl(210_40%_96%)]"
    >
      {children}
      <ChevronDown className="h-3 w-3 text-[hsl(215.4_16.3%_46.9%)]" />
    </button>
  )
}

function DCard({
  children,
  className = '',
  style,
}: {
  children: React.ReactNode
  className?: string
  style?: React.CSSProperties
}) {
  return (
    <section
      className={`rounded-lg border bg-white ${className}`}
      style={{ borderColor: BORDER, ...style }}
    >
      {children}
    </section>
  )
}

function KpiCard() {
  return (
    <DCard className="px-[18px] py-4">
      <div className="flex items-center justify-between gap-2 pb-3 pt-0.5">
        <h3 className="m-0 text-[14px] font-semibold">Today's results</h3>
        <span
          className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-[3px] text-[12px] font-medium"
          style={{ background: DONE_SOFT, color: DONE_INK }}
        >
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2.5} strokeLinecap="round" strokeLinejoin="round" className="h-3 w-3">
            <circle cx="12" cy="12" r="10" />
            <polyline points="12 6 12 12 16 14" />
          </svg>
          Live · updated 12s ago
        </span>
      </div>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4">
        <Kpi
          icon={<DollarSign className="h-4 w-4" />}
          iconBg={BLUE_SOFT}
          iconColor={BLUE_INK}
          label="Revenue"
          value="$4,287"
          unit=".40"
          sub="142 orders · 4 registers"
        />
        <Kpi
          icon={<Undo2 className="h-4 w-4" />}
          iconBg={LOW_SOFT}
          iconColor={LOW_INK}
          label="Refunds"
          value="$18"
          unit=".50"
          sub="1 order · Alex"
          divider
        />
        <Kpi
          icon={<ArrowUpRight className="h-4 w-4" />}
          iconBg={IDLE_SOFT}
          iconColor={IDLE_INK}
          label="Net revenue"
          value="$4,268"
          unit=".90"
          subColor="hsl(0 84% 60%)"
          subNode={
            <>
              <TrendingUp className="h-3 w-3" style={{ transform: 'scaleY(-1)' }} />
              −2.1% vs yesterday
            </>
          }
          divider
        />
        <Kpi
          icon={<TrendingUp className="h-4 w-4" />}
          iconBg={DONE_SOFT}
          iconColor={DONE_INK}
          label="Net margin"
          value="29"
          unit=".4%"
          subColor={DONE_INK}
          subNode={
            <>
              <TrendingUp className="h-3 w-3" />
              +4.2% vs last week
            </>
          }
          divider
        />
      </div>
    </DCard>
  )
}

function Kpi({
  icon,
  iconBg,
  iconColor,
  label,
  value,
  unit,
  sub,
  subNode,
  subColor,
  divider,
}: {
  icon: React.ReactNode
  iconBg: string
  iconColor: string
  label: string
  value: string
  unit?: string
  sub?: string
  subNode?: React.ReactNode
  subColor?: string
  divider?: boolean
}) {
  return (
    <div
      className="flex items-start gap-3 px-4 py-1 lg:pl-4 lg:pr-5"
      style={{
        borderLeft: divider ? `1px dashed ${BORDER}` : undefined,
      }}
    >
      <div
        className="grid h-[34px] w-[34px] flex-shrink-0 place-items-center rounded-lg"
        style={{ background: iconBg, color: iconColor }}
      >
        {icon}
      </div>
      <div className="min-w-0">
        <div className="text-[12px] font-medium text-[hsl(215.4_16.3%_46.9%)]">
          {label}
        </div>
        <div className="mt-0.5 flex items-baseline gap-1 text-[22px] font-bold leading-[1.1] tracking-[-0.015em] tabular-nums">
          {value}
          {unit && (
            <span className="text-[13px] font-medium text-[hsl(215.4_16.3%_46.9%)]">
              {unit}
            </span>
          )}
        </div>
        <div
          className="mt-1 inline-flex items-center gap-1 text-[11.5px]"
          style={{ color: subColor ?? 'hsl(215.4 16.3% 46.9%)' }}
        >
          {subNode ?? sub}
        </div>
      </div>
    </div>
  )
}

function ChartCard() {
  const [view, setView] = useState<ChartView>('day')
  const data = CHART_DATA[view]
  const peak = Math.max(...data.values)
  const ySteps = useMemo(() => {
    const arr: number[] = []
    for (let v = data.yMax; v >= 0; v -= data.yStep) arr.push(v)
    return arr
  }, [data.yMax, data.yStep])
  const yLabel = (v: number) => {
    if (view === 'day') return `$${v}k`
    if (view === 'hour') return `$${v}`
    return `$${(v / 1000).toFixed(0)}k`
  }

  return (
    <DCard className="px-[18px] py-4">
      <div className="flex items-center justify-between px-0.5 pb-2.5 pt-1">
        <div className="flex items-baseline gap-3">
          <span className="text-[14px] font-semibold">Net revenue</span>
          <span
            className="rounded-md px-2.5 py-0.5 text-[18px] font-bold tabular-nums tracking-[-0.01em]"
            style={{ background: BLUE_SOFT, color: BLUE }}
          >
            $2,125,485
          </span>
        </div>
        <SelectBtn>This month</SelectBtn>
      </div>

      <div
        className="mx-0.5 mb-3.5 mt-1.5 flex items-center gap-0.5 border-b pb-1"
        style={{ borderColor: BORDER }}
      >
        {(['day', 'hour', 'weekday'] as ChartView[]).map((v) => (
          <button
            key={v}
            type="button"
            onClick={() => setView(v)}
            className="-mb-px cursor-pointer border-b-2 px-3.5 pb-2.5 pt-2 text-[13px] transition"
            style={{
              color: view === v ? BLUE : 'hsl(215.4 16.3% 46.9%)',
              borderBottomColor: view === v ? BLUE : 'transparent',
              fontWeight: view === v ? 600 : 500,
            }}
          >
            {v === 'day' ? 'By day' : v === 'hour' ? 'By hour' : 'By weekday'}
          </button>
        ))}
      </div>

      <div className="relative mr-1 h-[280px] py-1.5 pl-12 pr-2">
        <div className="absolute bottom-[22px] left-0 top-0 flex w-10 flex-col justify-between pr-1.5 text-right font-mono text-[10.5px] text-[hsl(215.4_16.3%_46.9%)]">
          {ySteps.map((v) => (
            <div key={v}>{yLabel(v)}</div>
          ))}
        </div>
        <div className="absolute bottom-[22px] left-12 right-1 top-0 flex flex-col justify-between">
          {ySteps.map((v) => (
            <div
              key={v}
              className="h-0 border-t border-dashed"
              style={{ borderColor: BORDER }}
            />
          ))}
        </div>
        <div className="relative z-10 flex h-[calc(100%-22px)] items-end justify-between gap-1.5">
          {data.values.map((v, i) => {
            const h = Math.max(2, (v / data.yMax) * 100)
            const isPeak = v === peak
            return (
              <div
                key={`${view}-${i}`}
                className="group relative flex h-full min-w-0 flex-1 cursor-pointer flex-col items-center justify-end gap-1.5"
              >
                <div
                  className="w-full max-w-[28px] rounded-t"
                  style={{
                    height: `${h}%`,
                    background: isPeak
                      ? `linear-gradient(180deg, ${BLUE}, ${BLUE_INK})`
                      : BLUE,
                    transition: 'background 120ms',
                  }}
                />
                <div className="pointer-events-none absolute -top-7 left-1/2 -translate-x-1/2 whitespace-nowrap rounded-md bg-[hsl(222_47%_11%)] px-2 py-1 font-mono text-[11px] text-white opacity-0 shadow-lg transition group-hover:opacity-100">
                  {data.labels[i]} · {data.fmt(v)}
                </div>
                <div className="font-mono text-[10.5px] text-[hsl(215.4_16.3%_46.9%)]">
                  {data.labels[i]}
                </div>
              </div>
            )
          })}
        </div>
      </div>
    </DCard>
  )
}

function TopTenRow() {
  return (
    <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
      <TopList
        title="Top 10 best-selling items"
        right={
          <div className="flex items-center gap-2">
            <SelectBtn>By net revenue</SelectBtn>
            <SelectBtn>This month</SelectBtn>
          </div>
        }
        data={TOP_ITEMS}
        axis={['$0', '$400', '$800', '$1.2k', '$1.6k']}
      />
      <TopList
        title="Top 10 customers"
        right={<SelectBtn>This month</SelectBtn>}
        data={TOP_CUST}
        axis={['$0', '$1k', '$2k', '$3k', '$4k']}
      />
    </div>
  )
}

function TopList({
  title,
  right,
  data,
  axis,
}: {
  title: string
  right: React.ReactNode
  data: Array<{ n: string; v: number }>
  axis: string[]
}) {
  const max = Math.max(...data.map((d) => d.v))
  const scaleMax = Math.ceil(max / 400) * 400
  return (
    <DCard className="px-[18px] py-4">
      <div className="flex items-center justify-between gap-2 pb-3 pt-0.5">
        <h3 className="m-0 text-[14px] font-semibold">{title}</h3>
        <div className="flex items-center gap-2">{right}</div>
      </div>
      <div className="py-0.5">
        {data.map((d, i) => (
          <div
            key={d.n}
            className="border-b py-2.5 last:border-none"
            style={{ borderColor: BORDER }}
          >
            <div className="flex items-center gap-2 text-[12.5px] font-medium">
              <span
                className="min-w-[22px] rounded px-1.5 py-px text-center font-mono text-[10.5px] text-[hsl(215.4_16.3%_46.9%)]"
                style={{ background: IDLE_SOFT }}
              >
                {String(i + 1).padStart(2, '0')}
              </span>
              <span className="truncate">{d.n}</span>
            </div>
            <div className="mt-1.5 flex items-center gap-2.5">
              <div
                className="relative h-3.5 flex-1 overflow-hidden rounded"
                style={{ background: BLUE_SOFT }}
              >
                <div
                  className="h-full rounded transition-[width] duration-[400ms] ease-out"
                  style={{
                    width: `${(d.v / scaleMax) * 100}%`,
                    background: BLUE,
                  }}
                />
              </div>
              <span className="min-w-[52px] text-right font-mono text-[11.5px] font-medium tabular-nums">
                ${d.v.toLocaleString()}
              </span>
            </div>
          </div>
        ))}
      </div>
      <div
        className="mt-1 flex justify-between border-t border-dashed pb-0.5 pl-0.5 pr-[52px] pt-2 font-mono text-[10px] text-[hsl(215.4_16.3%_46.9%)]"
        style={{ borderColor: BORDER }}
      >
        {axis.map((a) => (
          <span key={a}>{a}</span>
        ))}
      </div>
    </DCard>
  )
}

function PromoCard({
  tone,
  title,
  sub,
}: {
  tone: 'pay' | 'loan'
  title: string
  sub: string
}) {
  const bg = tone === 'pay' ? DONE_SOFT : LOW_SOFT
  const color = tone === 'pay' ? DONE_INK : LOW_INK
  const Icon = tone === 'pay' ? Package : DollarSign
  return (
    <DCard className="flex cursor-pointer items-center gap-3 px-3.5 py-3.5 transition hover:bg-[hsl(210_40%_96%)]">
      <div
        className="grid h-[38px] w-[38px] flex-shrink-0 place-items-center rounded-lg"
        style={{ background: bg, color }}
      >
        <Icon className="h-[18px] w-[18px]" strokeWidth={2} />
      </div>
      <div className="min-w-0 flex-1">
        <div className="text-[13px] font-semibold">{title}</div>
        <div className="mt-0.5 text-[11.5px] text-[hsl(215.4_16.3%_46.9%)]">
          {sub}
        </div>
      </div>
      <ChevronRight className="h-4 w-4 flex-shrink-0 text-[hsl(215.4_16.3%_46.9%)]" />
    </DCard>
  )
}

function AlertCard() {
  return (
    <DCard
      className="flex gap-3 px-3.5 py-3.5"
      style={{ background: LOW_SOFT, borderColor: LOW_BORDER }}
    >
      <div
        className="grid h-8 w-8 flex-shrink-0 place-items-center rounded-lg bg-white"
        style={{ color: LOW_INK }}
      >
        <AlertTriangle className="h-[15px] w-[15px]" strokeWidth={2} />
      </div>
      <div>
        <div className="text-[13px] leading-[1.45]" style={{ color: LOW_INK }}>
          We spotted <strong className="font-bold">14 unusual sign-in attempts</strong> that need review.
        </div>
        <a
          href="#"
          onClick={(e) => e.preventDefault()}
          className="mt-1 inline-block text-[12px] font-semibold underline"
          style={{ color: LOW_INK }}
        >
          Review now →
        </a>
      </div>
    </DCard>
  )
}

function ActivityCard() {
  return (
    <DCard className="px-4 py-3.5">
      <h3 className="m-0 mb-2.5 mt-0.5 text-[14px] font-semibold">
        Recent activity
      </h3>
      <div
        className="-mx-1.5 flex max-h-[720px] flex-col overflow-y-auto px-1.5"
      >
        {ACTIVITY.map((a, i) => (
          <ActivityItem key={i} a={a} last={i === ACTIVITY.length - 1} />
        ))}
      </div>
    </DCard>
  )
}

function ActivityItem({ a, last }: { a: Activity; last: boolean }) {
  const palette =
    a.kind === 'sale'
      ? { bg: BLUE_SOFT, color: BLUE_INK, Icon: ShoppingCart }
      : a.kind === 'stock'
        ? { bg: DONE_SOFT, color: DONE_INK, Icon: Package }
        : { bg: LOW_SOFT, color: LOW_INK, Icon: ClipboardCheck }
  const Icon = palette.Icon
  return (
    <div
      className="flex gap-2.5 px-1 py-3"
      style={{ borderBottom: last ? undefined : `1px solid ${BORDER}` }}
    >
      <div
        className="mt-0.5 grid h-7 w-7 flex-shrink-0 place-items-center rounded-full"
        style={{ background: palette.bg, color: palette.color }}
      >
        <Icon className="h-[13px] w-[13px]" strokeWidth={2} />
      </div>
      <div className="min-w-0 flex-1 text-[12.5px] leading-[1.5]">
        <span className="font-semibold">{a.who}</span>
        {a.kind === 'check' ? (
          <span className="font-medium" style={{ color: BLUE }}>
            {' '}
            {a.what}
          </span>
        ) : (
          <>
            <span> just </span>
            <span className="font-medium" style={{ color: BLUE }}>
              {a.what}
            </span>
            {a.amt !== '—' && (
              <>
                {' '}
                worth{' '}
                <span className="whitespace-nowrap font-semibold tabular-nums">
                  {a.amt}
                </span>
              </>
            )}
          </>
        )}
        <div className="mt-0.5 text-[11px] text-[hsl(215.4_16.3%_46.9%)]">
          {a.t}
        </div>
      </div>
    </div>
  )
}

