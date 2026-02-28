# Bootstrap Verification Checklist

This document verifies that all requirements from the issue have been implemented.

## ✅ Docker Infrastructure

- [x] `docker-compose.yml` with 3 services: db, api, web
- [x] `Makefile` with `make dev` command
- [x] Backend `Dockerfile` (multi-stage build with Go 1.22)
- [x] Frontend `Dockerfile` (Node 20 with Vite dev server)
- [x] `.gitignore` configured for Go, Node, and Docker artifacts

## ✅ Backend (Go)

### Structure
- [x] `/backend/cmd/api/main.go` - Entry point
- [x] `/backend/internal/db/db.go` - Database connection & migrations
- [x] `/backend/internal/handlers/handlers.go` - HTTP handlers
- [x] `/backend/internal/llm/client.go` - LLM client
- [x] `/backend/internal/models/models.go` - Data models
- [x] `/backend/migrations/001_initial_schema.sql` - Database schema

### API Endpoints (All Implemented)
Arc Management:
- [x] `POST /api/arcs` - Create arc
- [x] `GET /api/arcs` - List arcs
- [x] `POST /api/arcs/:id/start` - Start arc (creates day 0 snapshot)
- [x] `POST /api/arcs/:id/advance` - Advance to next day

Region:
- [x] `POST /api/arcs/:id/regions/init` - Initialize region

Packets:
- [x] `POST /api/arcs/:id/packets/generate` - Generate candidates
- [x] `GET /api/packets/:id` - Get packet
- [x] `GET /api/packets/:id/candidates` - Get candidates

Review:
- [x] `POST /api/candidates/:id/select` - Select candidate
- [x] `POST /api/candidates/:id/reject` - Reject candidate
- [x] `POST /api/packets/:id/rank` - Rank candidates
- [x] `POST /api/candidates/:id/edit` - Edit candidate
- [x] `POST /api/candidates/:id/publish` - Publish article

Public:
- [x] `GET /api/public/latest` - Get latest articles
- [x] `GET /api/public/article/:id` - Get article

### LLM Client
- [x] OpenAI-compatible API client
- [x] Uses `LLM_BASE_URL`, `LLM_API_KEY`, `LLM_MODEL` env vars
- [x] Structured JSON output required
- [x] Retry once on JSON parse failure
- [x] Always stores raw LLM output

### Database
- [x] Automatic migration runner on startup
- [x] Uses pgx/v5 with connection pooling
- [x] All 10 tables implemented (see below)

## ✅ Database Schema

All required tables with UUID + JSONB:
- [x] `arcs` - slug unique, global_rules/escalation_model JSONB
- [x] `world_state_snapshots` - unique(arc_id, day_index), global_state/event_log JSONB
- [x] `region_state` - unique(arc_id, region_key), state JSONB
- [x] `draft_packets` - generation_config JSONB
- [x] `draft_candidates` - structured/validator_flags JSONB
- [x] `articles` - headline/body_md/tags[], canonical bool
- [x] `review_actions` - action, reason_tags[], notes
- [x] `candidate_rankings` - ranked_candidate_ids UUID[]
- [x] `edit_diffs` - before_json/after_json/diff_meta JSONB
- [x] `media_assets` - article_id, kind, path, meta JSONB

Indexes added:
- [x] arcs: slug, status
- [x] world_state_snapshots: (arc_id, day_index)
- [x] region_state: (arc_id, region_key)
- [x] draft_packets: (arc_id, day_index), region_key
- [x] draft_candidates: packet_id
- [x] articles: arc_id, status, canonical, published_at DESC, tags (GIN)
- [x] review_actions: candidate_id, action
- [x] candidate_rankings: packet_id
- [x] edit_diffs: candidate_id
- [x] media_assets: article_id

## ✅ Frontend (React/Vite)

### Structure
- [x] `/frontend/src/App.tsx` - Main app with routing
- [x] `/frontend/src/main.tsx` - Entry point
- [x] `/frontend/src/lib/api.ts` - API client
- [x] `/frontend/src/pages/PublicSite.tsx` - Public article browsing
- [x] `/frontend/src/pages/Console.tsx` - Editorial console
- [x] Material-UI (MUI) components used throughout

