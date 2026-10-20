import { useCallback, useState } from 'react'
import { Field, Input } from '@/components/ui/Field'
import { Button } from '@/components/ui/Button'
import { formatApiTime, fromDateTimeLocal, toDateTimeLocal } from '@/lib/format'

export interface ApiTimeRange {
  startTime: string
  endTime: string
}

/** 数据服务查询的公共时间范围状态。 */
export function useTimeRange(hours = 24) {
  const [start, setStart] = useState(() =>
    toDateTimeLocal(new Date(Date.now() - hours * 60 * 60 * 1000)),
  )
  const [end, setEnd] = useState(() => toDateTimeLocal(new Date()))

  const applyPreset = useCallback((presetHours: number) => {
    const now = new Date()
    setEnd(toDateTimeLocal(now))
    setStart(toDateTimeLocal(new Date(now.getTime() - presetHours * 60 * 60 * 1000)))
  }, [])

  const toApi = useCallback((): ApiTimeRange | null => {
    const startDate = fromDateTimeLocal(start)
    const endDate = fromDateTimeLocal(end)
    if (!startDate || !endDate || startDate.getTime() >= endDate.getTime()) return null
    return { startTime: formatApiTime(startDate), endTime: formatApiTime(endDate) }
  }, [start, end])

  return { start, end, setStart, setEnd, applyPreset, toApi, valid: Boolean(toApi()) }
}

const PRESETS = [
  { hours: 1, label: '1 小时' },
  { hours: 24, label: '24 小时' },
  { hours: 24 * 7, label: '7 天' },
]

export function TimeRangeFields({
  range,
}: {
  range: ReturnType<typeof useTimeRange>
}) {
  return (
    <div className="flex flex-wrap items-end gap-3">
      <Field label="开始时间" className="w-52">
        <Input
          type="datetime-local"
          value={range.start}
          onChange={(event) => range.setStart(event.target.value)}
        />
      </Field>
      <Field label="结束时间" className="w-52">
        <Input
          type="datetime-local"
          value={range.end}
          onChange={(event) => range.setEnd(event.target.value)}
        />
      </Field>
      <div className="flex items-center gap-1 pb-0.5">
        {PRESETS.map((preset) => (
          <Button
            key={preset.hours}
            size="sm"
            variant="ghost"
            onClick={() => range.applyPreset(preset.hours)}
          >
            {preset.label}
          </Button>
        ))}
      </div>
      {!range.valid ? (
        <p className="pb-2 text-xs text-danger">开始时间必须早于结束时间</p>
      ) : null}
    </div>
  )
}
