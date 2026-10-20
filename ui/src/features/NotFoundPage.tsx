import { Link } from 'react-router-dom'
import { Compass } from 'lucide-react'
import { Card } from '@/components/ui/Card'

export function NotFoundPage() {
  return (
    <Card className="mx-auto mt-10 flex max-w-lg flex-col items-center px-6 py-16 text-center">
      <div className="mb-5 flex size-12 items-center justify-center rounded-2xl border border-border bg-surface-2 text-muted">
        <Compass className="size-5" />
      </div>
      <p className="text-4xl font-semibold tracking-tight tabular">404</p>
      <h1 className="mt-2 text-sm font-semibold">页面不存在</h1>
      <p className="mt-1 text-xs text-muted">你访问的地址可能已被移动或删除。</p>
      <Link
        to="/"
        className="mt-5 inline-flex h-9.5 items-center rounded-lg bg-primary px-4 text-sm font-medium text-primary-fg shadow-sm transition-colors hover:bg-primary-hover focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
      >
        返回概览
      </Link>
    </Card>
  )
}
