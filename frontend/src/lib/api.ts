const API_BASE = import.meta.env.VITE_API_BASE || 'http://localhost:8080/api'

export interface Arc {
  id: string
  slug: string
  title: string
  description?: string
  global_rules: Record<string, unknown>
  escalation_model: Record<string, unknown>
  status: string
  created_at: string
  updated_at: string
}

export interface WorldStateSnapshot {
  id: string
  arc_id: string
  day_index: number
  global_state: Record<string, unknown>
  event_log: unknown[]
  created_at: string
}

export interface DraftPacket {
  id: string
  arc_id: string
  day_index: number
  region_key?: string
  generation_config: Record<string, unknown>
  status: string
  created_at: string
}

export interface DraftCandidate {
  id: string
  packet_id: string
  structured: Record<string, unknown>
  raw_text: string
  validator_flags: Record<string, unknown>
  created_at: string
}

export interface Article {
  id: string
  arc_id: string
  candidate_id?: string
  headline: string
  subhead?: string
  byline: string
  dateline?: string
  body_md: string
  tags: string[]
  status: string
  canonical: boolean
  published_at?: string
  created_at: string
  updated_at: string
}

// API client
export const api = {
  // Arcs
  async createArc(data: { slug: string; title: string; description?: string }): Promise<Arc> {
    const res = await fetch(`${API_BASE}/arcs`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    })
    if (!res.ok) throw new Error('Failed to create arc')
    return res.json()
  },

  async listArcs(): Promise<Arc[]> {
    const res = await fetch(`${API_BASE}/arcs`)
    if (!res.ok) throw new Error('Failed to list arcs')
    return res.json()
  },

  async startArc(id: string): Promise<WorldStateSnapshot> {
    const res = await fetch(`${API_BASE}/arcs/${id}/start`, { method: 'POST' })
    if (!res.ok) throw new Error('Failed to start arc')
    return res.json()
  },

  async advanceArc(id: string): Promise<WorldStateSnapshot> {
    const res = await fetch(`${API_BASE}/arcs/${id}/advance`, { method: 'POST' })
    if (!res.ok) throw new Error('Failed to advance arc')
    return res.json()
  },

  // Regions
  async initRegion(arcId: string, regionKey: string): Promise<unknown> {
    const res = await fetch(`${API_BASE}/arcs/${arcId}/regions/init`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ region_key: regionKey }),
    })
    if (!res.ok) throw new Error('Failed to init region')
    return res.json()
  },

  // Packets
  async generatePacket(arcId: string, dayIndex: number, regionKey?: string): Promise<DraftPacket> {
    const res = await fetch(`${API_BASE}/arcs/${arcId}/packets/generate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ day_index: dayIndex, region_key: regionKey }),
    })
    if (!res.ok) throw new Error('Failed to generate packet')
    return res.json()
  },

  async getPacket(id: string): Promise<DraftPacket> {
    const res = await fetch(`${API_BASE}/packets/${id}`)
    if (!res.ok) throw new Error('Failed to get packet')
    return res.json()
  },

  async getPacketCandidates(id: string): Promise<DraftCandidate[]> {
    const res = await fetch(`${API_BASE}/packets/${id}/candidates`)
    if (!res.ok) throw new Error('Failed to get candidates')
    return res.json()
  },

  // Review actions
  async selectCandidate(id: string, data: { reason_tags: string[]; notes: string }): Promise<unknown> {
    const res = await fetch(`${API_BASE}/candidates/${id}/select`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    })
    if (!res.ok) throw new Error('Failed to select candidate')
    return res.json()
  },

  async rejectCandidate(id: string, data: { reason_tags: string[]; notes: string }): Promise<unknown> {
    const res = await fetch(`${API_BASE}/candidates/${id}/reject`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    })
    if (!res.ok) throw new Error('Failed to reject candidate')
    return res.json()
  },

  async rankCandidates(packetId: string, rankedIds: string[]): Promise<unknown> {
    const res = await fetch(`${API_BASE}/packets/${packetId}/rank`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ ranked_candidate_ids: rankedIds }),
    })
    if (!res.ok) throw new Error('Failed to rank candidates')
    return res.json()
  },

  async editCandidate(id: string, after: Record<string, unknown>): Promise<unknown> {
    const res = await fetch(`${API_BASE}/candidates/${id}/edit`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ after }),
    })
    if (!res.ok) throw new Error('Failed to edit candidate')
    return res.json()
  },

  async publishCandidate(id: string): Promise<Article> {
    const res = await fetch(`${API_BASE}/candidates/${id}/publish`, { method: 'POST' })
    if (!res.ok) throw new Error('Failed to publish candidate')
    return res.json()
  },

  // Public
  async getLatestArticles(region?: string, limit = 20): Promise<Article[]> {
    const params = new URLSearchParams()
    if (region) params.append('region', region)
    params.append('limit', String(limit))

    const res = await fetch(`${API_BASE}/public/latest?${params}`)
    if (!res.ok) throw new Error('Failed to get articles')
    return res.json()
  },

  async getArticle(id: string): Promise<Article> {
    const res = await fetch(`${API_BASE}/public/article/${id}`)
    if (!res.ok) throw new Error('Failed to get article')
    return res.json()
  },
}
