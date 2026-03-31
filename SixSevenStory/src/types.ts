export interface Item {
  name: string
  type: string
}

export interface Stats {
  health: number
  mana: number
  strength: number
  dexterity: number
  intelligence: number
  wisdom: number
  charisma: number
}

export interface Player {
  name: string
  info: string
  role: string
  race: string
  items: Item[]
  stats: Stats
}

export interface HistoryEvent {
  action: string
  map?: string
  damage?: number
}

export interface GameState {
  history: HistoryEvent[]
  players: Player[]
}