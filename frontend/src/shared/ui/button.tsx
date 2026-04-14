import * as React from 'react'
import { Slot } from '@radix-ui/react-slot'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/shared/lib/cn'

const buttonVariants = cva(
  'inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[color:var(--color-ring)] focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50',
  {
    variants: {
      variant: {
        default:
          'bg-[color:var(--color-primary)] text-[color:var(--color-primary-foreground)] hover:bg-[color:var(--color-primary)]/90',
        destructive:
          'bg-[color:var(--color-destructive)] text-[color:var(--color-destructive-foreground)] hover:bg-[color:var(--color-destructive)]/90',
        outline:
          'border border-[color:var(--color-input)] bg-[color:var(--color-background)] hover:bg-[color:var(--color-accent)] hover:text-[color:var(--color-accent-foreground)]',
        secondary:
          'bg-[color:var(--color-secondary)] text-[color:var(--color-secondary-foreground)] hover:bg-[color:var(--color-secondary)]/80',
        ghost:
          'hover:bg-[color:var(--color-accent)] hover:text-[color:var(--color-accent-foreground)]',
        link: 'text-[color:var(--color-primary)] underline-offset-4 hover:underline',
      },
      size: {
        default: 'h-11 px-6 py-2',
        sm: 'h-9 rounded-md px-3',
        lg: 'h-14 rounded-md px-8 text-base',
        icon: 'h-11 w-11',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'default',
    },
  }
)

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, asChild = false, ...props }, ref) => {
    const Comp = asChild ? Slot : 'button'
    return (
      <Comp
        className={cn(buttonVariants({ variant, size, className }))}
        ref={ref}
        {...props}
      />
    )
  }
)
Button.displayName = 'Button'

export { buttonVariants }
