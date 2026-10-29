import { Brain, TerminalSquare } from 'lucide-react'
import { useEffect, useState } from 'react'
import type { Activity } from '../types'

const STALL_SECONDS = 5 * 60

export function ActivityBadge({ activity, compact = false }: { activity: Activity; compact?: boolean }) {
  const seconds = useElapsedSeconds(activity.since)
  const stalled = seconds >= STALL_SECONDS
  const label = activity.kind === 'tool' ? activity.label || 'Rodando comando' : 'Pensando'

  return (
    <span
      className={`activity ${stalled ? 'stalled' : ''} ${compact ? 'compact' : ''}`}
      title={stalled ? 'Está nisso há bastante tempo, pode estar travado' : label}
    >
      {activity.kind === 'tool' ? <TerminalSquare size={14} /> : <Brain size={14} />}
      <span className="activity-label">{label}</span>
      <span className="activity-time">{formatElapsed(seconds)}</span>
      {stalled && !compact && <span className="activity-warning">pode estar travado</span>}
    </span>
  )
}

function useElapsedSeconds(since: string) {
  const [now, setNow] = useState(Date.now)
  useEffect(() => {
    const timer = setInterval(() => setNow(Date.now()), 1000)
    return () => clearInterval(timer)
  }, [])
  return Math.max(0, Math.floor((now - Date.parse(since)) / 1000))
}

function formatElapsed(total: number) {
  const hours = Math.floor(total / 3600)
  const minutes = Math.floor((total % 3600) / 60)
  const seconds = String(total % 60).padStart(2, '0')
  return hours > 0 ? `${hours}:${String(minutes).padStart(2, '0')}:${seconds}` : `${minutes}:${seconds}`
}
