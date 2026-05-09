import { useEffect } from 'react';
import { useSearchParams, Link } from 'react-router-dom';
import { useSearch } from '../hooks/useQuestions';
import { SearchBar } from '../components/SearchBar';
import { QuestionCard } from '../components/QuestionCard';

export function SearchPage() {
  const [searchParams] = useSearchParams();
  const query = searchParams.get('q') || '';
  const { questions, total, loading, search } = useSearch();

  useEffect(() => {
    search(query);
  }, [query, search]);

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <Link
        to="/"
        className="inline-flex items-center gap-1 text-sm text-[var(--color-text-muted)] hover:text-[var(--color-accent)] transition-colors mb-6"
      >
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
        </svg>
        返回首页
      </Link>

      <div className="mb-8">
        <SearchBar />
      </div>

      {loading && (
        <div className="text-center py-10 text-[var(--color-text-muted)]">搜索中...</div>
      )}

      {!loading && query && (
        <p className="text-sm text-[var(--color-text-secondary)] mb-4">
          找到 {total} 条相关题目
        </p>
      )}

      {!loading && query && questions.length === 0 && (
        <div className="text-center py-20">
          <p className="text-[var(--color-text-muted)] text-lg mb-2">未找到相关题目</p>
          <p className="text-sm text-[var(--color-text-muted)]">尝试其他关键词</p>
        </div>
      )}

      <div className="space-y-3">
        {questions.map(q => (
          <QuestionCard key={q.id} question={q} defaultOpen />
        ))}
      </div>
    </div>
  );
}
