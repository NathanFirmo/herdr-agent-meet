import { useCallback, useEffect, useRef, useState } from 'react'
import type { MeetEvent, Participant, Speech } from './types'

const MIN_CAPTION_MS = 2500
const MS_PER_CHAR = 60

export interface Toast {
  id: string
  text: string
}

export function useMeeting(joined: boolean) {
  const [participants, setParticipants] = useState<Participant[]>([])
  const [speeches, setSpeeches] = useState<Speech[]>([])
  const [connected, setConnected] = useState(false)
  const [current, setCurrent] = useState<Speech | null>(null)
  const [muted, setMuted] = useState(false)
  const [toasts, setToasts] = useState<Toast[]>([])

  const queue = useRef<Speech[]>([])
  const playing = useRef(false)
  const audio = useRef<HTMLAudioElement | null>(null)
  const joinedRef = useRef(joined)
  const mutedRef = useRef(muted)
  const statuses = useRef(new Map<string, string>())

  const pushToast = useCallback((text: string) => {
    const id = `${Date.now()}-${Math.random()}`
    setToasts((all) => [...all, { id, text }])
    setTimeout(() => setToasts((all) => all.filter((toast) => toast.id !== id)), 6000)
  }, [])

  const playNext = useCallback(function next() {
    const speech = queue.current.shift()
    if (!speech) {
      playing.current = false
      setCurrent(null)
      return
    }
    playing.current = true
    setCurrent(speech)

    let finished = false
    const finish = () => {
      if (finished) return
      finished = true
      setTimeout(next, 400)
    }
    const showCaptionOnly = () =>
      setTimeout(finish, Math.max(MIN_CAPTION_MS, speech.text.length * MS_PER_CHAR))

    if (!speech.audioUrl || mutedRef.current) {
      showCaptionOnly()
      return
    }
    const element = new Audio(speech.audioUrl)
    audio.current = element
    element.onended = finish
    element.onerror = showCaptionOnly
    element.play().catch(showCaptionOnly)
  }, [])

  useEffect(() => {
    joinedRef.current = joined
    if (!joined) {
      queue.current = []
      audio.current?.pause()
    }
  }, [joined])

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
          break
        case 'participants':
          trackStatuses(event.participants ?? [], true)
          setParticipants(event.participants ?? [])
          break
        case 'speech':
          setSpeeches((all) => [...all.slice(-99), event.speech])
          if (joinedRef.current) {
            queue.current.push(event.speech)
            if (!playing.current) playNext()
          }
          break
      }
    }
    return () => source.close()
  }, [playNext, trackStatuses])

  return { participants, speeches, connected, current, muted, setMuted, toasts }
}
