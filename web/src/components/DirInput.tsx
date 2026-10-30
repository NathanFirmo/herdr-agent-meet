import { useEffect, useRef, useState, type KeyboardEvent } from 'react'
import { t } from '../i18n'

const DEBOUNCE_MS = 120

interface DirInputProps {
  value: string
  onChange: (value: string) => void
  known: string[]
}

// Completes directories like a shell: Tab fills the highlighted one and keeps
// going into it, Enter picks it, arrows move. An empty field suggests the
// folders the agents already work in.
export function DirInput({ value, onChange, known }: DirInputProps) {
  const [suggestions, setSuggestions] = useState<string[]>([])
  const [open, setOpen] = useState(false)
  const [highlight, setHighlight] = useState(0)
  const request = useRef(0)

  useEffect(() => {
    if (!open) return
    const id = ++request.current
    const timer = setTimeout(async () => {
      const fromServer = await fetch(`/api/dirs?path=${encodeURIComponent(value)}`)
        .then((response) => (response.ok ? (response.json() as Promise<string[]>) : []))
        .catch(() => [] as string[])
      if (id !== request.current) return
      const knownFirst = value.trim() === '' ? known.map((dir) => (dir.endsWith('/') ? dir : `${dir}/`)) : []
      setSuggestions([...new Set([...knownFirst, ...fromServer])])
      setHighlight(0)
    }, DEBOUNCE_MS)
    return () => clearTimeout(timer)
  }, [value, open, known])

  const visible = open ? suggestions.filter((suggestion) => suggestion !== value) : []

  const choose = (suggestion: string, keepGoing: boolean) => {
    onChange(keepGoing ? suggestion : suggestion.replace(/\/$/, ''))
    setOpen(keepGoing)
  }

  const onKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (visible.length === 0) return
    switch (event.key) {
      case 'ArrowDown':
      case 'ArrowUp': {
        event.preventDefault()
        const step = event.key === 'ArrowDown' ? 1 : -1
        setHighlight((current) => (current + step + visible.length) % visible.length)
        break
      }
      case 'Tab':
        event.preventDefault()
        choose(visible[highlight], true)
        break
      case 'Enter':
        event.preventDefault()
        choose(visible[highlight], false)
        break
      case 'Escape':
        event.preventDefault()
        event.stopPropagation()
        setOpen(false)
        break
    }
  }

  return (
    <div className="dir-input">
      <input
        className="mono"
        value={value}
        spellCheck={false}
        autoComplete="off"
        placeholder={t.folderPlaceholder}
        onChange={(event) => {
          onChange(event.target.value)
          setOpen(true)
        }}
        onFocus={() => setOpen(true)}
        onBlur={() => setOpen(false)}
        onKeyDown={onKeyDown}
      />
      {visible.length > 0 && (
        <ul className="dir-suggestions" role="listbox">
          {visible.map((suggestion, index) => (
            <li
              key={suggestion}
              role="option"
              aria-selected={index === highlight}
              className={index === highlight ? 'selected' : ''}
              onMouseDown={(event) => {
                event.preventDefault()
                choose(suggestion, true)
              }}
            >
              {suggestion}
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
