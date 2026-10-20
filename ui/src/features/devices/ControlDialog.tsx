import { useEffect, useMemo, useState } from 'react'
import { Send } from 'lucide-react'
import { toast } from 'sonner'
import { useControlDevice } from '@/api/hooks'
import type { AccessMode, DataType, DeviceAction, DeviceDetail } from '@/api/types'
import { errorMessage } from '@/lib/http'
import { Button } from '@/components/ui/Button'
import { Dialog } from '@/components/ui/Dialog'
import { Field, Input, Select } from '@/components/ui/Field'

const NUMERIC_TYPES: DataType[] = ['int16', 'uint16', 'int32', 'int64', 'float32', 'float64', 'number']

function rawValue(value: unknown): string {
  if (value === null || value === undefined) return ''
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value)
}

function parseValue(type: DataType, raw: string): unknown {
  if (type === 'string') return raw
  if (type === 'bool') {
    if (raw === 'true' || raw === '1') return true
    if (raw === 'false' || raw === '0') return false
    throw new Error('布尔值只能是 true 或 false')
  }
  if (NUMERIC_TYPES.includes(type)) {
    const parsed = Number(raw)
    if (raw.trim() === '' || Number.isNaN(parsed)) throw new Error(`「${raw}」不是有效数字`)
    return parsed
  }
  return raw
}

export function ControlDialog({
  open,
  device,
  only,
  onClose,
}: {
  open: boolean
  device: DeviceDetail
  /** 仅显示指定变量名(从变量行进入时使用) */
  only?: string
  onClose: () => void
}) {
  const control = useControlDevice()
  const variables = useMemo(() => {
    // 三种协议的变量结构不同,反控只关心公共字段。
    const all = device.variables as Array<{
      name: string
      dataType: DataType
      accessMode: AccessMode
      value?: unknown
    }>
    const writable = all.filter((variable) => variable.accessMode === 'rw')
    return only ? writable.filter((variable) => variable.name === only) : writable
  }, [device.variables, only])

  const [values, setValues] = useState<Record<string, string>>({})
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!open) return
    const next: Record<string, string> = {}
    for (const variable of variables) {
      next[variable.name] = rawValue(variable.value)
    }
    setValues(next)
    setError(null)
  }, [open, variables])

  const submit = () => {
    setError(null)
    const actions: DeviceAction[] = []
    try {
      for (const variable of variables) {
        actions.push({ [variable.name]: parseValue(variable.dataType, values[variable.name] ?? '') })
      }
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : '输入值无效')
      return
    }

    if (actions.length === 0) {
      setError('没有可下发的读写变量')
      return
    }

    control.mutate(
      { id: device.id, actions },
      {
        onSuccess: () => {
          toast.success(`已向 ${device.name} 下发 ${actions.length} 个变量`)
          onClose()
        },
        onError: (cause) => setError(errorMessage(cause)),
      },
    )
  }

  return (
    <Dialog
      open={open}
      onClose={onClose}
      title="变量反控"
      description="仅读写 (rw) 变量可下发。值会按变量数据类型转换后提交。"
      size="lg"
      footer={
        <>
          <Button variant="ghost" onClick={onClose}>
            取消
          </Button>
          <Button
            variant="primary"
            loading={control.isPending}
            icon={<Send className="size-4" />}
            onClick={submit}
            disabled={variables.length === 0}
          >
            下发写入
          </Button>
        </>
      }
    >
      {variables.length === 0 ? (
        <p className="rounded-lg border border-border bg-surface-2 px-4 py-6 text-center text-sm text-muted">
          该设备没有配置 accessMode 为 rw 的变量,无法反控。
        </p>
      ) : (
        <div className="space-y-3">
          {variables.map((variable) => (
            <Field
              key={variable.name}
              label={
                <span className="flex items-center gap-2">
                  <span className="font-medium text-fg">{variable.name}</span>
                  <span className="rounded bg-surface-2 px-1.5 py-0.5 font-mono text-[10px] text-faint">
                    {variable.dataType}
                  </span>
                </span>
              }
              htmlFor={`control-${variable.name}`}
            >
              {variable.dataType === 'bool' ? (
                <Select
                  id={`control-${variable.name}`}
                  value={values[variable.name] ?? 'false'}
                  onChange={(event) =>
                    setValues((prev) => ({ ...prev, [variable.name]: event.target.value }))
                  }
                >
                  <option value="true">true (开)</option>
                  <option value="false">false (关)</option>
                </Select>
              ) : (
                <Input
                  id={`control-${variable.name}`}
                  inputMode={NUMERIC_TYPES.includes(variable.dataType) ? 'decimal' : 'text'}
                  value={values[variable.name] ?? ''}
                  onChange={(event) =>
                    setValues((prev) => ({ ...prev, [variable.name]: event.target.value }))
                  }
                  placeholder={variable.dataType === 'string' ? '输入字符串值' : '输入数值'}
                />
              )}
            </Field>
          ))}
          {error ? (
            <p className="rounded-lg border border-danger/30 bg-danger/5 px-3 py-2 text-xs text-danger">
              {error}
            </p>
          ) : null}
          <p className="text-[11px] leading-relaxed text-faint">
            提交格式为 <code className="font-mono">{`[{"变量名": 值}]`}</code>,由网关转发给设备。
            写入结果取决于设备是否在线与寄存器是否可写。
          </p>
        </div>
      )}
    </Dialog>
  )
}
