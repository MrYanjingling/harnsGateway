import type { DeviceStatus } from '@/api/types'
import { DEVICE_STATUS, UNKNOWN_STATUS } from '@/lib/constants'
import { cn } from '@/lib/cn'

export function StatusBadge({ status, className }: { status: DeviceStatus; className?: string }) {
  const meta = DEVICE_STATUS[status] ?? UNKNOWN_STATUS
  return (
    <span
      title={meta.description}
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full border px-2 py-0.5 text-[11px] font-medium whitespace-nowrap',
        meta.className,
        className,
      )}
    >
      <span className={cn('size-1.5 rounded-full', meta.dotClassName)} />
      {meta.label}
    </span>
  )
}
