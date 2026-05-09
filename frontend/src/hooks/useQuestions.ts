import { useState, useEffect, useCallback } from 'react';
import { api } from '../api';
import type { Question, Category } from '../types';

export function useQuestionsByCategory(slug: string) {
  const [questions, setQuestions] = useState<Question[]>([]);
  const [category, setCategory] = useState<Category | null>(null);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    api.getQuestionsByCategory(slug)
      .then(res => {
        setQuestions(res.data.questions);
        setCategory(res.data.category);
        setTotal(res.meta.total);
      })
      .catch(err => setError(err.message))
      .finally(() => setLoading(false));
  }, [slug]);

  return { questions, category, total, loading, error };
}

export function useSearch(query: string) {
  const [questions, setQuestions] = useState<Question[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const search = useCallback((q: string) => {
    if (!q.trim()) {
      setQuestions([]);
      setTotal(0);
      return;
    }
    setLoading(true);
    api.searchQuestions(q)
      .then(res => {
        setQuestions(res.data);
        setTotal(res.meta.total);
      })
      .catch(err => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  return { questions, total, loading, error, search };
}