### Console Features (All Implemented)
- [x] Create/list/select arcs
- [x] Start arc (initialize world state)
- [x] Advance arc (next day)
- [x] Generate draft packet
- [x] View candidates with full structured data
- [x] Select/reject candidates with reason tags
- [x] Publish candidates directly
- [x] Polling for candidate generation completion

### Public Site Features
- [x] List latest published articles
- [x] Display headline, subhead, byline, dateline, body, tags
- [x] Fiction warning banner
- [x] Clean typography (Georgia serif font)

## ✅ Configuration

- [x] `.env.example` with all required variables
- [x] Docker Compose uses host.docker.internal for LLM access
- [x] CORS configured for localhost:5173
- [x] Postgres healthcheck in docker-compose.yml

## ✅ Documentation

- [x] README.md updated with:
  - Quick start instructions
  - Architecture overview
  - API endpoint documentation
  - Project structure
  - Development commands
  - LLM setup guidance
- [x] `/docs` directory preserved with existing docs

## ✅ Hard Requirements Met

- [x] Docker-first: `make dev` runs `docker compose up --build`
- [x] Services: db (postgres:16), api (Go), web (React/Vite)
- [x] Postgres schema uses UUID + JSONB
- [x] Migrations run automatically on API startup
- [x] LLM client is OpenAI-compatible
- [x] POST to `${LLM_BASE_URL}/chat/completions`
- [x] Uses `LLM_MODEL` env var
- [x] Structured JSON output with retry
- [x] Always stores raw output
- [x] Editorial logging implemented:
  - review_actions (select/reject/edited/published)
  - candidate_rankings (per packet)
  - edit_diffs (before/after JSON)

## ✅ Deliverables Met

### Dev Experience
- [x] `make dev` starts all services
- [x] Public site: http://localhost:5173/
- [x] Console: http://localhost:5173/console
- [x] API: http://localhost:8080/api
- [x] Postgres: localhost:5432

### Repo Structure
- [x] `/backend` - Go app
- [x] `/frontend` - React app
- [x] `/docs` - Preserved documentation
- [x] `docker-compose.yml` in root
- [x] `Makefile` in root
- [x] `.env.example` in root

## Code Quality

- [x] Go code compiles without errors
- [x] TypeScript configured with strict mode
- [x] No auth/users (as specified)
- [x] No video generation (as specified)
- [x] No message queues (as specified)
- [x] Minimal but clean code

## Testing Steps

To verify the implementation:

1. **Start services:**
   ```bash
   make dev
   ```

2. **Check services are running:**
   - Database: `docker ps | grep postgres`
   - API: `curl http://localhost:8080/health`
   - Frontend: Visit http://localhost:5173

3. **Test Console workflow:**
   - Go to http://localhost:5173/console
   - Create a new arc
   - Start the arc
   - Generate a packet (will fail without LLM, this is expected)

4. **Test Public site:**
   - Go to http://localhost:5173/
   - Should show empty state with fiction warning

5. **Test API directly:**
   ```bash
   # Create arc
   curl -X POST http://localhost:8080/api/arcs \
     -H "Content-Type: application/json" \
     -d '{"slug":"test","title":"Test Arc"}'

   # List arcs
   curl http://localhost:8080/api/arcs
   ```

## Known Limitations (Expected)

- LLM generation requires a local LLM server (LM Studio or llama.cpp) running on host at port 1234
- Without LLM, packet generation will create packets but candidates won't be generated
- Region filtering in public API is placeholder (marked with comment)
- No authentication (by design)

## Summary

All requirements from the issue have been successfully implemented:
- ✅ Docker-first MVP with Go + Postgres + React
- ✅ Complete database schema with migrations
- ✅ All API endpoints functional
- ✅ Console UI with full editorial workflow
- ✅ Public site for browsing articles
- ✅ OpenAI-compatible LLM integration
- ✅ Editorial logging for AI training
- ✅ Comprehensive documentation

The project is ready for testing and development!
