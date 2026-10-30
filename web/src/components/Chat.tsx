import { Paperclip, SendHorizontal, X } from 'lucide-react'
import { Fragment, useEffect, useLayoutEffect, useMemo, useRef, useState, type ClipboardEvent, type DragEvent, type KeyboardEvent } from 'react'
import { avatarColor, initials } from '../avatar'
import { formatTime, mentionAll, t } from '../i18n'
import type { ChatMessage, Participant, Speech } from '../types'

const MAX_INPUT_HEIGHT = 120
const MAX_IMAGES = 5
const MAX_IMAGE_BYTES = 10 * 1024 * 1024
const IMAGE_TYPES = ['image/png', 'image/jpeg', 'image/gif', 'image/webp']

interface ChatProps {
  participants: Participant[]
  speeches: Speech[]
  messages: ChatMessage[]
  onSend: (text: string, mentions: string[], images: File[]) => Promise<void>
  draft: string
  onDraftChange: (draft: string) => void
}

type Entry =
  | { kind: 'speech'; id: string; at: number; speech: Speech }
  | { kind: 'message'; id: string; at: number; message: ChatMessage }

interface Suggestion {
  id: string
  name: string
}

export function Chat({ participants, speeches, messages, onSend, draft: text, onDraftChange: setText }: ChatProps) {
  const [query, setQuery] = useState<string | null>(null)
  const [highlight, setHighlight] = useState(0)
  const [sending, setSending] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [images, setImages] = useState<{ file: File; preview: string }[]>([])
  const [dragging, setDragging] = useState(false)
  const input = useRef<HTMLTextAreaElement>(null)
  const picker = useRef<HTMLInputElement>(null)
  const end = useRef<HTMLDivElement>(null)

  const entries = useMemo<Entry[]>(() => {
    const all: Entry[] = [
      ...joinTurns(speeches).map((speech) => ({ kind: 'speech' as const, id: speech.id, at: Date.parse(speech.createdAt), speech })),
      ...messages.map((message) => ({ kind: 'message' as const, id: message.id, at: Date.parse(message.createdAt), message })),
    ]
    return all.sort((a, b) => a.at - b.at)
  }, [speeches, messages])

  useEffect(() => end.current?.scrollIntoView({ behavior: 'smooth' }), [entries.length])


  useLayoutEffect(() => {
    const element = input.current
    if (!element) return
    element.style.height = 'auto'
    element.style.height = `${Math.min(element.scrollHeight, MAX_INPUT_HEIGHT)}px`
    element.style.overflowY = element.scrollHeight > MAX_INPUT_HEIGHT ? 'auto' : 'hidden'
  }, [text])

  const suggestions = useMemo<Suggestion[]>(() => {
    if (query === null) return []
    const all = [{ id: mentionAll, name: mentionAll }, ...participants.map(({ id, name }) => ({ id, name }))]
    const normalized = query.toLowerCase()
    return all.filter((option) => option.name.toLowerCase().includes(normalized)).slice(0, 6)
  }, [query, participants])

  const names = useMemo(() => new Map(participants.map((p) => [p.id, p.name])), [participants])

  const updateQuery = (value: string, caret: number) => {
    const match = /@([^@\n]{0,40})$/.exec(value.slice(0, caret))
    setQuery(match ? match[1] : null)
    setHighlight(0)
  }

  const choose = (suggestion: Suggestion) => {
    const element = input.current
    const caret = element?.selectionStart ?? text.length
    const before = text.slice(0, caret).replace(/@([^@\n]{0,40})$/, `@${suggestion.name} `)
    const next = before + text.slice(caret)
    setText(next)
    setQuery(null)
    requestAnimationFrame(() => {
      element?.focus()
      element?.setSelectionRange(before.length, before.length)
    })
  }

  const addImages = (files: Iterable<File>) => {
    const accepted: { file: File; preview: string }[] = []
    for (const file of files) {
      if (!IMAGE_TYPES.includes(file.type)) {
        setError(t.imageTypes)
        continue
      }
      if (file.size > MAX_IMAGE_BYTES) {
        setError(t.imageTooBig(file.name))
        continue
      }
      accepted.push({ file, preview: URL.createObjectURL(file) })
    }
    setImages((current) => {
      const next = [...current, ...accepted]
      if (next.length > MAX_IMAGES) setError(t.tooManyImages(MAX_IMAGES))
      next.slice(MAX_IMAGES).forEach((image) => URL.revokeObjectURL(image.preview))
      return next.slice(0, MAX_IMAGES)
    })
  }

  const removeImage = (index: number) => {
    setImages((current) => {
      URL.revokeObjectURL(current[index].preview)
      return current.filter((_, i) => i !== index)
    })
  }

  const onPaste = (event: ClipboardEvent<HTMLTextAreaElement>) => {
    const files = Array.from(event.clipboardData.files)
    if (files.length === 0) return
    event.preventDefault()
    addImages(files)
  }

  const onDrop = (event: DragEvent<HTMLDivElement>) => {
    event.preventDefault()
    setDragging(false)
    addImages(Array.from(event.dataTransfer.files))
  }

  const mentionsIn = (value: string) => {
    if (value.includes(`@${mentionAll}`)) return participants.map((p) => p.id)
    return participants.filter((p) => value.includes(`@${p.name}`)).map((p) => p.id)
  }

  const hasContent = text.trim() !== '' || images.length > 0
  const mentioned = mentionsIn(text).length > 0
  const canSend = hasContent && mentioned && !sending

  const send = async () => {
    const value = text.trim()
    if (!canSend) return
    setSending(true)
    setError(null)
    try {
      await onSend(value, mentionsIn(value), images.map((image) => image.file))
      images.forEach((image) => URL.revokeObjectURL(image.preview))
      setImages([])
      setText('')
    } catch {
      setError(t.sendFailed)
    } finally {
      setSending(false)
    }
  }

  const onKeyDown = (event: KeyboardEvent<HTMLTextAreaElement>) => {
    if (suggestions.length > 0) {
      if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
        event.preventDefault()
        const step = event.key === 'ArrowDown' ? 1 : -1
        setHighlight((current) => (current + step + suggestions.length) % suggestions.length)
        return
      }
      if (event.key === 'Enter' || event.key === 'Tab') {
        event.preventDefault()
        choose(suggestions[highlight])
        return
      }
      if (event.key === 'Escape') {
        setQuery(null)
        return
      }
    }
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault()
      send()
    }
  }

  return (
    <div className="chat">
      <div className="panel-body chat-log">
        {entries.length === 0 && <p className="empty">{t.chatEmpty}</p>}
        {entries.map((entry) =>
          entry.kind === 'speech' ? (
            <div key={entry.id} className="message">
              <Head name={entry.speech.name} at={entry.speech.createdAt} color={avatarColor(entry.speech.participantId)} />
              <p>{entry.speech.text}</p>
            </div>
          ) : (
            <div key={entry.id} className="message mine">
              <Head name={entry.message.authorName} at={entry.message.createdAt} />
              <p>
                <Highlighted text={entry.message.text} names={[mentionAll, ...entry.message.mentions.map((id) => names.get(id) ?? id)]} />
              </p>
              {entry.message.attachments && entry.message.attachments.length > 0 && (
                <div className="attachments">
                  {entry.message.attachments.map((attachment) => (
                    <a key={attachment.url} href={attachment.url} target="_blank" rel="noreferrer">
                      <img src={attachment.url} alt={t.sentImage} />
                    </a>
                  ))}
                </div>
              )}
              <Delivery message={entry.message} names={names} />
            </div>
          ),
        )}
        <div ref={end} />
      </div>

      <div
        className={`composer ${dragging ? 'dragging' : ''}`}
        onDragOver={(event) => {
          event.preventDefault()
          setDragging(true)
        }}
        onDragLeave={() => setDragging(false)}
        onDrop={onDrop}
      >
        {suggestions.length > 0 && (
          <ul className="mentions" role="listbox">
            {suggestions.map((suggestion, index) => (
              <li
                key={suggestion.id}
                role="option"
                aria-selected={index === highlight}
                className={index === highlight ? 'selected' : ''}
                onMouseDown={(event) => {
                  event.preventDefault()
                  choose(suggestion)
                }}
              >
                <span className="mini-avatar" style={{ background: suggestion.id === mentionAll ? '#5f6368' : avatarColor(suggestion.id) }}>
                  {suggestion.id === mentionAll ? '@' : initials(suggestion.name)}
                </span>
                {suggestion.id === mentionAll ? t.allAgents : suggestion.name}
              </li>
            ))}
          </ul>
        )}
        {error && <small className="composer-error">{error}</small>}
        {hasContent && !mentioned && <small className="composer-hint">{t.mentionToSend}</small>}
        {images.length > 0 && (
          <div className="pending-images">
            {images.map((image, index) => (
              <div key={image.preview} className="pending-image">
                <img src={image.preview} alt={image.file.name} />
                <button onClick={() => removeImage(index)} aria-label={t.removeImage(image.file.name)}>
                  <X size={14} />
                </button>
              </div>
            ))}
          </div>
        )}
        <div className="composer-row">
          <input
            ref={picker}
            type="file"
            accept={IMAGE_TYPES.join(',')}
            multiple
            hidden
            onChange={(event) => {
              addImages(Array.from(event.target.files ?? []))
              event.target.value = ''
            }}
          />
          <button className="icon attach" onClick={() => picker.current?.click()} aria-label={t.attachImage} title={t.attachImage}>
            <Paperclip size={18} />
          </button>
          <textarea
            ref={input}
            rows={1}
            value={text}
            placeholder={t.messagePlaceholder}
            onChange={(event) => {
              setText(event.target.value)
              updateQuery(event.target.value, event.target.selectionStart)
            }}
            onClick={(event) => updateQuery(text, event.currentTarget.selectionStart)}
            onKeyDown={onKeyDown}
            onPaste={onPaste}
          />
          <button className="icon send" onClick={send} disabled={!canSend} aria-label={t.send}>
            <SendHorizontal size={20} />
          </button>
        </div>
      </div>
    </div>
  )
}

