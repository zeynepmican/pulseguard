import { cn } from '@/lib/utils'
import { levelStyles, type MetricEvent } from '@/lib/pulseguard'

interface UsageBarProps {
  label: string
  metric: MetricEvent | null
}

export function UsageBar({ label, metric }: UsageBarProps) {
  if (!metric) {
    return (
      <div className="w-full min-w-0 sm:w-32">
        <div className="mb-1.5 flex items-center justify-between text-xs">
          <span className="font-medium text-muted-foreground">{label}</span>
          <span className="tabular-nums text-muted-foreground/60">—%</span>
        </div>
        <div className="h-2 rounded-full bg-secondary" />
      </div>
    )
  }

  const styles = levelStyles(metric.level)
  const value = Math.min(Math.max(metric.metric_value, 0), 100)

  return (
    <div className="w-full min-w-0 sm:w-32">
      <div className="mb-1.5 flex items-center justify-between text-xs">
        <span className="font-medium text-muted-foreground">{label}</span>
        <span className={cn('font-semibold tabular-nums', styles.text)}>{value.toFixed(0)}%</span>
      </div>
      <div
        className="h-2 overflow-hidden rounded-full bg-secondary"
        role="progressbar"
        aria-label={`${label} usage`}
        aria-valuenow={Math.round(value)}
        aria-valuemin={0}
        aria-valuemax={100}
      >
        <div
          className={cn('h-full rounded-full transition-[width,background-color] duration-500 ease-out', styles.bar)}
          style={{ width: `${value}%` }}
        />
      </div>
    </div>
  )
}
