import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { ArrowLeft, ArrowRight, Cable, Cpu, Globe2 } from 'lucide-react'
import { toast } from 'sonner'
import { useCreateDevice } from '@/api/hooks'
import type { DeviceType } from '@/api/types'
import { PROTOCOLS } from '@/lib/constants'
import { errorMessage } from '@/lib/http'
import { cn } from '@/lib/cn'
import { Button } from '@/components/ui/Button'
import { PageHeader } from '@/components/shared/PageHeader'
import { DeviceForm } from './DeviceForm'
import { defaultFormValues, toPayload } from './schema'

const ICONS: Record<DeviceType, typeof Cable> = {
  modbus: Cable,
  opcUa: Globe2,
  s7: Cpu,
}

export function DeviceCreatePage() {
  const navigate = useNavigate()
  const createDevice = useCreateDevice()
  const [type, setType] = useState<DeviceType | null>(null)

  const submit = (values: ReturnType<typeof defaultFormValues>) => {
    createDevice.mutate(toPayload(values), {
      onSuccess: (device) => {
        toast.success(`设备「${device.name}」已创建`)
        navigate(`/devices/${device.id}`)
      },
      onError: (error) => toast.error(errorMessage(error)),
    })
  }

  return (
    <div>
      <PageHeader
        backTo="/devices"
        backLabel="返回设备列表"
        title={type ? `新建设备 · ${PROTOCOLS[type].label}` : '新建设备'}
        description={
          type
            ? '填写设备信息与采集变量,创建后可随时启停或修改。'
            : '选择设备使用的协议,不同协议的连接参数与变量字段不同。'
        }
        actions={
          type ? (
            <Button
              variant="outline"
              icon={<ArrowLeft className="size-4" />}
              onClick={() => setType(null)}
            >
              更换协议
            </Button>
          ) : undefined
        }
      />

      {!type ? (
        <div className="grid gap-4 md:grid-cols-3">
          {(Object.keys(PROTOCOLS) as DeviceType[]).map((key) => {
            const protocol = PROTOCOLS[key]
            const Icon = ICONS[key]
            return (
              <button
                key={key}
                type="button"
                onClick={() => setType(key)}
                className={cn(
                  'group flex flex-col items-start gap-3 rounded-2xl border border-border bg-surface p-5 text-left',
                  'transition-all hover:-translate-y-0.5 hover:border-primary/40 hover:shadow-md',
                  'focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary',
                )}
              >
                <span className="flex size-10 items-center justify-center rounded-xl bg-primary-soft text-primary">
                  <Icon className="size-5" />
                </span>
                <span>
                  <span className="block text-sm font-semibold">{protocol.label}</span>
                  <span className="mt-1 block text-xs leading-relaxed text-muted">
                    {protocol.description}
                  </span>
                </span>
                <span className="mt-auto w-full space-y-1.5 border-t border-border pt-3">
                  {protocol.models.map((model) => (
                    <span key={model.value} className="flex items-center justify-between gap-2">
                      <span className="font-mono text-[11px] text-muted">{model.value}</span>
                      <span className="text-[10px] text-faint">{model.hint}</span>
                    </span>
                  ))}
                </span>
                <span className="flex items-center gap-1 text-xs font-medium text-primary">
                  开始配置
                  <ArrowRight className="size-3.5 transition-transform group-hover:translate-x-0.5" />
                </span>
              </button>
            )
          })}
        </div>
      ) : (
        <DeviceForm
          key={type}
          mode="create"
          type={type}
          initialValues={defaultFormValues(type)}
          submitting={createDevice.isPending}
          onSubmit={submit}
          onCancel={() => navigate('/devices')}
        />
      )}
    </div>
  )
}
