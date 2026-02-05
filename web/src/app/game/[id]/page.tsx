"use client";

import { useEffect, useMemo, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import styles from "../game.module.css";

interface ServerMessage {
  type: string;
  payload: any;
}

export default function GamePage() {
  const params = useParams();
  const gameId = useMemo(
    () => (Array.isArray(params?.id) ? params?.id[0] : params?.id),
    [params]
  );
  const [status, setStatus] = useState("Waiting for opponent...");
  const [player, setPlayer] = useState<string | null>(null);
  const [connected, setConnected] = useState(false);

  const wsUrl = process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8080/ws";

  useEffect(() => {
    if (!gameId) return;

    const token = localStorage.getItem("ws_token") || "";
    setPlayer(localStorage.getItem("ws_player"));

    const url = new URL(wsUrl);
    url.searchParams.set("token", token);
    const socket = new WebSocket(url.toString());

    socket.onopen = () => {
      setConnected(true);
    };

    socket.onmessage = (event) => {
      const msg: ServerMessage = JSON.parse(event.data);
      if (msg.type === "opponent_joined") {
        setStatus("Opponent joined. Ready to place ships.");
      }
      if (msg.type === "game_state") {
        if (msg.payload?.status === "waiting") {
          setStatus("Waiting for opponent...");
        }
        if (msg.payload?.status === "placing") {
          setStatus("Both players connected. Place your ships.");
        }
        if (msg.payload?.status === "active") {
          setStatus("Match in progress.");
        }
      }
    };

    socket.onclose = () => {
      setConnected(false);
    };

    return () => {
      socket.close();
    };
  }, [gameId, wsUrl]);

  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <Link href="/" className={styles.backLink}>
          ← Back to lobby
        </Link>
        <div>
          <p className={styles.label}>Game</p>
          <h1>{gameId}</h1>
        </div>
      </header>

      <section className={styles.panel}>
        <p className={styles.label}>Status</p>
        <p className={styles.statusText}>{status}</p>
        <div className={styles.metaRow}>
          <span>Player: {player || "-"}</span>
          <span>Socket: {connected ? "connected" : "disconnected"}</span>
        </div>
        <div className={styles.waitCard}>
          <p>Share this game ID or join code with your opponent.</p>
          <p>Once they join, you will be notified here.</p>
        </div>
      </section>
    </div>
  );
}
