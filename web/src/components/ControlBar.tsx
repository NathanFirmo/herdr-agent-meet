import { Captions, Info, MessageSquareText, Mic, MicOff, PhoneOff, Users, Volume2, VolumeX } from 'lucide-react'
import { useEffect, useState } from 'react'

export type Panel = 'people' | 'chat' | 'info' | null

interface ControlBarProps {
  micSupported: boolean
  micOn: boolean
  onToggleMic: () => void
  muted: boolean
  onToggleMuted: () => void
  captions: boolean
  onToggleCaptions: () => void
  onLeave: () => void
  panel: Panel
  onPanel: (panel: Panel) => void
  participantCount: number
}

export function ControlBar(props: ControlBarProps) {
  const now = useClock()
  const togglePanel = (panel: Panel) => props.onPanel(props.panel === panel ? null : panel)

  return (
    <footer className="control-bar">
      <div className="meeting-info">
        <span>{now}</span>
        <span className="divider">|</span>
        <span>Agent Meet</span>
      </div>

      <div className="controls">
        <RoundButton
          active={props.micOn}
          onClick={props.onToggleMic}
          disabled={!props.micSupported}
          label={!props.micSupported ? 'Ditado por voz só funciona no Chrome' : props.micOn ? 'Desativar microfone' : 'Ditar mensagem por voz'}
        >
          {props.micOn ? <Mic size={20} /> : <MicOff size={20} />}
        </RoundButton>
        <RoundButton active={!props.muted} onClick={props.onToggleMuted} label={props.muted ? 'Ouvir agentes' : 'Silenciar agentes'}>
          {props.muted ? <VolumeX size={20} /> : <Volume2 size={20} />}
        </RoundButton>
        <RoundButton active={!props.captions} highlighted={props.captions} onClick={props.onToggleCaptions} label="Legendas">
          <Captions size={20} />
        </RoundButton>
        <button className="leave" onClick={props.onLeave} title="Sair da chamada">
          <PhoneOff size={20} />
        </button>
      </div>

      <div className="side-controls">
        <IconButton active={props.panel === 'info'} onClick={() => togglePanel('info')} label="Detalhes da reunião">
          <Info size={22} />
        </IconButton>
        <IconButton active={props.panel === 'people'} onClick={() => togglePanel('people')} label="Participantes">
          <Users size={22} />
          <span className="count">{props.participantCount}</span>
        </IconButton>
        <IconButton active={props.panel === 'chat'} onClick={() => togglePanel('chat')} label="Chat">
          <MessageSquareText size={22} />
        </IconButton>
      </div>
    </footer>
  )
}

function RoundButton(props: {
  active: boolean
  highlighted?: boolean
  disabled?: boolean
  onClick: () => void
  label: string
  children: React.ReactNode
}) {
  const className = props.highlighted ? 'round highlighted' : props.active ? 'round' : 'round off'
  return (
    <button className={className} onClick={props.onClick} disabled={props.disabled} title={props.label} aria-label={props.label}>
      {props.children}
    </button>
  )
}

function IconButton(props: { active: boolean; onClick: () => void; label: string; children: React.ReactNode }) {
  return (
    <button className={`icon ${props.active ? 'active' : ''}`} onClick={props.onClick} title={props.label} aria-label={props.label}>
      {props.children}
    </button>
  )
}

function useClock() {
  const format = () => new Date().toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' })
  const [now, setNow] = useState(format)
  useEffect(() => {
    const timer = setInterval(() => setNow(format()), 10_000)
    return () => clearInterval(timer)
  }, [])
  return now
}
