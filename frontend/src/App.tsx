import { useState, useEffect, useRef } from 'react'
import './App.css'

interface Game {
  id: string
  name: string
  created_at: string
  updated_at: string
}

interface LeaderboardEntry {
  rank: number
  player_id: string
  score: number
}

interface LeaderboardSnapshot {
  type: string
  game_id: number
  leaderboard: LeaderboardEntry[]
  timestamp: number
}

interface ScoreUpdate {
  type: string
  player_id: string
  score: number
  rank: number
}

type WSMessage = LeaderboardSnapshot | ScoreUpdate

function App() {
  const [games, setGames] = useState<Game[]>([])
  const [selectedGameId, setSelectedGameId] = useState<string>('')
  const [leaderboard, setLeaderboard] = useState<LeaderboardEntry[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string>('')
  const wsRef = useRef<WebSocket | null>(null)

  // Fetch games on component mount
  useEffect(() => {
    fetchGames()
  }, [])

  // Handle game selection and websocket connection
  useEffect(() => {
    if (selectedGameId) {
      connectWebSocket(selectedGameId)
    } else {
      disconnectWebSocket()
    }

    return () => {
      disconnectWebSocket()
    }
  }, [selectedGameId])

  const fetchGames = async () => {
    try {
      setLoading(true)
      const response = await fetch('http://localhost/api/v1/games')
      if (!response.ok) {
        throw new Error('Failed to fetch games')
      }
      const data = await response.json()
      setGames(data.items || [])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch games')
    } finally {
      setLoading(false)
    }
  }

  const connectWebSocket = (gameId: string) => {
    disconnectWebSocket()

    const ws = new WebSocket(`ws://localhost/api/v1/games/${gameId}/leaderboard/ws`)
    wsRef.current = ws

    ws.onopen = () => {
      console.log('WebSocket connected')
      setError('')
    }

    ws.onmessage = (event) => {
      try {
        const message: WSMessage = JSON.parse(event.data)
        if (message.type === 'leaderboard_snapshot') {
          setLeaderboard((message as LeaderboardSnapshot).leaderboard)
        } else if (message.type === 'score_update') {
          // For simplicity, we'll refetch the leaderboard on updates
          // In a more sophisticated implementation, you could update the local state
          const update = message as ScoreUpdate
          setLeaderboard(prev => {
            const newLeaderboard = [...prev]
            const existingIndex = newLeaderboard.findIndex(entry => entry.player_id === update.player_id)
            if (existingIndex >= 0) {
              newLeaderboard[existingIndex] = {
                rank: update.rank,
                player_id: update.player_id,
                score: update.score
              }
            } else {
              newLeaderboard.push({
                rank: update.rank,
                player_id: update.player_id,
                score: update.score
              })
            }
            return newLeaderboard.sort((a, b) => b.score - a.score).map((entry, index) => ({
              ...entry,
              rank: index + 1
            }))
          })
        }
      } catch (err) {
        console.error('Failed to parse WebSocket message:', err)
      }
    }

    ws.onerror = (error) => {
      console.error('WebSocket error:', error)
      setError('WebSocket connection failed')
    }

    ws.onclose = () => {
      console.log('WebSocket disconnected')
    }
  }

  const disconnectWebSocket = () => {
    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }
  }

  return (
    <div className="app">
      <header>
        <h1>Gaming Leaderboard</h1>
      </header>

      <main>
        <div className="controls">
          <label htmlFor="game-select">Select Game:</label>
          <select
            id="game-select"
            value={selectedGameId}
            onChange={(e) => setSelectedGameId(e.target.value)}
            disabled={loading}
          >
            <option value="">Choose a game...</option>
            {games.map((game) => (
              <option key={game.id} value={game.id}>
                {game.name}
              </option>
            ))}
          </select>
        </div>

        {error && (
          <div className="error">
            {error}
          </div>
        )}

        {selectedGameId && (
          <div className="leaderboard">
            <h2>Leaderboard</h2>
            {leaderboard.length === 0 ? (
              <p>No scores yet</p>
            ) : (
              <table>
                <thead>
                  <tr>
                    <th>Rank</th>
                    <th>Player ID</th>
                    <th>Score</th>
                  </tr>
                </thead>
                <tbody>
                  {leaderboard.map((entry) => (
                    <tr key={entry.player_id} className={`rank-${entry.rank <= 3 ? entry.rank : ''}`}>
                      <td>{entry.rank}</td>
                      <td>{entry.player_id}</td>
                      <td>{entry.score.toLocaleString()}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        )}
      </main>
    </div>
  )
}

export default App
