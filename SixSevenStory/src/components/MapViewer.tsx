interface Props { map: string }

export default function MapViewer({ map }: Props) {
  return (
    <div className="map-viewer">
      <div className="map-viewer__label">🗺️ Current Map</div>

      {map ? (
        <div className="map-viewer__wrap">
          <pre className="map-viewer__pre">{map}</pre>
        </div>
      ) : (
        <p className="map-viewer__empty">No hay mapa disponible para este evento.</p>
      )}

      <style>{`
        .map-viewer { width: 100%; }
        .map-viewer__label { font-size: 0.85rem; font-weight: 600; color: var(--gold-light); letter-spacing: 0.08em; margin-bottom: 0.8rem; text-align: center; text-transform: uppercase; }
        .map-viewer__wrap { width: 100%; max-height: clamp(220px, 40vh, 500px); overflow: auto; display: flex; justify-content: center; }
        .map-viewer__pre { font-family: var(--font-mono); font-size: clamp(11px,1.2vw,15px); line-height: 1.2; white-space: pre; color: #4dff80; background: #040906; padding: 1.2rem 1.5rem; margin: 0; border-radius: 8px; border: 1px solid #0d2212; text-shadow: 0 0 4px rgba(77,255,128,0.8),0 0 12px rgba(77,255,128,0.25); min-width: fit-content; }
        .map-viewer__empty { text-align: center; font-style: italic; color: var(--text-dim); padding: 0.5rem; }
      `}</style>
    </div>
  )
}