import { useParams, useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import { useDevice, useUpdateDevice } from '@/api/hooks'
import { errorMessage } from '@/lib/http'
import { Card, CardContent } from '@/components/ui/Card'
import { PageLoading } from '@/components/ui/Spinner'
import { ErrorState } from '@/components/shared/ErrorState'
import { PageHeader } from '@/components/shared/PageHeader'
import { DeviceForm } from './DeviceForm'
import { toFormValues, toPayload, type DeviceFormValues } from './schema'

export function DeviceEditPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const device = useDevice(id)
  const updateDevice = useUpdateDevice()

  if (device.isLoading) {
    return <PageLoading label="正在加载设备…" />
  }

  if (device.isError || !device.data) {
    return (
      <div>
        <PageHeader backTo="/devices" backLabel="返回设备列表" title="编辑设备" />
        <ErrorState error={device.error} onRetry={device.refetch} />
      </div>
    )
  }

  const detail = device.data

  const submit = (values: DeviceFormValues) => {
    updateDevice.mutate(
      { id: detail.id, payload: toPayload(values), etag: detail.eTag },
      {
        onSuccess: () => {
          toast.success(`设备「${values.name}」已保存`)
          navigate(`/devices/${detail.id}`)
        },
        onError: (error) => toast.error(errorMessage(error)),
      },
    )
  }

  return (
    <div>
      <PageHeader
        backTo={`/devices/${detail.id}`}
        backLabel="返回设备详情"
        title={`编辑 · ${detail.name}`}
        description="协议类型与型号不可修改。保存时携带当前 eTag 进行乐观锁校验。"
      />
      <Card className="mb-4">
        <CardContent className="flex flex-wrap items-center gap-x-6 gap-y-2 pt-4 text-xs text-muted">
          <span>
            ID <code className="ml-1 font-mono text-fg">{detail.id}</code>
          </span>
          <span>
            eTag <code className="ml-1 font-mono text-fg">{detail.eTag}</code>
          </span>
          <span>
            最后修改 <span className="text-fg">{detail.modTime}</span>
          </span>
        </CardContent>
      </Card>
      <DeviceForm
        mode="edit"
        type={detail.deviceType}
        initialValues={toFormValues(detail)}
        submitting={updateDevice.isPending}
        onSubmit={submit}
        onCancel={() => navigate(`/devices/${detail.id}`)}
      />
    </div>
  )
}
