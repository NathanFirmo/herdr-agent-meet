import { useCallback, useEffect, useRef, useState } from 'react'
import type { ChatMessage, MeetEvent, Participant, Speech } from './types'

const MIN_CAPTION_MS = 2500
const MS_PER_CHAR = 60
// Upper bound for one speech, so a clip that never loads or ends cannot stall the queue.
const MAX_SPEECH_MS = 60_000
// Speeches older than this when their turn comes are skipped, so the call stays live; they remain in the chat.
const MAX_SPEECH_LAG_MS = 45_000
// How long to wait for the next chunk of a turn before letting someone else talk.
const MAX_CHUNK_WAIT_MS = 4_000
const GAP_BETWEEN_TURNS_MS = 400
const GAP_BETWEEN_CHUNKS_MS = 80

export interface Toast {
  id: string
  text: string
}

// While held, queued speeches wait, e.g. so dictation does not transcribe the
// agents' voices coming out of the speakers.
export function useMeeting(joined: boolean, held: boolean) {
  const [participants, setParticipants] = useState<Participant[]>([])
  const [speeches, setSpeeches] = useState<Speech[]>([])
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [connected, setConnected] = useState(false)
  const [current, setCurrent] = useState<Speech | null>(null)
  const [muted, setMuted] = useState(false)
  const [toasts, setToasts] = useState<Toast[]>([])
  const [pending, setPending] = useState(0)

  const queue = useRef<Speech[]>([])
  const playing = useRef(false)
  const audio = useRef<HTMLAudioElement | null>(null)
  const joinedRef = useRef(joined)
  const mutedRef = useRef(muted)
  const heldRef = useRef(held)
  const statuses = useRef(new Map<string, string>())
  const openTurn = useRef<string | null>(null)
  const chunkWait = useRef<ReturnType<typeof setTimeout> | undefined>(undefined)
  const clips = useRef(new Map<string, Promise<Blob>>())

  const prefetch = useCallback((speech: Speech) => {
    if (!speech.audioUrl || clips.current.has(speech.id)) return
    const clip = fetch(speech.audioUrl).then((response) =>
      response.ok ? response.blob() : Promise.reject(new Error(response.statusText)),
    )
    clip.catch(() => undefined)
    clips.current.set(speech.id, clip)
  }, [])

  // Chunks of the turn being spoken go first, so agents do not interleave
  // mid-sentence. Returns undefined while that turn's next chunk is pending.
  const pickNext = useCallback(() => {
    const now = Date.now()
    const fresh = queue.current.filter((speech) => now - Date.parse(speech.createdAt) <= MAX_SPEECH_LAG_MS)
    queue.current
      .filter((speech) => !fresh.includes(speech))
      .forEach((speech) => clips.current.delete(speech.id))
    queue.current = fresh

    const index = openTurn.current ? fresh.findIndex((speech) => speech.turnId === openTurn.current) : 0
    if (index < 0) {
      setPending(queue.current.length)
      return undefined
    }
    const speech = queue.current.splice(index, 1)[0]
    setPending(queue.current.length)
    return speech
  }, [])

  const pushToast = useCallback((text: string) => {
    const id = `${Date.now()}-${Math.random()}`
    setToasts((all) => [...all, { id, text }])
    setTimeout(() => setToasts((all) => all.filter((toast) => toast.id !== id)), 6000)
  }, [])

  const playNext = useCallback(function next() {
    if (heldRef.current) {
      setPending(queue.current.length)
      playing.current = false
      setCurrent(null)
      return
    }
    const speech = pickNext()
    if (!speech) {
      playing.current = false
      if (openTurn.current) {
        chunkWait.current ??= setTimeout(() => {
          chunkWait.current = undefined
          openTurn.current = null
          if (!playing.current) next()
        }, MAX_CHUNK_WAIT_MS)
        return
      }
      setCurrent(null)
      return
    }
    clearTimeout(chunkWait.current)
    chunkWait.current = undefined
    playing.current = true
    openTurn.current = speech.final ? null : speech.turnId
    setCurrent(speech)
    const clip = clips.current.get(speech.id)
    clips.current.delete(speech.id)

    let finished = false
    let objectUrl: string | null = null
    const watchdog = setTimeout(() => {
      audio.current?.pause()
      finish()
    }, MAX_SPEECH_MS)
    const finish = () => {
      if (finished) return
      finished = true
      clearTimeout(watchdog)
      if (objectUrl) URL.revokeObjectURL(objectUrl)
      setTimeout(next, speech.final ? GAP_BETWEEN_TURNS_MS : GAP_BETWEEN_CHUNKS_MS)
    }
    const showCaptionOnly = () =>
      setTimeout(finish, Math.max(MIN_CAPTION_MS, speech.text.length * MS_PER_CHAR))

    if (!clip || mutedRef.current) {
      showCaptionOnly()
      return
    }
    clip
      .then((blob) => {
        if (finished || mutedRef.current) {
          showCaptionOnly()
          return
        }
        objectUrl = URL.createObjectURL(blob)
        const element = new Audio(objectUrl)
        audio.current = element
        element.onpause = finish
        element.onerror = showCaptionOnly
        return element.play()
      })
      .catch((error) => {
        console.warn('agent speech playback failed', error)
        showCaptionOnly()
      })
  }, [pickNext])

  useEffect(() => {
    joinedRef.current = joined
    if (!joined) {
      queue.current = []
      clips.current.clear()
      openTurn.current = null
      audio.current?.pause()
    }
  }, [joined])

  useEffect(() => {
    heldRef.current = held
    if (!held && !playing.current && queue.current.length > 0) playNext()
  }, [held, playNext])

  useEffect(() => {
    mutedRef.current = muted
    if (muted) audio.current?.pause()
  }, [muted])

  const trackStatuses = useCallback(
    (list: Participant[], announce: boolean) => {
      for (const participant of list) {
        const previous = statuses.current.get(participant.id)
        if (announce && previous && previous !== 'blocked' && participant.status === 'blocked') {
          pushToast(`${participant.name} levantou a mão`)
        }
        statuses.current.set(participant.id, participant.status)
      }
    },
    [pushToast],
  )

  useEffect(() => {
    const source = new EventSource('/api/events')
    source.onopen = () => setConnected(true)
    source.onerror = () => setConnected(false)
    source.onmessage = (message) => {
      const event = JSON.parse(message.data) as MeetEvent
      switch (event.type) {
        case 'snapshot':
          trackStatuses(event.participants ?? [], false)
          setParticipants(event.participants ?? [])
          setSpeeches(event.speeches ?? [])
          setMessages(event.messages ?? [])
          break
        case 'message':
          setMessages((all) => [...all.filter((message) => message.id !== event.message.id).slice(-99), event.message])
          break
        case 'participants':
          trackStatuses(event.participants ?? [], true)
          setParticipants(event.participants ?? [])
          break
        case 'speech':
          setSpeeches((all) => [...all.slice(-99), event.speech])
          if (joinedRef.current) {
            queue.current.push(event.speech)
            setPending(queue.current.length)
            prefetch(event.speech)
            if (!playing.current) playNext()
          }
          break
      }
    }
    return () => source.close()
  }, [playNext, prefetch, trackStatuses])

  const sendMessage = useCallback(async (text: string, mentions: string[], images: File[]) => {
    const form = new FormData()
    form.append('text', text)
    mentions.forEach((id) => form.append('mentions', id))
    images.forEach((image) => form.append('images', image))
    const response = await fetch('/api/chat', { method: 'POST', body: form })
    if (!response.ok) throw new Error(await response.text())
    const message = (await response.json()) as ChatMessage
    setMessages((all) => (all.some((existing) => existing.id === message.id) ? all : [...all, message]))
  }, [])

  const renameParticipant = useCallback(async (id: string, name: string) => {
    const response = await fetch(`/api/participants/${encodeURIComponent(id)}/name`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name }),
    })
    if (!response.ok) throw new Error(await response.text())
  }, [])

  const answerPrompt = useCallback(async (id: string, option: number, text = '') => {
    const response = await fetch(`/api/participants/${encodeURIComponent(id)}/answer`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ option, text }),
    })
    if (!response.ok) throw new Error((await response.text()).trim())
  }, [])

  const removeParticipant = useCallback(async (id: string) => {
    const response = await fetch(`/api/participants/${encodeURIComponent(id)}`, { method: 'DELETE' })
    if (!response.ok) throw new Error(await response.text())
  }, [])

  return { pending, participants, renameParticipant, removeParticipant, answerPrompt, speeches, messages, sendMessage, connected, current, muted, setMuted, toasts }
}
