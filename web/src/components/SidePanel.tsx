import { Hand, Loader2, MicOff, X } from 'lucide-react'
import { useEffect, useRef } from 'react'
import { avatarColor, initials } from '../avatar'
import type { Participant, Speech } from '../types'
import type { Panel } from './ControlBar'

interface SidePanelProps {
  panel: Exclude<Panel, null>
  participants: Participant[]
  speeches: Speech[]
  connected: boolean
  onClose: () => void
}

const TITLES = { people: 'Participantes', chat: 'Transcrição da chamada', info: 'Detalhes da reunião' }

export function SidePanel({ panel, participants, speeches, connected, onClose }: SidePanelProps) {
  return (
    <aside className="side-panel">
      <header>
        <h2>{TITLES[panel]}</h2>
        <button className="icon" onClick={onClose} aria-label="Fechar">
          <X size={20} />
        </button>
      </header>
      {panel === 'people' && <People participants={participants} />}
      {panel === 'chat' && <Transcript speeches={speeches} />}
      {panel === 'info' && (
        <div className="panel-body info">
          <p>Seus agentes do Herdr entram na chamada automaticamente e reportam o que estão fazendo.</p>
          <p>
            Conexão com o servidor: <strong>{connected ? 'ativa' : 'reconectando…'}</strong>
          </p>
        </div>
      )}
    </aside>
  )
}

function People({ participants }: { participants: Participant[] }) {
  return (
    <ul className="panel-body people">
      <li>
        <span className="mini-avatar" style={{ background: avatarColor('self') }}>
          V
        </span>
        <div>
          <strong>Você</strong>
          <small>Organizador da reunião</small>
        </div>
      </li>
      {participants.map((participant) => (
        <li key={participant.id}>
          <span className="mini-avatar" style={{ background: avatarColor(participant.id) }}>
            {initials(participant.name)}
          </span>
          <div>
            <strong>{participant.name}</strong>
            <small>
              {participant.kind} · {participant.id}
            </small>
          </div>
          {participant.status === 'blocked' ? (
            <Hand size={18} className="hand-icon" />
          ) : participant.status === 'working' ? (
            <Loader2 size={18} className="spin" />
          ) : (
            <MicOff size={18} />
          )}
        </li>
      ))}
    </ul>
  )
}

function Transcript({ speeches }: { speeches: Speech[] }) {
  const end = useRef<HTMLDivElement>(null)
  useEffect(() => end.current?.scrollIntoView({ behavior: 'smooth' }), [speeches.length])

  if (speeches.length === 0) {
    return <div className="panel-body empty">Ninguém falou ainda.</div>
  }
  return (
    <div className="panel-body transcript">
      {speeches.map((speech) => (
        <div key={speech.id} className="message">
          <div className="message-head">
            <strong>{speech.name}</strong>
            <small>{new Date(speech.createdAt).toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' })}</small>
          </div>
          <p>{speech.text}</p>
        </div>
      ))}
      <div ref={end} />
    </div>
  )
}
