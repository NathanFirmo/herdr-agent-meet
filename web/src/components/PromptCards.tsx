import { ChevronDown, ChevronUp, Hand } from 'lucide-react'
import { useState } from 'react'
import { avatarColor, initials } from '../avatar'
import type { Participant, Prompt, PromptOption } from '../types'

type Answer = (id: string, option: number, text?: string) => Promise<void>

export function PromptCards({ participants, onAnswer }: { participants: Participant[]; onAnswer: Answer }) {
  const asking = participants.filter((participant) => participant.prompt)
  if (asking.length === 0) return null
  return (
    <div className="prompt-cards">
      {asking.map((participant) => (
        <PromptCard
          key={`${participant.id}-${participant.prompt!.question}`}
          participant={participant}
          prompt={participant.prompt!}
          onAnswer={onAnswer}
        />
      ))}
    </div>
  )
}

function PromptCard({ participant, prompt, onAnswer }: { participant: Participant; prompt: Prompt; onAnswer: Answer }) {
  const [collapsed, setCollapsed] = useState(false)
  const [typing, setTyping] = useState<PromptOption | null>(null)
  const [text, setText] = useState('')
  const [sending, setSending] = useState(false)
  const [error, setError] = useState<string | null>(null)
  // Question dialogs show their tab as "☐ Topic"; approval dialogs show the command to run.
  const tab = prompt.context?.match(/^[☐☒✔✓]\s*(.+)$/)
  const topic = tab?.[1]
  const command = tab ? undefined : prompt.context

  const answer = async (option: number, typed = '') => {
    setSending(true)
    setError(null)
    try {
      await onAnswer(participant.id, option, typed)
    } catch (failure) {
      setError(failure instanceof Error && failure.message ? failure.message : 'Não foi possível responder.')
      setSending(false)
    }
  }

  return (
    <section className="prompt-card" aria-label={`${participant.name} precisa de você`}>
      <header>
        <span className="mini-avatar" style={{ background: avatarColor(participant.id) }}>
          {initials(participant.name)}
        </span>
        <div>
          <strong>{participant.name}</strong>
          <small>
            <Hand size={12} /> precisa de você
          </small>
        </div>
        <button className="icon" onClick={() => setCollapsed(!collapsed)} aria-label={collapsed ? 'Expandir' : 'Recolher'}>
          {collapsed ? <ChevronUp size={18} /> : <ChevronDown size={18} />}
        </button>
      </header>

      {!collapsed && (
        <>
          {topic && <span className="prompt-topic">{topic}</span>}
          <p className="prompt-question">{prompt.question}</p>
          {command && <p className="prompt-context">{command}</p>}

          {typing ? (
            <form
              className="prompt-text"
              onSubmit={(event) => {
                event.preventDefault()
                if (text.trim()) answer(typing.number, text.trim())
              }}
            >
              <textarea autoFocus rows={3} value={text} onChange={(event) => setText(event.target.value)} placeholder="Sua resposta" />
              <div className="prompt-actions">
                <button type="button" className="text-button" onClick={() => setTyping(null)} disabled={sending}>
                  Voltar
                </button>
                <button type="submit" className="primary-button" disabled={sending || !text.trim()}>
                  Enviar
                </button>
              </div>
            </form>
          ) : (
            <ul className="prompt-options">
              {prompt.options.map((option) => (
                <li key={option.number}>
                  <button disabled={sending} onClick={() => (option.freeText ? setTyping(option) : answer(option.number))}>
                    <strong>
                      {option.number}. {option.label}
                    </strong>
                    {option.description && <small>{option.description}</small>}
                  </button>
                </li>
              ))}
            </ul>
          )}

          {error && <p className="prompt-error">{error}</p>}
          {!typing && (
            <div className="prompt-actions">
              <button className="text-button" onClick={() => answer(0)} disabled={sending}>
                Cancelar (Esc)
              </button>
            </div>
          )}
        </>
      )}
    </section>
  )
}
