import { forwardRef, type TextareaHTMLAttributes } from 'react'

import { cn } from '@/shared/lib/cn'

export const Textarea = forwardRef<HTMLTextAreaElement, TextareaHTMLAttributes<HTMLTextAreaElement>>(
  ({ className, ...props }, ref) => (
    <textarea
      ref={ref}
      className={cn(
        'flex min-h-[80px] w-full rounded-lg border border-[color:var(--color-input)] bg-[color:var(--color-background)] px-3 py-2 text-[13.5px] shadow-sm placeholder:text-[color:var(--color-muted-foreground)]',
        'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[color:var(--color-ring)]',
        'disabled:cursor-not-allowed disabled:opacity-50',
        className,
      )}
      {...props}
    />
  ),
)
Textarea.displayName = 'Textarea'
