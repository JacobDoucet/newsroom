package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/JacobDoucet/newsroom/internal/llm"
	"github.com/JacobDoucet/newsroom/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db        *pgxpool.Pool
	llmClient *llm.Client
}

func New(db *pgxpool.Pool) *Handler {
	return &Handler{
		db:        db,
		llmClient: llm.NewClient(),
	}
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// CreateArc creates a new story arc
func (h *Handler) CreateArc(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Slug            string                 `json:"slug"`
		Title           string                 `json:"title"`
		Description     string                 `json:"description"`
		GlobalRules     map[string]interface{} `json:"global_rules"`
		EscalationModel map[string]interface{} `json:"escalation_model"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Slug == "" || req.Title == "" {
		respondError(w, http.StatusBadRequest, "Slug and title are required")
		return
	}

	// Set defaults
	if req.GlobalRules == nil {
		req.GlobalRules = map[string]interface{}{}
	}
	if req.EscalationModel == nil {
		req.EscalationModel = map[string]interface{}{}
	}

	var arc models.Arc
	err := h.db.QueryRow(r.Context(), `
		INSERT INTO arcs (slug, title, description, global_rules, escalation_model, status)
		VALUES ($1, $2, $3, $4, $5, 'draft')
		RETURNING id, slug, title, description, global_rules, escalation_model, status, created_at, updated_at
	`, req.Slug, req.Title, req.Description, req.GlobalRules, req.EscalationModel).Scan(
		&arc.ID, &arc.Slug, &arc.Title, &arc.Description, &arc.GlobalRules,
		&arc.EscalationModel, &arc.Status, &arc.CreatedAt, &arc.UpdatedAt,
	)

	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create arc: %v", err))
		return
	}

	respondJSON(w, http.StatusCreated, arc)
}

// ListArcs lists all arcs
func (h *Handler) ListArcs(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(r.Context(), `
		SELECT id, slug, title, description, global_rules, escalation_model, status, created_at, updated_at
		FROM arcs
		ORDER BY created_at DESC
	`)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list arcs: %v", err))
		return
	}
	defer rows.Close()

	arcs := []models.Arc{}
	for rows.Next() {
		var arc models.Arc
		if err := rows.Scan(&arc.ID, &arc.Slug, &arc.Title, &arc.Description,
			&arc.GlobalRules, &arc.EscalationModel, &arc.Status, &arc.CreatedAt, &arc.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to scan arc: %v", err))
			return
		}
		arcs = append(arcs, arc)
	}

	respondJSON(w, http.StatusOK, arcs)
}

// StartArc activates an arc
func (h *Handler) StartArc(w http.ResponseWriter, r *http.Request) {
	arcID := chi.URLParam(r, "id")
	id, err := uuid.Parse(arcID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid arc ID")
		return
	}

	// Update arc status
	_, err = h.db.Exec(r.Context(), `
		UPDATE arcs SET status = 'active', updated_at = NOW()
		WHERE id = $1
	`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to start arc: %v", err))
		return
	}

	// Create initial world state snapshot (day 0)
	initialState := map[string]interface{}{
		"phase":        "initial",
		"severity":     1,
		"known_facts":  []string{},
		"unknowns":     []string{},
		"threads":      []string{},
		"global_focus": "detection",
	}

	var snapshot models.WorldStateSnapshot
	err = h.db.QueryRow(r.Context(), `
		INSERT INTO world_state_snapshots (arc_id, day_index, global_state, event_log)
		VALUES ($1, 0, $2, '[]'::jsonb)
		RETURNING id, arc_id, day_index, global_state, event_log, created_at
	`, id, initialState).Scan(
		&snapshot.ID, &snapshot.ArcID, &snapshot.DayIndex,
		&snapshot.GlobalState, &snapshot.EventLog, &snapshot.CreatedAt,
	)

	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create initial snapshot: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, snapshot)
}

// AdvanceArc creates the next day's world state snapshot
func (h *Handler) AdvanceArc(w http.ResponseWriter, r *http.Request) {
	arcID := chi.URLParam(r, "id")
	id, err := uuid.Parse(arcID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid arc ID")
		return
	}

	// Get the latest snapshot
	var latestDay int
	err = h.db.QueryRow(r.Context(), `
		SELECT COALESCE(MAX(day_index), -1)
		FROM world_state_snapshots
		WHERE arc_id = $1
	`, id).Scan(&latestDay)

	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get latest day: %v", err))
		return
	}

	// Create next day's snapshot (simplified - in real version would use LLM to evolve state)
	nextDay := latestDay + 1
	newState := map[string]interface{}{
		"phase":        "unfolding",
		"severity":     nextDay + 1,
		"known_facts":  []string{fmt.Sprintf("KF%d", nextDay)},
		"unknowns":     []string{fmt.Sprintf("UNK%d", nextDay)},
		"threads":      []string{fmt.Sprintf("T%d", nextDay)},
		"global_focus": "investigation",
		"day":          nextDay,
	}

	var snapshot models.WorldStateSnapshot
	err = h.db.QueryRow(r.Context(), `
		INSERT INTO world_state_snapshots (arc_id, day_index, global_state, event_log)
		VALUES ($1, $2, $3, '[]'::jsonb)
		RETURNING id, arc_id, day_index, global_state, event_log, created_at
	`, id, nextDay, newState).Scan(
		&snapshot.ID, &snapshot.ArcID, &snapshot.DayIndex,
		&snapshot.GlobalState, &snapshot.EventLog, &snapshot.CreatedAt,
	)

	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create snapshot: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, snapshot)
}

// InitRegion initializes a new region
func (h *Handler) InitRegion(w http.ResponseWriter, r *http.Request) {
	arcID := chi.URLParam(r, "id")
	id, err := uuid.Parse(arcID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid arc ID")
		return
	}

	var req struct {
		RegionKey string `json:"region_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.RegionKey == "" {
		respondError(w, http.StatusBadRequest, "region_key is required")
		return
	}

	// Initialize region state
	regionState := map[string]interface{}{
		"stance":    "cautious",
		"modifiers": []string{"official_sources_only"},
		"focus":     []string{"local_impact", "expert_analysis"},
	}

	var region models.RegionState
	err = h.db.QueryRow(r.Context(), `
		INSERT INTO region_state (arc_id, region_key, state)
		VALUES ($1, $2, $3)
		ON CONFLICT (arc_id, region_key) DO UPDATE SET state = $3, updated_at = NOW()
		RETURNING id, arc_id, region_key, state, created_at, updated_at
	`, id, req.RegionKey, regionState).Scan(
		&region.ID, &region.ArcID, &region.RegionKey,
		&region.State, &region.CreatedAt, &region.UpdatedAt,
	)

	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to init region: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, region)
}

