/**
 * src/components/VolumeSlider.tsx
 * Volume control with icon and range input.
 */

interface VolumeSliderProps {
  volume: number;
  onVolumeChange: (pct: number) => void;
}

export function VolumeSlider({ volume, onVolumeChange }: VolumeSliderProps) {
  const icon = volume === 0 ? '🔇' : volume < 50 ? '🔉' : '🔊';

  return (
    <div className="flex items-center gap-2">
      <span className="text-base">{icon}</span>
      <input
        type="range"
        min={0}
        max={100}
        value={volume}
        onChange={(e) => onVolumeChange(Number(e.target.value))}
        className="w-20 accent-[#7C3AED] cursor-pointer"
        aria-label="Volume"
      />
    </div>
  );
}
