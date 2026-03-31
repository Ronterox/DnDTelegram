import type { HistoryEvent } from '../types'

interface Props {
  history: HistoryEvent[]
  selectedIndex: number | null
  onSelect: (index: number) => void
}

/** Detecta tipo de acción para emoji + borde */
function classify(action: string): { emoji: string; cls: string } {
  const a = action.toLowerCase()
  if (/attack|fight|strike|slash|sword|combat/.test(a)) return { emoji: '⚔️', cls: 'combat' }
  if (/magic|spell|cast|enchant|arcane/.test(a)) return { emoji: '🪄', cls: 'magic' }
  if (/damage|poison|hit|wound|hurt/.test(a)) return { emoji: '💥', cls: 'damage' }
  if (/move|went|walk|travel|enter|arrive|left/.test(a)) return { emoji: '🏃', cls: 'move' }
  if (/heal|cure|recover|rest|restore/.test(a)) return { emoji: '❤️', cls: 'heal' }
  if (/found|loot|treasure|chest/.test(a)) return { emoji: '🔍', cls: '' }
  if (/met|greet|speak|say|talk/.test(a)) return { emoji: '💬', cls: '' }
  if (/death|die|dead|fall|perish/.test(a)) return { emoji: '💀', cls: 'damage' }
  return { emoji: '📖', cls: '' }
}

export default function History({ history, selectedIndex, onSelect }: Props) {
  return (
    <div className="hist">
      {history.length === 0 && <p className="hist__empty">No hay eventos registrados.</p>}

      <div className="hist__scroll">
        {history.map((ev, i) => {
          const { emoji, cls } = classify(ev.action)
          const isSelected = i === selectedIndex

          return (
            <div
              key={i}
              className={`he fade-in ${cls ? `he--${cls}` : ''} ${isSelected ? 'he--selected' : ''}`}
              style={{ animationDelay: `${i * 0.03}s` }}
            >
              <span className="he__icon">{emoji}</span>
              <div className="he__body">
                <p className="he__text">{ev.action}</p>
                {ev.damage !== undefined && (
                  <p className="he__dmg">💥 <strong>-{ev.damage}</strong> daño</p>
                )}
                {ev.map && (
                  <span
                    className="he__map"
                    onClick={() => onSelect(i)}
                    style={{ cursor: 'pointer' }}
                    title="Mostrar mapa de este evento"
                  >
                    🗺️ cambio de mapa
                  </span>
                )}
              </div>
              <span className="he__num">#{i + 1}</span>
            </div>
          )
        })}
      </div>

      <style>{`
        .hist__scroll { display: flex; flex-direction: column; gap: 0.4rem; max-height: 430px; overflow-y: auto; padding-right: 0.2rem; }
        .hist__empty { color: var(--text-dim); font-style: italic; font-size: 0.88rem; text-align: center; }

        .he { display: flex; align-items: flex-start; gap: 0.5rem; padding: 0.5rem 0.65rem; border-radius: 6px; border: 1px solid var(--border-dim); background: rgba(0,0,0,0.22); border-left: 3px solid var(--border-dim); transition: background 0.15s; }
        .he:hover { background: rgba(201,168,76,0.06); }

        .he--selected { background: rgba(255,215,0,0.12); border-left-color: var(--gold); }

        .he--combat { border-left-color: #c0392b; }
        .he--damage { border-left-color: #e74c3c; background: rgba(231,76,60,0.05); }
        .he--magic  { border-left-color: #8e44ad; }
        .he--heal   { border-left-color: #27ae60; }
        .he--move   { border-left-color: #2980b9; }

        .he__icon { font-size: 1.1rem; line-height: 1; flex-shrink: 0; margin-top: 0.1rem; }
        .he__body { flex: 1; min-width: 0; }
        .he__text { font-size: 0.88rem; color: var(--text); line-height: 1.4; margin: 0; }
        .he__dmg  { font-size: 0.78rem; color: var(--red-light); margin: 0.15rem 0 0; font-family: var(--font-heading); }

        .he__map { display: inline-block; margin-top: 0.2rem; font-size: 0.7rem; color: var(--text-dim); font-style: italic; background: rgba(255,255,255,0.03); border: 1px solid var(--border-dim); border-radius: 10px; padding: 0.05em 0.5em; cursor: pointer; }
        .he__num { font-family: var(--font-heading); font-size: 0.62rem; color: var(--border); flex-shrink: 0; margin-top: 0.1rem; }
      `}</style>
    </div>
  )
}