// GeneratePacket creates a draft packet with LLM-generated candidates
func (h *Handler) GeneratePacket(w http.ResponseWriter, r *http.Request) {
	arcID := chi.URLParam(r, "id")
	id, err := uuid.Parse(arcID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid arc ID")
		return
	}

	var req struct {
		DayIndex  int     `json:"day_index"`
		RegionKey *string `json:"region_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Create packet
	genConfig := map[string]interface{}{
		"num_candidates": 3,
		"style":          "reuters",
	}

	var packet models.DraftPacket
	err = h.db.QueryRow(r.Context(), `
		INSERT INTO draft_packets (arc_id, day_index, region_key, generation_config, status)
		VALUES ($1, $2, $3, $4, 'pending')
		RETURNING id, arc_id, day_index, region_key, generation_config, status, created_at
	`, id, req.DayIndex, req.RegionKey, genConfig).Scan(
		&packet.ID, &packet.ArcID, &packet.DayIndex, &packet.RegionKey,
		&packet.GenerationConfig, &packet.Status, &packet.CreatedAt,
	)

	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create packet: %v", err))
		return
	}

	// Generate candidates using LLM
	go h.generateCandidates(context.Background(), packet.ID, id, req.DayIndex, req.RegionKey)

	respondJSON(w, http.StatusAccepted, packet)
}

// generateCandidates runs asynchronously to generate article candidates
func (h *Handler) generateCandidates(ctx context.Context, packetID, arcID uuid.UUID, dayIndex int, regionKey *string) {
	systemPrompt := `You are a professional news writer for a Reuters-style news organization. Generate a fictional news article in JSON format with the exact structure specified. The article should be restrained, well-sourced, and acknowledge uncertainty.`

	userPrompt := fmt.Sprintf(`Generate a fictional news article for day %d. Return ONLY valid JSON with this exact structure:
{
  "headline": "Restrained, factual headline",
  "subhead": "Additional context",
  "byline": "Staff Reporter",
  "dateline": "LONDON",
  "story_type": "update",
  "tags": ["World", "Science"],
  "lede": "Opening paragraph with key facts",
  "body_md": "Full article body in markdown with multiple paragraphs",
  "what_we_know": ["Known fact 1", "Known fact 2"],
  "what_we_dont_know": ["Unknown 1", "Unknown 2"],
  "sources": [{"type":"official","name":"Government spokesperson","confidence":0.8}],
  "references": {"known_facts":["KF1"],"threads":["T1"]},
  "risk_flags": []
}`, dayIndex)

	// Generate 3 candidates
	for i := 0; i < 3; i++ {
		structured, rawText, err := h.llmClient.GenerateStructured(ctx, systemPrompt, userPrompt)
		if err != nil {
			fmt.Printf("Failed to generate candidate %d: %v\n", i, err)
			continue
		}

		// Store candidate
		_, err = h.db.Exec(ctx, `
			INSERT INTO draft_candidates (packet_id, structured, raw_text, validator_flags)
			VALUES ($1, $2, $3, '{}'::jsonb)
		`, packetID, structured, rawText)

		if err != nil {
			fmt.Printf("Failed to store candidate %d: %v\n", i, err)
		}
	}

	// Update packet status
	h.db.Exec(ctx, `
		UPDATE draft_packets SET status = 'generated'
		WHERE id = $1
	`, packetID)
}

// GetPacket retrieves a draft packet
func (h *Handler) GetPacket(w http.ResponseWriter, r *http.Request) {
	packetID := chi.URLParam(r, "id")
	id, err := uuid.Parse(packetID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid packet ID")
		return
	}

	var packet models.DraftPacket
	err = h.db.QueryRow(r.Context(), `
		SELECT id, arc_id, day_index, region_key, generation_config, status, created_at
		FROM draft_packets
		WHERE id = $1
	`, id).Scan(&packet.ID, &packet.ArcID, &packet.DayIndex, &packet.RegionKey,
		&packet.GenerationConfig, &packet.Status, &packet.CreatedAt)

	if err == pgx.ErrNoRows {
		respondError(w, http.StatusNotFound, "Packet not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get packet: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, packet)
}

// GetPacketCandidates retrieves all candidates for a packet
func (h *Handler) GetPacketCandidates(w http.ResponseWriter, r *http.Request) {
	packetID := chi.URLParam(r, "id")
	id, err := uuid.Parse(packetID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid packet ID")
		return
	}

	rows, err := h.db.Query(r.Context(), `
		SELECT id, packet_id, structured, raw_text, validator_flags, created_at
		FROM draft_candidates
		WHERE packet_id = $1
		ORDER BY created_at ASC
	`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get candidates: %v", err))
		return
	}
	defer rows.Close()

	candidates := []models.DraftCandidate{}
	for rows.Next() {
		var c models.DraftCandidate
		if err := rows.Scan(&c.ID, &c.PacketID, &c.Structured, &c.RawText,
			&c.ValidatorFlags, &c.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to scan candidate: %v", err))
			return
		}
		candidates = append(candidates, c)
	}

	respondJSON(w, http.StatusOK, candidates)
}

// SelectCandidate marks a candidate as selected
func (h *Handler) SelectCandidate(w http.ResponseWriter, r *http.Request) {
	candidateID := chi.URLParam(r, "id")
	id, err := uuid.Parse(candidateID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid candidate ID")
		return
	}

	var req struct {
		ReasonTags []string `json:"reason_tags"`
		Notes      string   `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var action models.ReviewAction
	err = h.db.QueryRow(r.Context(), `
		INSERT INTO review_actions (candidate_id, action, reason_tags, notes)
		VALUES ($1, 'select', $2, $3)
		RETURNING id, candidate_id, action, reason_tags, notes, created_at
	`, id, req.ReasonTags, req.Notes).Scan(
		&action.ID, &action.CandidateID, &action.Action,
		&action.ReasonTags, &action.Notes, &action.CreatedAt,
	)

	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to record action: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, action)
}

