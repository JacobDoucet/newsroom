import { Routes, Route } from 'react-router-dom'
import { ThemeProvider, createTheme, CssBaseline } from '@mui/material'
import PublicSite from './pages/PublicSite'
import Console from './pages/Console'

const theme = createTheme({
  palette: {
    mode: 'light',
    primary: {
      main: '#1976d2',
    },
  },
  typography: {
    fontFamily: '"Georgia", "Times New Roman", serif',
  },
})

function App() {
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Routes>
        <Route path="/" element={<PublicSite />} />
        <Route path="/console/*" element={<Console />} />
      </Routes>
    </ThemeProvider>
  )
}

export default App
