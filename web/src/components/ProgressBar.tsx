/**
 * src/components/ProgressBar.tsx
 * Audio seek bar component.
 */
import { useRef } from 'react';
import { formatTime } from '../utils/formatTime';

interface ProgressBarProps {
  progress: number;
  duration: number;
  onSeek: (seconds: number) => void;
}

export function ProgressBar({ progress, duration, onSeek }: ProgressBarProps) {
  const barRef = useRef<HTMLDivElement>(null);

  const ratio = duration > 0 ? Math.min(progress / duration, 1) : 0;

  const handleClick = (e: React.MouseEvent<HTMLDivElement>) => {
    if (!barRef.current || duration <= 0) return;
    const rect = barRef.current.getBoundingClientRect();
    const x = e.clientX - rect.left;
    const pct = x / rect.width;
    onSeek(pct * duration);
  };

  return (
    <div className="flex items-center gap-2 w-full">
      <span className="text-[#6B7280] text-xs w-10 text-right shrink-0">{formatTime(progress)}</span>
      <div
        ref={barRef}
        onClick={handleClick}
        className="relative flex-1 h-1.5 rounded-full bg-[#374151] cursor-pointer group"
      >
        <div
          className="absolute left-0 top-0 h-full rounded-full bg-[#7C3AED] transition-all duration-150"
          style={{ width: `${ratio * 100}%` }}
        />
        <div
          className="absolute top-1/2 -translate-y-1/2 -translate-x-1/2 w-3 h-3 rounded-full bg-white opacity-0 group-hover:opacity-100 transition-opacity"
          style={{ left: `${ratio * 100}%` }}
        />
      </div>
      <span className="text-[#6B7280] text-xs w-10 shrink-0">{formatTime(duration)}</span>
    </div>
  );
}
