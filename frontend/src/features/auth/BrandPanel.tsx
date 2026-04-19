function BrandGlyph() {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={2.2}
      strokeLinecap="round"
      strokeLinejoin="round"
      className="h-4 w-4"
      aria-hidden
    >
      <rect x="3" y="4" width="18" height="16" rx="2" />
      <path d="M3 9h18" />
      <path d="M8 14h3" />
      <path d="M15 14h2" />
    </svg>
  )
}

function GlobeIcon() {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={2}
      strokeLinecap="round"
      strokeLinejoin="round"
      className="h-3 w-3"
      aria-hidden
    >
      <circle cx="12" cy="12" r="10" />
      <path d="M2 12h20" />
      <path d="M12 2a15 15 0 0 1 0 20" />
      <path d="M12 2a15 15 0 0 0 0 20" />
    </svg>
  )
}

function TrendUp() {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={2.5}
      strokeLinecap="round"
      strokeLinejoin="round"
      className="h-3 w-3"
      aria-hidden
    >
      <polyline points="23 6 13.5 15.5 8.5 10.5 1 18" />
      <polyline points="17 6 23 6 23 12" />
    </svg>
  )
}

const HOUR_BARS = [
  { h: '7a', v: 22 },
  { h: '8a', v: 38 },
  { h: '9a', v: 54 },
  { h: '10a', v: 44 },
  { h: '11a', v: 68 },
  { h: '12p', v: 92 },
  { h: '1p', v: 100 },
  { h: '2p', v: 76 },
  { h: '3p', v: 58 },
  { h: '4p', v: 62 },
  { h: '5p', v: 48 },
  { h: '6p', v: 30 },
]

