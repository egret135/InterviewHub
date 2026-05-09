import { useState } from 'react';
import { Link } from 'react-router-dom';
import { DifficultyBadge } from './DifficultyBadge';
import { MarkdownContent } from './MarkdownContent';
import type { Question } from '../types';

export function QuestionCard({ question, defaultOpen = false }: { question: Question; defaultOpen?: boolean }) {
  const [open, setOpen] = useState(defaultOpen);

  return (
    <div className="card overflow-hidden transition-all duration-200 hover:shadow-[var(--shadow-card-hover)]">
      <div className="flex items-start gap-4 p-5">
        <span className="text-[var(--color-accent)] font-mono text-sm mt-0.5 shrink-0 font-semibold">
          Q{question.sort_order}
        </span>

        <div className="flex-1 min-w-0">
          {/* Title links to detail page */}
          <Link
            to={`/question/${question.id}`}
            className="text-[var(--color-text-primary)] leading-relaxed hover:text-[var(--color-accent)] transition-colors"
          >
            {question.title}
          </Link>
          <div className="flex items-center gap-2 mt-2 flex-wrap">
            <DifficultyBadge level={question.difficulty} />
            {question.tags?.slice(0, 3).map(tag => (
              <span key={tag.id} className="text-xs text-[var(--color-text-muted)]">
                #{tag.name}
              </span>
            ))}
            <Link
              to={`/question/${question.id}`}
              className="text-xs text-[var(--color-accent)] ml-auto hover:underline"
            >
              详情
            </Link>
          </div>
        </div>

        {/* Expand toggle */}
        <button
          onClick={() => setOpen(!open)}
          className="shrink-0 p-1 rounded hover:bg-[var(--color-accent-dim)] transition-colors cursor-pointer"
          title={open ? '收起答案' : '展开答案'}
        >
          <svg
            className={`w-5 h-5 text-[var(--color-text-muted)] transition-transform duration-200 ${open ? 'rotate-180' : ''}`}
            fill="none" stroke="currentColor" viewBox="0 0 24 24"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
          </svg>
        </button>
      </div>

      {open && question.answer && (
        <div className="border-t border-[var(--color-border)] px-5 pb-5 pt-4 bg-[var(--color-bg-primary)]/50">
          <MarkdownContent content={question.answer.content_md} />
          {question.tags && question.tags.length > 0 && (
            <div className="flex flex-wrap gap-1.5 mt-4 pt-4 border-t border-[var(--color-border)]">
              {question.tags.map(tag => (
                <span
                  key={tag.id}
                  className="text-xs px-2 py-1 rounded-md bg-[var(--color-accent-dim)] text-[var(--color-accent)] font-medium"
                >
                  {tag.name}
                </span>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
