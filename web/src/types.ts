export type AgentStatus = 'idle' | 'working' | 'blocked' | 'done' | 'unknown'

export interface Participant {
  id: string
  name: string
  kind: string
  status: AgentStatus
  voice: string
}

export interface Speech {
  id: string
  participantId: string
  name: string
  text: string
  audioUrl?: string
  createdAt: string
}

export type MeetEvent =
  | { type: 'snapshot'; participants?: Participant[]; speeches?: Speech[] }
  | { type: 'participants'; participants?: Participant[] }
  | { type: 'speech'; speech: Speech }
