const pt = {
  you: 'Você',
  unlockAudio: 'Clique em qualquer lugar para ouvir os agentes',
  heldSpeeches: (count: number) =>
    `Microfone ligado: ${count} ${count === 1 ? 'fala em espera' : 'falas em espera'}. Desligue para ouvir os agentes.`,
  raisedHand: (name: string) => `${name} levantou a mão`,

  runningCommand: 'Rodando comando',
  thinking: 'Pensando',
  stalledTitle: 'Está nisso há bastante tempo, pode estar travado',
  stalled: 'pode estar travado',

  ttsOff: 'A voz está desativada (TTS=none). Só legendas.',
  ttsDown: 'O serviço de voz não respondeu. Se acabou de subir, aguarde alguns segundos.',
  playFailed: 'Não consegui tocar o áudio neste navegador.',
  playing: 'Tocando…',
  testAudio: 'Testar áudio',

  leftCall: 'Você saiu da chamada',
  readyToJoin: 'Pronto para participar?',
  connecting: 'Conectando ao Herdr…',
  noAgents: 'Nenhum agente na chamada ainda',
  agentsInCall: (count: number) => `${count} ${count === 1 ? 'agente está' : 'agentes estão'} na chamada`,
  rejoin: 'Participar novamente',
  joinNow: 'Participar agora',

  newAgent: 'Novo agente',
  close: 'Fechar',
  cancel: 'Cancelar',
  send: 'Enviar',
  launchFailed: 'Não foi possível iniciar o agente.',
  newAgentHint: 'Abre um workspace novo no Herdr e o agente entra na chamada.',
  name: 'Nome',
  namePlaceholder: 'Ex.: Revisor do checkout',
  command: 'Comando',
  commandHint: 'Roda no shell do workspace novo: aliases e flags funcionam.',
  folder: 'Pasta',
  folderHint: 'Tab completa a pasta, como no terminal.',
  folderPlaceholder: '~/projetos/meu-app',
  firstMessage: 'Primeira mensagem',
  optional: '(opcional)',
  firstMessagePlaceholder: 'O que ele deve fazer ao entrar',
  launching: 'Iniciando…',
  launchAgent: 'Iniciar agente',

  imageTypes: 'Só dá para anexar PNG, JPEG, GIF ou WebP.',
  imageTooBig: (name: string) => `${name} passa de 10 MB.`,
  tooManyImages: (max: number) => `No máximo ${max} imagens por mensagem.`,
  sendFailed: 'Não foi possível enviar a mensagem.',
  chatEmpty: 'Ninguém falou ainda. Marque um agente com @ para falar com ele.',
  sentImage: 'Imagem enviada',
  allAgents: 'todos os agentes',
  mentionToSend: 'Marque um agente com @ para enviar',
  removeImage: (name: string) => `Remover ${name}`,
  attachImage: 'Anexar imagem',
  messagePlaceholder: 'Mensagem (@ para marcar)',
  nobodyMentioned: 'Ninguém marcado, ficou só no chat',
  deliveredTo: (names: string) => `Entregue para ${names}`,
  notDeliveredTo: (name: string) => `Não entregue para ${name}`,
  waitingApproval: ': está esperando aprovação',

  micUnsupported: 'Ditado por voz só funciona no Chrome',
  micOff: 'Desativar microfone',
  dictate: 'Ditar mensagem por voz',
  listenAgents: 'Ouvir agentes',
  muteAgents: 'Silenciar agentes',
  captions: 'Legendas',
  leave: 'Sair da chamada',
  meetingDetails: 'Detalhes da reunião',
  participants: 'Participantes',
  chat: 'Chat',
  chatTitle: 'Mensagens na chamada',

  answerFailed: 'Não foi possível responder.',
  needsYouLabel: (name: string) => `${name} precisa de você`,
  needsYou: 'precisa de você',
  expand: 'Expandir',
  collapse: 'Recolher',
  yourAnswer: 'Sua resposta',
  back: 'Voltar',
  cancelEsc: 'Cancelar (Esc)',

  infoJoin: 'Seus agentes do Herdr entram na chamada automaticamente e reportam o que estão fazendo.',
  infoChat: 'No chat, marque um agente com @ para mandar uma mensagem direto para a sessão dele.',
  serverConnection: 'Conexão com o servidor:',
  connected: 'ativa',
  reconnecting: 'reconectando…',
  organizer: 'Organizador da reunião',
  renameFrom: (original: string) => `Renomear (original: ${original})`,
  rename: 'Renomear',
  closeFailed: 'Não foi possível fechar o pane.',
  endAgentWarning: 'Encerrar agente? O pane será fechado no Herdr.',
  ending: 'Encerrando…',
  end: 'Encerrar',
  endNamed: (name: string) => `Encerrar ${name}`,
  endAgent: 'Encerrar agente',

  needsApproval: 'Precisa de aprovação',
  talking: 'Falando',
  working: 'Trabalhando',
  silent: 'Em silêncio',

  micBlocked: 'O navegador bloqueou o microfone.',
  recognitionUnavailable: 'O reconhecimento de voz não está disponível neste navegador.',
  noMicrophone: 'Nenhum microfone encontrado.',
  recognitionOffline: 'Sem conexão com o serviço de reconhecimento de voz.',
  recognitionError: (error: string) => `Erro no reconhecimento de voz: ${error}`,
}

type Messages = typeof pt

