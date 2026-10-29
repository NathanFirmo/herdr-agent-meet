export type AgentStatus = 'idle' | 'working' | 'blocked' | 'done' | 'unknown'

export interface Activity {
  kind: 'tool' | 'thinking'
  label?: string
  since: string
}

export interface PromptOption {
  number: number
  label: string
  description?: string
  freeText?: boolean
}

export interface Prompt {
  question: string
  context?: string
  options: PromptOption[]
}

export interface Participant {
  id: string
  name: string
  kind: string
  status: AgentStatus
  voice: string
  activity?: Activity
  originalName?: string
  prompt?: Prompt
  dir?: string
}

export interface Speech {
  id: string
  turnId: string
  final: boolean
  participantId: string
  name: string
  text: string
  audioUrl?: string
  createdAt: string
}

export interface ChatMessage {
  id: string
  authorId: string
  authorName: string
  text: string
  mentions: string[]
  attachments?: { url: string }[]
  failed?: Record<string, string>
  createdAt: string
}

export type MeetEvent =
  | { type: 'snapshot'; participants?: Participant[]; speeches?: Speech[]; messages?: ChatMessage[] }
  | { type: 'participants'; participants?: Participant[] }
  | { type: 'speech'; speech: Speech }
  | { type: 'message'; message: ChatMessage }
