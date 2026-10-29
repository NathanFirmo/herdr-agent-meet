import { useState } from 'react'
import { ControlBar, type Panel } from './components/ControlBar'
import { JoinScreen } from './components/JoinScreen'
import { SidePanel } from './components/SidePanel'
import { AgentTile, SelfTile } from './components/Tile'
import { avatarColor, initials } from './avatar'
import { useMeeting } from './useMeeting'

export default function App() {
  const [joined, setJoined] = useState(false)
  const [left, setLeft] = useState(false)
  const [captions, setCaptions] = useState(true)
  const [panel, setPanel] = useState<Panel>(null)
  const [camera, setCamera] = useState<MediaStream | null>(null)
  const meeting = useMeeting(joined)

  const toggleCamera = async () => {
    if (camera) {
      camera.getTracks().forEach((track) => track.stop())
      setCamera(null)
      return
    }
    try {
      setCamera(await navigator.mediaDevices.getUserMedia({ video: true }))
    } catch {
      setCamera(null)
    }
  }

  const leave = () => {
    camera?.getTracks().forEach((track) => track.stop())
    setCamera(null)
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
        <section className="grid" style={{ gridTemplateColumns: `repeat(${columns}, 1fr)`, gridTemplateRows: `repeat(${rows}, 1fr)` }}>
          {meeting.participants.map((participant) => (
            <AgentTile key={participant.id} participant={participant} speaking={speaker?.participantId === participant.id} />
          ))}
          <SelfTile stream={camera} />
        </section>
        {panel && (
          <SidePanel
            panel={panel}
            participants={meeting.participants}
            speeches={meeting.speeches}
            connected={meeting.connected}
            onClose={() => setPanel(null)}
          />
        )}
      </div>

      {captions && speaker && (
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

      <div className="toasts">
        {meeting.toasts.map((toast) => (
          <div key={toast.id} className="toast">
            ✋ {toast.text}
          </div>
        ))}
      </div>

      <ControlBar
        muted={meeting.muted}
        onToggleMuted={() => meeting.setMuted(!meeting.muted)}
        captions={captions}
        onToggleCaptions={() => setCaptions(!captions)}
        camera={camera !== null}
        onToggleCamera={toggleCamera}
        onLeave={leave}
        panel={panel}
        onPanel={setPanel}
        participantCount={tileCount}
      />
    </div>
  )
}
