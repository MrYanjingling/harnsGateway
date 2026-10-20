import { cn } from '@/lib/cn'

const TONES = {
  primary: 'text-primary',
  success: 'text-emerald-500',
  warning: 'text-amber-500',
  danger: 'text-rose-500',
} as const

export type GaugeTone = keyof typeof TONES

export function toneForPercent(value: number): GaugeTone {
  if (value >= 90) return 'danger'
  if (value >= 75) return 'warning'
  return 'success'
}

export function GaugeRing({
  value,
  label,
  caption,
  tone,
  size = 92,
  strokeWidth = 8,
  className,
}: {
  value: number
  label: string
  caption?: string
  tone?: GaugeTone
  size?: number
  strokeWidth?: number
  className?: string
}) {
  const clamped = Math.min(100, Math.max(0, Number.isFinite(value) ? value : 0))
  const radius = (size - strokeWidth) / 2
  const circumference = 2 * Math.PI * radius
  const dash = (clamped / 100) * circumference
  const resolvedTone = tone ?? toneForPercent(clamped)

  return (
    <div className={cn('flex flex-col items-center gap-2', className)}>
      <div className="relative" style={{ width: size, height: size }}>
        <svg width={size} height={size} className="-rotate-90">
          <circle
            cx={size / 2}
            cy={size / 2}
            r={radius}
            fill="none"
            strokeWidth={strokeWidth}
            className="stroke-surface-3"
          />
          <circle
            cx={size / 2}
            cy={size / 2}
            r={radius}
            fill="none"
            strokeWidth={strokeWidth}
            strokeLinecap="round"
            strokeDasharray={`${dash} ${circumference - dash}`}
            className={cn('transition-all duration-700', TONES[resolvedTone])}
            stroke="currentColor"
          />
        </svg>
        <div className="absolute inset-0 flex flex-col items-center justify-center">
          <span className="text-lg font-semibold tabular">{Math.round(clamped)}%</span>
          {caption ? <span className="text-[10px] text-faint">{caption}</span> : null}
        </div>
      </div>
      <span className="text-[11px] font-medium text-muted">{label}</span>
    </div>
  )
}
