-- Migration: 001_initial_schema.sql
-- Create core tables for newsroom simulator

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Arcs table: storyline containers
CREATE TABLE arcs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    slug TEXT UNIQUE NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    global_rules JSONB NOT NULL DEFAULT '{}',
    escalation_model JSONB NOT NULL DEFAULT '{}',
    status TEXT NOT NULL DEFAULT 'draft', -- draft, active, completed
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_arcs_slug ON arcs(slug);
CREATE INDEX idx_arcs_status ON arcs(status);

-- World state snapshots: daily progression of global state
CREATE TABLE world_state_snapshots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    arc_id UUID NOT NULL REFERENCES arcs(id) ON DELETE CASCADE,
    day_index INTEGER NOT NULL,
    global_state JSONB NOT NULL DEFAULT '{}',
    event_log JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(arc_id, day_index)
);

CREATE INDEX idx_world_state_arc_day ON world_state_snapshots(arc_id, day_index);

-- Region state: persistent regional canon
CREATE TABLE region_state (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    arc_id UUID NOT NULL REFERENCES arcs(id) ON DELETE CASCADE,
    region_key TEXT NOT NULL, -- e.g., GB, US, CN
    state JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(arc_id, region_key)
);

CREATE INDEX idx_region_state_arc_region ON region_state(arc_id, region_key);

-- Draft packets: generation batches
CREATE TABLE draft_packets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    arc_id UUID NOT NULL REFERENCES arcs(id) ON DELETE CASCADE,
    day_index INTEGER NOT NULL,
    region_key TEXT, -- NULL for global
    generation_config JSONB NOT NULL DEFAULT '{}',
    status TEXT NOT NULL DEFAULT 'pending', -- pending, generated, reviewed
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_draft_packets_arc_day ON draft_packets(arc_id, day_index);
CREATE INDEX idx_draft_packets_region ON draft_packets(region_key);

-- Draft candidates: generated article candidates
CREATE TABLE draft_candidates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    packet_id UUID NOT NULL REFERENCES draft_packets(id) ON DELETE CASCADE,
    structured JSONB NOT NULL, -- full article structure
    raw_text TEXT NOT NULL, -- raw LLM output
    validator_flags JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_draft_candidates_packet ON draft_candidates(packet_id);

-- Articles: published content
CREATE TABLE articles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    arc_id UUID NOT NULL REFERENCES arcs(id) ON DELETE CASCADE,
    candidate_id UUID REFERENCES draft_candidates(id),
    headline TEXT NOT NULL,
    subhead TEXT,
    byline TEXT NOT NULL DEFAULT 'Staff Reporter',
    dateline TEXT,
    body_md TEXT NOT NULL,
    tags TEXT[] NOT NULL DEFAULT '{}',
    status TEXT NOT NULL DEFAULT 'published', -- draft, published, archived
    canonical BOOLEAN NOT NULL DEFAULT true,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_articles_arc ON articles(arc_id);
CREATE INDEX idx_articles_status ON articles(status);
CREATE INDEX idx_articles_canonical ON articles(canonical) WHERE canonical = true;
CREATE INDEX idx_articles_published_at ON articles(published_at DESC) WHERE published_at IS NOT NULL;
CREATE INDEX idx_articles_tags ON articles USING GIN(tags);

-- Review actions: editorial decisions
CREATE TABLE review_actions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    candidate_id UUID NOT NULL REFERENCES draft_candidates(id) ON DELETE CASCADE,
    action TEXT NOT NULL, -- select, reject, edited, published
    reason_tags TEXT[] NOT NULL DEFAULT '{}',
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_review_actions_candidate ON review_actions(candidate_id);
CREATE INDEX idx_review_actions_action ON review_actions(action);

-- Candidate rankings: preference signals
CREATE TABLE candidate_rankings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    packet_id UUID NOT NULL REFERENCES draft_packets(id) ON DELETE CASCADE,
    ranked_candidate_ids UUID[] NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_candidate_rankings_packet ON candidate_rankings(packet_id);

-- Edit diffs: before/after changes
CREATE TABLE edit_diffs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    candidate_id UUID NOT NULL REFERENCES draft_candidates(id) ON DELETE CASCADE,
    before_json JSONB NOT NULL,
    after_json JSONB NOT NULL,
    diff_meta JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_edit_diffs_candidate ON edit_diffs(candidate_id);

-- Media assets: images/videos
CREATE TABLE media_assets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    article_id UUID NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    kind TEXT NOT NULL, -- image, video
    provider TEXT NOT NULL, -- llm_image, external
    path TEXT NOT NULL,
    prompt TEXT,
    meta JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_media_assets_article ON media_assets(article_id);
