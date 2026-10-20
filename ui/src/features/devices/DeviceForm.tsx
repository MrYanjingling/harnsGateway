import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { useMemo } from 'react'
import { Info, Save, X } from 'lucide-react'
import type { DeviceType } from '@/api/types'
import { BAUD_RATES, DATA_BITS, MEMORY_LAYOUTS, PARITIES, PROTOCOLS, STOP_BITS } from '@/lib/constants'
import { Button } from '@/components/ui/Button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { Field, Input, Select } from '@/components/ui/Field'
import { deviceFormSchema, type DeviceFormValues } from './schema'
import { VariablesEditor } from './VariablesEditor'

export interface DeviceFormProps {
  mode: 'create' | 'edit'
  type: DeviceType
  initialValues: DeviceFormValues
  submitting: boolean
  onSubmit: (values: DeviceFormValues) => void
  onCancel: () => void
}

export function DeviceForm({
  mode,
  type,
  initialValues,
  submitting,
  onSubmit,
  onCancel,
}: DeviceFormProps) {
  const protocol = PROTOCOLS[type]
  const form = useForm<DeviceFormValues>({
    resolver: zodResolver(deviceFormSchema),
    defaultValues: initialValues,
    mode: 'onTouched',
  })
  const {
    register,
    handleSubmit,
    watch,
    setValue,
    formState: { errors, isDirty },
  } = form

  const model = watch('deviceModel')
  const isSerial = type === 'modbus' && model === 'modbusRtu'
  const title = useMemo(() => (mode === 'create' ? '新建设备' : '编辑设备'), [mode])

  return (
    <form
      onSubmit={handleSubmit(onSubmit)}
      className="space-y-4 pb-2"
      noValidate
      aria-label={title}
    >
      <Card>
        <CardHeader>
          <div>
            <CardTitle>基本信息</CardTitle>
            <CardDescription>设备标识与采集节奏,创建后协议类型不可更改</CardDescription>
          </div>
          <span className="rounded-md bg-surface-2 px-2 py-1 text-[11px] font-medium text-muted">
            {protocol.label}
          </span>
        </CardHeader>
        <CardContent className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <Field label="设备名称" htmlFor="device-name" required error={errors.name?.message}>
            <Input
              id="device-name"
              placeholder="例如: 1 号空压机"
              {...register('name')}
              aria-invalid={!!errors.name}
            />
          </Field>
          <Field
            label="设备编码"
            htmlFor="device-code"
            required
            hint="唯一标识,建议使用资产编号"
            error={errors.deviceCode?.message}
          >
            <Input
              id="device-code"
              placeholder="例如: AC-001"
              {...register('deviceCode')}
              aria-invalid={!!errors.deviceCode}
            />
          </Field>
          <Field label="设备型号" htmlFor="device-model" required error={errors.deviceModel?.message}>
            {mode === 'create' && protocol.models.length > 1 ? (
              <Select
                id="device-model"
                {...register('deviceModel', {
                  onChange: (event) => {
                    // 切换型号时补齐该型号的默认端口,避免沿用另一个型号的端口。
                    if (event.target.value === 'modbusRtu') {
                      setValue('modbusLocation', '')
                    }
                  },
                })}
              >
                {protocol.models.map((item) => (
                  <option key={item.value} value={item.value}>
                    {item.label} — {item.hint}
                  </option>
                ))}
              </Select>
            ) : (
              <Input id="device-model" value={model} readOnly disabled className="font-mono" />
            )}
          </Field>

          <Field
            label="采集周期 (秒)"
            htmlFor="collector-cycle"
            required
            hint="整机轮询一次变量的间隔,最小 1 秒"
            error={errors.collectorCycle?.message}
          >
            <Input
              id="collector-cycle"
              type="number"
              min={1}
              max={86400}
              {...register('collectorCycle', { valueAsNumber: true })}
            />
          </Field>
          <Field
            label="变量间隔 (毫秒)"
            htmlFor="variable-interval"
            hint="变量之间的采集间隔,0 表示不等待"
            error={errors.variableInterval?.message}
          >
            <Input
              id="variable-interval"
              type="number"
              min={0}
              max={86400}
              {...register('variableInterval', { valueAsNumber: true })}
            />
          </Field>
          <Field
            label="Topic (可选)"
            htmlFor="topic"
            hint="MQTT 上报主题,留空使用网关默认规则"
            error={errors.topic?.message}
          >
            <Input id="topic" placeholder="留空则自动生成" {...register('topic')} />
          </Field>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <div>
            <CardTitle>连接参数</CardTitle>
            <CardDescription>{protocol.description}</CardDescription>
          </div>
        </CardHeader>
        <CardContent className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {type === 'modbus' ? (
            <>
              <Field
                label={isSerial ? '串口设备' : '设备地址'}
                htmlFor="mb-location"
                required
                hint={isSerial ? '例如: /dev/ttyS0 或 COM1' : 'IP 地址或主机名'}
                error={errors.modbusLocation?.message}
              >
                <Input
                  id="mb-location"
                  placeholder={isSerial ? '/dev/ttyS0' : '192.168.1.10'}
                  {...register('modbusLocation')}
                />
              </Field>
              {isSerial ? (
                <>
                  <Field label="波特率" htmlFor="mb-baud" error={errors.modbusBaudRate?.message}>
                    <Select
                      id="mb-baud"
                      {...register('modbusBaudRate', { valueAsNumber: true })}
                    >
                      {BAUD_RATES.map((rate) => (
                        <option key={rate} value={rate}>
                          {rate}
                        </option>
                      ))}
                    </Select>
                  </Field>
                  <Field label="数据位" htmlFor="mb-databits" error={errors.modbusDataBits?.message}>
                    <Select
                      id="mb-databits"
                      {...register('modbusDataBits', { valueAsNumber: true })}
                    >
                      {DATA_BITS.map((bits) => (
                        <option key={bits} value={bits}>
                          {bits}
                        </option>
                      ))}
                    </Select>
                  </Field>
                  <Field label="校验位" htmlFor="mb-parity" error={errors.modbusParity?.message}>
                    <Select id="mb-parity" {...register('modbusParity')}>
                      {PARITIES.map((item) => (
                        <option key={item.value} value={item.value}>
                          {item.label}
                        </option>
                      ))}
                    </Select>
                  </Field>
                  <Field label="停止位" htmlFor="mb-stopbits" error={errors.modbusStopBits?.message}>
                    <Select id="mb-stopbits" {...register('modbusStopBits')}>
                      {STOP_BITS.map((item) => (
                        <option key={item.value} value={item.value}>
                          {item.label}
                        </option>
                      ))}
                    </Select>
                  </Field>
                </>
              ) : (
                <Field label="端口" htmlFor="mb-port" required error={errors.modbusPort?.message}>
                  <Input
                    id="mb-port"
                    type="number"
                    min={1}
                    max={65535}
                    {...register('modbusPort', { valueAsNumber: true })}
                  />
                </Field>
              )}
              <Field
                label="从站地址 (Slave ID)"
                htmlFor="mb-slave"
                required
                error={errors.slave?.message}
              >
                <Input
                  id="mb-slave"
                  type="number"
                  min={1}
                  max={65535}
                  {...register('slave', { valueAsNumber: true })}
                />
              </Field>
              <Field
                label="内存布局"
                htmlFor="mb-memory"
                hint="32/64 位数据的字节序"
                error={errors.memoryLayout?.message}
              >
                <Select id="mb-memory" {...register('memoryLayout')}>
                  {MEMORY_LAYOUTS.map((layout) => (
                    <option key={layout} value={layout}>
                      {layout}
                    </option>
                  ))}
                </Select>
              </Field>
              <Field
                label="起始地址偏移"
                htmlFor="mb-position"
                hint="寄存器地址是否额外偏移"
                error={errors.positionAddress?.message}
              >
                <Input
                  id="mb-position"
                  type="number"
                  min={0}
                  max={65535}
                  {...register('positionAddress', { valueAsNumber: true })}
                />
              </Field>
            </>
          ) : null}

          {type === 'opcUa' ? (
            <>
              <Field
                label="Endpoint"
                htmlFor="ua-location"
                required
                hint="形如 opc.tcp://192.168.1.20"
                error={errors.opcLocation?.message}
                className="sm:col-span-2"
              >
                <Input
                  id="ua-location"
                  placeholder="opc.tcp://192.168.1.20"
                  className="font-mono"
                  {...register('opcLocation')}
                />
              </Field>
              <Field label="端口" htmlFor="ua-port" required error={errors.opcPort?.message}>
                <Input
                  id="ua-port"
                  type="number"
                  min={1}
                  max={65535}
                  {...register('opcPort', { valueAsNumber: true })}
                />
              </Field>
              <Field label="用户名 (可选)" htmlFor="ua-username" error={errors.opcUsername?.message}>
                <Input
                  id="ua-username"
                  autoComplete="off"
                  placeholder="匿名登录可留空"
                  {...register('opcUsername')}
                />
              </Field>
              <Field label="密码 (可选)" htmlFor="ua-password" error={errors.opcPassword?.message}>
                <Input
                  id="ua-password"
                  type="password"
                  autoComplete="new-password"
                  placeholder="匿名登录可留空"
                  {...register('opcPassword')}
                />
              </Field>
            </>
          ) : null}

          {type === 's7' ? (
            <>
              <Field
                label="PLC 地址"
                htmlFor="s7-location"
                required
                hint="IP 地址或主机名"
                error={errors.s7Location?.message}
              >
                <Input
                  id="s7-location"
                  placeholder="192.168.1.30"
                  {...register('s7Location')}
                />
              </Field>
              <Field label="端口" htmlFor="s7-port" required error={errors.s7Port?.message}>
                <Input
                  id="s7-port"
                  type="number"
                  min={1}
                  max={65535}
                  {...register('s7Port', { valueAsNumber: true })}
                />
              </Field>
              <Field label="Rack" htmlFor="s7-rack" error={errors.s7Rack?.message}>
                <Input
                  id="s7-rack"
                  type="number"
                  min={0}
                  max={7}
                  {...register('s7Rack', { valueAsNumber: true })}
                />
              </Field>
              <Field label="Slot" htmlFor="s7-slot" error={errors.s7Slot?.message}>
                <Input
                  id="s7-slot"
                  type="number"
                  min={0}
                  max={31}
                  {...register('s7Slot', { valueAsNumber: true })}
                />
              </Field>
            </>
          ) : null}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <div>
            <CardTitle>采集变量</CardTitle>
            <CardDescription>
              变量名会作为测点标识,读写模式为 rw 的变量可在详情页下发写入
            </CardDescription>
          </div>
        </CardHeader>
        <CardContent>
          <VariablesEditor form={form} type={type} />
          {errors.variables?.root?.message ? (
            <p className="mt-2 text-xs text-danger">{errors.variables.root.message}</p>
          ) : null}
        </CardContent>
      </Card>

      <div className="sticky bottom-0 z-10 flex flex-wrap items-center justify-between gap-3 rounded-xl border border-border bg-surface/85 px-4 py-3 shadow-sm backdrop-blur">
        <p className="flex items-center gap-1.5 text-xs text-faint">
          <Info className="size-3.5 shrink-0" />
          {mode === 'create'
            ? '保存后设备将以停止状态创建,可在列表页启动采集。'
            : '保存时使用当前版本号校验,若设备已被他人修改会提示刷新。'}
          {isDirty ? <span className="ml-1 text-warning">· 有未保存的修改</span> : null}
        </p>
        <div className="flex items-center gap-2">
          <Button variant="ghost" icon={<X className="size-4" />} onClick={onCancel}>
            取消
          </Button>
          <Button
            type="submit"
            variant="primary"
            loading={submitting}
            icon={<Save className="size-4" />}
          >
            {mode === 'create' ? '创建设备' : '保存修改'}
          </Button>
        </div>
      </div>
    </form>
  )
}
