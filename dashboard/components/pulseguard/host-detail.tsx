'use client'

import { useMemo } from 'react'
import { Cpu, Download, HardDrive, MemoryStick, Settings2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'
import { MetricCard } from './metric-card'
import { MetricChart } from './metric-chart'
import { ThresholdSettings } from './threshold-settings'
import {
  downloadAllCSV,
  type AgentConfig,
  type ChartPoint,
  type MetricEvent,
  type MetricType,
} from '@/lib/pulseguard'

const ROWS: { type: MetricType; title: string; chartTitle: string; icon: React.ReactNode }[] = [
  { type: 'cpu', title: 'CPU Usage', chartTitle: 'CPU', icon: <Cpu className="size-4" aria-hidden="true" /> },
  { type: 'ram', title: 'RAM Usage', chartTitle: 'RAM', icon: <MemoryStick className="size-4" aria-hidden="true" /> },
  { type: 'disk', title: 'Disk Usage', chartTitle: 'Disk', icon: <HardDrive className="size-4" aria-hidden="true" /> },
]

interface HostDetailProps {
  hostId: string
  hostEvents: MetricEvent[]
  loading: boolean
  timeRange: number
  onTimeRangeChange: (value: number) => void
  agentConfig: AgentConfig | null
  isSettingsOpen: boolean
  onToggleSettings: () => void
  isSaving: boolean
  onFieldChange: (key: string, value: string) => void
  onSave: () => void
  onCancelSettings: () => void
}

export function HostDetail({
  hostId,
  hostEvents,
  loading,
  timeRange,
  onTimeRangeChange,
  agentConfig,
  isSettingsOpen,
  onToggleSettings,
  isSaving,
  onFieldChange,
  onSave,
  onCancelSettings,
}: HostDetailProps) {
  const getChartData = (type: MetricType): ChartPoint[] => {
    const events = hostEvents.filter((e) => e.check_type === type)
    if (events.length === 0) return []
    const latestTime = new Date(events[0].timestamp).getTime()
    const cutoffTime = latestTime - timeRange * 60 * 1000
    return events
      .filter((e) => new Date(e.timestamp).getTime() >= cutoffTime)
      .reverse()
      .map((e) => ({ timestamp: new Date(e.timestamp).getTime(), value: e.metric_value, level: e.level }))
  }

  const getLatestMetric = (type: MetricType): MetricEvent | null => {
    const events = hostEvents.filter((e) => e.check_type === type)
    return events.length > 0 ? events[0] : null
  }

  const rowData = useMemo(
    () =>
      ROWS.map((row) => ({
        ...row,
        latest: getLatestMetric(row.type),
        data: getChartData(row.type),
      })),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [hostEvents, timeRange],
  )

  return (
    <section aria-labelledby="host-heading">
      <div className="mb-6 flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
        <div className="flex items-center gap-3">
          <span className="relative flex size-3">
            <span className="absolute inline-flex size-full animate-ping rounded-full bg-success opacity-60" />
            <span className="relative inline-flex size-3 rounded-full bg-success" />
          </span>
          <h2 id="host-heading" className="text-xl font-semibold tracking-tight text-foreground">
            {hostId} <span className="font-normal text-muted-foreground">— Live Stream</span>
          </h2>
        </div>

        <div className="flex flex-wrap items-center gap-3">
          <Button
            variant="outline"
            className="gap-1.5 border-success/40 bg-success/10 text-success hover:bg-success/20 hover:text-success"
            onClick={() => downloadAllCSV(hostId, hostEvents, timeRange, agentConfig)}
            title="Download all CPU, RAM, and Disk logs as a single CSV report"
          >
            <Download className="size-4" aria-hidden="true" />
            Full Report (CSV)
          </Button>

          <Button
            variant={isSettingsOpen ? 'default' : 'outline'}
            className="gap-1.5"
            aria-pressed={isSettingsOpen}
            onClick={onToggleSettings}
          >
            <Settings2 className="size-4" aria-hidden="true" />
            Edit Thresholds
          </Button>

          <div className="flex items-center gap-2">
            <label htmlFor="time-range" className="text-sm text-muted-foreground">
              Time Range:
            </label>
            <Select value={String(timeRange)} onValueChange={(v) => onTimeRangeChange(Number(v))}>
              <SelectTrigger id="time-range" className="w-[150px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="1">Last 1 minute</SelectItem>
                <SelectItem value="5">Last 5 minutes</SelectItem>
                <SelectItem value="15">Last 15 minutes</SelectItem>
                <SelectItem value="60">Last 1 hour</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
      </div>

      {isSettingsOpen && (
        <ThresholdSettings
          hostId={hostId}
          config={agentConfig}
          isSaving={isSaving}
          onFieldChange={onFieldChange}
          onSave={onSave}
          onCancel={onCancelSettings}
        />
      )}

      {loading ? (
        <div className="flex flex-col gap-6">
          {[0, 1, 2].map((i) => (
            <div key={i} className="flex flex-col gap-5 lg:flex-row">
              <Skeleton className="h-40 w-full rounded-xl lg:w-60 lg:shrink-0" />
              <Skeleton className="h-[236px] flex-1 rounded-xl" />
            </div>
          ))}
        </div>
      ) : (
        <div className={cn('flex flex-col gap-6')}>
          {rowData.map((row) => (
            <div key={row.type} className="flex flex-col items-stretch gap-5 lg:flex-row">
              <MetricCard title={row.title} unit="%" type={row.type} latest={row.latest} icon={row.icon} />
              <MetricChart
                title={row.chartTitle}
                type={row.type}
                data={row.data}
                timeRange={timeRange}
                hostId={hostId}
                config={agentConfig}
              />
            </div>
          ))}
        </div>
      )}
    </section>
  )
}
