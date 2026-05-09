import { useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';

export function SearchBar() {
  const [value, setValue] = useState('');
  const navigate = useNavigate();

  const handleSubmit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault();
      if (value.trim()) {
        navigate(`/search?q=${encodeURIComponent(value.trim())}`);
      }
    },
    [value, navigate]
  );

  return (
    <form onSubmit={handleSubmit} className="w-full max-w-2xl mx-auto">
      <div className="relative">
        <svg
          className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5"
          style={{ color: 'var(--color-text-muted)' }}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        <input
          type="text"
          value={value}
          onChange={e => setValue(e.target.value)}
          placeholder="搜索题目、分类或标签..."
          className="w-full pl-12 pr-4 py-3 rounded-xl text-sm
                     bg-[var(--color-bg-panel)] border border-[var(--color-border)]
                     text-[var(--color-text-primary)] placeholder-[var(--color-text-muted)]
                     focus:outline-none focus:border-[var(--color-accent)] focus:ring-1 focus:ring-[var(--color-accent)]
                     transition-all duration-200"
        />
      </div>
    </form>
  );
}
