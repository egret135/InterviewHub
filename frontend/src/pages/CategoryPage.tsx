import { useParams, Link } from 'react-router-dom';
import { useQuestionsByCategory } from '../hooks/useQuestions';
import { QuestionCard } from '../components/QuestionCard';

export function CategoryPage() {
  const { slug } = useParams<{ slug: string }>();
  const { questions, category, loading, error } = useQuestionsByCategory(slug!);

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

      {loading && (
        <div className="text-center py-20 text-[var(--color-text-muted)]">加载中...</div>
      )}

      {error && (
        <div className="text-center py-20">
          <p className="text-[var(--color-difficulty-hard)] mb-2">加载失败</p>
          <p className="text-sm text-[var(--color-text-muted)]">{error}</p>
        </div>
      )}

      {category && (
        <>
          <div className="mb-8">
            <h1 className="text-2xl font-bold text-[var(--color-text-primary)] mb-2">
              {category.name}
            </h1>
            <p className="text-[var(--color-text-secondary)]">{category.description}</p>
          </div>

          {questions.length === 0 && (
            <p className="text-center py-10 text-[var(--color-text-muted)]">暂无题目</p>
          )}

          <div className="space-y-3">
            {questions.map(q => (
              <QuestionCard key={q.id} question={q} />
            ))}
          </div>
        </>
      )}
    </div>
  );
}
