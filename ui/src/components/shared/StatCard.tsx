import type { ReactNode } from 'react'
import { cn } from '@/lib/cn'
import { Card } from '@/components/ui/Card'

export function StatCard({
  label,
  value,
  hint,
  icon,
  className,
  footer,
}: {
  label: ReactNode
  value: ReactNode
  hint?: ReactNode
  icon?: ReactNode
  className?: string
  footer?: ReactNode
}) {
  return (
    <Card className={cn('p-5', className)}>
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="text-xs font-medium text-muted">{label}</p>
          <p className="mt-2 text-2xl font-semibold tracking-tight tabular">{value}</p>
          {hint ? <p className="mt-1 text-xs text-faint">{hint}</p> : null}
        </div>
        {icon ? (
          <div className="flex size-9 shrink-0 items-center justify-center rounded-xl border border-border bg-surface-2 text-muted">
            {icon}
          </div>
        ) : null}
      </div>
      {footer ? <div className="mt-4">{footer}</div> : null}
    </Card>
  )
}
