/* =====================================================
   Players.tsx — Collapsible player list wrapper
   Renders a <PlayerCard> for each player in the game state
   ===================================================== */

import { useState } from 'react'
import type { Player } from '../types'
import PlayerCard from './PlayerCard'

interface Props {
  players: Player[]
}

/**
 * Wrapper component that shows a "Ver jugadores" toggle button
 * and renders one <PlayerCard> per player when expanded.
 */
export default function Players({ players }: Props) {
  const [visible, setVisible] = useState(true)

  return (
    <div className="players">
      {/* Section header with toggle */}
      <div className="players__header">
        <h2 className="players__title">
          🧙 Adventurers
          <span className="players__count">{players.length}</span>
        </h2>
        <button
          className="players__toggle"
          onClick={() => setVisible(v => !v)}
          aria-expanded={visible}
        >
          {visible ? '👁️ Ocultar' : '👁️ Ver jugadores'}
        </button>
      </div>

      {/* Player cards — collapsible */}
      {visible && (
        <div className="players__grid fade-in">
          {players.length === 0 && (
            <p className="players__empty">No hay aventureros en la partida.</p>
          )}
          {players.map(player => (
            <PlayerCard key={player.name} player={player} />
          ))}
        </div>
      )}

      <style>{`
        .players {
          height: 100%;
        }

        .players__header {
          display: flex;
          align-items: center;
          justify-content: space-between;
          flex-wrap: wrap;
          gap: 0.5rem;
          margin-bottom: 0.9rem;
        }

        .players__title {
          display: flex;
          align-items: center;
          gap: 0.5rem;
          margin: 0;
        }

        .players__count {
          background: var(--color-border);
          color: var(--color-gold-light);
          border-radius: 12px;
          padding: 0 0.55em;
          font-size: 0.75rem;
          line-height: 1.7;
          font-family: var(--font-heading);
        }

        .players__toggle {
          font-size: 0.78rem;
          padding: 0.35em 0.9em;
        }

        .players__grid {
          display: flex;
          flex-direction: column;
          gap: 0.8rem;
          max-height: 460px;
          overflow-y: auto;
          padding-right: 0.3rem;
        }

        .players__empty {
          color: var(--color-text-dim);
          font-style: italic;
          font-size: 0.9rem;
          text-align: center;
        }
      `}</style>
    </div>
  )
}