// RejectCandidate marks a candidate as rejected
func (h *Handler) RejectCandidate(w http.ResponseWriter, r *http.Request) {
	candidateID := chi.URLParam(r, "id")
	id, err := uuid.Parse(candidateID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid candidate ID")
		return
	}

	var req struct {
		ReasonTags []string `json:"reason_tags"`
		Notes      string   `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var action models.ReviewAction
	err = h.db.QueryRow(r.Context(), `
		INSERT INTO review_actions (candidate_id, action, reason_tags, notes)
		VALUES ($1, 'reject', $2, $3)
		RETURNING id, candidate_id, action, reason_tags, notes, created_at
	`, id, req.ReasonTags, req.Notes).Scan(
		&action.ID, &action.CandidateID, &action.Action,
		&action.ReasonTags, &action.Notes, &action.CreatedAt,
	)

	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to record action: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, action)
}

// RankCandidates records a ranking of candidates
func (h *Handler) RankCandidates(w http.ResponseWriter, r *http.Request) {
	packetID := chi.URLParam(r, "id")
	id, err := uuid.Parse(packetID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid packet ID")
		return
	}

	var req struct {
		RankedCandidateIDs []uuid.UUID `json:"ranked_candidate_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var ranking models.CandidateRanking
	err = h.db.QueryRow(r.Context(), `
		INSERT INTO candidate_rankings (packet_id, ranked_candidate_ids)
		VALUES ($1, $2)
		RETURNING id, packet_id, ranked_candidate_ids, created_at
	`, id, req.RankedCandidateIDs).Scan(
		&ranking.ID, &ranking.PacketID, &ranking.RankedCandidateIDs, &ranking.CreatedAt,
	)

	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to record ranking: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, ranking)
}

