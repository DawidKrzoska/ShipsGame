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
              <button className={styles.primaryButton}>Create Game</button>
              <button className={styles.secondaryButton}>Join with Code</button>
            </div>
            <div className={styles.guestCard}>
              <div>
                <p className={styles.cardLabel}>Guest Login</p>
                <p className={styles.cardHelper}>
                  Pick a callsign and jump into the lobby.
                </p>
              </div>
              <div className={styles.inputRow}>
                <input placeholder="Callsign" />
                <button className={styles.ghostButton}>Enter</button>
              </div>
              <div className={styles.joinRow}>
                <input placeholder="Join code" />
                <button className={styles.ghostButton}>Join</button>
              </div>
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
