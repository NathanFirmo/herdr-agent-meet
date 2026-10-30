import { avatarColor, initials } from '../avatar'
import { AudioTest } from './AudioTest'
import type { Participant } from '../types'
import { t } from '../i18n'

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
          {initials(t.you)}
        </div>
      </div>
      <div className="join-side">
        <h1>{left ? t.leftCall : t.readyToJoin}</h1>
        <div className="join-people">
          {participants.slice(0, 5).map((participant) => (
            <span key={participant.id} className="mini-avatar" style={{ background: avatarColor(participant.id) }}>
              {initials(participant.name)}
            </span>
          ))}
        </div>
        <p>
          {!connected
            ? t.connecting
            : participants.length === 0
              ? t.noAgents
              : t.agentsInCall(participants.length)}
        </p>
        <button className="join-button" onClick={onJoin}>
          {left ? t.rejoin : t.joinNow}
        </button>
        <AudioTest />
      </div>
    </main>
  )
}