export function BrandPanel() {
  const year = new Date().getFullYear()
  const max = Math.max(...HOUR_BARS.map((b) => b.v))
  const peak = HOUR_BARS.findIndex((b) => b.v === max)

  return (
    <section className="relative isolate hidden flex-col overflow-hidden bg-[#0B0D10] p-8 text-[hsl(210_40%_95%)] lg:flex lg:px-10 lg:py-8">
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 z-0"
        style={{
          backgroundImage:
            'linear-gradient(hsl(215 20% 100% / 0.035) 1px, transparent 1px), linear-gradient(90deg, hsl(215 20% 100% / 0.035) 1px, transparent 1px)',
          backgroundSize: '32px 32px',
          WebkitMaskImage:
            'radial-gradient(ellipse 80% 70% at 30% 40%, black 50%, transparent 100%)',
          maskImage:
            'radial-gradient(ellipse 80% 70% at 30% 40%, black 50%, transparent 100%)',
        }}
      />
      <div
        aria-hidden
        className="pointer-events-none absolute -left-40 -top-44 z-0 h-[560px] w-[560px] rounded-full blur-[20px]"
        style={{
          background:
            'radial-gradient(circle, hsl(221 83% 53% / 0.28) 0%, hsl(261 83% 58% / 0.08) 40%, transparent 70%)',
        }}
      />

      <div className="relative z-10 flex items-center justify-between">
        <div className="inline-flex items-center gap-2.5 text-[15px] font-bold tracking-tight">
          <span
            className="grid h-[30px] w-[30px] place-items-center rounded-lg text-white"
            style={{
              background: 'hsl(221 83% 53%)',
              boxShadow:
                '0 6px 16px hsl(221 83% 53% / 0.45), inset 0 1px 0 hsl(0 0% 100% / 0.25)',
            }}
          >
            <BrandGlyph />
          </span>
          GenPos
        </div>
        <div className="inline-flex items-center gap-1.5 rounded-full border border-white/10 bg-white/[0.06] px-2.5 py-1.5 text-[12px] font-medium text-[hsl(215_15%_75%)]">
          <GlobeIcon />
          EN · USD
        </div>
      </div>

      <div className="relative z-10 flex min-h-0 flex-1 flex-col justify-center py-8">
        <div className="mb-5 inline-flex items-center gap-2 text-[11.5px] font-semibold uppercase tracking-[0.12em] text-[hsl(215_20%_70%)]">
          <span
            className="h-[7px] w-[7px] animate-pulse rounded-full"
            style={{
              background: 'hsl(142 76% 36%)',
              boxShadow: '0 0 0 3px hsl(142 76% 36% / 0.25)',
            }}
          />
          Live · 4 registers up
        </div>

        <h1 className="m-0 mb-4 max-w-[520px] text-[44px] font-bold leading-[1.05] tracking-[-0.025em] text-balance">
          Every sale, every shift —{' '}
          <em
            className="not-italic"
            style={{
              background:
                'linear-gradient(120deg, hsl(221 83% 53%) 0%, hsl(261 95% 72%) 100%)',
              WebkitBackgroundClip: 'text',
              backgroundClip: 'text',
              color: 'transparent',
            }}
          >
            one register away.
          </em>
        </h1>
        <p className="mb-9 max-w-[440px] text-[15px] leading-[1.55] text-[hsl(215_15%_70%)] text-pretty">
          GenPos is the point-of-sale your team actually wants to use. Ring up
          orders in two taps, track inventory in real time, and close the day
          in under a minute.
        </p>

        <div
          className="relative w-full max-w-[520px] rounded-[14px] border p-6"
          style={{
            background:
              'linear-gradient(180deg, hsl(220 18% 12%) 0%, hsl(220 18% 9%) 100%)',
            borderColor: 'hsl(215 20% 100% / 0.06)',
            boxShadow: '0 30px 60px -20px hsl(222 47% 4% / 0.6)',
          }}
        >
          <div className="text-[11.5px] font-semibold uppercase tracking-[0.1em] text-[hsl(215_15%_55%)]">
            Today's sales · Main st.
          </div>
          <div className="mt-1 text-[44px] font-bold leading-none tracking-[-0.025em] tabular-nums text-white">
            $4,287.
            <span className="text-[28px] text-[hsl(215_15%_55%)]">40</span>
          </div>
          <div
            className="mt-2.5 inline-flex items-center gap-1 whitespace-nowrap rounded-full px-2.5 py-[3px] text-[12px] font-semibold"
            style={{
              background: 'hsl(142 76% 36% / 0.12)',
              color: 'hsl(142 76% 36%)',
            }}
          >
            <TrendUp />
            +18.4% vs. last Tue
          </div>

          <div
            className="mt-7 grid grid-cols-12 items-end gap-1.5 border-b border-dashed pb-5"
            style={{
              height: 120,
              borderColor: 'hsl(215 20% 100% / 0.08)',
            }}
          >
            {HOUR_BARS.map((b, i) => {
              const h = Math.max(6, (b.v / max) * 100)
              const isPeak = i === peak
              return (
                <div
                  key={b.h}
                  className="rounded-t transition-opacity"
                  style={{
                    height: `${h}%`,
                    background:
                      'linear-gradient(180deg, hsl(221 83% 53%), hsl(261 83% 58%))',
                    opacity: isPeak ? 1 : 0.55,
                    boxShadow: isPeak
                      ? '0 0 20px hsl(221 83% 53% / 0.5)'
                      : undefined,
                    borderRadius: '4px 4px 2px 2px',
                  }}
                />
              )
            })}
          </div>
          <div className="mt-1.5 grid grid-cols-12 gap-1.5 text-center font-mono text-[9.5px] text-[hsl(215_15%_45%)]">
            {HOUR_BARS.map((b) => (
              <div key={b.h}>{b.h}</div>
            ))}
          </div>

          <div
            className="mt-5 grid grid-cols-3 gap-3.5 border-t pt-4"
            style={{ borderColor: 'hsl(215 20% 100% / 0.06)' }}
          >
            <div>
              <div className="text-[18px] font-semibold tabular-nums text-white">
                142
              </div>
              <div className="mt-0.5 text-[11px] text-[hsl(215_15%_55%)]">
                Orders today
              </div>
            </div>
            <div>
              <div className="text-[18px] font-semibold tabular-nums text-white">
                $30.19
              </div>
              <div className="mt-0.5 text-[11px] text-[hsl(215_15%_55%)]">
                Avg ticket
              </div>
            </div>
            <div>
              <div className="text-[18px] font-semibold tabular-nums text-white">
                4
              </div>
              <div className="mt-0.5 text-[11px] text-[hsl(215_15%_55%)]">
                Registers up
              </div>
            </div>
          </div>
        </div>
      </div>

      <div className="relative z-10 flex items-center justify-between text-[12px] text-[hsl(215_15%_55%)]">
        <span>© {year} GenPos</span>
        <div className="flex items-center gap-2">
          <a href="#" className="hover:text-[hsl(215_15%_85%)]">
            Privacy
          </a>
          <span className="opacity-40">·</span>
          <a href="#" className="hover:text-[hsl(215_15%_85%)]">
            Terms
          </a>
          <span className="opacity-40">·</span>
          <a href="#" className="hover:text-[hsl(215_15%_85%)]">
            Status
          </a>
        </div>
      </div>
    </section>
  )
}
