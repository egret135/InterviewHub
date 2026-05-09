const labels: Record<string, string> = {
  easy: '简单',
  medium: '中等',
  hard: '困难',
};

export function DifficultyBadge({ level }: { level: string }) {
  const colors: Record<string, string> = {
    easy: 'text-[var(--color-difficulty-easy)] bg-[var(--color-difficulty-easy)]/10',
    medium: 'text-[var(--color-difficulty-medium)] bg-[var(--color-difficulty-medium)]/10',
    hard: 'text-[var(--color-difficulty-hard)] bg-[var(--color-difficulty-hard)]/10',
  };

  return (
    <span className={`inline-block text-xs px-2 py-0.5 rounded-full font-medium ${colors[level] || colors.medium}`}>
      {labels[level] || level}
    </span>
  );
}
