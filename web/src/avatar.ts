const COLORS = ['#e8710a', '#1e8e3e', '#a142f4', '#d93025', '#12a4af', '#f538a0', '#5f6368', '#1a73e8']

export function avatarColor(id: string) {
  let hash = 0
  for (const char of id) hash = (hash * 31 + char.charCodeAt(0)) | 0
  return COLORS[Math.abs(hash) % COLORS.length]
}

export function initials(name: string) {
  return name.trim().charAt(0).toUpperCase() || '?'
}
