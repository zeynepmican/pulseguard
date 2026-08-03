'use client'

import { CartesianGrid, Line, LineChart, XAxis, YAxis } from 'recharts'
import { Download } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { ChartContainer, ChartTooltip, ChartTooltipContent, type ChartConfig } from '@/components/ui/chart'
import {
  downloadCSV,
  type AgentConfig,
  type ChartPoint,
  type MetricType,
} from '@/lib/pulseguard'

interface MetricChartProps {
  title: string
  type: MetricType
  data: ChartPoint[]
  timeRange: number
  hostId: string
  config: AgentConfig | null
}

const CHART_CONFIG: Record<MetricType, ChartConfig> = {
  cpu: { value: { label: 'CPU', color: 'var(--chart-1)' } },
  ram: { value: { label: 'RAM', color: 'var(--chart-2)' } },
  disk: { value: { label: 'Disk', color: 'var(--chart-3)' } },
}

export function MetricChart({ title, type, data, timeRange, hostId, config }: MetricChartProps) {
  const rangeLabel = timeRange >= 60 ? `${timeRange / 60} hour(s)` : `${timeRange} minute(s)`

  if (!data || data.length === 0) {
    return (
      <Card className="flex flex-1 items-center justify-center p-6">
        <p className="text-sm text-muted-foreground">No data available for this time range.</p>
      </Card>
    )
  }

  const maxTime = data[data.length - 1].timestamp
  const minTime = maxTime - timeRange * 60 * 1000

  return (
    <Card className="flex flex-1 flex-col gap-0 py-4">
      <CardHeader className="flex flex-row items-center justify-between gap-3 space-y-0 px-4 pb-4">
        <CardTitle className="text-sm font-medium text-foreground">
          {title} Trend{' '}
          <span className="font-normal text-muted-foreground">(Last {rangeLabel})</span>
        </CardTitle>
        <Button
          variant="outline"
          size="sm"
          className="h-8 gap-1.5"
          onClick={() => downloadCSV(hostId, type, data, timeRange, config)}
          title={`Download ${title} data as CSV`}
        >
          <Download className="size-3.5" aria-hidden="true" />
          Download CSV
        </Button>
      </CardHeader>
      <CardContent className="px-2 pb-1">
        <ChartContainer config={CHART_CONFIG[type]} className="h-[180px] w-full">
          <LineChart data={data} margin={{ left: 4, right: 12, top: 4, bottom: 0 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" vertical={false} />
            <XAxis
              dataKey="timestamp"
              type="number"
              domain={[minTime, maxTime]}
              allowDataOverflow
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              tick={{ fontSize: 11, fill: 'var(--muted-foreground)' }}
              tickFormatter={(unixTime: number) => {
                const formatOptions: Intl.DateTimeFormatOptions = { hour: '2-digit', minute: '2-digit' }
                if (timeRange === 1) formatOptions.second = '2-digit'
                return new Date(unixTime).toLocaleTimeString([], formatOptions)
              }}
            />
            <YAxis
              domain={[0, 100]}
              tickLine={false}
              axisLine={false}
              tickMargin={8}
              width={32}
              tick={{ fontSize: 11, fill: 'var(--muted-foreground)' }}
            />
            <ChartTooltip
              content={
                <ChartTooltipContent
                  labelFormatter={(_, payload) => {
                    const ts = payload?.[0]?.payload?.timestamp
                    return ts
                      ? new Date(ts).toLocaleTimeString([], {
                          hour: '2-digit',
                          minute: '2-digit',
                          second: '2-digit',
                        })
                      : ''
                  }}
                  formatter={(value) => (
                    <span className="tabular-nums">{Number(value).toFixed(1)}%</span>
                  )}
                />
              }
            />
            <Line
              type="monotone"
              dataKey="value"
              stroke="var(--color-value)"
              strokeWidth={2.5}
              dot={false}
              isAnimationActive={false}
            />
          </LineChart>
        </ChartContainer>
      </CardContent>
    </Card>
  )
}
