import { Hand, Loader2, Pencil, Trash2, UserPlus, X } from 'lucide-react'
import { useState } from 'react'
import { avatarColor, initials } from '../avatar'
import { agentLabel } from '../model'
import type { ChatMessage, Participant, Speech } from '../types'
import { ActivityBadge } from './ActivityBadge'
import { AudioTest } from './AudioTest'
import { Chat } from './Chat'
import { NewAgentDialog } from './NewAgentDialog'
import type { Panel } from './ControlBar'
import { t } from '../i18n'

interface SidePanelProps {
  panel: Exclude<Panel, null>
  participants: Participant[]
  speeches: Speech[]
  messages: ChatMessage[]
  connected: boolean
  onSend: (text: string, mentions: string[], images: File[]) => Promise<void>
  onRename: (id: string, name: string) => Promise<void>
  onRemove: (id: string) => Promise<void>
  onLaunch: (agent: { name: string; command: string; dir: string; prompt: string }) => Promise<void>
  draft: string
  onDraftChange: (draft: string) => void
  onClose: () => void
}


export function SidePanel({ panel, participants, speeches, messages, connected, onSend, onRename, onRemove, onLaunch, draft, onDraftChange, onClose }: SidePanelProps) {
  return (
    <aside className="side-panel">
      <header>
        <h2>{{ people: t.participants, chat: t.chatTitle, info: t.meetingDetails }[panel]}</h2>
        <button className="icon" onClick={onClose} aria-label={t.close}>
          <X size={20} />
        </button>
      </header>
      {panel === 'people' && <People participants={participants} onRename={onRename} onRemove={onRemove} onLaunch={onLaunch} />}
      {panel === 'chat' && <Chat participants={participants} speeches={speeches} messages={messages} onSend={onSend} draft={draft} onDraftChange={onDraftChange} />}
      {panel === 'info' && (
        <div className="panel-body info">
          <p>{t.infoJoin}</p>
          <p>{t.infoChat}</p>
          <p>
            {t.serverConnection} <strong>{connected ? t.connected : t.reconnecting}</strong>
          </p>
          <AudioTest variant="light" />
        </div>
      )}
    </aside>
  )
}

interface PeopleProps {
  participants: Participant[]
  onRename: (id: string, name: string) => Promise<void>
  onRemove: (id: string) => Promise<void>
  onLaunch: (agent: { name: string; command: string; dir: string; prompt: string }) => Promise<void>
}

function People({ participants, onRename, onRemove, onLaunch }: PeopleProps) {
  const [adding, setAdding] = useState(false)
  const dirs = [...new Set(participants.map((participant) => participant.dir).filter((dir): dir is string => Boolean(dir)))]

  return (
    <ul className="panel-body people">
      <li className="add-agent">
        <button className="add-agent-button" onClick={() => setAdding(true)}>
          <UserPlus size={18} /> {t.newAgent}
        </button>
        {adding && <NewAgentDialog dirs={dirs} onLaunch={onLaunch} onClose={() => setAdding(false)} />}
      </li>
      <li>
        <span className="mini-avatar" style={{ background: avatarColor('self') }}>
          {initials(t.you)}
        </span>
        <div>
          <strong>{t.you}</strong>
          <small>{t.organizer}</small>
        </div>
      </li>
      {participants.map((participant) => (
        <PersonRow key={participant.id} participant={participant} onRename={onRename} onRemove={onRemove} />
      ))}
    </ul>
  )
}

function EditableName({ participant, onRename }: { participant: Participant; onRename: (id: string, name: string) => Promise<void> }) {
  const [editing, setEditing] = useState(false)
  const [value, setValue] = useState(participant.name)

  const startEditing = () => {
    setValue(participant.name)
    setEditing(true)
  }

  const save = async () => {
    setEditing(false)
    if (value.trim() !== participant.name) await onRename(participant.id, value).catch(() => undefined)
  }

  if (editing) {
    return (
      <input
        className="rename"
        autoFocus
        value={value}
        maxLength={40}
        placeholder={participant.originalName ?? participant.name}
        onChange={(event) => setValue(event.target.value)}
        onBlur={save}
        onKeyDown={(event) => {
          if (event.key === 'Enter') event.currentTarget.blur()
          if (event.key === 'Escape') setEditing(false)
        }}
      />
    )
  }

  return (
    <button
      className="name-button"
      onClick={startEditing}
      title={participant.originalName ? t.renameFrom(participant.originalName) : t.rename}
    >
      <strong>{participant.name}</strong>
      <Pencil size={13} className="pencil" />
    </button>
  )
}

interface PersonRowProps {
  participant: Participant
  onRename: (id: string, name: string) => Promise<void>
  onRemove: (id: string) => Promise<void>
}

function PersonRow({ participant, onRename, onRemove }: PersonRowProps) {
  const [confirming, setConfirming] = useState(false)
  const [removing, setRemoving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const remove = async () => {
    setRemoving(true)
    setError(null)
    try {
      await onRemove(participant.id)
    } catch {
      setError(t.closeFailed)
      setRemoving(false)
    }
  }

  return (
    <li className={confirming ? 'confirming' : ''}>
      <span className="mini-avatar" style={{ background: avatarColor(participant.id) }}>
        {initials(participant.name)}
      </span>
      <div>
        <EditableName participant={participant} onRename={onRename} />
        {confirming ? (
          <small className="remove-warning">{error ?? t.endAgentWarning}</small>
        ) : participant.activity ? (
          <ActivityBadge activity={participant.activity} compact />
        ) : (
          <small>
            {agentLabel(participant)} · {participant.id}
          </small>
        )}
        {confirming && (
          <span className="remove-actions">
            <button className="text-button" onClick={() => setConfirming(false)} disabled={removing}>
              {t.cancel}
            </button>
            <button className="danger-button" onClick={remove} disabled={removing}>
              {removing ? t.ending : t.end}
            </button>
          </span>
        )}
      </div>
      {participant.status === 'blocked' && <Hand size={18} className="hand-icon" />}
      {participant.status === 'working' && <Loader2 size={18} className="spin" />}
      {!confirming && (
        <button className="icon remove" onClick={() => setConfirming(true)} aria-label={t.endNamed(participant.name)} title={t.endAgent}>
          <Trash2 size={18} />
        </button>
      )}
    </li>
  )
}
