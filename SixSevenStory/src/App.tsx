import { useEffect, useRef, useState } from "react";
import "./App.css";
import AuthPage from "./components/AuthPage";
import Players from "./components/Players";
import type { GameState } from "./types";

const FETCH_INTERVAL_MS = 5000;

function getChatIdFromUrl(): string | null {
    const params = new URLSearchParams(window.location.search);
    const idFromParam = params.get("id");
    if (idFromParam && /^-?\d+$/.test(idFromParam)) {
        return idFromParam;
    }
    const path = window.location.pathname;
    const cleanPath = path.replace(/^\//, "");
    if (cleanPath && /^-?\d+$/.test(cleanPath)) {
        return cleanPath;
    }
    return params.get("game") || params.get("chatId");
}

function transformDataToGameState(data: {
    history?: unknown[];
    players?: unknown[];
}): GameState {
    return {
        history: (data.history as GameState["history"]) || [],
        players: (data.players as GameState["players"]) || [],
    };
}

function App() {
    const [game, setGame] = useState<GameState | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [lastSync, setLastSync] = useState<Date | null>(null);
    const [loading, setLoading] = useState(true);
    const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
    const chatId = getChatIdFromUrl();

    const [authenticated, setAuthenticated] = useState(false);

    useEffect(() => {
        const token = localStorage.getItem("jwt_token");
        if (token) {
            setAuthenticated(true);
        }
    }, [localStorage]);

    const fetchGame = async () => {
        if (!chatId) {
            setError(
                "No se ha especificado un juego. Añade el ID a la URL (ej: ?id=123456)",
            );
            setLoading(false);
            return;
        }
        try {
            const res = await fetch(`/data_${chatId}.json`);
            if (!res.ok) {
                if (res.status === 404) {
                    setError(`Juego no encontrado: ${chatId}`);
                } else {
                    setError(`HTTP ${res.status}`);
                }
                return;
            }
            const data = await res.json();
            setGame(transformDataToGameState(data));
            setError(null);
            setLastSync(new Date());
            setLoading(false);
        } catch (err) {
            setError(err instanceof Error ? err.message : "Error desconocido");
            setLoading(false);
        }
    };

    useEffect(() => {
        if (!authenticated) return;
        fetchGame();
        intervalRef.current = setInterval(fetchGame, FETCH_INTERVAL_MS);
        return () => {
            if (intervalRef.current) clearInterval(intervalRef.current);
        };
    }, [chatId]);

    const syncLabel = lastSync ? lastSync.toLocaleTimeString() : "—";

    if (!authenticated) {
        return <AuthPage setAuthenticated={setAuthenticated} />;
    }

    if (!chatId) {
        return (
            <div className="app">
                <div className="state-error">
                    ⚠️ No se ha especificado un juego.
                    <br />
                    Añade el ID a la URL (ej: /123456)
                </div>
            </div>
        );
    }

    if (!game && loading) {
        return (
            <div className="app">
                <div className="state-loading">
                    <div className="spinner" />
                    🗺️ Cargando partida…
                </div>
            </div>
        );
    }

    if (error && !game) {
        return (
            <div className="app">
                <div className="state-error">
                    ⚠️ No se pudo cargar la partida:
                    <br />
                    <code>{error}</code>
                </div>
            </div>
        );
    }

    return (
        <div className="app">
            <header className="app-header">
                <div>
                    <h1>⚔️ SixSevenStory</h1>
                    <span className="app-header__subtitle">
                        Partida #{chatId}
                    </span>
                </div>

                <div className="app-header__controls">
                    <div className="refresh-indicator">
                        <div className="refresh-dot" />
                        Actualizado: {syncLabel}
                    </div>
                    <button onClick={fetchGame}>🔄 Actualizar</button>
                </div>
            </header>

            <div className="bottom-layout">
                <section className="panel scroll-panel" style={{ flex: 1 }}>
                    {game && <Players players={game.players} />}
                </section>
            </div>
        </div>
    );
}

export default App;
