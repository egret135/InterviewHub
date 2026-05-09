import { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { api } from '../api';
import { DifficultyBadge } from '../components/DifficultyBadge';
import { MarkdownContent } from '../components/MarkdownContent';
import type { Question } from '../types';

export function QuestionPage() {
  const { id } = useParams<{ id: string }>();
  const [question, setQuestion] = useState<Question | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    api.getQuestion(Number(id))
      .then(res => setQuestion(res.data))
      .catch(err => setError(err.message))
      .finally(() => setLoading(false));
  }, [id]);

  if (loading) {
    return (
      <div className="max-w-4xl mx-auto px-4 py-20 text-center text-[var(--color-text-muted)]">
        加载中...
      </div>
    );
  }

  if (error || !question) {
    return (
      <div className="max-w-4xl mx-auto px-4 py-20 text-center">
        <p className="text-[var(--color-difficulty-hard)] text-lg mb-2">
          {error || '题目不存在'}
        </p>
        <Link to="/" className="text-sm text-[var(--color-accent)] hover:underline">
          返回首页
        </Link>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      {/* Breadcrumb */}
      <div className="flex items-center gap-2 text-sm mb-6">
        <Link to="/" className="text-[var(--color-text-muted)] hover:text-[var(--color-accent)] transition-colors">
          首页
        </Link>
        <span className="text-[var(--color-text-muted)]">/</span>
        <Link
          to={`/category/${question.category.slug}`}
          className="text-[var(--color-text-muted)] hover:text-[var(--color-accent)] transition-colors"
        >
          {question.category.name}
        </Link>
        <span className="text-[var(--color-text-muted)]">/</span>
        <span className="text-[var(--color-text-secondary)]">Q{question.sort_order}</span>
      </div>

      {/* Question header */}
      <div className="card p-6 mb-6">
        <div className="flex items-start gap-4">
          <span className="text-[var(--color-accent)] font-mono text-lg font-semibold mt-0.5 shrink-0">
            Q{question.sort_order}
          </span>
          <div>
            <h1 className="text-xl font-semibold text-[var(--color-text-primary)] leading-relaxed mb-3">
              {question.title}
            </h1>
            <div className="flex items-center gap-3 flex-wrap">
              <DifficultyBadge level={question.difficulty} />
              {question.tags?.map(tag => (
                <span key={tag.id} className="text-xs text-[var(--color-text-muted)]">
                  #{tag.name}
                </span>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Answer */}
      {question.answer ? (
        <div className="card p-6">
          <h2 className="text-sm font-semibold text-[var(--color-text-muted)] uppercase tracking-wide mb-4">
            参考答案
          </h2>
          <MarkdownContent content={question.answer.content_md} />
        </div>
      ) : (
        <div className="card p-6 text-center text-[var(--color-text-muted)]">
          暂无答案
        </div>
      )}

      {/* Tags footer */}
      {question.tags && question.tags.length > 0 && (
        <div className="mt-4 flex flex-wrap gap-2">
          {question.tags.map(tag => (
            <span
              key={tag.id}
              className="text-xs px-2.5 py-1 rounded-md bg-[var(--color-accent-dim)] text-[var(--color-accent)] font-medium"
            >
              {tag.name}
            </span>
          ))}
        </div>
      )}
    </div>
  );
}
