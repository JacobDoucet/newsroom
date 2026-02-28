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

## MVP (v1)
- Postgres (local in Docker)
- Go API (chi + pgxpool)
- React/Vite frontend (TypeScript)
- Console UI: MUI
- Public site: clean typography/layout
- Local LLM support via OpenAI-compatible endpoint (LM Studio or llama.cpp) on host Mac:
  - `LLM_BASE_URL=http://host.docker.internal:1234/v1`

  ## Core workflow (private)
  1. Start an Arc (season)
  2. Advance World State (daily)
  3. Generate a draft packet (global and/or region)
  4. Review candidates (rank/select/reject with reason tags)
  5. Edit chosen candidate
  6. Publish
  7. Log all actions for future preference training (AI Editor)

  See `/docs` for details.