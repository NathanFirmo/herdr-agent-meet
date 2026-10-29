import { avatarColor, initials } from '../avatar'
import type { Participant } from '../types'

interface JoinScreenProps {
  participants: Participant[]
  connected: boolean
  left: boolean
  onJoin: () => void
}

export function JoinScreen({ participants, connected, left, onJoin }: JoinScreenProps) {
  return (
    <main className="join">
      <div className="join-preview">
        <div className="avatar" style={{ background: avatarColor('self') }}>
          V
        </div>
      </div>
      <div className="join-side">
        <h1>{left ? 'Você saiu da chamada' : 'Pronto para participar?'}</h1>
        <div className="join-people">
          {participants.slice(0, 5).map((participant) => (
            <span key={participant.id} className="mini-avatar" style={{ background: avatarColor(participant.id) }}>
              {initials(participant.name)}
            </span>
          ))}
        </div>
        <p>
          {!connected
            ? 'Conectando ao Herdr…'
            : participants.length === 0
              ? 'Nenhum agente na chamada ainda'
              : `${participants.length} ${participants.length === 1 ? 'agente está' : 'agentes estão'} na chamada`}
        </p>
        <button className="join-button" onClick={onJoin}>
          {left ? 'Participar novamente' : 'Participar agora'}
        </button>
      </div>
    </main>
  )
}
