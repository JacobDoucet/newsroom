package models

import (
	"time"

	"github.com/google/uuid"
)

type Arc struct {
	ID              uuid.UUID              `json:"id"`
	Slug            string                 `json:"slug"`
	Title           string                 `json:"title"`
	Description     *string                `json:"description,omitempty"`
	GlobalRules     map[string]interface{} `json:"global_rules"`
	EscalationModel map[string]interface{} `json:"escalation_model"`
	Status          string                 `json:"status"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

type WorldStateSnapshot struct {
	ID          uuid.UUID              `json:"id"`
	ArcID       uuid.UUID              `json:"arc_id"`
	DayIndex    int                    `json:"day_index"`
	GlobalState map[string]interface{} `json:"global_state"`
	EventLog    []interface{}          `json:"event_log"`
	CreatedAt   time.Time              `json:"created_at"`
}

type RegionState struct {
	ID        uuid.UUID              `json:"id"`
	ArcID     uuid.UUID              `json:"arc_id"`
	RegionKey string                 `json:"region_key"`
	State     map[string]interface{} `json:"state"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

type DraftPacket struct {
	ID               uuid.UUID              `json:"id"`
	ArcID            uuid.UUID              `json:"arc_id"`
	DayIndex         int                    `json:"day_index"`
	RegionKey        *string                `json:"region_key,omitempty"`
	GenerationConfig map[string]interface{} `json:"generation_config"`
	Status           string                 `json:"status"`
	CreatedAt        time.Time              `json:"created_at"`
}

type DraftCandidate struct {
	ID             uuid.UUID              `json:"id"`
	PacketID       uuid.UUID              `json:"packet_id"`
	Structured     map[string]interface{} `json:"structured"`
	RawText        string                 `json:"raw_text"`
	ValidatorFlags map[string]interface{} `json:"validator_flags"`
	CreatedAt      time.Time              `json:"created_at"`
}

type Article struct {
	ID          uuid.UUID  `json:"id"`
	ArcID       uuid.UUID  `json:"arc_id"`
	CandidateID *uuid.UUID `json:"candidate_id,omitempty"`
	Headline    string     `json:"headline"`
	Subhead     *string    `json:"subhead,omitempty"`
	Byline      string     `json:"byline"`
	Dateline    *string    `json:"dateline,omitempty"`
	BodyMd      string     `json:"body_md"`
	Tags        []string   `json:"tags"`
	Status      string     `json:"status"`
	Canonical   bool       `json:"canonical"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type ReviewAction struct {
	ID          uuid.UUID `json:"id"`
	CandidateID uuid.UUID `json:"candidate_id"`
	Action      string    `json:"action"`
	ReasonTags  []string  `json:"reason_tags"`
	Notes       *string   `json:"notes,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type CandidateRanking struct {
	ID                  uuid.UUID   `json:"id"`
	PacketID            uuid.UUID   `json:"packet_id"`
	RankedCandidateIDs  []uuid.UUID `json:"ranked_candidate_ids"`
	CreatedAt           time.Time   `json:"created_at"`
}

type EditDiff struct {
	ID          uuid.UUID              `json:"id"`
	CandidateID uuid.UUID              `json:"candidate_id"`
	BeforeJSON  map[string]interface{} `json:"before_json"`
	AfterJSON   map[string]interface{} `json:"after_json"`
	DiffMeta    map[string]interface{} `json:"diff_meta"`
	CreatedAt   time.Time              `json:"created_at"`
}

type MediaAsset struct {
	ID        uuid.UUID              `json:"id"`
	ArticleID uuid.UUID              `json:"article_id"`
	Kind      string                 `json:"kind"`
	Provider  string                 `json:"provider"`
	Path      string                 `json:"path"`
	Prompt    *string                `json:"prompt,omitempty"`
	Meta      map[string]interface{} `json:"meta"`
	CreatedAt time.Time              `json:"created_at"`
}

// Article structure from LLM
type ArticleStructure struct {
	Headline        string                   `json:"headline"`
	Subhead         string                   `json:"subhead"`
	Byline          string                   `json:"byline"`
	Dateline        string                   `json:"dateline"`
	StoryType       string                   `json:"story_type"`
	Tags            []string                 `json:"tags"`
	Lede            string                   `json:"lede"`
	BodyMd          string                   `json:"body_md"`
	WhatWeKnow      []string                 `json:"what_we_know"`
	WhatWeDontKnow  []string                 `json:"what_we_dont_know"`
	Sources         []ArticleSource          `json:"sources"`
	References      ArticleReferences        `json:"references"`
	RiskFlags       []string                 `json:"risk_flags"`
}

type ArticleSource struct {
	Type       string  `json:"type"`
	Name       string  `json:"name"`
	Confidence float64 `json:"confidence"`
}

type ArticleReferences struct {
	KnownFacts []string `json:"known_facts"`
	Threads    []string `json:"threads"`
}
