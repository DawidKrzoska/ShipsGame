# Battleships Online – Build Order (MVP)

This document defines the **recommended order of work** for building the Battleships Online MVP.
Following this sequence minimizes rework, keeps the scope under control, and ensures that each
step builds on a stable foundation.

---

## 1. Repository Setup & Local Stack

**Goal:** Establish a stable local development environment.

### Tasks

- Create repository structure:
  - `backend/`
  - `web/`
  - `infra/`
- Add `docker-compose.yml` with:
  - PostgreSQL
  - Redis
- Add `Makefile` targets:
  - `up` / `down`
  - `dev`
  - `test`
  - `lint`
- Verify Postgres and Redis start correctly and are reachable.

### Output

- Local infrastructure running
- Backend can connect to Redis and Postgres

---

## 2. Core Game Engine (Pure Go)

**Goal:** Build the Battleships ruleset independent of networking or storage.

### Tasks

- Board representation
- Ship placement validation
- Hit / miss / sunk logic
- Win condition detection
- Comprehensive unit tests

### Output

- Tested `internal/game` package
- Deterministic and validated game logic

---

## 3. Redis Game-State Layer

**Goal:** Enable scalable, stateless backend design.

### Tasks

- Define Redis key structure and state schema
- Implement functions:
  - Create game (returns `gameId` and `joinCode`)
  - Join game by code
  - Place ships
  - Fire at opponent
- Implement atomic updates using `WATCH/MULTI` or Lua

### Output

- Fully functional multiplayer game state stored in Redis
- Safe concurrent access

---

## 4. Backend API Skeleton

**Goal:** Prepare the HTTP foundation of the backend.

### Tasks

- HTTP server setup
- Routing structure
- Configuration loading from environment variables
- Logging setup
- Health endpoint (`/healthz`)

### Output

- Running backend service with health checks
- Clean separation of concerns

---

## 5. WebSocket Gameplay

**Goal:** Enable real-time multiplayer gameplay.

### Tasks

- WebSocket authentication via JWT
- Message routing:
  - `place_ships`
  - `fire`
- Server-side validation using Redis state
- Broadcast updates to both players:
  - `game_state`
  - `shot_result`
  - `turn_changed`
  - `game_finished`
- Reconnection support (send full state on connect)

### Output

- Playable real-time game using WebSockets

---

## 6. PostgreSQL Persistence

**Goal:** Persist completed games and statistics.

### Tasks

- Create database migrations
- Define schema for:
  - users
  - games
  - game_events
  - leaderboard
- On game completion:
  - Write game summary
  - Persist ordered game events
  - Update leaderboard
- Wrap writes in a single transaction

### Output

- Durable game history
- Accurate leaderboard

---

## 7. Minimal Web UI

**Goal:** Provide a usable browser-based interface.

### Tasks

- Guest login flow
- Create / Join game screen
- Game board UI:
  - Ship placement
  - Firing interaction
- Leaderboard screen

### Output

- Full end-to-end gameplay in browser

---

## 8. Production Packaging

**Goal:** Prepare the application for deployment.

### Tasks

- Backend Dockerfile
- Environment-based configuration
- CORS configuration
- WebSocket URL configuration for production
- Basic documentation

### Output

- Production-ready container images
- Clear run and deploy instructions

---

## 9. Terraform AWS Deployment

**Goal:** Deploy scalable backend infrastructure.

### Tasks

- Terraform configuration for:
  - VPC and subnets
  - Application Load Balancer
  - ECS Fargate service
  - RDS PostgreSQL
  - ElastiCache Redis
- Configure environment variables and secrets
- Deploy backend and verify endpoints

### Output

- Backend running on AWS
- Infrastructure reproducible via Terraform

---

## 10. Scale Verification & CV Polish

**Goal:** Validate scalability and finalize portfolio quality.

### Tasks

- Run multiple ECS tasks
- Verify gameplay works across instances
- Add architecture diagram
- Document scaling approach and tradeoffs
- (Optional) Add GitHub Actions CI/CD pipeline

### Output

- Proven horizontally scalable system
- Strong, interview-ready project

---