const en: Messages = {
  you: 'You',
  unlockAudio: 'Click anywhere to hear the agents',
  heldSpeeches: (count: number) =>
    `Microphone on: ${count} ${count === 1 ? 'update' : 'updates'} on hold. Turn it off to hear the agents.`,
  raisedHand: (name: string) => `${name} raised their hand`,

  runningCommand: 'Running command',
  thinking: 'Thinking',
  stalledTitle: 'It has been at this for a long time, it may be stuck',
  stalled: 'may be stuck',

  ttsOff: 'Voice is turned off (TTS=none). Captions only.',
  ttsDown: 'The voice service did not respond. If it just started, wait a few seconds.',
  playFailed: 'Could not play audio in this browser.',
  playing: 'Playing…',
  testAudio: 'Test audio',

  leftCall: 'You left the call',
  readyToJoin: 'Ready to join?',
  connecting: 'Connecting to Herdr…',
  noAgents: 'No agents in the call yet',
  agentsInCall: (count: number) => `${count} ${count === 1 ? 'agent is' : 'agents are'} in the call`,
  rejoin: 'Rejoin',
  joinNow: 'Join now',

  newAgent: 'New agent',
  close: 'Close',
  cancel: 'Cancel',
  send: 'Send',
  launchFailed: 'Could not start the agent.',
  newAgentHint: 'Opens a new Herdr workspace and the agent joins the call.',
  name: 'Name',
  namePlaceholder: 'E.g. Checkout reviewer',
  command: 'Command',
  commandHint: "Runs in the new workspace's shell: aliases and flags work.",
  folder: 'Folder',
  folderHint: 'Tab completes the folder, like in the terminal.',
  folderPlaceholder: '~/projects/my-app',
  firstMessage: 'First message',
  optional: '(optional)',
  firstMessagePlaceholder: 'What it should do when it joins',
  launching: 'Starting…',
  launchAgent: 'Start agent',

  imageTypes: 'Only PNG, JPEG, GIF or WebP images can be attached.',
  imageTooBig: (name: string) => `${name} is over 10 MB.`,
  tooManyImages: (max: number) => `At most ${max} images per message.`,
  sendFailed: 'Could not send the message.',
  chatEmpty: 'Nobody has spoken yet. Mention an agent with @ to talk to it.',
  sentImage: 'Sent image',
  allAgents: 'all agents',
  mentionToSend: 'Mention an agent with @ to send',
  removeImage: (name: string) => `Remove ${name}`,
  attachImage: 'Attach image',
  messagePlaceholder: 'Message (@ to mention)',
  nobodyMentioned: 'Nobody mentioned, it stayed in the chat',
  deliveredTo: (names: string) => `Delivered to ${names}`,
  notDeliveredTo: (name: string) => `Not delivered to ${name}`,
  waitingApproval: ': waiting for approval',

  micUnsupported: 'Voice dictation only works in Chrome',
  micOff: 'Turn off microphone',
  dictate: 'Dictate a message',
  listenAgents: 'Hear agents',
  muteAgents: 'Mute agents',
  captions: 'Captions',
  leave: 'Leave call',
  meetingDetails: 'Meeting details',
  participants: 'People',
  chat: 'Chat',
  chatTitle: 'In-call messages',

  answerFailed: 'Could not answer.',
  needsYouLabel: (name: string) => `${name} needs you`,
  needsYou: 'needs you',
  expand: 'Expand',
  collapse: 'Collapse',
  yourAnswer: 'Your answer',
  back: 'Back',
  cancelEsc: 'Cancel (Esc)',

  infoJoin: 'Your Herdr agents join the call automatically and report what they are doing.',
  infoChat: "In the chat, mention an agent with @ to send a message straight into its session.",
  serverConnection: 'Server connection:',
  connected: 'active',
  reconnecting: 'reconnecting…',
  organizer: 'Meeting host',
  renameFrom: (original: string) => `Rename (original: ${original})`,
  rename: 'Rename',
  closeFailed: 'Could not close the pane.',
  endAgentWarning: 'End this agent? Its pane will be closed in Herdr.',
  ending: 'Ending…',
  end: 'End',
  endNamed: (name: string) => `End ${name}`,
  endAgent: 'End agent',

  needsApproval: 'Needs approval',
  talking: 'Talking',
  working: 'Working',
  silent: 'Silent',

  micBlocked: 'The browser blocked the microphone.',
  recognitionUnavailable: 'Speech recognition is not available in this browser.',
  noMicrophone: 'No microphone found.',
  recognitionOffline: 'No connection to the speech recognition service.',
  recognitionError: (error: string) => `Speech recognition error: ${error}`,
}

const catalogs: Record<string, Messages> = { pt, en }

// The server's MEET_LANGUAGE picks these before the app renders; languages
// without a catalog use English text but still dictate and format dates in
// their own language.
export let language = 'pt-BR'
export let mentionAll = 'todos'
export let t: Messages = pt

export async function loadLocale() {
  try {
    const response = await fetch('/api/config')
    if (response.ok) {
      const config = (await response.json()) as { language: string; mentionAll: string }
      language = config.language
      mentionAll = config.mentionAll
    }
  } catch {
    // Keep the defaults; the call itself reports the lost connection.
  }
  t = catalogs[language.split(/[-_]/)[0].toLowerCase()] ?? en
  document.documentElement.lang = language
}

export function formatTime(date: Date) {
  try {
    return date.toLocaleTimeString(language, { hour: '2-digit', minute: '2-digit' })
  } catch {
    return date.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' })
  }
}
