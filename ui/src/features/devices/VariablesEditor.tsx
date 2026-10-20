import { useFieldArray, type UseFormReturn } from 'react-hook-form'
import { Copy, Plus, Trash2 } from 'lucide-react'
import type { DataType, DeviceType } from '@/api/types'
import { DATA_TYPES } from '@/lib/constants'
import { cn } from '@/lib/cn'
import { Button } from '@/components/ui/Button'
import { Select } from '@/components/ui/Field'
import { emptyVariable, type DeviceFormValues } from './schema'

const ALLOWED: Record<DeviceType, DataType[]> = {
  modbus: ['bool', 'int16', 'uint16', 'int32', 'int64', 'float32', 'float64'],
  opcUa: ['number', 'string'],
  s7: ['bool', 'int16', 'uint16', 'int32', 'int64', 'float32', 'float64'],
}

const cellInput =
  'h-8.5 w-full rounded-md border border-border bg-surface px-2 text-xs text-fg placeholder:text-faint ' +
  'transition-colors hover:border-border-strong focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/15'

export function VariablesEditor({
  form,
  type,
}: {
  form: UseFormReturn<DeviceFormValues>
  type: DeviceType
}) {
  const { fields, append, insert, remove } = useFieldArray({
    control: form.control,
    name: 'variables',
  })
  const {
    register,
    formState: { errors },
  } = form

  const rowErrors = errors.variables
  const rootError = rowErrors?.root?.message ?? rowErrors?.message
  const options = DATA_TYPES.filter((item) => ALLOWED[type].includes(item.value))

  return (
    <div>
      <div className="mb-3 flex items-center justify-between gap-3">
        <div>
          <p className="text-xs text-muted">
            共 <span className="font-medium text-fg tabular">{fields.length}</span> 个变量
            {type === 'modbus' && ' · 地址从 0 开始'}
          </p>
        </div>
        <Button
          size="sm"
          variant="outline"
          icon={<Plus className="size-3.5" />}
          onClick={() => append(emptyVariable(type))}
        >
          添加变量
        </Button>
      </div>

      <div className="overflow-x-auto rounded-xl border border-border">
        <table className="w-full min-w-[860px] border-collapse text-xs">
          <thead>
            <tr className="border-b border-border bg-surface-2/70 text-left text-[10px] font-medium tracking-wide text-muted uppercase">
              <th className="w-9 px-2 py-2 text-center">#</th>
              <th className="min-w-36 px-2 py-2">变量名</th>
              <th className="w-28 px-2 py-2">数据类型</th>
              <th className="w-28 px-2 py-2">地址</th>
              {type === 'modbus' ? (
                <>
                  <th className="w-24 px-2 py-2">功能码</th>
                  <th className="w-20 px-2 py-2">位</th>
                  <th className="w-20 px-2 py-2">倍率</th>
                  <th className="w-20 px-2 py-2">数量</th>
                </>
              ) : null}
              {type === 'opcUa' ? <th className="w-24 px-2 py-2">Namespace</th> : null}
              {type === 's7' ? <th className="w-20 px-2 py-2">倍率</th> : null}
              <th className="w-24 px-2 py-2">读写</th>
              <th className="w-20 px-2 py-2 text-right">操作</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border">
            {fields.map((field, index) => {
              const err = rowErrors?.[index]
              const dataType = form.watch(`variables.${index}.dataType`)
              return (
                <tr key={field.id} className="align-top hover:bg-surface-2/40">
                  <td className="px-2 pt-3 text-center text-[10px] text-faint tabular">{index + 1}</td>
                  <td className="px-2 py-2">
                    <input
                      {...register(`variables.${index}.name`)}
                      placeholder="例如: 排气压力"
                      className={cn(cellInput, err?.name && 'border-danger')}
                    />
                    <CellError message={err?.name?.message} />
                  </td>
                  <td className="px-2 py-2">
                    <Select
                      className="h-8.5 rounded-md px-2 text-xs"
                      {...register(`variables.${index}.dataType`)}
                    >
                      {options.map((option) => (
                        <option key={option.value} value={option.value}>
                          {option.label}
                        </option>
                      ))}
                    </Select>
                    <CellError message={err?.dataType?.message} />
                  </td>
                  <td className="px-2 py-2">
                    <input
                      {...register(`variables.${index}.address`)}
                      placeholder={type === 's7' ? 'DB1.DBW0' : type === 'opcUa' ? 'ns=2;s=Tag' : '0'}
                      className={cn(cellInput, err?.address && 'border-danger', 'font-mono')}
                    />
                    <CellError message={err?.address?.message} />
                  </td>

                  {type === 'modbus' ? (
                    <>
                      <td className="px-2 py-2">
                        <Select
                          className="h-8.5 rounded-md px-2 text-xs"
                          {...register(`variables.${index}.functionCode`, { valueAsNumber: true })}
                        >
                          {[1, 2, 3, 4].map((code) => (
                            <option key={code} value={code}>
                              {code}
                            </option>
                          ))}
                        </Select>
                        <CellError message={err?.functionCode?.message} />
                      </td>
                      <td className="px-2 py-2">
                        <input
                          type="number"
                          min={0}
                          max={7}
                          disabled={dataType !== 'bool'}
                          {...register(`variables.${index}.bits`, { valueAsNumber: true })}
                          className={cn(cellInput, 'disabled:opacity-40', err?.bits && 'border-danger')}
                        />
                        <CellError message={err?.bits?.message} />
                      </td>
                      <td className="px-2 py-2">
                        <input
                          type="number"
                          step="any"
                          {...register(`variables.${index}.rate`, { valueAsNumber: true })}
                          className={cn(cellInput, err?.rate && 'border-danger')}
                        />
                        <CellError message={err?.rate?.message} />
                      </td>
                      <td className="px-2 py-2">
                        <input
                          type="number"
                          min={0}
                          {...register(`variables.${index}.amount`, { valueAsNumber: true })}
                          className={cn(cellInput, err?.amount && 'border-danger')}
                        />
                        <CellError message={err?.amount?.message} />
                      </td>
                    </>
                  ) : null}

                  {type === 'opcUa' ? (
                    <td className="px-2 py-2">
                      <input
                        type="number"
                        min={1}
                        {...register(`variables.${index}.namespace`, { valueAsNumber: true })}
                        className={cn(cellInput, err?.namespace && 'border-danger')}
                      />
                      <CellError message={err?.namespace?.message} />
                    </td>
                  ) : null}

                  {type === 's7' ? (
                    <td className="px-2 py-2">
                      <input
                        type="number"
                        step="any"
                        {...register(`variables.${index}.rate`, { valueAsNumber: true })}
                        className={cn(cellInput, err?.rate && 'border-danger')}
                      />
                      <CellError message={err?.rate?.message} />
                    </td>
                  ) : null}

                  <td className="px-2 py-2">
                    <Select
                      className="h-8.5 rounded-md px-2 text-xs"
                      {...register(`variables.${index}.accessMode`)}
                    >
                      <option value="r">只读 (r)</option>
                      <option value="rw">读写 (rw)</option>
                    </Select>
                    <CellError message={err?.accessMode?.message} />
                  </td>

                  <td className="px-2 py-2">
                    <div className="flex items-center justify-end gap-0.5">
                      <button
                        type="button"
                        title="复制"
                        onClick={() => {
                          const row = form.getValues(`variables.${index}`)
                          insert(index + 1, { ...row, name: `${row.name} 副本` })
                        }}
                        className="rounded-md p-1.5 text-faint transition-colors hover:bg-surface-2 hover:text-fg"
                      >
                        <Copy className="size-3.5" />
                      </button>
                      <button
                        type="button"
                        title="删除"
                        disabled={fields.length <= 1}
                        onClick={() => remove(index)}
                        className="rounded-md p-1.5 text-faint transition-colors hover:bg-danger/10 hover:text-danger disabled:pointer-events-none disabled:opacity-30"
                      >
                        <Trash2 className="size-3.5" />
                      </button>
                    </div>
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
      {rootError ? <p className="mt-2 text-xs text-danger">{rootError}</p> : null}
    </div>
  )
}

function CellError({ message }: { message?: string }) {
  if (!message) return null
  return <p className="mt-1 text-[10px] leading-tight text-danger">{message}</p>
}