function Head({ name, at, color }: { name: string; at: string; color?: string }) {
  return (
    <div className="message-head">
      <strong style={color ? { color } : undefined}>{name}</strong>
      <small>{formatTime(new Date(at))}</small>
    </div>
  )
}

function Highlighted({ text, names }: { text: string; names: string[] }) {
  const tokens = names.filter(Boolean).map((name) => `@${name}`)
  if (tokens.length === 0) return <>{text}</>
  const pattern = new RegExp(`(${tokens.map(escapeRegExp).join('|')})`, 'g')
  return (
    <>
      {text.split(pattern).map((part, index) => (
        <Fragment key={index}>{tokens.includes(part) ? <span className="mention">{part}</span> : part}</Fragment>
      ))}
    </>
  )
}

function Delivery({ message, names }: { message: ChatMessage; names: Map<string, string> }) {
  if (message.mentions.length === 0) return <small className="delivery">{t.nobodyMentioned}</small>
  const failed = message.failed ?? {}
  const delivered = message.mentions.filter((id) => !failed[id])
  return (
    <>
      {delivered.length > 0 && (
        <small className="delivery">{t.deliveredTo(delivered.map((id) => names.get(id) ?? id).join(', '))}</small>
      )}
      {Object.entries(failed).map(([id, reason]) => (
        <small key={id} className="delivery failed" title={reason}>
          {t.notDeliveredTo(names.get(id) ?? id)}
          {reason.includes('agent_blocked') ? t.waitingApproval : ''}
        </small>
      ))}
    </>
  )
}

function escapeRegExp(value: string) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

// Speech arrives in sentence chunks; the chat shows each turn as one message.
function joinTurns(speeches: Speech[]): Speech[] {
  const turns = new Map<string, Speech>()
  for (const speech of speeches) {
    const turn = turns.get(speech.turnId)
    turns.set(speech.turnId, turn ? { ...turn, text: `${turn.text} ${speech.text}` } : speech)
  }
  return [...turns.values()]
}
