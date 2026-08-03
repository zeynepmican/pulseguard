'use client'

import { AlertTriangle, Save, X } from 'lucide-react'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { cn } from '@/lib/utils'
import type { AgentConfig, MetricType } from '@/lib/pulseguard'

const GROUPS: { type: MetricType; label: string; tint: string }[] = [
  { type: 'cpu', label: 'CPU Thresholds', tint: 'text-[var(--chart-1)]' },
  { type: 'ram', label: 'RAM Thresholds', tint: 'text-[var(--chart-2)]' },
  { type: 'disk', label: 'Disk Thresholds', tint: 'text-[var(--chart-3)]' },
]

interface ThresholdSettingsProps {
  hostId: string
  config: AgentConfig | null
  isSaving: boolean
  onFieldChange: (key: string, value: string) => void
  onSave: () => void
  onCancel: () => void
}

export function ThresholdSettings({
  hostId,
  config,
  isSaving,
  onFieldChange,
  onSave,
  onCancel,
}: ThresholdSettingsProps) {
  const moveCursorToEnd = (e: React.FocusEvent<HTMLInputElement> | React.MouseEvent<HTMLInputElement>) => {
    const target = e.target as HTMLInputElement
    const len = target.value.length
    target.setSelectionRange(len, len)
  }

  return (
    <Card className="mb-6 gap-5 p-6">
      <div className="flex items-center gap-2.5">
        <span className="flex size-8 items-center justify-center rounded-lg bg-warning/15 text-warning">
          <AlertTriangle className="size-4" aria-hidden="true" />
        </span>
        <h3 className="text-base font-semibold text-foreground">
          {hostId} Threshold Values <span className="font-normal text-muted-foreground">(Thresholds)</span>
        </h3>
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {GROUPS.map(({ type, label, tint }) => (
          <div key={type} className="rounded-xl border border-border bg-background/40 p-4">
            <h4 className={cn('mb-3 text-sm font-semibold', tint)}>{label}</h4>
            <div className="flex flex-col gap-3">
              <div className="flex flex-col gap-1.5">
                <Label htmlFor={`${type}_warn`} className="text-xs text-muted-foreground">
                  Warning Limit (Warning %)
                </Label>
                <Input
                  id={`${type}_warn`}
                  type="text"
                  inputMode="numeric"
                  value={config?.[`${type}_warn` as keyof AgentConfig] ?? ''}
                  onChange={(e) => onFieldChange(`${type}_warn`, e.target.value)}
                  onFocus={moveCursorToEnd}
                  onClick={moveCursorToEnd}
                  className="border-warning/40 tabular-nums focus-visible:border-warning focus-visible:ring-warning/30"
                />
              </div>
              <div className="flex flex-col gap-1.5">
                <Label htmlFor={`${type}_err`} className="text-xs text-muted-foreground">
                  Critical Limit (Error %)
                </Label>
                <Input
                  id={`${type}_err`}
                  type="text"
                  inputMode="numeric"
                  value={config?.[`${type}_err` as keyof AgentConfig] ?? ''}
                  onChange={(e) => onFieldChange(`${type}_err`, e.target.value)}
                  onFocus={moveCursorToEnd}
                  onClick={moveCursorToEnd}
                  className="border-error/40 tabular-nums focus-visible:border-error focus-visible:ring-error/30"
                />
              </div>
            </div>
          </div>
        ))}
      </div>

      <div className="flex justify-end gap-3">
        <Button variant="secondary" onClick={onCancel} className="gap-1.5">
          <X className="size-4" aria-hidden="true" />
          Cancel
        </Button>
        <Button
          onClick={onSave}
          disabled={isSaving}
          className="gap-1.5 bg-success text-success-foreground hover:bg-success/90"
        >
          <Save className="size-4" aria-hidden="true" />
          {isSaving ? 'Saving...' : 'Save Settings'}
        </Button>
      </div>
    </Card>
  )
}
