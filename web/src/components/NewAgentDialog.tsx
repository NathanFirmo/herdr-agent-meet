import { Loader2, X } from 'lucide-react'
import { useState, type FormEvent } from 'react'

const DEFAULT_COMMAND = 'claude'
const LAST_COMMAND_KEY = 'agent-meet:last-command'

function lastCommand() {
  try {
    return localStorage.getItem(LAST_COMMAND_KEY) || DEFAULT_COMMAND
  } catch {
    return DEFAULT_COMMAND
  }
}

function rememberCommand(command: string) {
  try {
    localStorage.setItem(LAST_COMMAND_KEY, command)
  } catch {
    // Remembering the command is a convenience; private windows may block storage.
  }
}

interface NewAgentDialogProps {
  dirs: string[]
  onLaunch: (agent: { name: string; command: string; dir: string; prompt: string }) => Promise<void>
  onClose: () => void
}

export function NewAgentDialog({ dirs, onLaunch, onClose }: NewAgentDialogProps) {
  const [name, setName] = useState('')
  const [command, setCommand] = useState(lastCommand)
  const [dir, setDir] = useState(dirs[0] ?? '')
  const [prompt, setPrompt] = useState('')
  const [launching, setLaunching] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const submit = async (event: FormEvent) => {
    event.preventDefault()
    if (!name.trim() || !command.trim() || launching) return
    setLaunching(true)
    setError(null)
    try {
      await onLaunch({ name: name.trim(), command: command.trim(), dir: dir.trim(), prompt: prompt.trim() })
      rememberCommand(command.trim())
      onClose()
    } catch (failure) {
      setError(failure instanceof Error && failure.message ? failure.message : 'Não foi possível iniciar o agente.')
      setLaunching(false)
    }
  }

  return (
    <div className="dialog-backdrop" onMouseDown={(event) => event.target === event.currentTarget && !launching && onClose()}>
      <form className="dialog" onSubmit={submit}>
        <header>
          <h2>Novo agente</h2>
          <button type="button" className="icon" onClick={onClose} disabled={launching} aria-label="Fechar">
            <X size={20} />
          </button>
        </header>
        <p className="dialog-hint">Abre um workspace novo no Herdr e o agente entra na chamada.</p>

        <label>
          Nome
          <input autoFocus value={name} maxLength={40} onChange={(event) => setName(event.target.value)} placeholder="Ex.: Revisor do checkout" />
        </label>

        <label>
          Comando
          <input
            className="mono"
            value={command}
            onChange={(event) => setCommand(event.target.value)}
            placeholder="claude"
            spellCheck={false}
          />
          <small className="field-hint">Roda no shell do workspace novo: aliases e flags funcionam.</small>
        </label>

        <label>
          Pasta
          <input value={dir} onChange={(event) => setDir(event.target.value)} list="agent-dirs" placeholder="/caminho/do/projeto" />
          <datalist id="agent-dirs">
            {dirs.map((option) => (
              <option key={option} value={option} />
            ))}
          </datalist>
        </label>

        <label>
          Primeira mensagem <small>(opcional)</small>
          <textarea rows={3} value={prompt} onChange={(event) => setPrompt(event.target.value)} placeholder="O que ele deve fazer ao entrar" />
        </label>

        {error && <p className="dialog-error">{error}</p>}

        <footer>
          <button type="button" className="text-button" onClick={onClose} disabled={launching}>
            Cancelar
          </button>
          <button type="submit" className="primary-button" disabled={!name.trim() || !command.trim() || launching}>
            {launching ? (
              <>
                <Loader2 size={16} className="spin" /> Iniciando…
              </>
            ) : (
              'Iniciar agente'
            )}
          </button>
        </footer>
      </form>
    </div>
  )
}
