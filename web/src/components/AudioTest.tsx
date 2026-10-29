import { Loader2, Volume2 } from 'lucide-react'
import { useEffect, useRef, useState } from 'react'

type State = { kind: 'idle' } | { kind: 'loading' } | { kind: 'playing' } | { kind: 'error'; message: string }

export function AudioTest({ variant = 'dark' }: { variant?: 'dark' | 'light' }) {
  const [state, setState] = useState<State>({ kind: 'idle' })
  const audio = useRef<HTMLAudioElement | null>(null)

  useEffect(() => () => audio.current?.pause(), [])

  const test = async () => {
    audio.current?.pause()
    setState({ kind: 'loading' })
    try {
      const response = await fetch('/api/audio-test')
      if (response.status === 204) {
        setState({ kind: 'error', message: 'A voz está desativada (TTS=none). Só legendas.' })
        return
      }
      if (!response.ok) {
        setState({ kind: 'error', message: 'O serviço de voz não respondeu. Se acabou de subir, aguarde alguns segundos.' })
        return
      }
      const url = URL.createObjectURL(await response.blob())
      const element = new Audio(url)
      audio.current = element
      element.onended = () => {
        URL.revokeObjectURL(url)
        setState({ kind: 'idle' })
      }
      await element.play()
      setState({ kind: 'playing' })
    } catch {
      setState({ kind: 'error', message: 'Não consegui tocar o áudio neste navegador.' })
    }
  }

  return (
    <div className={`audio-test ${variant}`}>
      <button onClick={test} disabled={state.kind === 'loading'}>
        {state.kind === 'loading' ? <Loader2 size={18} className="spin" /> : <Volume2 size={18} />}
        {state.kind === 'playing' ? 'Tocando…' : 'Testar áudio'}
      </button>
      {state.kind === 'error' && <small>{state.message}</small>}
    </div>
  )
}
