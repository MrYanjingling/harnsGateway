import { AlertTriangle, RefreshCw } from 'lucide-react'
import { errorMessage } from '@/lib/http'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'

export function ErrorState({
  error,
  onRetry,
  compact = false,
}: {
  error: unknown
  onRetry?: () => void
  compact?: boolean
}) {
  const message = errorMessage(error)

  if (compact) {
    return (
      <div className="flex items-center gap-3 rounded-xl border border-danger/25 bg-danger/5 px-4 py-3 text-sm">
        <AlertTriangle className="size-4 shrink-0 text-danger" />
        <span className="min-w-0 flex-1 text-fg">{message}</span>
        {onRetry ? (
          <Button size="sm" variant="ghost" onClick={onRetry} icon={<RefreshCw className="size-3.5" />}>
            重试
          </Button>
        ) : null}
      </div>
    )
  }

  return (
    <Card className="flex flex-col items-center justify-center px-6 py-14 text-center">
      <div className="mb-4 flex size-12 items-center justify-center rounded-2xl border border-danger/25 bg-danger/5 text-danger">
        <AlertTriangle className="size-5" />
      </div>
      <h3 className="text-sm font-semibold">加载失败</h3>
      <p className="mt-1 max-w-md text-xs leading-relaxed text-muted">{message}</p>
      {onRetry ? (
        <Button className="mt-4" variant="outline" onClick={onRetry} icon={<RefreshCw className="size-4" />}>
          重新加载
        </Button>
      ) : null}
    </Card>
  )
}
