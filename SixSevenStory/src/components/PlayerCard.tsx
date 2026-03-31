/* =====================================================
   PlayerCard.tsx — Individual character card
   ===================================================== */

import type { Player } from '../types'
import Inventory from './Inventory'

interface Props { player: Player }

/** Role → emoji */
function roleEmoji(role: string): string {
    const r = role.toLowerCase()
    if (r.includes('warrior') || r.includes('fighter') || r.includes('knight')) return '⚔️'
    if (r.includes('mage') || r.includes('wizard')) return '🪄'
    if (r.includes('archer') || r.includes('ranger')) return '🏹'
    if (r.includes('sorcerer') || r.includes('cleric') || r.includes('priest')) return '✨'
    if (r.includes('rogue')) return '🗡️'
    if (r.includes('paladin')) return '🛡️'
    return '🧙'
}

/** Race → emoji */
function raceEmoji(race: string): string {
    const r = race.toLowerCase()
    if (r.includes('human')) return '👤'
    if (r.includes('elf')) return '🧝'
    if (r.includes('dwarf')) return '⛏️'
    if (r.includes('orc')) return '👹'
    if (r.includes('gnome')) return '🔮'
    return ''
}

/** A simple percentage bar for HP or Mana */
function Bar({ label, value, max = 100, emoji, fillColor }: {
    label: string; value: number; max?: number
    emoji: string; fillColor: string
}) {
    const pct = Math.max(0, Math.min(100, (value / max) * 100))
    return (
        <div className="bar">
            <div className="bar__row">
                <span className="bar__label">{emoji} {label}</span>
                <span className="bar__value">{value}<span className="bar__max">/{max}</span></span>
            </div>
            <div className="bar__track">
                <div className="bar__fill" style={{ width: `${pct}%`, background: fillColor }} />
            </div>
        </div>
    )
}

export default function PlayerCard({ player }: Props) {
    const miniStats = [
        { key: 'strength' as const, label: 'STR', emoji: '💪' },
        { key: 'dexterity' as const, label: 'DEX', emoji: '🌀' },
        { key: 'intelligence' as const, label: 'INT', emoji: '📚' },
        { key: 'wisdom' as const, label: 'WIS', emoji: '🦉' },
        { key: 'charisma' as const, label: 'CHA', emoji: '💬' },
    ]

    return (
        <div className="pc fade-in">
            {/* Header */}
            <div className="pc__head">
                <span className="pc__emoji">{roleEmoji(player.role)}</span>
                <div>
                    <h3 className="pc__name">{player.name}</h3>
                    <div className="pc__sub">{raceEmoji(player.race)} {player.race} · {player.role}</div>
                </div>
            </div>

            {/* Lore */}
            <p className="pc__info">{player.info}</p>

            <hr />

            {/* HP + Mana bars */}
            <Bar label="HP" value={player.stats.health} emoji="❤️" fillColor="linear-gradient(90deg,#7b0000,#e74c3c)" />
            <Bar label="Mana" value={player.stats.mana} emoji="💧" fillColor="linear-gradient(90deg,#002060,#5dade2)" />

            {/* Mini stat row */}
            <div className="pc__stats">
                {miniStats.map(({ key, label, emoji }) => (
                    <div key={key} className="pc__stat">
                        <span>{emoji}</span>
                        <span className="pc__stat-lbl">{label}</span>
                        <strong>{player.stats[key]}</strong>
                    </div>
                ))}
            </div>

            {/* Collapsible inventory */}
            <Inventory items={player.items} />

            <style>{`
        .pc {
          background: var(--surface2);
          border: 1px solid var(--border-dim);
          border-left: 3px solid var(--gold);
          border-radius: 8px;
          padding: 0.85rem 1rem;
          transition: box-shadow 0.2s;
        }
        .pc:hover { box-shadow: 0 0 14px rgba(201,168,76,0.15); }

        .pc__head { display: flex; align-items: center; gap: 0.6rem; margin-bottom: 0.45rem; }
        .pc__emoji { font-size: 1.8rem; line-height: 1; }
        .pc__name  { font-size: 1rem; color: var(--gold-light); margin: 0; }
        .pc__sub   { font-size: 0.75rem; color: var(--text-dim); font-style: italic; }

        .pc__info  { font-size: 0.83rem; color: var(--text-dim); font-style: italic; line-height: 1.4; margin: 0.4rem 0 0.5rem; }

        /* Bars */
        .bar { margin-top: 0.45rem; }
        .bar__row   { display: flex; justify-content: space-between; font-size: 0.76rem; color: var(--text-dim); margin-bottom: 0.18rem; }
        .bar__value { font-family: var(--font-heading); color: var(--text); font-size: 0.76rem; }
        .bar__max   { opacity: 0.5; }
        .bar__track { height: 6px; background: rgba(0,0,0,0.5); border-radius: 3px; overflow: hidden; }
        .bar__fill  { height: 100%; border-radius: 3px; transition: width 0.5s ease; }

        /* Mini stats */
        .pc__stats {
          display: flex; flex-wrap: wrap; gap: 0.3rem; margin-top: 0.7rem;
        }
        .pc__stat {
          display: flex; flex-direction: column; align-items: center;
          background: rgba(0,0,0,0.3); border: 1px solid var(--border-dim);
          border-radius: 5px; padding: 0.25rem 0.5rem; flex: 1; min-width: 40px;
          font-size: 0.75rem;
        }
        .pc__stat-lbl { font-family: var(--font-heading); font-size: 0.62rem; color: var(--text-dim); text-transform: uppercase; }
        .pc__stat strong { font-family: var(--font-heading); font-size: 0.85rem; color: var(--gold-light); }
      `}</style>
        </div>
    )
}
