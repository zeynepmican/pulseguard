'use client'

import { useEffect, useState } from 'react'
import { Activity, ArrowLeft } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { FleetView } from './fleet-view'
import { HostDetail } from './host-detail'
import { API_BASE, DEFAULT_CONFIG, type AgentConfig, type HostSummary, type MetricEvent } from '@/lib/pulseguard'

export function Dashboard() {
  const [fleetData, setFleetData] = useState<HostSummary[]>([])
  const [selectedHost, setSelectedHost] = useState<string | null>(null)
  const [hostEvents, setHostEvents] = useState<MetricEvent[]>([])
  const [loading, setLoading] = useState(true)
  const [timeRange, setTimeRange] = useState(5)

  const [agentConfig, setAgentConfig] = useState<AgentConfig | null>(null)
  const [isSettingsOpen, setIsSettingsOpen] = useState(false)
  const [isSaving, setIsSaving] = useState(false)

  // 1. SILENTLY REFRESH FLEET DATA EVERY 5 SECONDS
  useEffect(() => {
    const fetchFleetData = async () => {
      try {
        const hostsRes = await fetch(`${API_BASE}/api/hosts`)
        if (!hostsRes.ok) return
        const hostIds: string[] = await hostsRes.json()

        const fleetPromises = hostIds.map(async (id) => {
          const eventsRes = await fetch(`${API_BASE}/api/events?agent_id=${id}`)
          const events: MetricEvent[] = await eventsRes.json()

          if (events && Array.isArray(events)) {
            events.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
          }

          const getLatest = (type: string) => events?.find((e) => e.check_type === type) || null
          let isOnline = false
          if (events && events.length > 0) {
            const latestTime = new Date(events[0].timestamp).getTime()
            const fiveMinsAgo = new Date().getTime() - 5 * 60 * 1000
            isOnline = latestTime > fiveMinsAgo
          }
          return { id, isOnline, cpu: getLatest('cpu'), ram: getLatest('ram'), disk: getLatest('disk') }
        })

        const results = await Promise.all(fleetPromises)
        results.sort((a, b) => (a.isOnline === b.isOnline ? 0 : a.isOnline ? -1 : 1))
        setFleetData(results)
        setLoading(false)
      } catch (err) {
        console.error('Error while fetching fleet data:', err)
      }
    }

    fetchFleetData()
    const interval = setInterval(fetchFleetData, 5000)
    return () => clearInterval(interval)
  }, [])

  // 2. SMOOTHLY REFRESH THE SELECTED HOST'S LOGS AND SETTINGS EVERY 5 SECONDS
  useEffect(() => {
    if (!selectedHost) return

    const fetchConfig = () => {
      fetch(`${API_BASE}/api/configs/${selectedHost}`)
        .then((res) => res.json())
        .then((data) => {
          if (!isSettingsOpen) {
            if (data && (data.cpu_warn || data.cpu_err)) {
              setAgentConfig(data)
            } else {
              setAgentConfig(DEFAULT_CONFIG(selectedHost))
            }
          }
        })
        .catch((err) => {
          console.error('Failed to fetch settings, applying defaults', err)
          if (!isSettingsOpen) {
            setAgentConfig(DEFAULT_CONFIG(selectedHost))
          }
        })
    }

    const fetchEvents = () => {
      fetch(`${API_BASE}/api/events?agent_id=${selectedHost}`)
        .then((res) => res.json())
        .then((data: MetricEvent[]) => {
          if (data && Array.isArray(data)) {
            data.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
          }
          setHostEvents(data || [])
          setLoading(false)
        })
        .catch((err) => console.error('Error while fetching events:', err))
    }

    setLoading(true)
    fetchConfig()
    fetchEvents()

    const interval = setInterval(() => {
      fetchEvents()
      fetchConfig()
    }, 5000)

    return () => clearInterval(interval)
  }, [selectedHost, isSettingsOpen])

  const saveConfig = async () => {
    if (!agentConfig) return
    const requiredKeys: (keyof AgentConfig)[] = ['cpu_warn', 'cpu_err', 'ram_warn', 'ram_err', 'disk_warn', 'disk_err']
    for (const key of requiredKeys) {
      if (agentConfig[key] === '' || agentConfig[key] === null || agentConfig[key] === undefined) {
        alert('Please fill in all threshold values. Fields cannot be left empty.')
        return
      }
    }

    setIsSaving(true)
    try {
      await fetch(`${API_BASE}/api/configs/${selectedHost}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ...agentConfig,
          cpu_warn: Number(agentConfig.cpu_warn),
          cpu_err: Number(agentConfig.cpu_err),
          ram_warn: Number(agentConfig.ram_warn),
          ram_err: Number(agentConfig.ram_err),
          disk_warn: Number(agentConfig.disk_warn),
          disk_err: Number(agentConfig.disk_err),
        }),
      })
      setIsSettingsOpen(false)
    } catch (error) {
      alert('An error occurred while saving.')
    }
    setIsSaving(false)
  }

  const handleNumberChange = (key: string, value: string) => {
    if (value === '' || /^[0-9]+$/.test(value)) {
      setAgentConfig((prev) => (prev ? { ...prev, [key]: value } : prev))
    }
  }

  return (
    <main className="min-h-screen bg-background">
      <div className="mx-auto w-full max-w-[1400px] px-4 py-8 sm:px-6 sm:py-10 lg:px-8 lg:py-12">
        <header className="mb-8 flex flex-col gap-4 sm:mb-10 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-center gap-3">
            <span className="flex size-11 items-center justify-center rounded-xl bg-primary/15 text-primary ring-1 ring-inset ring-primary/25">
              <Activity className="size-6" aria-hidden="true" />
            </span>
            <div>
              <h1 className="text-2xl font-semibold tracking-tight text-foreground">PulseGuard</h1>
              <p className="text-sm text-muted-foreground">Enterprise Fleet Monitoring Center</p>
            </div>
          </div>

          {selectedHost && (
            <Button
              variant="secondary"
              className="gap-1.5 self-start sm:self-auto"
              onClick={() => {
                setSelectedHost(null)
                setIsSettingsOpen(false)
              }}
            >
              <ArrowLeft className="size-4" aria-hidden="true" />
              Back to Fleet View
            </Button>
          )}
        </header>

        {!selectedHost ? (
          <FleetView hosts={fleetData} loading={loading} onSelect={(id) => setSelectedHost(id)} />
        ) : (
          <HostDetail
            hostId={selectedHost}
            hostEvents={hostEvents}
            loading={loading}
            timeRange={timeRange}
            onTimeRangeChange={setTimeRange}
            agentConfig={agentConfig}
            isSettingsOpen={isSettingsOpen}
            onToggleSettings={() => setIsSettingsOpen((v) => !v)}
            isSaving={isSaving}
            onFieldChange={handleNumberChange}
            onSave={saveConfig}
            onCancelSettings={() => setIsSettingsOpen(false)}
          />
        )}
      </div>
    </main>
  )
}