// EditCandidate records an edit to a candidate
func (h *Handler) EditCandidate(w http.ResponseWriter, r *http.Request) {
	candidateID := chi.URLParam(r, "id")
	id, err := uuid.Parse(candidateID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid candidate ID")
		return
	}

	var req struct {
		After map[string]interface{} `json:"after"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get current candidate
	var beforeJSON map[string]interface{}
	err = h.db.QueryRow(r.Context(), `
		SELECT structured FROM draft_candidates WHERE id = $1
	`, id).Scan(&beforeJSON)

	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get candidate: %v", err))
		return
	}

	// Record edit diff
	diffMeta := map[string]interface{}{
		"editor": "human",
	}

	var diff models.EditDiff
	err = h.db.QueryRow(r.Context(), `
		INSERT INTO edit_diffs (candidate_id, before_json, after_json, diff_meta)
		VALUES ($1, $2, $3, $4)
		RETURNING id, candidate_id, before_json, after_json, diff_meta, created_at
	`, id, beforeJSON, req.After, diffMeta).Scan(
		&diff.ID, &diff.CandidateID, &diff.BeforeJSON,
		&diff.AfterJSON, &diff.DiffMeta, &diff.CreatedAt,
	)

	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to record edit: %v", err))
		return
	}

	// Update candidate
	_, err = h.db.Exec(r.Context(), `
		UPDATE draft_candidates SET structured = $1 WHERE id = $2
	`, req.After, id)

	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update candidate: %v", err))
		return
	}

	// Record edit action
	h.db.Exec(r.Context(), `
		INSERT INTO review_actions (candidate_id, action, reason_tags, notes)
		VALUES ($1, 'edited', '{}', '')
	`, id)

	respondJSON(w, http.StatusOK, diff)
}

// PublishCandidate creates a published article from a candidate
func (h *Handler) PublishCandidate(w http.ResponseWriter, r *http.Request) {
	candidateID := chi.URLParam(r, "id")
	id, err := uuid.Parse(candidateID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid candidate ID")
		return
	}

	// Get candidate and packet info
	var candidate models.DraftCandidate
	var arcID uuid.UUID
	err = h.db.QueryRow(r.Context(), `
		SELECT c.id, c.packet_id, c.structured, c.raw_text, c.validator_flags, c.created_at, p.arc_id
		FROM draft_candidates c
		JOIN draft_packets p ON c.packet_id = p.id
		WHERE c.id = $1
	`, id).Scan(&candidate.ID, &candidate.PacketID, &candidate.Structured,
		&candidate.RawText, &candidate.ValidatorFlags, &candidate.CreatedAt, &arcID)

	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get candidate: %v", err))
		return
	}

	// Extract fields from structured JSON
	headline := getString(candidate.Structured, "headline", "Untitled")
	subhead := getString(candidate.Structured, "subhead", "")
	byline := getString(candidate.Structured, "byline", "Staff Reporter")
	dateline := getString(candidate.Structured, "dateline", "")
	bodyMd := getString(candidate.Structured, "body_md", "")
	tags := getStringArray(candidate.Structured, "tags")

	now := time.Now()
	var article models.Article
	err = h.db.QueryRow(r.Context(), `
		INSERT INTO articles (arc_id, candidate_id, headline, subhead, byline, dateline, body_md, tags, status, canonical, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'published', true, $9)
		RETURNING id, arc_id, candidate_id, headline, subhead, byline, dateline, body_md, tags, status, canonical, published_at, created_at, updated_at
	`, arcID, id, headline, subhead, byline, dateline, bodyMd, tags, now).Scan(
		&article.ID, &article.ArcID, &article.CandidateID, &article.Headline,
		&article.Subhead, &article.Byline, &article.Dateline, &article.BodyMd,
		&article.Tags, &article.Status, &article.Canonical, &article.PublishedAt,
		&article.CreatedAt, &article.UpdatedAt,
	)

	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to publish article: %v", err))
		return
	}

	// Record publish action
	h.db.Exec(r.Context(), `
		INSERT INTO review_actions (candidate_id, action, reason_tags, notes)
		VALUES ($1, 'published', '{}', '')
	`, id)

	// Update packet status
	h.db.Exec(r.Context(), `
		UPDATE draft_packets SET status = 'reviewed' WHERE id = $1
	`, candidate.PacketID)

	respondJSON(w, http.StatusOK, article)
}

// GetLatestArticles returns published articles
func (h *Handler) GetLatestArticles(w http.ResponseWriter, r *http.Request) {
	_ = r.URL.Query().Get("region") // region filtering not yet implemented
	limit := 20

	query := `
		SELECT id, arc_id, candidate_id, headline, subhead, byline, dateline, body_md, tags, status, canonical, published_at, created_at, updated_at
		FROM articles
		WHERE status = 'published' AND canonical = true
		ORDER BY published_at DESC
		LIMIT $1
	`

	rows, err := h.db.Query(r.Context(), query, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get articles: %v", err))
		return
	}
	defer rows.Close()

	articles := []models.Article{}
	for rows.Next() {
		var a models.Article
		if err := rows.Scan(&a.ID, &a.ArcID, &a.CandidateID, &a.Headline, &a.Subhead,
			&a.Byline, &a.Dateline, &a.BodyMd, &a.Tags, &a.Status, &a.Canonical,
			&a.PublishedAt, &a.CreatedAt, &a.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to scan article: %v", err))
			return
		}
		articles = append(articles, a)
	}

	respondJSON(w, http.StatusOK, articles)
}

// GetArticle returns a single article
func (h *Handler) GetArticle(w http.ResponseWriter, r *http.Request) {
	articleID := chi.URLParam(r, "id")
	id, err := uuid.Parse(articleID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid article ID")
		return
	}

	var article models.Article
	err = h.db.QueryRow(r.Context(), `
		SELECT id, arc_id, candidate_id, headline, subhead, byline, dateline, body_md, tags, status, canonical, published_at, created_at, updated_at
		FROM articles
		WHERE id = $1
	`, id).Scan(&article.ID, &article.ArcID, &article.CandidateID, &article.Headline,
		&article.Subhead, &article.Byline, &article.Dateline, &article.BodyMd,
		&article.Tags, &article.Status, &article.Canonical, &article.PublishedAt,
		&article.CreatedAt, &article.UpdatedAt)

	if err == pgx.ErrNoRows {
		respondError(w, http.StatusNotFound, "Article not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get article: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, article)
}

// Helper functions
func getString(m map[string]interface{}, key, defaultVal string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return defaultVal
}

func getStringArray(m map[string]interface{}, key string) []string {
	if val, ok := m[key].([]interface{}); ok {
		result := make([]string, 0, len(val))
		for _, v := range val {
			if s, ok := v.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return []string{}
}
