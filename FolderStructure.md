# Project Folder Structure

This document describes the folder layout for the **Battleships Online (MVP)** project.

The structure follows:

- Go project best practices
- Clear separation of concerns
- Cloud-native and scalable architecture
- Readability for reviewers and recruiters

---

## Root Structure

.
├── backend/ # Go backend (REST + WebSocket)
├── web/ # Next.js web frontend
├── infra/ # Terraform infrastructure (AWS)
├── docker-compose.yml # Local development stack
├── Makefile # Common development commands
├── PROJECT_PLAN.md
├── FOLDER_STRUCTURE.md
└── README.md

---

## Backend (Go)

backend/
├── cmd/
│ └── api/
│ └── main.go # Application entrypoint
│
├── internal/ # Private application code
│ ├── auth/ # Authentication and JWT
│ │ ├── handler.go
│ │ └── middleware.go
│ │
│ ├── game/ # Core Battleships game logic
│ │ ├── engine.go # Board, ships, rules
│ │ ├── validator.go
│ │ └── engine_test.go
│ │
│ ├── ws/ # WebSocket layer
│ │ ├── hub.go # Connection management
│ │ ├── handler.go # Message routing
│ │ └── messages.go # WebSocket message schemas
│ │
│ ├── http/ # REST API handlers
│ │ ├── games.go
│ │ ├── leaderboard.go
│ │ └── health.go
│ │
│ ├── store/ # Data access layer
│ │ ├── redis/
│ │ │ ├── client.go
│ │ │ ├── game_state.go
│ │ │ └── locks.go
│ │ │
│ │ └── postgres/
│ │ ├── client.go
│ │ ├── games.go
│ │ ├── events.go
│ │ └── leaderboard.go
│ │
│ ├── config/ # Configuration loading
│ │ └── config.go
│ │
│ └── observability/ # Logging and tracing (optional)
│ └── logger.go
│
├── migrations/ # SQL migrations
│ ├── 001_init.sql
│ ├── 002_games.sql
│ └── 003_leaderboard.sql
│
├── pkg/ # Reusable public packages (optional)
│
├── Dockerfile
├── go.mod
├── go.sum
└── README.mod

---

## Web Frontend (Next.js)

web/
├── app/ or pages/ # Next.js routing
│ ├── index.tsx # Lobby / Home
│ ├── game/
│ │ └── [id].tsx # Game view
│ └── leaderboard.tsx
│
├── components/ # Reusable UI components
│ ├── Board.tsx
│ ├── ShipPlacement.tsx
│ └── StatusBar.tsx
│
├── lib/
│ ├── api.ts # REST client
│ ├── ws.ts # WebSocket client
│ └── auth.ts
│
├── styles/
│ └── globals.css
│
├── public/
│
├── next.config.js
├── package.json
└── README.md

---

## Infrastructure (Terraform)

infra/
└── terraform/
├── providers.tf
├── main.tf
├── variables.tf
├── outputs.tf
│
├── vpc.tf
├── alb.tf
├── ecs.tf
├── rds.tf
├── redis.tf
├── iam.tf
└── security_groups.tf

---

## Local Development

docker-compose.yml

Includes:

- Go backend
- PostgreSQL
- Redis

Used for full local gameplay testing with the same architecture as production.
