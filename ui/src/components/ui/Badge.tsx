import type { HTMLAttributes } from 'react'
import { cn } from '@/lib/cn'

type Tone = 'neutral' | 'primary' | 'success' | 'warning' | 'danger' | 'info' | 'violet'

const tones: Record<Tone, string> = {
  neutral: 'bg-surface-2 text-muted border-border',
  primary: 'bg-primary-soft text-primary border-primary/20',
  success:
    'bg-emerald-500/10 text-emerald-700 border-emerald-600/20 dark:text-emerald-300',
  warning: 'bg-amber-500/10 text-amber-700 border-amber-600/20 dark:text-amber-300',
  danger: 'bg-rose-500/10 text-rose-700 border-rose-600/20 dark:text-rose-300',
  info: 'bg-sky-500/10 text-sky-700 border-sky-600/20 dark:text-sky-300',
  violet: 'bg-violet-500/10 text-violet-700 border-violet-600/20 dark:text-violet-300',
}

export interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  tone?: Tone
  dot?: boolean
}

export function Badge({ className, tone = 'neutral', dot, children, ...props }: BadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full border px-2 py-0.5 text-[11px] font-medium whitespace-nowrap',
        tones[tone],
        className,
      )}
      {...props}
    >
      {dot ? <span className="size-1.5 rounded-full bg-current" /> : null}
      {children}
    </span>
  )
}
