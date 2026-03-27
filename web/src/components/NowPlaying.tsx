/**
 * src/components/NowPlaying.tsx
 * Fixed bottom player bar — visible on all authenticated pages.
 * Shows current track info, play/pause controls, seek bar, and volume.
 */
import { usePlayer } from '../context/PlayerContext';
import { AlbumArt } from './AlbumArt';
import { ProgressBar } from './ProgressBar';
import { VolumeSlider } from './VolumeSlider';

export function NowPlaying() {
  const { currentTrack, isPlaying, progress, duration, volume, pause, resume, next, previous, seek, setVolume } =
    usePlayer();

  if (!currentTrack) {
    return (
      <footer className="fixed bottom-0 left-0 right-0 h-20 bg-[#1F2937] border-t border-[#374151] flex items-center px-4">
        <p className="text-[#6B7280] text-sm">No track playing</p>
      </footer>
    );
  }

  return (
    <footer className="fixed bottom-0 left-0 right-0 h-20 bg-[#1F2937] border-t border-[#374151] flex items-center px-4 gap-4 z-50">
      {/* Left: album art + info */}
      <div className="flex items-center gap-3 w-1/4 min-w-0">
        <AlbumArt trackId={currentTrack.id} title={currentTrack.title} size={48} />
        <div className="min-w-0">
          <p className="text-sm font-medium truncate text-[#F9FAFB]">{currentTrack.title}</p>
          <p className="text-xs text-[#6B7280] truncate">{currentTrack.artist}</p>
        </div>
      </div>

      {/* Center: controls + progress */}
      <div className="flex-1 flex flex-col items-center gap-1">
        <div className="flex items-center gap-4">
          <button
            onClick={previous}
            className="text-[#6B7280] hover:text-[#F9FAFB] transition-colors text-lg"
            aria-label="Previous"
          >
            ⏮
          </button>
          <button
            onClick={isPlaying ? pause : resume}
            className="w-10 h-10 rounded-full bg-[#7C3AED] hover:bg-[#6D28D9] flex items-center justify-center text-white transition-colors"
            aria-label={isPlaying ? 'Pause' : 'Play'}
          >
            {isPlaying ? '⏸' : '▶'}
          </button>
          <button
            onClick={next}
            className="text-[#6B7280] hover:text-[#F9FAFB] transition-colors text-lg"
            aria-label="Next"
          >
            ⏭
          </button>
        </div>
        <ProgressBar progress={progress} duration={duration} onSeek={seek} />
      </div>

      {/* Right: volume */}
      <div className="w-1/4 flex justify-end">
        <VolumeSlider volume={volume} onVolumeChange={setVolume} />
      </div>
    </footer>
  );
}
