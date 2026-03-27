/**
 * src/App.tsx
 * Application router — uses React Router v6.
 *
 * Public routes:  /login, /register
 * Protected routes (require token):  /library, /playlists, /stats
 * Default redirect: / → /library if authenticated, /login otherwise.
 */
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import { StatsProvider } from './context/StatsContext';
import { PlayerProvider } from './context/PlayerContext';
import { ProtectedRoute } from './components/ProtectedRoute';
import { Layout } from './components/Layout';
import { LoginPage } from './pages/LoginPage';
import { RegisterPage } from './pages/RegisterPage';
import { LibraryPage } from './pages/LibraryPage';
import { PlaylistsPage } from './pages/PlaylistsPage';
import { StatsPage } from './pages/StatsPage';

export default function App() {
  return (
    <AuthProvider>
      <StatsProvider>
        <PlayerProvider>
          <BrowserRouter>
            <Routes>
              {/* Public */}
              <Route path="/login" element={<LoginPage />} />
              <Route path="/register" element={<RegisterPage />} />

              {/* Protected */}
              <Route element={<ProtectedRoute />}>
                <Route element={<Layout />}>
                  <Route path="/library" element={<LibraryPage />} />
                  <Route path="/playlists" element={<PlaylistsPage />} />
                  <Route path="/stats" element={<StatsPage />} />
                </Route>
              </Route>

              {/* Default redirect */}
              <Route path="/" element={<Navigate to="/library" replace />} />
              <Route path="*" element={<Navigate to="/library" replace />} />
            </Routes>
          </BrowserRouter>
        </PlayerProvider>
      </StatsProvider>
    </AuthProvider>
  );
}

