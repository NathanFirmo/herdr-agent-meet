import { Hand, Loader2, MicOff } from 'lucide-react'
import { useEffect, useRef } from 'react'
import { avatarColor, initials } from '../avatar'
import type { Participant } from '../types'

interface AgentTileProps {
  participant: Participant
  speaking: boolean
}

export function AgentTile({ participant, speaking }: AgentTileProps) {
  return (
    <div className={`tile ${speaking ? 'speaking' : ''}`}>
      <Avatar id={participant.id} name={participant.name} speaking={speaking} />
      <div className="tile-badges">
        {participant.status === 'blocked' && (
          <span className="badge hand" title="Precisa de aprovação">
            <Hand size={16} />
          </span>
        )}
        {speaking ? (
          <span className="badge talking" aria-label="Falando">
            <i />
            <i />
            <i />
          </span>
        ) : participant.status === 'working' ? (
          <span className="badge" title="Trabalhando">
            <Loader2 size={16} className="spin" />
          </span>
        ) : (
          <span className="badge" title="Em silêncio">
            <MicOff size={16} />
          </span>
        )}
      </div>
      <div className="tile-name">
        <span>{participant.name}</span>
        <small>{participant.kind}</small>
      </div>
    </div>
  )
}

interface SelfTileProps {
  stream: MediaStream | null
}

export function SelfTile({ stream }: SelfTileProps) {
  const video = useRef<HTMLVideoElement>(null)

  useEffect(() => {
    if (video.current) video.current.srcObject = stream
  }, [stream])

  return (
    <div className="tile self">
      {stream ? <video ref={video} autoPlay muted playsInline /> : <Avatar id="self" name="Você" speaking={false} />}
      <div className="tile-badges">
        <span className="badge">
          <MicOff size={16} />
        </span>
      </div>
      <div className="tile-name">
        <span>Você</span>
      </div>
    </div>
  )
}

function Avatar({ id, name, speaking }: { id: string; name: string; speaking: boolean }) {
  return (
    <div className={`avatar ${speaking ? 'pulse' : ''}`} style={{ background: avatarColor(id) }}>
      {initials(name)}
    </div>
  )
}
