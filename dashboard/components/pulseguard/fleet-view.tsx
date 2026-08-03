'use client'

import { ChevronRight, Server } from 'lucide-react'
import { Card } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'
import { UsageBar } from './usage-bar'
import type { HostSummary } from '@/lib/pulseguard'

interface FleetViewProps {
  hosts: HostSummary[]
  loading: boolean
  onSelect: (id: string) => void
}

export function FleetView({ hosts, loading, onSelect }: FleetViewProps) {
  const onlineCount = hosts.filter((h) => h.isOnline).length

  return (
    <section aria-labelledby="fleet-heading">
      <div className="mb-5 flex items-center justify-between gap-4">
        <h2 id="fleet-heading" className="text-lg font-semibold text-foreground">
          Connected Servers <span className="text-muted-foreground">(Live)</span>
        </h2>
        {hosts.length > 0 && (
          <span className="text-sm tabular-nums text-muted-foreground">
            <span className="font-semibold text-success">{onlineCount}</span> / {hosts.length} online
          </span>
        )}
      </div>

      {loading && hosts.length === 0 && (
        <div className="flex flex-col gap-3">
          {[0, 1, 2].map((i) => (
            <Skeleton key={i} className="h-[76px] w-full rounded-xl" />
          ))}
        </div>
      )}

      <div className="flex flex-col gap-3">
        {hosts.map((host) => (
          <Card
            key={host.id}
            role="button"
            tabIndex={0}
            aria-label={`View server ${host.id}`}
            onClick={() => onSelect(host.id)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault()
                onSelect(host.id)
              }
            }}
            className={cn(
              'group relative cursor-pointer overflow-hidden p-0 transition-all duration-200',
              'hover:-translate-y-0.5 hover:border-border/80 hover:bg-accent/40 hover:shadow-lg',
              'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background',
              !host.isOnline && 'opacity-60',
            )}
          >
            <div
              className={cn('absolute inset-y-0 left-0 w-1', host.isOnline ? 'bg-success' : 'bg-muted-foreground/30')}
              aria-hidden="true"
            />
            <div className="flex flex-col gap-4 py-5 pl-6 pr-5 sm:flex-row sm:items-center sm:justify-between sm:gap-6">
              <div className="flex items-center gap-3 sm:w-52">
                <span className="relative flex size-2.5 shrink-0">
                  {host.isOnline && (
                    <span className="absolute inline-flex size-full animate-ping rounded-full bg-success opacity-60" />
                  )}
                  <span
                    className={cn(
                      'relative inline-flex size-2.5 rounded-full',
                      host.isOnline ? 'bg-success' : 'bg-muted-foreground/40',
                    )}
                  />
                </span>
                <Server className="size-4 shrink-0 text-muted-foreground" aria-hidden="true" />
                <span
                  className={cn(
                    'truncate font-medium',
                    host.isOnline ? 'text-foreground' : 'text-muted-foreground',
                  )}
                >
                  {host.id}
                </span>
              </div>

              <div className="grid flex-1 grid-cols-3 gap-4 sm:flex sm:justify-center sm:gap-8">
                <UsageBar label="CPU" metric={host.cpu} />
                <UsageBar label="RAM" metric={host.ram} />
                <UsageBar label="DISK" metric={host.disk} />
              </div>

              <ChevronRight
                className="hidden size-5 shrink-0 text-muted-foreground transition-transform duration-200 group-hover:translate-x-0.5 group-hover:text-foreground sm:block"
                aria-hidden="true"
              />
            </div>
          </Card>
        ))}
      </div>
    </section>
  )
}
