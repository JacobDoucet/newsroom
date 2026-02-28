import { useState, useEffect } from 'react'
import {
  Box,
  Container,
  Typography,
  Button,
  Card,
  CardContent,
  Stack,
  TextField,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Chip,
  Alert,
  CircularProgress,
  Tabs,
  Tab,
  Paper,
  Divider,
  IconButton,
} from '@mui/material'
import { Refresh, Add, PlayArrow, FastForward } from '@mui/icons-material'
import { api, Arc, DraftPacket, DraftCandidate } from '../lib/api'

interface TabPanelProps {
  children?: React.ReactNode
  index: number
  value: number
}

function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props
  return (
    <div role="tabpanel" hidden={value !== index} {...other}>
      {value === index && <Box sx={{ pt: 3 }}>{children}</Box>}
    </div>
  )
}

export default function Console() {
  const [tab, setTab] = useState(0)
  const [arcs, setArcs] = useState<Arc[]>([])
  const [selectedArc, setSelectedArc] = useState<Arc | null>(null)
  const [packets, setPackets] = useState<DraftPacket[]>([])
  const [selectedPacket, setSelectedPacket] = useState<DraftPacket | null>(null)
  const [candidates, setCandidates] = useState<DraftCandidate[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  // Dialogs
  const [createArcDialog, setCreateArcDialog] = useState(false)
  const [generatePacketDialog, setGeneratePacketDialog] = useState(false)
  const [newArc, setNewArc] = useState({ slug: '', title: '', description: '' })
  const [generatePacketData, setGeneratePacketData] = useState({ dayIndex: 1, regionKey: '' })

  useEffect(() => {
    loadArcs()
  }, [])

  const loadArcs = async () => {
    try {
      setLoading(true)
      const data = await api.listArcs()
      setArcs(data)
      if (data.length > 0 && !selectedArc) {
        setSelectedArc(data[0])
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load arcs')
    } finally {
      setLoading(false)
    }
  }

  const handleCreateArc = async () => {
    try {
      setLoading(true)
      const arc = await api.createArc(newArc)
      setArcs([arc, ...arcs])
      setSelectedArc(arc)
      setCreateArcDialog(false)
      setNewArc({ slug: '', title: '', description: '' })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create arc')
    } finally {
      setLoading(false)
    }
  }

  const handleStartArc = async () => {
    if (!selectedArc) return
    try {
      setLoading(true)
      await api.startArc(selectedArc.id)
      await loadArcs()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to start arc')
    } finally {
      setLoading(false)
    }
  }

  const handleAdvanceArc = async () => {
    if (!selectedArc) return
    try {
      setLoading(true)
      await api.advanceArc(selectedArc.id)
      await loadArcs()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to advance arc')
    } finally {
      setLoading(false)
    }
  }

  const handleGeneratePacket = async () => {
    if (!selectedArc) return
    try {
      setLoading(true)
      const packet = await api.generatePacket(
        selectedArc.id,
        generatePacketData.dayIndex,
        generatePacketData.regionKey || undefined
      )
      setGeneratePacketDialog(false)
      setSelectedPacket(packet)
      setTab(1)

      // Poll for candidates
      pollForCandidates(packet.id)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to generate packet')
    } finally {
      setLoading(false)
    }
  }

  const pollForCandidates = async (packetId: string) => {
    let attempts = 0
    const maxAttempts = 30

    const poll = async () => {
      if (attempts >= maxAttempts) {
        setError('Candidate generation timed out')
        return
      }

      try {
        const packet = await api.getPacket(packetId)
        if (packet.status === 'generated') {
          const candidatesData = await api.getPacketCandidates(packetId)
          setCandidates(candidatesData)
          return
        }

        attempts++
        setTimeout(poll, 2000)
      } catch (err) {
        console.error('Error polling candidates:', err)
      }
    }

    poll()
  }

  const handleLoadCandidates = async (packetId: string) => {
    try {
      setLoading(true)
      const candidatesData = await api.getPacketCandidates(packetId)
      setCandidates(candidatesData)
      const packet = await api.getPacket(packetId)
      setSelectedPacket(packet)
      setTab(1)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load candidates')
    } finally {
      setLoading(false)
    }
  }

  const handlePublishCandidate = async (candidateId: string) => {
    try {
      setLoading(true)
      await api.publishCandidate(candidateId)
      alert('Article published successfully!')
      setCandidates(candidates.filter(c => c.id !== candidateId))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to publish candidate')
    } finally {
      setLoading(false)
    }
  }

  const handleSelectCandidate = async (candidateId: string) => {
    try {
      await api.selectCandidate(candidateId, { reason_tags: ['good_quality'], notes: 'Selected for publication' })
      alert('Candidate selected')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to select candidate')
    }
  }

  const handleRejectCandidate = async (candidateId: string) => {
    try {
      await api.rejectCandidate(candidateId, { reason_tags: ['poor_quality'], notes: 'Not suitable' })
      setCandidates(candidates.filter(c => c.id !== candidateId))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to reject candidate')
    }
  }

  return (
    <Box sx={{ minHeight: '100vh', bgcolor: '#fafafa' }}>
      <Box sx={{ bgcolor: '#1976d2', color: 'white', py: 2 }}>
        <Container maxWidth="xl">
          <Typography variant="h5" component="h1" sx={{ fontWeight: 600 }}>
            Editorial Console
          </Typography>
          <Typography variant="body2" sx={{ opacity: 0.9 }}>
            Human-in-the-loop editorial workflow
          </Typography>
        </Container>
      </Box>

      <Container maxWidth="xl" sx={{ py: 3 }}>
        {error && (
          <Alert severity="error" onClose={() => setError('')} sx={{ mb: 2 }}>
            {error}
          </Alert>
        )}

        <Paper sx={{ mb: 3 }}>
          <Tabs value={tab} onChange={(_, v) => setTab(v)}>
            <Tab label="Arc Management" />
            <Tab label="Review Candidates" />
          </Tabs>
        </Paper>

        <TabPanel value={tab} index={0}>
          <Stack spacing={3}>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <Typography variant="h6">Story Arcs</Typography>
              <Button
                variant="contained"
                startIcon={<Add />}
                onClick={() => setCreateArcDialog(true)}
              >
                Create Arc
              </Button>
            </Box>

            {selectedArc && (
              <Card>
                <CardContent>
                  <Typography variant="h6">{selectedArc.title}</Typography>
                  <Typography variant="body2" color="text.secondary" gutterBottom>
                    {selectedArc.description}
                  </Typography>
                  <Chip label={selectedArc.status} size="small" sx={{ mt: 1 }} />

                  <Box sx={{ mt: 2, display: 'flex', gap: 1 }}>
                    {selectedArc.status === 'draft' && (
                      <Button
                        variant="outlined"
                        startIcon={<PlayArrow />}
                        onClick={handleStartArc}
                        disabled={loading}
                      >
                        Start Arc
                      </Button>
                    )}
                    {selectedArc.status === 'active' && (
                      <>
                        <Button
                          variant="outlined"
                          startIcon={<FastForward />}
                          onClick={handleAdvanceArc}
                          disabled={loading}
                        >
                          Advance Day
                        </Button>
                        <Button
                          variant="contained"
                          onClick={() => setGeneratePacketDialog(true)}
                          disabled={loading}
                        >
                          Generate Draft Packet
                        </Button>
                      </>
                    )}
                  </Box>
                </CardContent>
              </Card>
            )}

            <Box>
              <Typography variant="h6" gutterBottom>All Arcs</Typography>
              <Stack spacing={2}>
                {arcs.map((arc) => (
                  <Card
                    key={arc.id}
                    sx={{
                      cursor: 'pointer',
                      borderLeft: selectedArc?.id === arc.id ? '4px solid #1976d2' : 'none'
                    }}
                    onClick={() => setSelectedArc(arc)}
                  >
                    <CardContent>
                      <Typography variant="subtitle1">{arc.title}</Typography>
                      <Typography variant="body2" color="text.secondary">
                        {arc.slug} • {arc.status}
                      </Typography>
                    </CardContent>
                  </Card>
                ))}
              </Stack>
            </Box>
          </Stack>
        </TabPanel>

        <TabPanel value={tab} index={1}>
          {selectedPacket ? (
            <Box>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 2 }}>
                <Typography variant="h6">
                  Candidates (Day {selectedPacket.day_index})
                </Typography>
                <Chip label={selectedPacket.status} />
              </Box>

              {candidates.length === 0 && selectedPacket.status === 'pending' && (
                <Box sx={{ textAlign: 'center', py: 4 }}>
                  <CircularProgress />
                  <Typography sx={{ mt: 2 }}>Generating candidates...</Typography>
                </Box>
              )}

              <Stack spacing={3}>
                {candidates.map((candidate) => {
                  const structured = candidate.structured as any
                  return (
                    <Card key={candidate.id}>
                      <CardContent>
                        <Typography variant="h5" gutterBottom>
                          {structured.headline || 'Untitled'}
                        </Typography>
                        {structured.subhead && (
                          <Typography variant="subtitle1" color="text.secondary" gutterBottom>
                            {structured.subhead}
                          </Typography>
                        )}
                        <Typography variant="body2" color="text.secondary" gutterBottom>
                          {structured.byline} • {structured.dateline}
                        </Typography>

                        <Divider sx={{ my: 2 }} />

                        <Typography variant="body1" paragraph>
                          {structured.lede}
                        </Typography>

                        <Typography variant="body2" sx={{ whiteSpace: 'pre-wrap' }}>
                          {structured.body_md?.substring(0, 400)}...
                        </Typography>

                        {structured.tags && (
                          <Box sx={{ mt: 2, display: 'flex', gap: 1, flexWrap: 'wrap' }}>
                            {structured.tags.map((tag: string) => (
                              <Chip key={tag} label={tag} size="small" />
                            ))}
                          </Box>
                        )}

                        <Box sx={{ mt: 2, display: 'flex', gap: 1 }}>
                          <Button
                            variant="contained"
                            color="success"
                            onClick={() => handlePublishCandidate(candidate.id)}
                          >
                            Publish
                          </Button>
                          <Button
                            variant="outlined"
                            onClick={() => handleSelectCandidate(candidate.id)}
                          >
                            Select
                          </Button>
                          <Button
                            variant="outlined"
                            color="error"
                            onClick={() => handleRejectCandidate(candidate.id)}
                          >
                            Reject
                          </Button>
                        </Box>
                      </CardContent>
                    </Card>
                  )
                })}
              </Stack>
            </Box>
          ) : (
            <Alert severity="info">
              Generate a draft packet from the Arc Management tab to review candidates.
            </Alert>
          )}
        </TabPanel>
      </Container>

      {/* Create Arc Dialog */}
      <Dialog open={createArcDialog} onClose={() => setCreateArcDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Create New Arc</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <TextField
              label="Slug"
              value={newArc.slug}
              onChange={(e) => setNewArc({ ...newArc, slug: e.target.value })}
              fullWidth
              required
            />
            <TextField
              label="Title"
              value={newArc.title}
              onChange={(e) => setNewArc({ ...newArc, title: e.target.value })}
              fullWidth
              required
            />
            <TextField
              label="Description"
              value={newArc.description}
              onChange={(e) => setNewArc({ ...newArc, description: e.target.value })}
              fullWidth
              multiline
              rows={3}
            />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setCreateArcDialog(false)}>Cancel</Button>
          <Button onClick={handleCreateArc} variant="contained" disabled={!newArc.slug || !newArc.title}>
            Create
          </Button>
        </DialogActions>
      </Dialog>

      {/* Generate Packet Dialog */}
      <Dialog open={generatePacketDialog} onClose={() => setGeneratePacketDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Generate Draft Packet</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <TextField
              label="Day Index"
              type="number"
              value={generatePacketData.dayIndex}
              onChange={(e) => setGeneratePacketData({ ...generatePacketData, dayIndex: parseInt(e.target.value) })}
              fullWidth
              required
            />
            <TextField
              label="Region Key (optional)"
              value={generatePacketData.regionKey}
              onChange={(e) => setGeneratePacketData({ ...generatePacketData, regionKey: e.target.value })}
              fullWidth
              placeholder="e.g., GB, US"
            />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setGeneratePacketDialog(false)}>Cancel</Button>
          <Button onClick={handleGeneratePacket} variant="contained" disabled={loading}>
            Generate
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  )
}
