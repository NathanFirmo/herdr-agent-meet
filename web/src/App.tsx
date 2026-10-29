import { useCallback, useState } from 'react'
import { ControlBar, type Panel } from './components/ControlBar'
import { JoinScreen } from './components/JoinScreen'
import { PromptCards } from './components/PromptCards'
import { SidePanel } from './components/SidePanel'
import { AgentTile, SelfTile } from './components/Tile'
import { avatarColor, initials } from './avatar'
import { useMeeting } from './useMeeting'
import { useSpeechRecognition } from './useSpeechRecognition'

export default function App() {
  const [joined, setJoined] = useState(false)
  const [left, setLeft] = useState(false)
  const [captions, setCaptions] = useState(false)
  const [panel, setPanel] = useState<Panel>(null)
  const [draft, setDraft] = useState('')
  const onDictated = useCallback((text: string) => {
    setDraft((current) => (current.trim() ? `${current.trimEnd()} ${text}` : text))
    setPanel('chat')
  }, [])
  const speech = useSpeechRecognition(onDictated)
  const meeting = useMeeting(joined, speech.listening)

  const leave = () => {
    speech.stop()
    setJoined(false)
    setLeft(true)
  }

  if (!joined) {
    return <JoinScreen participants={meeting.participants} connected={meeting.connected} left={left} onJoin={() => setJoined(true)} />
  }

  const tileCount = meeting.participants.length + 1
  const columns = Math.ceil(Math.sqrt(tileCount))
  const rows = Math.ceil(tileCount / columns)
  const speaker = meeting.current

  return (
    <div className="meeting">
      <div className="stage">
        <PromptCards participants={meeting.participants} onAnswer={meeting.answerPrompt} />
        <section className="grid" style={{ gridTemplateColumns: `repeat(${columns}, 1fr)`, gridTemplateRows: `repeat(${rows}, 1fr)` }}>
          {meeting.participants.map((participant) => (
            <AgentTile key={participant.id} participant={participant} speaking={speaker?.participantId === participant.id} />
          ))}
          <SelfTile listening={speech.listening} speaking={speech.interim !== ''} />
        </section>
        {panel && (
          <SidePanel
            panel={panel}
            participants={meeting.participants}
            speeches={meeting.speeches}
            messages={meeting.messages}
            connected={meeting.connected}
            onSend={meeting.sendMessage}
            onRename={meeting.renameParticipant}
            onRemove={meeting.removeParticipant}
            draft={draft}
            onDraftChange={setDraft}
            onClose={() => setPanel(null)}
          />
        )}
      </div>

      {captions && speech.interim && (
        <div className="captions">
          <span className="mini-avatar" style={{ background: avatarColor('self') }}>
            V
          </span>
          <div>
            <strong>Você</strong>
            <p>{speech.interim}</p>
          </div>
        </div>
      )}

      {captions && speaker && !speech.interim && (
        <div className="captions">
          <span className="mini-avatar" style={{ background: avatarColor(speaker.participantId) }}>
            {initials(speaker.name)}
          </span>
          <div>
            <strong>{speaker.name}</strong>
            <p>{speaker.text}</p>
          </div>
        </div>
      )}

      {speech.listening && meeting.pending > 0 && (
        <div className="held-notice">
          Microfone ligado: {meeting.pending} {meeting.pending === 1 ? 'fala em espera' : 'falas em espera'}. Desligue para ouvir os agentes.
        </div>
      )}

      <div className="toasts">
        {speech.error && <div className="toast">🎤 {speech.error}</div>}
        {meeting.toasts.map((toast) => (
          <div key={toast.id} className="toast">
            ✋ {toast.text}
          </div>
        ))}
      </div>

      <ControlBar
        micSupported={speech.supported}
        micOn={speech.listening}
        onToggleMic={speech.listening ? speech.stop : speech.start}
        muted={meeting.muted}
        onToggleMuted={() => meeting.setMuted(!meeting.muted)}
        captions={captions}
        onToggleCaptions={() => setCaptions(!captions)}
        onLeave={leave}
        panel={panel}
        onPanel={setPanel}
        participantCount={tileCount}
      />
    </div>
  )
}
