import { useCallback, useEffect, useRef, useState } from 'react'

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

const ERRORS: Record<string, string> = {
  'not-allowed': 'O navegador bloqueou o microfone.',
  'service-not-allowed': 'O reconhecimento de voz não está disponível neste navegador.',
  'audio-capture': 'Nenhum microfone encontrado.',
  network: 'Sem conexão com o serviço de reconhecimento de voz.',
}

// Chrome's Web Speech API. Recognition stops on its own after silence, so it
// is restarted while the user keeps the microphone on.
export function useSpeechRecognition(onFinal: (text: string) => void, lang = 'pt-BR') {
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
      setError(ERRORS[event.error] ?? `Erro no reconhecimento de voz: ${event.error}`)
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
