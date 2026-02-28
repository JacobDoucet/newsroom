# Fictional Newsroom Simulator

A serialized speculative fiction project presented as a BBC/Reuters-style newsroom.

This project generates and publishes fictional news articles about a large-scale fictional global event (e.g., alien signal, outbreak, anomaly). It is **not real news** and is **not intended to mislead**. The goal is immersive storytelling in a news format, with global canon + region-specific editions.

## What this is
- A “newsroom simulation engine” that advances a global world state over time
- A private **Newsroom Console** (Editor-in-Chief workflow) to generate candidate stories, review, edit, and publish
- A public **read-only site** to browse published articles

## What this is not
- Not real journalism
- Not intended to trick or deceive
- Not a political satire of real current events

## Quick Start

### Prerequisites
- Docker and Docker Compose
- (Optional) Local LLM server (LM Studio or llama.cpp) running on `http://localhost:1234/v1`

### Running the Application

1. Start all services:
```bash
make dev
```

This will start:
- **Database** (Postgres) on `localhost:5432`
- **API** (Go) on `http://localhost:8080`
- **Frontend** (React/Vite) on `http://localhost:5173`

2. Access the application:
- **Public Site**: http://localhost:5173/
- **Editorial Console**: http://localhost:5173/console
- **API Health**: http://localhost:8080/health

### Basic Workflow

1. **Open the Console**: Navigate to http://localhost:5173/console
2. **Create an Arc**: Click "Create Arc" and fill in the details
3. **Start the Arc**: Click "Start Arc" to initialize the world state
4. **Generate Candidates**: Click "Generate Draft Packet" to create article candidates (requires LLM)
5. **Review Candidates**: Review generated articles, select/reject them
6. **Publish**: Click "Publish" on a candidate to make it public
7. **View Public Site**: Check http://localhost:5173/ to see published articles

### Configuration

Copy `.env.example` to `.env` and adjust settings:

```bash
# API
PORT=8080
DATABASE_URL=postgres://newsroom:newsroom@db:5432/newsroom?sslmode=disable

# LLM (local host server)
LLM_BASE_URL=http://host.docker.internal:1234/v1
LLM_API_KEY=local-dev
LLM_MODEL=local-model-name

# Frontend
VITE_API_BASE=http://localhost:8080/api
```

### Using Without Local LLM

If you don't have a local LLM running, the packet generation will fail. You have two options:

1. Set up LM Studio or llama.cpp on your host machine
2. Mock the LLM responses by modifying the `generateCandidates` function in `backend/internal/handlers/handlers.go`

### Development Commands

```bash
# Start all services
make dev

# View logs
make logs

# Stop all services
make down

# Clean up (remove volumes)
make clean
```

## Architecture

### Components
- **Planner**: advances world state (phase/severity/known facts/unknowns/threads)
- **Generator (Writers)**: produces multiple candidate articles from world state + region state
- **Editor**: human-in-loop for v1; later an AI model trained on preferences
- **Validators**: tone linter + canon consistency checks
- **Publisher**: creates immutable published articles

### Data stores
- Postgres: arcs, world snapshots, region state, packets, candidates, review actions, edit diffs, published articles
- Local filesystem: media assets in `./data/media`

### Tech Stack
- **Backend**: Go 1.22, chi router, pgx (Postgres driver)
- **Frontend**: React 18, TypeScript, Vite, Material-UI
- **Database**: PostgreSQL 16
- **LLM Integration**: OpenAI-compatible API (local LM Studio/llama.cpp)

## Project Structure

```
.
├── backend/              # Go API server
│   ├── cmd/api/         # Main application entry
│   ├── internal/        # Internal packages
│   │   ├── db/         # Database connection & migrations
│   │   ├── handlers/   # HTTP handlers
│   │   ├── llm/        # LLM client
│   │   └── models/     # Data models
│   └── migrations/     # SQL migration files
├── frontend/            # React/Vite application
│   └── src/
│       ├── pages/      # Console & Public site
│       ├── lib/        # API client
│       └── components/ # Reusable components
├── docs/               # Documentation
├── docker-compose.yml  # Docker services config
└── Makefile           # Development commands
```

## Database Schema

Key tables:
- `arcs`: Story arcs/seasons
- `world_state_snapshots`: Daily world state progression
- `region_state`: Regional perspectives and canon
- `draft_packets`: Generation batches
- `draft_candidates`: Generated article candidates
- `articles`: Published articles
- `review_actions`: Editorial decisions (for AI training)
- `candidate_rankings`: Preference signals
- `edit_diffs`: Before/after changes
- `media_assets`: Images and media

## API Endpoints

### Arc Management
- `POST /api/arcs` - Create new arc
- `GET /api/arcs` - List all arcs
- `POST /api/arcs/:id/start` - Initialize arc
- `POST /api/arcs/:id/advance` - Advance to next day

### Packet Generation
- `POST /api/arcs/:id/packets/generate` - Generate article candidates
- `GET /api/packets/:id` - Get packet details
- `GET /api/packets/:id/candidates` - Get candidates

### Review Actions
- `POST /api/candidates/:id/select` - Select candidate
- `POST /api/candidates/:id/reject` - Reject candidate
- `POST /api/packets/:id/rank` - Rank candidates
- `POST /api/candidates/:id/edit` - Edit candidate
- `POST /api/candidates/:id/publish` - Publish article

### Public API
- `GET /api/public/latest` - Get latest articles
- `GET /api/public/article/:id` - Get single article

## Editorial Philosophy (Reuters-like)

- Restrained headlines
- Attribution and sourcing
- Avoid early certainty
- "Signal density" is controlled: most days are analysis and incremental updates
- Breaking news is rare

## Roadmap

- v1: Articles + images, human-in-the-loop editor workflow ✅
- v2: AI editor that ranks/rewrites to match preferences
- v3: Fully autonomous newsroom (planner → writers → editor → validators → publisher)

## Documentation

See `/docs` for detailed documentation:
- `ARCHITECTURE.md` - System architecture
- `WORKFLOW.md` - Editorial workflow
- `PRODUCT_VISION.md` - Product vision and goals
- `LOCAL_LLM.md` - Local LLM setup guide
- `STYLE_GUIDE.md` - Writing style guide
- `RUBRIC.md` - Editorial rubric

## License

This is a fictional storytelling project. All generated content is fiction.