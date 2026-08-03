import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import { levelStyles, type MetricEvent, type MetricType } from '@/lib/pulseguard'

const ICON_TINT: Record<MetricType, string> = {
  cpu: 'text-[var(--chart-1)]',
  ram: 'text-[var(--chart-2)]',
  disk: 'text-[var(--chart-3)]',
}

interface MetricCardProps {
  title: string
  unit: string
  type: MetricType
  latest: MetricEvent | null
  icon: React.ReactNode
}

export function MetricCard({ title, unit, type, latest, icon }: MetricCardProps) {
  if (!latest) return null

  const styles = levelStyles(latest.level)
  const measureTime = new Date(latest.timestamp).toLocaleTimeString('en-US')

  return (
    <Card className="flex w-full flex-col gap-4 p-5 lg:w-60 lg:shrink-0">
      <div className="flex items-start justify-between gap-2">
        <div className="flex items-center gap-2.5">
          <span className={cn('flex size-8 items-center justify-center rounded-lg bg-secondary', ICON_TINT[type])}>
            {icon}
          </span>
          <h3 className="text-xs font-medium uppercase tracking-wide text-muted-foreground">{title}</h3>
        </div>
        <Badge variant="outline" className={cn('border font-semibold', styles.badge)}>
          {latest.level}
        </Badge>
      </div>

      <div className="flex flex-1 items-center justify-center py-2">
        <div className={cn('flex items-baseline', styles.text)}>
          <span className="text-5xl font-semibold leading-none tabular-nums tracking-tight">
            {latest.metric_value.toFixed(1)}
          </span>
          <span className="ml-1.5 text-lg font-medium text-muted-foreground">{unit}</span>
        </div>
      </div>

      <p className="text-center text-xs tabular-nums text-muted-foreground/70">Last measurement: {measureTime}</p>
    </Card>
  )
}
