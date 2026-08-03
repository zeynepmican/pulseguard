export const API_BASE = 'http://localhost:8081'

export type MetricType = 'cpu' | 'ram' | 'disk'

export interface MetricEvent {
  agent_id?: string
  check_type: string
  metric_value: number
  level: string
  timestamp: string
}

export interface HostSummary {
  id: string
  isOnline: boolean
  cpu: MetricEvent | null
  ram: MetricEvent | null
  disk: MetricEvent | null
}

export interface AgentConfig {
  agent_id: string
  cpu_warn: number | string
  cpu_err: number | string
  ram_warn: number | string
  ram_err: number | string
  disk_warn: number | string
  disk_err: number | string
}

export interface ChartPoint {
  timestamp: number
  value: number
  level: string
}

export const DEFAULT_CONFIG = (agentId: string): AgentConfig => ({
  agent_id: agentId,
  cpu_warn: 75,
  cpu_err: 90,
  ram_warn: 80,
  ram_err: 95,
  disk_warn: 85,
  disk_err: 95,
})

/**
 * Semantic Tailwind classes for a given severity level.
 * Mirrors the original getStatusColor() logic (ERROR / WARNING / else).
 */
export function levelStyles(level: string) {
  if (level === 'ERROR') {
    return {
      key: 'error' as const,
      text: 'text-error',
      bar: 'bg-error',
      badge: 'bg-error/15 text-error border-error/25',
      dot: 'bg-error',
      glow: 'shadow-[0_0_0_1px_var(--error)]',
    }
  }
  if (level === 'WARNING') {
    return {
      key: 'warning' as const,
      text: 'text-warning',
      bar: 'bg-warning',
      badge: 'bg-warning/15 text-warning border-warning/25',
      dot: 'bg-warning',
      glow: 'shadow-[0_0_0_1px_var(--warning)]',
    }
  }
  return {
    key: 'success' as const,
    text: 'text-success',
    bar: 'bg-success',
    badge: 'bg-success/15 text-success border-success/25',
    dot: 'bg-success',
    glow: 'shadow-[0_0_0_1px_var(--success)]',
  }
}

export const CHART_VAR: Record<MetricType, string> = {
  cpu: 'var(--chart-1)',
  ram: 'var(--chart-2)',
  disk: 'var(--chart-3)',
}

// 1. SINGLE-METRIC CSV EXPORT (button in the chart header)
export function downloadCSV(
  hostId: string,
  type: string,
  data: ChartPoint[],
  timeRange: number,
  config: AgentConfig | null,
) {
  if (!data || data.length === 0) {
    alert('No data available to export.')
    return
  }

  const warnThresh = config ? (config as any)[`${type}_warn`] : 'N/A'
  const errThresh = config ? (config as any)[`${type}_err`] : 'N/A'
  const timeLabel = timeRange >= 60 ? `Last ${timeRange / 60} hour(s)` : `Last ${timeRange} minute(s)`
  const exportDate = new Date().toLocaleString('en-US')

  const metadataBlock = [
    `"========================================="`,
    `"PULSEGUARD METRIC AUDIT REPORT"`,
    `"========================================="`,
    `"Host ID:","${hostId}"`,
    `"Check Type:","${type.toUpperCase()}"`,
    `"Active Warning Threshold (Warning Limit):","${warnThresh}%"`,
    `"Active Critical Threshold (Error Limit):","${errThresh}%"`,
    `"Reporting Range:","${timeLabel}"`,
    `"Report Generated On:","${exportDate}"`,
    `"========================================="`,
    '',
  ]

  const headers = ['Date / Time', 'Measured Value (%)', 'Status Level']

  const rows = data.map((item) => {
    const readableDate = new Date(item.timestamp).toLocaleString('en-US')
    return [`"${readableDate}"`, `"${item.value.toFixed(2)}"`, `"${item.level}"`]
  })

  const csvContent =
    '\uFEFF' + [...metadataBlock, headers.join(','), ...rows.map((e) => e.join(','))].join('\n')

  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')

  const fileTimeLabel = timeRange >= 60 ? `${timeRange / 60}Hour` : `${timeRange}Min`
  link.setAttribute('href', url)
  link.setAttribute('download', `PulseGuard_${hostId}_${type.toUpperCase()}_${fileTimeLabel}.csv`)

  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

// 2. BULK REPORT CSV EXPORT (button in the top bar)
export function downloadAllCSV(
  hostId: string,
  allEvents: MetricEvent[],
  timeRange: number,
  config: AgentConfig | null,
) {
  if (!allEvents || allEvents.length === 0) {
    alert('No data available to export.')
    return
  }

  const latestTime = new Date(allEvents[0].timestamp).getTime()
  const cutoffTime = latestTime - timeRange * 60 * 1000
  const filteredEvents = allEvents.filter((e) => new Date(e.timestamp).getTime() >= cutoffTime)

  if (filteredEvents.length === 0) {
    alert('No data available to export for the selected time range.')
    return
  }

  const timeLabel = timeRange >= 60 ? `Last ${timeRange / 60} hour(s)` : `Last ${timeRange} minute(s)`
  const exportDate = new Date().toLocaleString('en-US')

  const getThresh = (type: string, level: string) =>
    config && (config as any)[`${type}_${level}`] ? `${(config as any)[`${type}_${level}`]}%` : 'N/A'

  const metadataBlock = [
    `"========================================="`,
    `"PULSEGUARD - FULL SYSTEM AUDIT REPORT"`,
    `"========================================="`,
    `"Host ID:","${hostId}"`,
    `"Reporting Range:","${timeLabel}"`,
    `"Report Generated On:","${exportDate}"`,
    `"-----------------------------------------"`,
    `"ACTIVE THRESHOLD LIMITS (LATEST THRESHOLDS)"`,
    `"CPU Warning / Critical Threshold:","Warn: ${getThresh('cpu', 'warn')} | Err: ${getThresh('cpu', 'err')}"`,
    `"RAM Warning / Critical Threshold:","Warn: ${getThresh('ram', 'warn')} | Err: ${getThresh('ram', 'err')}"`,
    `"DISK Warning / Critical Threshold:","Warn: ${getThresh('disk', 'warn')} | Err: ${getThresh('disk', 'err')}"`,
    `"========================================="`,
    '',
  ]

  const headers = ['Date / Time', 'Check Type', 'Measured Value (%)', 'Status Level']

  const rows = filteredEvents.map((item) => {
    const readableDate = new Date(item.timestamp).toLocaleString('en-US')
    return [
      `"${readableDate}"`,
      `"${item.check_type.toUpperCase()}"`,
      `"${item.metric_value.toFixed(2)}"`,
      `"${item.level}"`,
    ]
  })

  const csvContent =
    '\uFEFF' + [...metadataBlock, headers.join(','), ...rows.map((e) => e.join(','))].join('\n')

  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')

  const fileTimeLabel = timeRange >= 60 ? `${timeRange / 60}Hour` : `${timeRange}Min`
  link.setAttribute('href', url)
  link.setAttribute('download', `PulseGuard_${hostId}_FULL_REPORT_${fileTimeLabel}.csv`)

  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}
