import { ArrowRight, X } from 'lucide-react'
import { useEffect, useLayoutEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'

const BORDER = 'hsl(214.3 31.8% 91.4%)'
const MUTED = 'hsl(210 40% 96.1%)'
const MUTED_FG = 'hsl(215.4 16.3% 46.9%)'
const FG = 'hsl(222.2 84% 4.9%)'
const RED = 'hsl(0 74% 42%)'
const RED_SOFT = 'hsl(0 93% 94%)'
const DONE_SOFT = 'hsl(138.5 76.5% 96.7%)'
const DONE_INK = 'hsl(142.1 70.6% 29.2%)'
const RING = 'hsl(221 83% 53%)'

const POP_W = 340
const GAP = 12

export type AdjustReason =
  | 'count'
  | 'received'
  | 'damaged'
  | 'loss'
  | 'return'
  | 'other'

const REASON_KEYS: AdjustReason[] = [
  'count',
  'received',
  'damaged',
  'loss',
  'return',
  'other',
]

export function AdjustOnHandPopover({
  productLabel,
  from,
  to,
  anchor,
  onCancel,
  onSave,
}: {
  productLabel: string
  from: number
  to: number
  anchor: DOMRect
  onCancel: () => void
  onSave: (reason: AdjustReason, note: string) => void
}) {
  const { t } = useTranslation()
  const [reason, setReason] = useState<AdjustReason>('count')
  const [note, setNote] = useState('')
  const popRef = useRef<HTMLDivElement>(null)
  const noteRef = useRef<HTMLTextAreaElement>(null)
  const [pos, setPos] = useState<{ left: number; top: number } | null>(null)

  const delta = to - from

  useLayoutEffect(() => {
    const el = popRef.current
    if (!el) return
    const ph = el.offsetHeight || 240

    // Prefer LEFT of the anchored input — easier scan, doesn't push off-screen.
    let left = anchor.left - POP_W - GAP
    if (left < 8) left = anchor.right + GAP
    if (left + POP_W > window.innerWidth - 8) {
      left = Math.max(8, window.innerWidth - POP_W - 8)
    }
    let top = anchor.top + anchor.height / 2 - ph / 2
    if (top < 8) top = 8
    if (top + ph > window.innerHeight - 8) top = window.innerHeight - ph - 8
    setPos({ left, top })
  }, [anchor])

  useEffect(() => {
    noteRef.current?.focus()
  }, [])

  useEffect(() => {
    function onDown(e: MouseEvent) {
      if (popRef.current && !popRef.current.contains(e.target as Node)) {
        onCancel()
      }
    }
    function onKey(e: KeyboardEvent) {
      if (e.key === 'Escape') onCancel()
    }
    document.addEventListener('mousedown', onDown)
    document.addEventListener('keydown', onKey)
    return () => {
      document.removeEventListener('mousedown', onDown)
      document.removeEventListener('keydown', onKey)
    }
  }, [onCancel])

  function save() {
    onSave(reason, note.trim())
  }

  return (
    <div
      ref={popRef}
      role="dialog"
      aria-label={t('inventory.adjustOnHand', 'Adjust on hand')}
      className="fixed z-[200] rounded-lg border bg-white p-3.5 transition"
      style={{
        width: POP_W,
        left: pos?.left ?? anchor.left,
        top: pos?.top ?? anchor.top,
        opacity: pos ? 1 : 0,
        visibility: pos ? 'visible' : 'hidden',
        borderColor: BORDER,
        boxShadow: '0 12px 40px hsl(222 47% 11% / 0.14)',
      }}
    >
      <div className="mb-2.5 flex items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="text-[13px] font-semibold" style={{ color: FG }}>
            {t('inventory.adjustOnHand', 'Adjust on hand')}
          </div>
          <div className="mt-px truncate text-[12px]" style={{ color: MUTED_FG }}>
            {productLabel}
          </div>
        </div>
        <button
          type="button"
          aria-label={t('common.cancel', 'Cancel')}
          onClick={onCancel}
          className="grid h-6 w-6 place-items-center rounded hover:bg-[hsl(210_40%_96%)]"
          style={{ color: MUTED_FG }}
        >
          <X className="h-3.5 w-3.5" />
        </button>
      </div>

      <div
        className="mb-3 flex items-center justify-between rounded-md px-3 py-2.5"
        style={{ background: 'hsl(210 40% 96.1% / 0.5)' }}
      >
        <div
          className="inline-flex items-center gap-2 text-[14px] font-medium tabular-nums"
          style={{ color: FG }}
        >
          <span>{from}</span>
          <ArrowRight className="h-3.5 w-3.5" style={{ color: MUTED_FG }} />
          <span>{to}</span>
        </div>
        <span
          className="rounded-full px-2 py-px text-[13px] font-semibold tabular-nums"
          style={{
            background: delta > 0 ? DONE_SOFT : delta < 0 ? RED_SOFT : MUTED,
            color: delta > 0 ? DONE_INK : delta < 0 ? RED : FG,
          }}
        >
          {delta > 0 ? `+${delta}` : delta}
        </span>
      </div>

      <div
        className="mb-1.5 text-[11px] font-semibold uppercase tracking-[0.06em]"
        style={{ color: MUTED_FG }}
      >
        {t('inventory.adjustReason', 'Reason')}
      </div>
      <div className="mb-3 flex flex-wrap gap-1.5">
        {REASON_KEYS.map((r) => {
          const active = reason === r
          return (
            <button
              key={r}
              type="button"
              onClick={() => setReason(r)}
              className="cursor-pointer rounded-full border px-2.5 py-1 text-[12px] font-medium transition"
              style={{
                borderColor: active ? FG : BORDER,
                background: active ? FG : 'white',
                color: active ? 'white' : FG,
              }}
            >
              {t(`inventory.reason_${r}`, r)}
            </button>
          )
        })}
      </div>

      <div
        className="mb-1.5 text-[11px] font-semibold uppercase tracking-[0.06em]"
        style={{ color: MUTED_FG }}
      >
        {t('inventory.adjustNote', 'Note')}{' '}
        <span
          className="font-normal"
          style={{ textTransform: 'none', letterSpacing: 0, color: MUTED_FG }}
        >
          {t('inventory.adjustNoteOptional', '(optional)')}
        </span>
      </div>
      <textarea
        ref={noteRef}
        rows={2}
        value={note}
        onChange={(e) => setNote(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
            e.preventDefault()
            save()
          }
        }}
        placeholder={t(
          'inventory.adjustNotePlaceholder',
          'Add a note for the history log…',
        )}
        className="mb-3 w-full resize-y rounded-md border px-2.5 py-2 text-[13px] outline-none transition"
        style={
          {
            borderColor: BORDER,
            color: FG,
            background: 'white',
            '--tw-ring-color': RING,
          } as React.CSSProperties
        }
        onFocus={(e) => {
          e.currentTarget.style.borderColor = RING
          e.currentTarget.style.boxShadow = `0 0 0 2px hsl(221 83% 53% / 0.15)`
        }}
        onBlur={(e) => {
          e.currentTarget.style.borderColor = BORDER
          e.currentTarget.style.boxShadow = 'none'
        }}
      />

      <div className="flex justify-end gap-2">
        <button
          type="button"
          onClick={onCancel}
          className="inline-flex h-9 items-center rounded-md border bg-white px-3.5 text-[13px] font-medium hover:bg-[hsl(210_40%_96%)]"
          style={{ borderColor: BORDER, color: FG }}
        >
          {t('common.cancel', 'Cancel')}
        </button>
        <button
          type="button"
          onClick={save}
          className="inline-flex h-9 items-center rounded-md px-3.5 text-[13px] font-medium text-white transition"
          style={{ background: 'hsl(222.2 47.4% 11.2%)' }}
          onMouseEnter={(e) => {
            e.currentTarget.style.background = 'hsl(222.2 47.4% 16%)'
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.background = 'hsl(222.2 47.4% 11.2%)'
          }}
        >
          {t('inventory.saveAdjustment', 'Save adjustment')}
        </button>
      </div>
    </div>
  )
}
