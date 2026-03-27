/**
 * src/pages/RegisterPage.tsx
 * New user registration form.
 */
import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { LoadingSpinner } from '../components/LoadingSpinner';

export function RegisterPage() {
  const { register } = useAuth();
  const navigate = useNavigate();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(''); setSuccess('');
    if (!username || !password) { setError('Please fill in all fields.'); return; }
    if (password.length < 6) { setError('Password must be at least 6 characters.'); return; }
    if (password !== confirm) { setError('Passwords do not match.'); return; }

    setLoading(true);
    try {
      await register({ username, password });
      setSuccess('Account created! Redirecting to login...');
      setTimeout(() => navigate('/login'), 1500);
    } catch (err: unknown) {
      const msg = (err as { response?: { data?: { error?: string } } })?.response?.data?.error;
      if (msg?.includes('409') || msg?.includes('exists')) {
        setError('Username already taken.');
      } else {
        setError(msg ?? 'Registration failed.');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-[#111827] flex items-center justify-center p-4">
      <div className="bg-[#1F2937] rounded-2xl p-8 w-full max-w-sm shadow-2xl border border-[#374151]">
        <div className="text-center mb-8">
          <div className="text-5xl mb-3">🎵</div>
          <h1 className="text-2xl font-bold text-[#F9FAFB]">Create Account</h1>
          <p className="text-[#6B7280] mt-1 text-sm">Join gtmpc today</p>
        </div>

        {error && (
          <div className="mb-4 px-3 py-2 bg-[#EF4444]/10 border border-[#EF4444]/30 rounded-lg text-[#EF4444] text-sm">
            {error}
          </div>
        )}
        {success && (
          <div className="mb-4 px-3 py-2 bg-[#10B981]/10 border border-[#10B981]/30 rounded-lg text-[#10B981] text-sm">
            {success}
          </div>
        )}

        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <div>
            <label className="block text-sm text-[#6B7280] mb-1">Username</label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="w-full px-4 py-2.5 bg-[#374151] border border-[#4B5563] rounded-lg text-[#F9FAFB] placeholder-[#6B7280] focus:outline-none focus:border-[#7C3AED] transition-colors"
              placeholder="Choose a username"
              disabled={loading}
            />
          </div>
          <div>
            <label className="block text-sm text-[#6B7280] mb-1">Password</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full px-4 py-2.5 bg-[#374151] border border-[#4B5563] rounded-lg text-[#F9FAFB] placeholder-[#6B7280] focus:outline-none focus:border-[#7C3AED] transition-colors"
              placeholder="Min 6 characters"
              disabled={loading}
            />
          </div>
          <div>
            <label className="block text-sm text-[#6B7280] mb-1">Confirm Password</label>
            <input
              type="password"
              value={confirm}
              onChange={(e) => setConfirm(e.target.value)}
              className="w-full px-4 py-2.5 bg-[#374151] border border-[#4B5563] rounded-lg text-[#F9FAFB] placeholder-[#6B7280] focus:outline-none focus:border-[#7C3AED] transition-colors"
              placeholder="Repeat password"
              disabled={loading}
            />
          </div>
          <button
            type="submit"
            disabled={loading}
            className="mt-2 w-full py-2.5 rounded-lg bg-[#7C3AED] hover:bg-[#6D28D9] disabled:bg-[#374151] text-white font-semibold transition-colors flex items-center justify-center gap-2"
          >
            {loading ? <LoadingSpinner size="sm" /> : 'Create Account'}
          </button>
        </form>

        <p className="mt-6 text-center text-sm text-[#6B7280]">
          Already have an account?{' '}
          <Link to="/login" className="text-[#7C3AED] hover:underline">
            Sign in
          </Link>
        </p>
      </div>
    </div>
  );
}
