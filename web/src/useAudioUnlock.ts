import { useEffect, useState } from 'react'

const SILENCE = 'data:audio/wav;base64,UklGRiQAAABXQVZFZm10IBAAAAABAAEARKwAAIhYAQACABAAZGF0YQAAAAA='

// Browsers only allow audio after the user interacts with the page, unless
// they trust the site. Tries right away, then on the first click or key.
export function useAudioUnlock() {
  const [unlocked, setUnlocked] = useState(false)

  useEffect(() => {
    let done = false
    const stopListening = () => {
      window.removeEventListener('pointerdown', tryUnlock)
      window.removeEventListener('keydown', tryUnlock)
    }
    function tryUnlock() {
      new Audio(SILENCE)
        .play()
        .then(() => {
          if (done) return
          done = true
          stopListening()
          setUnlocked(true)
        })
        .catch(() => undefined)
    }

    window.addEventListener('pointerdown', tryUnlock)
    window.addEventListener('keydown', tryUnlock)
    tryUnlock()
    return () => {
      done = true
      stopListening()
    }
  }, [])

  return unlocked
}
