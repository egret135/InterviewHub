import { Link } from 'react-router-dom';

export function Logo() {
  return (
    <Link to="/" className="flex items-center gap-2.5 group">
      {/* Icon: Stacked code brackets representing interview question bank */}
      <svg
        className="w-8 h-8 shrink-0"
        viewBox="0 0 32 32"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <rect
          x="1" y="1" width="30" height="30" rx="7"
          className="fill-[var(--color-accent-dim)] stroke-[var(--color-accent)] stroke-[1.5]"
        />
        <path
          d="M12 11L8 16l4 5"
          className="stroke-[var(--color-accent)]"
          strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
        />
        <path
          d="M20 11l4 5-4 5"
          className="stroke-[var(--color-accent)]"
          strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
        />
        <line
          x1="18" y1="9" x2="15" y2="23"
          className="stroke-[var(--color-accent)]"
          strokeWidth="1.5" strokeLinecap="round"
        />
      </svg>

      {/* Text */}
      <span className="text-lg font-semibold tracking-tight text-[var(--color-text-primary)]">
        InterviewHub
      </span>
    </Link>
  );
}
