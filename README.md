# 🎬 FanFare Backend

> **Pop Culture Prediction & Sentiment Platform** — A Golang + MongoDB REST API powering fan predictions, community sentiment dashboards, crowdsourced context cards, and merit-based reputation.

---

## 🏗️ Architecture Overview

```
fanfare-backend/
├── cmd/api/main.go                    # Server entrypoint
├── internal/
│   ├── config/config.go               # Environment config loader
│   ├── db/
│   │   ├── mongo.go                   # MongoDB connection + auto-index creation
│   │   └── seeder.go                  # First-run sample data seeder
│   ├── models/models.go               # BSON/JSON structs & request/response types
│   ├── repository/
│   │   ├── interfaces.go              # Clean architecture interfaces
│   │   ├── errors.go                  # Shared repository errors
│   │   └── mongodb/                   # MongoDB implementations
│   │       ├── user_repo.go
│   │       ├── content_repo.go
│   │       ├── prediction_repo.go
│   │       └── context_repo.go
│   ├── services/
│   │   ├── auth_service.go            # Bcrypt + JWT auth
│   │   ├── auth_service_test.go       # Unit tests
│   │   ├── scoring_service.go         # Prediction resolution + point engine
│   │   ├── scoring_service_test.go    # Unit tests
│   │   ├── reputation_service.go      # Dynamic badges + leaderboards
│   │   └── reputation_service_test.go # Unit tests
│   ├── handlers/                      # HTTP route handlers
│   │   ├── auth_handler.go
│   │   ├── content_handler.go
│   │   ├── prediction_handler.go
│   │   ├── sentiment_handler.go
│   │   └── context_handler.go
│   └── middleware/
│       ├── auth_middleware.go          # JWT Bearer verification
│       ├── cors.go                    # Cross-origin support
│       └── logging.go                 # Request logging
├── docs/swagger.yaml                  # OpenAPI 3.0 specification
├── .env                               # Local environment config
├── go.mod / go.sum                    # Go module dependencies
└── README.md                          # This file
```

---

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+** installed
- **MongoDB** running locally or a **MongoDB Atlas** connection string

### 1. Clone & Configure

```bash
# Clone the repository
cd fanfare-backend

# Edit .env with your MongoDB connection string
# For MongoDB Atlas:
#   MONGODB_URI=mongodb+srv://user:pass@cluster.mongodb.net/fanfare
# For local MongoDB:
#   MONGODB_URI=mongodb://localhost:27017
```

### 2. Run the Server

```bash
go run ./cmd/api/main.go
```

The server starts on `http://localhost:8080` and automatically:
- Connects to MongoDB and creates all necessary indexes
- Seeds the database with sample universes, titles, moments, predictions, and polls

### 3. Run Tests

```bash
go test -v ./internal/services/...
```

Expected output: **5 PASS, 0 FAIL**

---

## 📡 API Endpoints

Full OpenAPI 3.0 spec: [`docs/swagger.yaml`](docs/swagger.yaml)

### Authentication (Public)
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `POST` | `/api/v1/auth/register` | No | Register a new account |
| `POST` | `/api/v1/auth/login` | No | Login and get JWT |

### Users & Profiles (JWT Required)
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `GET` | `/api/v1/users/me` | JWT | Get **your own** profile and stats |
| `GET` | `/api/v1/users/me/reputation` | JWT | Your full reputation point history |
| `GET` | `/api/v1/users/:username` | JWT | Get any user's public profile |
| `POST` | `/api/v1/users/follow-fandom` | JWT | Follow a universe |
| `GET` | `/api/v1/scoreboard` | JWT | Global/fandom leaderboard |

### Content Catalog (JWT Required)
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `GET` | `/api/v1/universes` | JWT | List all franchises |
| `GET` | `/api/v1/universes/:slug` | JWT | Universe with titles |
| `GET` | `/api/v1/titles/:slug` | JWT | Title details |

### News Feed & Debates (JWT Required)
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `GET` | `/api/v1/moments` | JWT | Paginated news feed |
| `GET` | `/api/v1/moments/:id` | JWT | Single moment |
| `GET` | `/api/v1/moments/:id/predictions` | JWT | Predictions attached to a moment |
| `GET` | `/api/v1/issues` | JWT | List debate hubs |
| `GET` | `/api/v1/issues/:slug` | JWT | Debate hub with timeline |
| `GET` | `/api/v1/issues/:slug/predictions` | JWT | Predictions within a debate hub |

### Predictions — Pillar 1 (JWT Required)
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `POST` | `/api/v1/predictions/:id/vote` | JWT | Place prediction vote |
| `POST` | `/api/v1/predictions/:id/resolve` | JWT | Resolve with outcome |

