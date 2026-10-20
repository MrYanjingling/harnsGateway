import type { DeviceType } from '@/api/types'
import { PROTOCOLS } from '@/lib/constants'
import { cn } from '@/lib/cn'

const accents: Record<DeviceType, string> = {
  modbus: 'bg-emerald-500/10 text-emerald-700 border-emerald-600/20 dark:text-emerald-300',
  opcUa: 'bg-sky-500/10 text-sky-700 border-sky-600/20 dark:text-sky-300',
  s7: 'bg-amber-500/10 text-amber-700 border-amber-600/20 dark:text-amber-300',
}

export function DeviceTypeBadge({
  type,
  model,
  className,
}: {
  type: DeviceType
  model?: string
  className?: string
}) {
  const meta = PROTOCOLS[type]
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 rounded-md border px-1.5 py-0.5 text-[11px] font-semibold',
        accents[type] ?? 'bg-surface-2 text-muted border-border',
        className,
      )}
      title={meta?.description ?? type}
    >
      {meta?.label ?? type}
      {model ? <span className="font-normal opacity-70">{model}</span> : null}
    </span>
  )
}
