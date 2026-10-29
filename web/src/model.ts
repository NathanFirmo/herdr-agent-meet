import type { Participant } from './types'

// "claude-opus-5-5" reads as "Opus 5.5"; ids that do not follow that pattern
// are shown as they are.
export function modelName(model: string) {
  const match = /^claude-([a-z]+)-(\d+)(?:-(\d{1,2}))?(?:-\d{8})?$/.exec(model)
  if (!match) return model
  const family = match[1].charAt(0).toUpperCase() + match[1].slice(1)
  return match[3] ? `${family} ${match[2]}.${match[3]}` : `${family} ${match[2]}`
}

export function agentLabel(participant: Participant) {
  return participant.model ? `${participant.kind} · ${modelName(participant.model)}` : participant.kind
}
