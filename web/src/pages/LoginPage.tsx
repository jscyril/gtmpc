/**
 * src/pages/LoginPage.tsx
 * Login & Register toggle page.
 */
import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { LoadingSpinner } from '../components/LoadingSpinner';

export function LoginPage() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    if (!username || !password) { setError('Please fill in all fields.'); return; }
    setLoading(true);
    try {
      await login({ username, password });
      navigate('/library');
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error;
      setError(msg ?? 'Login failed. Check your credentials.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-[#111827] flex items-center justify-center p-4">
      <div className="bg-[#1F2937] rounded-2xl p-8 w-full max-w-sm shadow-2xl border border-[#374151]">
        <div className="text-center mb-8">
          <div className="text-5xl mb-3">🎵</div>
          <h1 className="text-2xl font-bold text-[#F9FAFB]">Welcome to gtmpc</h1>
          <p className="text-[#6B7280] mt-1 text-sm">Sign in to your music player</p>
        </div>

        {error && (
          <div className="mb-4 px-3 py-2 bg-[#EF4444]/10 border border-[#EF4444]/30 rounded-lg text-[#EF4444] text-sm">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <div>
            <label className="block text-sm text-[#6B7280] mb-1">Username</label>
            <input
              id="username"
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              autoComplete="username"
              className="w-full px-4 py-2.5 bg-[#374151] border border-[#4B5563] rounded-lg text-[#F9FAFB] placeholder-[#6B7280] focus:outline-none focus:border-[#7C3AED] transition-colors"
              placeholder="Enter username"
              disabled={loading}
            />
          </div>
          <div>
            <label className="block text-sm text-[#6B7280] mb-1">Password</label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              autoComplete="current-password"
              className="w-full px-4 py-2.5 bg-[#374151] border border-[#4B5563] rounded-lg text-[#F9FAFB] placeholder-[#6B7280] focus:outline-none focus:border-[#7C3AED] transition-colors"
              placeholder="Enter password"
              disabled={loading}
            />
          </div>
          <button
            type="submit"
            disabled={loading}
            className="mt-2 w-full py-2.5 rounded-lg bg-[#7C3AED] hover:bg-[#6D28D9] disabled:bg-[#374151] text-white font-semibold transition-colors flex items-center justify-center gap-2"
          >
            {loading ? <LoadingSpinner size="sm" /> : 'Sign In'}
          </button>
        </form>

        <p className="mt-6 text-center text-sm text-[#6B7280]">
          No account?{' '}
          <Link to="/register" className="text-[#7C3AED] hover:underline">
            Register
          </Link>
        </p>
      </div>
    </div>
  );
}
