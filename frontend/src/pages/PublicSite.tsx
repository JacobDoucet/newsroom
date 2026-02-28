import { useState, useEffect } from 'react'
import { Box, Container, Typography, Card, CardContent, Chip, Stack, Alert } from '@mui/material'
import { api, Article } from '../lib/api'

export default function PublicSite() {
  const [articles, setArticles] = useState<Article[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    loadArticles()
  }, [])

  const loadArticles = async () => {
    try {
      setLoading(true)
      const data = await api.getLatestArticles()
      setArticles(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load articles')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Box sx={{ minHeight: '100vh', bgcolor: '#f5f5f5' }}>
      <Box sx={{ bgcolor: '#1a1a1a', color: 'white', py: 2, borderBottom: '4px solid #c00' }}>
        <Container maxWidth="lg">
          <Typography variant="h4" component="h1" sx={{ fontWeight: 700, fontFamily: '"Arial", sans-serif' }}>
            FICTIONAL NEWS
          </Typography>
          <Typography variant="body2" sx={{ color: '#ccc', mt: 0.5 }}>
            A simulated newsroom — all content is fiction
          </Typography>
        </Container>
      </Box>

      <Container maxWidth="lg" sx={{ py: 4 }}>
        <Alert severity="warning" sx={{ mb: 3 }}>
          This is a fictional newsroom simulator. All articles are generated content and are not real news.
        </Alert>

        {loading && (
          <Typography>Loading articles...</Typography>
        )}

        {error && (
          <Alert severity="error">{error}</Alert>
        )}

        <Stack spacing={3}>
          {articles.map((article) => (
            <Card key={article.id} elevation={2}>
              <CardContent>
                <Typography variant="h5" component="h2" gutterBottom sx={{ fontWeight: 600 }}>
                  {article.headline}
                </Typography>

                {article.subhead && (
                  <Typography variant="subtitle1" color="text.secondary" gutterBottom>
                    {article.subhead}
                  </Typography>
                )}

                <Box sx={{ display: 'flex', gap: 2, mb: 2, alignItems: 'center' }}>
                  <Typography variant="body2" color="text.secondary">
                    {article.byline}
                  </Typography>
                  {article.dateline && (
                    <>
                      <Typography variant="body2" color="text.secondary">•</Typography>
                      <Typography variant="body2" color="text.secondary">
                        {article.dateline}
                      </Typography>
                    </>
                  )}
                  {article.published_at && (
                    <>
                      <Typography variant="body2" color="text.secondary">•</Typography>
                      <Typography variant="body2" color="text.secondary">
                        {new Date(article.published_at).toLocaleDateString()}
                      </Typography>
                    </>
                  )}
                </Box>

                <Typography
                  variant="body1"
                  sx={{
                    mb: 2,
                    whiteSpace: 'pre-wrap',
                    lineHeight: 1.7
                  }}
                >
                  {article.body_md.substring(0, 300)}...
                </Typography>

                <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap' }}>
                  {article.tags.map((tag) => (
                    <Chip key={tag} label={tag} size="small" variant="outlined" />
                  ))}
                </Box>
              </CardContent>
            </Card>
          ))}
        </Stack>

        {articles.length === 0 && !loading && !error && (
          <Typography variant="body1" sx={{ textAlign: 'center', py: 4, color: 'text.secondary' }}>
            No articles published yet. Use the Console to generate and publish articles.
          </Typography>
        )}
      </Container>

      <Box sx={{ bgcolor: '#1a1a1a', color: 'white', py: 3, mt: 6 }}>
        <Container maxWidth="lg">
          <Typography variant="body2" sx={{ color: '#999' }}>
            Fictional Newsroom Simulator — This is a storytelling platform. All content is fiction.
          </Typography>
        </Container>
      </Box>
    </Box>
  )
}
