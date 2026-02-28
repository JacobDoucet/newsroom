# Architecture

## Components
- **Planner**: advances world state (phase/severity/known facts/unknowns/threads)
- **Generator (Writers)**: produces multiple candidate articles from world state + region state
- **Editor**: human-in-loop for v1; later an AI model trained on preferences
- **Validators**: tone linter + canon consistency checks
- **Publisher**: creates immutable published articles

## Data stores
- Postgres: arcs, world snapshots, region state, packets, candidates, review actions, edit diffs, published articles
- Local filesystem: media assets in `./data/media`

## Views
- Public: Latest feed, Article page, Story hub (summary)
- Console: generate packets, review/rank/reject/edit/publish, inspect validator warnings