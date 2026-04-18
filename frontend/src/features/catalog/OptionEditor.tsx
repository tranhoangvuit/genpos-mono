import { Trash2, X } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/shared/ui/button'
import { Input } from '@/shared/ui/input'
import { Label } from '@/shared/ui/label'

import type { OptionFormValue } from './types'

type Props = {
  option: OptionFormValue
  index: number
  onChange: (next: OptionFormValue) => void
  onRemove: () => void
}

// OptionEditor edits a single option axis: its name + list of value chips.
export function OptionEditor({ option, index, onChange, onRemove }: Props) {
  const { t } = useTranslation()
  const [newValue, setNewValue] = useState('')

  const setName = (name: string) => onChange({ ...option, name })
  const addValue = () => {
    const v = newValue.trim()
    if (!v) return
    if (option.values.includes(v)) {
      setNewValue('')
      return
    }
    onChange({ ...option, values: [...option.values, v] })
    setNewValue('')
  }
  const removeValue = (i: number) => {
    onChange({ ...option, values: option.values.filter((_, idx) => idx !== i) })
  }

  return (
    <div className="rounded-lg border border-[color:var(--color-border)] bg-[color:var(--color-card)] p-4 space-y-3">
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1 space-y-2">
          <Label htmlFor={`option-name-${index}`}>{t('catalog.optionName')}</Label>
          <Input
            id={`option-name-${index}`}
            placeholder={t('catalog.optionNamePlaceholder')}
            value={option.name}
            onChange={(e) => setName(e.target.value)}
          />
        </div>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          onClick={onRemove}
          aria-label={t('common.remove')}
          className="mt-7"
        >
          <Trash2 className="h-4 w-4" />
        </Button>
      </div>

      <div className="space-y-2">
        <Label>{t('catalog.optionValues')}</Label>
        <div className="flex flex-wrap gap-2">
          {option.values.map((v, i) => (
            <span
              key={v + i}
              className="inline-flex items-center gap-1 rounded-full bg-[color:var(--color-secondary)] px-3 py-1 text-xs text-[color:var(--color-secondary-foreground)]"
            >
              {v}
              <button
                type="button"
                onClick={() => removeValue(i)}
                className="rounded-full hover:bg-[color:var(--color-muted)]"
                aria-label={t('common.remove')}
              >
                <X className="h-3 w-3" />
              </button>
            </span>
          ))}
        </div>
        <div className="flex gap-2">
          <Input
            value={newValue}
            onChange={(e) => setNewValue(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                e.preventDefault()
                addValue()
              }
            }}
            placeholder={t('catalog.optionValuePlaceholder')}
          />
          <Button type="button" variant="outline" onClick={addValue}>
            {t('common.add')}
          </Button>
        </div>
      </div>
    </div>
  )
}
