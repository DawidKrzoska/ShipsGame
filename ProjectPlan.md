# Battleships Online – Project Plan (MVP)

## 1. Project Overview

**Battleships Online** is a scalable, multiplayer web-based Battleships game built as a production-grade portfolio project.

The project demonstrates:

- Backend engineering in **Go**
- Real-time communication via **WebSockets**
- **Redis** for distributed live game state
- **PostgreSQL** for persistence (leaderboards, history)
- Cloud-native deployment on **AWS** using **Terraform**
- Modern frontend deployment on **Vercel**

The MVP prioritizes correctness, scalability, and clean architecture over feature breadth.

---

## 2. Project Objectives

### Primary Objective

Build a multiplayer Battleships game where two players can play in real time via a web browser, with the backend capable of horizontal scaling behind a load balancer.

### Technical Objectives

- Stateless Go backend behind an AWS Application Load Balancer
- Redis used as the authoritative store for **live game state**
- PostgreSQL used for **finished games, event history, and leaderboard**
- WebSocket-based gameplay
- Infrastructure fully provisioned using Terraform
- Web UI deployed separately (Vercel)

### Definition of Done (MVP)

- Two players can create/join a game and complete a full match
- Game logic is validated server-side
- Leaderboard updates after each finished game
- Backend runs with multiple ECS tasks without breaking gameplay
- Infrastructure can be recreated from scratch via Terraform

---

## 3. Architecture Overview

### High-Level Architecture

Browser (Next.js on Vercel)

| HTTPS / WebSocket
v
AWS Application Load Balancer
|
v
ECS Fargate (Go Backend – multiple instances)
|
+--> Redis (ElastiCache) – live game state
|
+--> PostgreSQL (RDS) – history & leaderboard

### Key Design Choice

The backend is **stateless**.  
All live game state is stored in Redis, allowing any ECS task to handle any request or WebSocket message.

---

## 4. MVP Feature Scope

### Included in MVP

- Guest authentication (JWT-based, no passwords)
- Create game and join via short join code
- Ship placement validation
- Turn-based firing logic
- Win detection
- Real-time updates via WebSocket
- Redis-backed live game state
- Persistent leaderboard and game history
- AWS deployment via Terraform

### Explicitly Excluded (Post-MVP)

- TUI client
- Ranked / ELO matchmaking
- Spectator mode
- Replay UI
- Metrics dashboards

---

## 5. Backend Design (Go)

### API Types

#### REST API

- `POST /auth/guest`
- `POST /games` – create new game
- `POST /games/join` – join by code
- `GET /leaderboard`
- `GET /games/{id}/summary`
- `GET /games/{id}/events`

#### WebSocket API

- `GET /ws?token=JWT`

**Client → Server**

- `place_ships`
- `fire`

**Server → Client**

- `game_state`
- `shot_result`
- `turn_changed`
- `game_finished`
- `error`

---

## 6. Redis Design (Live Game State)

Redis is the **authoritative store** for all active games.

### Key Structure

- `game:{gameId}:state`
- `game:{gameId}:players`
- `join:{joinCode} -> gameId` (TTL)

### State Contents

- Game status (`waiting`, `active`, `finished`)
- Player boards (ships and hits)
- Current turn
- Readiness flags

### Concurrency

- Updates performed atomically using `WATCH/MULTI` or Lua scripts
- Guarantees correctness under concurrent access and horizontal scaling

---

## 7. PostgreSQL Design (Persistence)

PostgreSQL is used only for **finished games and analytics**.

### Core Tables

- `users`
- `games`
- `game_players`
- `game_events`
- `leaderboard`

### Persistence Strategy

- Live gameplay uses Redis only
- On game finish:
  - Persist game summary
  - Store ordered game events
  - Update leaderboard
- All writes executed inside a single transaction

---

## 8. Frontend (Web UI)

### Stack

- Next.js (React)
- WebSocket for gameplay
- REST for authentication and data
- Deployed on Vercel

### MVP Screens

- Guest login
- Create / Join game
- Game board
- Leaderboard

---

## 9. Infrastructure (Terraform + AWS)

### AWS Components

- VPC and subnets
- Application Load Balancer
- ECS Fargate service (Go backend)
- RDS PostgreSQL
- ElastiCache Redis
- IAM roles and security groups

### Scaling

- ECS Service autoscaling based on CPU or request count
- No sticky sessions required due to Redis-backed state

### Secrets Management

- JWT signing key
- Database credentials
- Redis endpoint
- Stored in AWS Secrets Manager or SSM Parameter Store

---

## 10. Local Development

### Tooling

- Docker Compose
- Local PostgreSQL
- Local Redis
- Go backend service

### Goals

- Full gameplay runnable locally
- Identical runtime behavior to cloud environment

---

## 11. Development Milestones

1. Core game engine (pure Go + unit tests)
2. Redis-backed game state management
3. WebSocket gameplay loop
4. PostgreSQL persistence on game finish
5. Minimal Next.js web UI
6. Terraform-based AWS deployment
7. Horizontal scale testing (multiple ECS tasks)

---

## 12. CV Highlights

- Designed stateless Go services with Redis-backed real-time state
- Implemented scalable WebSocket multiplayer gameplay
- Deployed autoscaling backend using AWS ALB and ECS Fargate
- Managed infrastructure entirely with Terraform
- Event-based persistence model for game history and analytics

---
