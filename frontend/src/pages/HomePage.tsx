import { useCategories } from '../hooks/useCategories';
import { CategoryCard } from '../components/CategoryCard';
import { SearchBar } from '../components/SearchBar';

export function HomePage() {
  const { categories, loading, error } = useCategories();

  return (
    <div className="max-w-6xl mx-auto px-4 py-12">
      {/* Hero */}
      <div className="text-center mb-12">
        <h1 className="text-4xl font-bold mb-4 tracking-tight">
          <span className="text-[var(--color-accent)]">Go</span>{' '}
          <span className="text-[var(--color-text-primary)]">后端面试题库</span>
        </h1>
        <p className="text-[var(--color-text-secondary)] text-lg mb-8">
          简洁优雅的面试题库，覆盖 Go 核心、微服务、分布式等 8 大领域
        </p>
        <SearchBar />
      </div>

      {/* Category Grid */}
      {loading && (
        <div className="text-center py-20 text-[var(--color-text-muted)]">
          加载中...
        </div>
      )}

      {error && (
        <div className="text-center py-20">
          <p className="text-[var(--color-difficulty-hard)] mb-2">加载失败</p>
          <p className="text-sm text-[var(--color-text-muted)]">{error}</p>
        </div>
      )}

      {!loading && !error && (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {categories.map(cat => (
            <CategoryCard key={cat.id} category={cat} />
          ))}
        </div>
      )}
    </div>
  );
}
