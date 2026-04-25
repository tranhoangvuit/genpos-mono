import { ConnectError } from '@connectrpc/connect'
import { Download, Upload } from 'lucide-react'
import { useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'

import type { CsvCustomerRow } from '@/gen/genpos/v1/customer_pb'
import { Button } from '@/shared/ui/button'
import { Checkbox } from '@/shared/ui/checkbox'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/shared/ui/dialog'

import { useImportCustomers, useParseImportCustomerCsv } from './hooks'

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
}

type PreviewState = {
  rows: CsvCustomerRow[]
  validCount: number
  errorCount: number
  warnings: string[]
  overrides: Set<number>
}

const SAMPLE_HEADER =
  'name,email,phone,code,address,company,tax_code,date_of_birth,gender,facebook,groups,notes,status'

const SAMPLE_CSV =
  SAMPLE_HEADER +
  '\n' +
  'Jane Doe,jane@example.com,+84 90 123 4567,C-0001,"12 Main St, D1",,,1990-01-15,female,,VIP,VIP customer,active\n' +
  'Acme LLC,billing@acme.vn,+84 28 555 0100,C-0002,"100 Ly Thuong Kiet",Acme LLC,0312345678,,,,,"Corporate account",active\n'

export function ImportCustomerDialog({ open, onOpenChange }: Props) {
  const { t } = useTranslation()
  const parseMut = useParseImportCustomerCsv()
  const importMut = useImportCustomers()
  const fileRef = useRef<HTMLInputElement>(null)

  const [preview, setPreview] = useState<PreviewState | null>(null)
  const [result, setResult] = useState<{
    created: number
    updated: number
    skipped: number
    errors: string[]
  } | null>(null)
  const [errorMsg, setErrorMsg] = useState<string | null>(null)

  const reset = () => {
    setPreview(null)
    setResult(null)
    setErrorMsg(null)
    parseMut.reset()
    importMut.reset()
    if (fileRef.current) fileRef.current.value = ''
  }

  const close = (next: boolean) => {
    if (!next) reset()
    onOpenChange(next)
  }

  const onFileSelected = async (file: File) => {
    setErrorMsg(null)
    setResult(null)
    try {
      const buf = await file.arrayBuffer()
      const res = await parseMut.mutateAsync(new Uint8Array(buf))
      setPreview({
        rows: res.rows,
        validCount: res.validCount,
        errorCount: res.errorCount,
        warnings: res.warnings,
        overrides: new Set<number>(),
      })
    } catch (err) {
      setErrorMsg(ConnectError.from(err).rawMessage)
    }
  }

  const toggleOverride = (idx: number) => {
    if (!preview) return
    const next = new Set(preview.overrides)
    if (next.has(idx)) next.delete(idx)
    else next.add(idx)
    setPreview({ ...preview, overrides: next })
  }

  const onImport = async () => {
    if (!preview) return
    setErrorMsg(null)
    try {
      const items = preview.rows.map((row, i) => ({
        row,
        overrideExisting: preview.overrides.has(i),
        existingId: row.existingId,
      }))
      const res = await importMut.mutateAsync({ items })
      setResult({
        created: res.created,
        updated: res.updated,
        skipped: res.skipped,
        errors: res.errors,
      })
    } catch (err) {
      setErrorMsg(ConnectError.from(err).rawMessage)
    }
  }

  const downloadSample = () => {
    const blob = new Blob([SAMPLE_CSV], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'customers_sample.csv'
    a.click()
    URL.revokeObjectURL(url)
  }

  const parsing = parseMut.isPending
  const importing = importMut.isPending

  return (
    <Dialog open={open} onOpenChange={close}>
      <DialogContent className="max-w-3xl">
        <DialogHeader>
          <DialogTitle>{t('customers.importCustomers')}</DialogTitle>
          <DialogDescription>{t('customers.importSubtitle')}</DialogDescription>
        </DialogHeader>

        {errorMsg && (
          <div className="rounded-md border border-[color:var(--color-destructive)]/30 bg-[color:var(--color-destructive)]/10 px-3 py-2 text-sm text-[color:var(--color-destructive)]">
            {errorMsg}
          </div>
        )}

        {!preview && !result && (
          <div className="space-y-4">
            <div className="flex flex-col items-center gap-3 rounded-lg border border-dashed border-[color:var(--color-border)] bg-[color:var(--color-muted)]/20 py-10">
              <Upload className="h-8 w-8 text-[color:var(--color-muted-foreground)]" />
              <input
                ref={fileRef}
                type="file"
                accept=".csv,text/csv"
                className="hidden"
                onChange={(e) => {
                  const f = e.target.files?.[0]
                  if (f) void onFileSelected(f)
                }}
              />
              <Button
                type="button"
                variant="outline"
                onClick={() => fileRef.current?.click()}
                disabled={parsing}
              >
                {parsing ? t('common.loading') : t('catalog.chooseCsv')}
              </Button>
              <button
                type="button"
                onClick={downloadSample}
                className="inline-flex items-center gap-1 text-xs text-[color:var(--color-muted-foreground)] hover:text-[color:var(--color-foreground)]"
              >
                <Download className="h-3 w-3" />
                {t('catalog.downloadSample')}
              </button>
            </div>
          </div>
        )}

        {preview && !result && (
          <div className="space-y-3">
            <div className="flex items-center justify-between text-sm">
              <span>
                {t('catalog.importSummary', {
                  valid: preview.validCount,
                  errors: preview.errorCount,
                })}
              </span>
              <button type="button" onClick={reset} className="text-xs underline">
                {t('catalog.chooseDifferent')}
              </button>
            </div>
            {preview.warnings.map((w, i) => (
              <p key={i} className="text-xs text-[color:var(--color-destructive)]">
                {w}
              </p>
            ))}
            <div className="max-h-80 overflow-auto rounded-lg border border-[color:var(--color-border)]">
              <table className="w-full text-xs">
                <thead className="bg-[color:var(--color-muted)]/40 text-[color:var(--color-muted-foreground)]">
                  <tr>
                    <th className="px-2 py-2 text-left">{t('customers.name')}</th>
                    <th className="px-2 py-2 text-left">Code</th>
                    <th className="px-2 py-2 text-left">{t('customers.phone')}</th>
                    <th className="px-2 py-2 text-left">{t('customers.email')}</th>
                    <th className="px-2 py-2 text-left">{t('customers.groups')}</th>
                    <th className="px-2 py-2 text-left">{t('catalog.status')}</th>
                    <th className="px-2 py-2 text-left">{t('catalog.override')}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-[color:var(--color-border)]">
                  {preview.rows.map((r, i) => (
                    <tr
                      key={i}
                      className={r.errors.length > 0 ? 'bg-[color:var(--color-destructive)]/5' : ''}
                    >
                      <td className="px-2 py-1">
                        {r.name}
                        {r.company ? (
                          <span className="ml-1 text-[color:var(--color-muted-foreground)]">
                            · {r.company}
                          </span>
                        ) : null}
                      </td>
                      <td className="px-2 py-1">{r.code}</td>
                      <td className="px-2 py-1">{r.phone}</td>
                      <td className="px-2 py-1">{r.email}</td>
                      <td className="px-2 py-1">{r.groups}</td>
                      <td className="px-2 py-1">
                        {r.errors.length > 0 ? (
                          <span className="text-[color:var(--color-destructive)]">
                            {r.errors.join(', ')}
                          </span>
                        ) : r.exists ? (
                          <span className="text-[color:var(--color-muted-foreground)]">
                            {t('catalog.exists')}
                          </span>
                        ) : (
                          <span className="text-[color:var(--color-success)]">
                            {t('catalog.new')}
                          </span>
                        )}
                      </td>
                      <td className="px-2 py-1">
                        {r.exists && r.errors.length === 0 && (
                          <Checkbox
                            checked={preview.overrides.has(i)}
                            onCheckedChange={() => toggleOverride(i)}
                          />
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => close(false)} disabled={importing}>
                {t('common.cancel')}
              </Button>
              <Button
                type="button"
                onClick={onImport}
                disabled={importing || preview.validCount === 0}
              >
                {importing
                  ? t('common.importing')
                  : t('customers.importNow', { n: preview.validCount })}
              </Button>
            </DialogFooter>
          </div>
        )}

        {result && (
          <div className="space-y-3 text-sm">
            <p>
              {t('catalog.importResult', {
                created: result.created,
                updated: result.updated,
                skipped: result.skipped,
              })}
            </p>
            {result.errors.length > 0 && (
              <ul className="list-disc space-y-1 pl-5 text-xs text-[color:var(--color-destructive)]">
                {result.errors.map((e, i) => (
                  <li key={i}>{e}</li>
                ))}
              </ul>
            )}
            <DialogFooter>
              <Button type="button" onClick={() => close(false)}>
                {t('common.done')}
              </Button>
            </DialogFooter>
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
