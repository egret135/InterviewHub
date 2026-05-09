import { Link } from 'react-router-dom';
import type { Category } from '../types';

const iconMap: Record<string, string> = {
  zap: '⚡',
  cpu: '🔬',
  layers: '🏗',
  package: '📦',
  database: '🗄',
  globe: '🌐',
  coffee: '☕',
  briefcase: '💼',
};

export function CategoryCard({ category }: { category: Category }) {
  return (
    <Link
      to={`/category/${category.slug}`}
      className="card p-5 block group cursor-pointer
                 transition-all duration-200 hover:-translate-y-0.5
                 hover:border-[var(--color-border-hover)] hover:shadow-[var(--shadow-card-hover)]"
    >
      <div className="flex items-start justify-between mb-3">
        <span className="text-2xl">{iconMap[category.icon] || '📋'}</span>
        <span className="text-xs px-2 py-0.5 rounded-full bg-[var(--color-accent-dim)] text-[var(--color-accent)] font-medium">
          {category.question_count} 题
        </span>
      </div>
      <h3 className="font-semibold text-[var(--color-text-primary)] mb-1 group-hover:text-[var(--color-accent)] transition-colors">
        {category.name}
      </h3>
      <p className="text-sm text-[var(--color-text-secondary)] leading-relaxed">{category.description}</p>
    </Link>
  );
}
