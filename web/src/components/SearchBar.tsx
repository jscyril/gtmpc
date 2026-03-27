/**
 * src/components/SearchBar.tsx
 * Inline search/filter input for the library.
 */

interface SearchBarProps {
  value: string;
  onChange: (v: string) => void;
  placeholder?: string;
}

export function SearchBar({ value, onChange, placeholder = 'Search tracks...' }: SearchBarProps) {
  return (
    <div className="relative">
      <span className="absolute left-3 top-1/2 -translate-y-1/2 text-[#6B7280]">🔍</span>
      <input
        type="text"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className="w-full pl-9 pr-4 py-2 bg-[#374151] border border-[#4B5563] rounded-lg text-[#F9FAFB] placeholder-[#6B7280] focus:outline-none focus:border-[#7C3AED] transition-colors"
      />
      {value && (
        <button
          onClick={() => onChange('')}
          className="absolute right-3 top-1/2 -translate-y-1/2 text-[#6B7280] hover:text-[#F9FAFB]"
        >
          ✕
        </button>
      )}
    </div>
  );
}