### Sentiment — Pillar 2 (JWT Required)
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `GET` | `/api/v1/sentiment/dashboards/:type/:id` | JWT | Aggregated metrics |
| `POST` | `/api/v1/sentiment/polls/:id/vote` | JWT | Vote in poll |

### Context Cards — Pillar 3 (JWT Required)
| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `POST` | `/api/v1/context` | JWT | Submit context card |
| `POST` | `/api/v1/context/:id/rate` | JWT | Rate helpful/biased |

---

## 🎯 Curl Walkthrough

### Register & Login

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"fanfan","email":"fan@example.com","password":"secret123"}'

# Login (save token)
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"fanfan","password":"secret123"}' | jq -r '.token')

echo "JWT: $TOKEN"
```

### Browse Content (requires token)

```bash
# List universes
curl http://localhost:8080/api/v1/universes \
  -H "Authorization: Bearer $TOKEN"

# Get MCU details with titles
curl http://localhost:8080/api/v1/universes/mcu \
  -H "Authorization: Bearer $TOKEN"

# News feed (page 1)
curl "http://localhost:8080/api/v1/moments?limit=10&offset=0" \
  -H "Authorization: Bearer $TOKEN"

# Debate hubs
curl http://localhost:8080/api/v1/issues \
  -H "Authorization: Bearer $TOKEN"
```

### Place a Prediction

```bash
# Get prediction IDs from moments
curl http://localhost:8080/api/v1/moments \
  -H "Authorization: Bearer $TOKEN"

# Vote on a prediction (replace PREDICTION_ID)
curl -X POST http://localhost:8080/api/v1/predictions/PREDICTION_ID/vote \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"selected_option":"Yes (Above $357M)","confidence":"high"}'
```

### View Your Profile

```bash
curl http://localhost:8080/api/v1/users/fanfan \
  -H "Authorization: Bearer $TOKEN"
```

---

## 🏆 Reputation & Badges System (Pillar 4)

Reputation is earned through **insight, not followers**:

| Action | Points |
|--------|--------|
| Correct prediction (medium confidence) | 100 pts |
| Correct prediction (high confidence) | 150 pts |
| Minority winner bonus (< 50% picked correctly) | Up to +100 pts |
| Submit context card | 5 pts |
| Context card rated "helpful" | 2 pts per rating |
| Vote in sentiment poll | 1 pt |

### Dynamic Badges

| Badge | Criteria |
|-------|----------|
| **Analyst** | ≥ 5 predictions with ≥ 60% accuracy |
| **Lore Master** | ≥ 3 context cards with ≥ 15 net helpful score |
| **Hype Leader** | ≥ 10 sentiment poll votes |
| **Fan** | Default badge for all users |

---

## 🗃️ MongoDB Collections

| Collection | Description | Unique Indexes |
|------------|-------------|----------------|
| `users` | User accounts & reputation | `username`, `email` |
| `universes` | Franchise IPs | `slug` |
| `titles` | Watchable units | `slug` |
| `moments` | News feed items | — |
| `issue_hubs` | Long-running debates | `slug` |
| `predictions` | Votable questions | — |
| `prediction_participations` | User votes | `{prediction_id, user_id}` |
| `sentiment_polls` | Opinion polls | — |
| `sentiment_votes` | Poll votes with cohorts | `{sentiment_poll_id, user_id}` |
| `context_cards` | Community fact-checks | — |
| `context_card_ratings` | Card peer reviews | `{context_card_id, user_id}` |
| `reputation_history` | Points audit trail | — |

---

## 🔧 Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `MONGODB_URI` | `mongodb://localhost:27017` | MongoDB connection string |
| `MONGODB_DB_NAME` | `fanfare` | Database name |
| `JWT_SECRET` | (insecure fallback) | JWT signing secret key |

---

## 🧪 Test Coverage

| Test Suite | Tests | Status |
|------------|-------|--------|
| Auth Service | Registration, duplicate checks, JWT signing/validation | ✅ PASS |
| Scoring Engine | Confidence multipliers, minority bonuses, point allocation | ✅ PASS |
| Reputation | Dynamic badge evaluation (Analyst, Lore Master, Hype Leader) | ✅ PASS |

Run all: `go test -v ./internal/services/...`

---

## 🤝 Multi-Developer Setup

This backend is designed for a **2-person team** (Go + TypeScript):

1. **MongoDB Atlas** — Shared cloud database both services connect to
2. **MongoDB Compass** — Desktop GUI for visual inspection on both machines
3. **Same `.env`** — Both developers use the same `MONGODB_URI` pointing to Atlas

The shared collection schemas and compound indexes ensure data consistency regardless of which backend writes to the database.

---

## 📄 License

MIT License — Built with ❤️ for pop culture fans everywhere.
