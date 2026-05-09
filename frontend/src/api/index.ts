const BASE = '/api/v1';

async function fetchJSON<T>(url: string): Promise<T> {
  const res = await fetch(url);
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: { message: 'Network error' } }));
    throw new Error(err.error?.message || `HTTP ${res.status}`);
  }
  return res.json();
}

export const api = {
  getCategories: () =>
    fetchJSON<{ data: import('../types').Category[] }>(`${BASE}/categories`),

  getQuestionsByCategory: (slug: string, page = 1, size = 20) =>
    fetchJSON<{
      data: { category: import('../types').Category; questions: import('../types').Question[] };
      meta: { page: number; size: number; total: number };
    }>(`${BASE}/categories/${slug}/questions?page=${page}&size=${size}`),

  getQuestion: (id: number) =>
    fetchJSON<{ data: import('../types').Question }>(`${BASE}/questions/${id}`),

  searchQuestions: (q: string, page = 1, size = 20) =>
    fetchJSON<{ data: import('../types').Question[]; meta: { total: number } }>(
      `${BASE}/questions/search?q=${encodeURIComponent(q)}&page=${page}&size=${size}`
    ),

  getTags: () =>
    fetchJSON<{ data: import('../types').Tag[] }>(`${BASE}/tags`),
};
