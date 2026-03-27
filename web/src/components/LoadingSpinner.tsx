/**
 * src/components/LoadingSpinner.tsx
 * Animated spinner for loading states.
 */
export function LoadingSpinner({ size = 'md', label }: { size?: 'sm' | 'md' | 'lg'; label?: string }) {
  const sizeClass = { sm: 'w-5 h-5', md: 'w-8 h-8', lg: 'w-12 h-12' }[size];
  return (
    <div className="flex flex-col items-center gap-3">
      <div
        className={`${sizeClass} rounded-full border-2 border-[#4B5563] border-t-[#7C3AED] animate-spin`}
      />
      {label && <p className="text-[#6B7280] text-sm">{label}</p>}
    </div>
  );
}
