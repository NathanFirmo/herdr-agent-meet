import { useCallback, useEffect, useRef, useState } from 'react'
import { language, t } from './i18n'

interface RecognitionResult {
  isFinal: boolean
  0: { transcript: string }
}

interface RecognitionEvent {
  resultIndex: number
  results: ArrayLike<RecognitionResult>
}

interface Recognition {
  lang: string
  continuous: boolean
  interimResults: boolean
  onresult: ((event: RecognitionEvent) => void) | null
  onerror: ((event: { error: string }) => void) | null
  onend: (() => void) | null
  start: () => void
  stop: () => void
}

type RecognitionConstructor = new () => Recognition

const Recognition: RecognitionConstructor | undefined =
  (window as unknown as { SpeechRecognition?: RecognitionConstructor }).SpeechRecognition ??
  (window as unknown as { webkitSpeechRecognition?: RecognitionConstructor }).webkitSpeechRecognition

const errorMessage = (error: string) =>
  ({
    'not-allowed': t.micBlocked,
    'service-not-allowed': t.recognitionUnavailable,
    'audio-capture': t.noMicrophone,
    network: t.recognitionOffline,
  })[error] ?? t.recognitionError(error)

// Chrome's Web Speech API. Recognition stops on its own after silence, so it
// is restarted while the user keeps the microphone on.
export function useSpeechRecognition(onFinal: (text: string) => void, lang = language) {
  const [listening, setListening] = useState(false)
  const [interim, setInterim] = useState('')
  const [error, setError] = useState<string | null>(null)
  const recognition = useRef<Recognition | null>(null)
  const wanted = useRef(false)
  const onFinalRef = useRef(onFinal)

  useEffect(() => {
    onFinalRef.current = onFinal
  }, [onFinal])

  const stop = useCallback(() => {
    wanted.current = false
    recognition.current?.stop()
    setListening(false)
    setInterim('')
  }, [])

  const start = useCallback(() => {
    if (!Recognition) return
    const instance = new Recognition()
    instance.lang = lang
    instance.continuous = true
    instance.interimResults = true
    instance.onresult = (event) => {
      let pending = ''
      for (let i = event.resultIndex; i < event.results.length; i++) {
        const result = event.results[i]
        const transcript = result[0].transcript.trim()
        if (result.isFinal) {
          if (transcript) onFinalRef.current(transcript)
        } else {
          pending += ` ${transcript}`
        }
      }
      setInterim(pending.trim())
    }
    instance.onerror = (event) => {
      if (event.error === 'no-speech' || event.error === 'aborted') return
      setError(errorMessage(event.error))
      wanted.current = false
    }
    instance.onend = () => {
      setInterim('')
      if (wanted.current) {
        instance.start()
      } else {
        setListening(false)
      }
    }

    recognition.current?.stop()
    recognition.current = instance
    wanted.current = true
    setError(null)
    setListening(true)
    instance.start()
  }, [lang])

  useEffect(() => () => {
    wanted.current = false
    recognition.current?.stop()
  }, [])

  return { supported: Boolean(Recognition), listening, interim, error, start, stop }
}
