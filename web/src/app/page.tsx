"use client";

import { useState } from "react";
import styles from "./page.module.css";

const grid = Array.from({ length: 100 }, (_, i) => i);
const shipCells = new Set([12, 13, 14, 15, 16, 44, 54, 64, 74, 34, 35, 36, 71, 81]);
const hitCells = new Set([44, 54, 12, 36]);
const missCells = new Set([7, 28, 57, 83, 91]);

const recentSignals = [
  { label: "Direct hit at E5", tone: "hit" },
  { label: "Splash at B8", tone: "miss" },
  { label: "Carrier spotted", tone: "intel" },
];

const leaderboard = [
  { name: "Astra", wins: 14, losses: 3 },
  { name: "Koda", wins: 11, losses: 5 },
  { name: "Nia", wins: 9, losses: 4 },
];

export default function Home() {
  const [joinCode, setJoinCode] = useState("");
  const [callSign, setCallSign] = useState("");
  const [gameInfo, setGameInfo] = useState<{
    gameId: string;
    joinCode?: string;
    player: string;
  } | null>(null);
  const [error, setError] = useState<string | null>(null);

  const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

  const handleCreate = async () => {
    setError(null);
    try {
      const res = await fetch(`${apiUrl}/games`, { method: "POST" });
      if (!res.ok) {
        throw new Error("Failed to create game");
      }
      const data = await res.json();
      localStorage.setItem("ws_token", data.token);
      setGameInfo({ gameId: data.game_id, joinCode: data.join_code, player: data.player });
    } catch (err) {
      setError("Could not create game");
    }
  };

  const handleJoin = async () => {
    setError(null);
    try {
      const res = await fetch(`${apiUrl}/games/join`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ join_code: joinCode }),
      });
      if (!res.ok) {
        throw new Error("Failed to join game");
      }
      const data = await res.json();
      localStorage.setItem("ws_token", data.token);
      setGameInfo({ gameId: data.game_id, player: data.player });
    } catch (err) {
      setError("Could not join game");
    }
  };

  return (
    <div className={styles.page}>
      <header className={styles.nav}>
        <div className={styles.brand}>
          <span className={styles.brandMark} />
          <div>
            <p className={styles.brandTitle}>WebShips</p>
            <p className={styles.brandSubtitle}>Live Battleship Arena</p>
          </div>
        </div>
        <nav className={styles.navLinks}>
          <a href="#lobby">Lobby</a>
          <a href="#play">Play</a>
          <a href="#leaderboard">Leaderboard</a>
        </nav>
        <button className={styles.navButton}>Guest Login</button>
      </header>

      <main className={styles.main}>
        <section className={styles.hero} id="lobby">
          <div className={styles.heroContent}>
            <p className={styles.eyebrow}>Command a 10x10 grid</p>
            <h1>Plot your fleet, fire in real time, and claim the tide.</h1>
            <p className={styles.lead}>
              Fast, tactical Battleship rounds with live updates, recon
              pings, and a clean lobby flow for guests.
            </p>
            <div className={styles.actions}>
              <button className={styles.primaryButton} onClick={handleCreate}>
                Create Game
              </button>
              <button className={styles.secondaryButton} onClick={handleJoin}>
                Join with Code
              </button>
            </div>
            <div className={styles.guestCard}>
              <div>
                <p className={styles.cardLabel}>Guest Login</p>
                <p className={styles.cardHelper}>
                  Pick a callsign and jump into the lobby.
                </p>
              </div>
              <div className={styles.inputRow}>
                <input
                  placeholder="Callsign"
                  value={callSign}
                  onChange={(event) => setCallSign(event.target.value)}
                />
                <button className={styles.ghostButton}>Enter</button>
              </div>
              <div className={styles.joinRow}>
                <input
                  placeholder="Join code"
                  value={joinCode}
                  onChange={(event) => setJoinCode(event.target.value)}
                />
                <button className={styles.ghostButton} onClick={handleJoin}>
                  Join
                </button>
              </div>
              {gameInfo ? (
                <div className={styles.sessionInfo}>
                  <p>Game: {gameInfo.gameId}</p>
                  <p>Player: {gameInfo.player}</p>
                  {gameInfo.joinCode ? <p>Join code: {gameInfo.joinCode}</p> : null}
                </div>
              ) : null}
              {error ? <p className={styles.errorText}>{error}</p> : null}
            </div>
          </div>

          <div className={styles.boardCard}>
            <div className={styles.boardHeader}>
              <div>
                <p className={styles.cardLabel}>Live Board</p>
                <p className={styles.cardHelper}>Turn: Player One</p>
              </div>
              <span className={styles.turnChip}>P1</span>
            </div>
            <div className={styles.board}>
              {grid.map((cell) => {
                const classes = [styles.cell];
                if (shipCells.has(cell)) classes.push(styles.ship);
                if (hitCells.has(cell)) classes.push(styles.hit);
                if (missCells.has(cell)) classes.push(styles.miss);
                return <div key={cell} className={classes.join(" ")} />;
              })}
            </div>
            <div className={styles.signalPanel}>
              <p className={styles.cardLabel}>Signals</p>
              <div className={styles.signalList}>
                {recentSignals.map((signal) => (
                  <div key={signal.label} className={styles.signalItem}>
                    <span className={`${styles.signalDot} ${styles[signal.tone]}`} />
                    <p>{signal.label}</p>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </section>

        <section className={styles.steps} id="play">
          <div className={styles.stepCard}>
            <p className={styles.stepTitle}>Place ships</p>
            <p>Drag your fleet into formation and lock placements.</p>
          </div>
          <div className={styles.stepCard}>
            <p className={styles.stepTitle}>Fire salvos</p>
            <p>Coordinate shots in turn-based bursts with live results.</p>
          </div>
          <div className={styles.stepCard}>
            <p className={styles.stepTitle}>Track momentum</p>
            <p>Realtime turn updates and recon keeps every duel honest.</p>
          </div>
        </section>

        <section className={styles.leaderboard} id="leaderboard">
          <div className={styles.leaderboardHeader}>
            <div>
              <p className={styles.eyebrow}>Captain rankings</p>
              <h2>Leaderboard</h2>
            </div>
            <button className={styles.secondaryButton}>View full board</button>
          </div>
          <div className={styles.leaderboardGrid}>
            {leaderboard.map((entry, index) => (
              <div key={entry.name} className={styles.leaderRow}>
                <span className={styles.rank}>#{index + 1}</span>
                <div>
                  <p className={styles.leaderName}>{entry.name}</p>
                  <p className={styles.leaderRecord}>
                    {entry.wins}W · {entry.losses}L
                  </p>
                </div>
                <span className={styles.rankBadge}>+{entry.wins * 12}</span>
              </div>
            ))}
          </div>
        </section>
      </main>
    </div>
  );
}
