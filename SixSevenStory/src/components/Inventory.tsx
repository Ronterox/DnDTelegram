/* =====================================================
   Inventory.tsx — Collapsible item list per player
   ===================================================== */

import { useState } from 'react'
import type { Item } from '../types'

interface Props { items: Item[] }

function itemEmoji(type: string): string {
  const t = type.toLowerCase()
  if (t.includes('weapon') || t.includes('sword')) return '⚔️'
  if (t.includes('armor') || t.includes('shield')) return '🛡️'
  if (t.includes('magic') || t.includes('staff') || t.includes('robe')) return '🪄'
  if (t.includes('potion')) return '🧪'
  if (t.includes('bow') || t.includes('arrow')) return '🏹'
  return '📦'
}

export default function Inventory({ items }: Props) {
  const [open, setOpen] = useState(false)

  return (
    <div style={{ marginTop: '0.7rem' }}>
      <button
        className="inv-btn"
        onClick={() => setOpen(o => !o)}
        aria-expanded={open}
      >
        🎒 {open ? 'Ocultar inventario' : 'Ver inventario'}
        <span className="inv-badge">{items.length}</span>
      </button>

      {open && (
        <ul className="inv-list fade-in">
          {items.length === 0 && <li className="inv-empty">Sin objetos</li>}
          {items.map((item, i) => (
            <li key={i} className="inv-item">
              <span>{itemEmoji(item.type)}</span>
              <span className="inv-name">{item.name}</span>
              <span className="inv-type">{item.type}</span>
            </li>
          ))}
        </ul>
      )}

      <style>{`
        .inv-btn {
          font-size: 0.76rem;
          padding: 0.32em 0.85em;
          display: flex;
          align-items: center;
          gap: 0.4rem;
          width: 100%;
          justify-content: center;
        }
        .inv-badge {
          background: var(--border);
          color: var(--gold-light);
          border-radius: 10px;
          padding: 0 0.45em;
          font-size: 0.7rem;
          line-height: 1.6;
        }
        .inv-list {
          list-style: none;
          border: 1px solid var(--border-dim);
          border-radius: 5px;
          overflow: hidden;
          margin-top: 0.4rem;
        }
        .inv-item {
          display: flex;
          align-items: center;
          gap: 0.45rem;
          padding: 0.36rem 0.6rem;
          border-bottom: 1px solid var(--border-dim);
          font-size: 0.84rem;
        }
        .inv-item:last-child { border-bottom: none; }
        .inv-item:hover { background: rgba(201,168,76,0.05); }

        .inv-name { flex: 1; color: var(--text); font-weight: 600; }
        .inv-type { font-size: 0.72rem; color: var(--text-dim); font-style: italic; }
        .inv-empty { padding: 0.4rem 0.6rem; color: var(--text-dim); font-style: italic; font-size: 0.84rem; }
      `}</style>
    </div>
  )
}
