import { Outlet, Link } from 'react-router-dom';
import { Logo } from './Logo';

export function Layout() {
  return (
    <div className="min-h-screen flex flex-col bg-[var(--color-bg-primary)]">
      <header className="border-b border-[var(--color-border)] bg-[var(--color-bg-panel)] shadow-sm sticky top-0 z-10">
        <div className="max-w-6xl mx-auto px-4 py-3 flex items-center justify-between">
          <Logo />
          <nav className="flex items-center gap-5 text-sm text-[var(--color-text-secondary)]">
            <Link to="/" className="hover:text-[var(--color-accent)] transition-colors font-medium">
              首页
            </Link>
            <Link to="/search" className="hover:text-[var(--color-accent)] transition-colors font-medium">
              搜索
            </Link>
          </nav>
        </div>
      </header>
      <main className="flex-1">
        <Outlet />
      </main>
      <footer className="border-t border-[var(--color-border)] py-6 text-center text-xs text-[var(--color-text-muted)]">
        InterviewHub · Go 后端面试题库
      </footer>
    </div>
  );
}
