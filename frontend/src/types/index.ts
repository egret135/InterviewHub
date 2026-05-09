export interface Category {
  id: number;
  name: string;
  slug: string;
  description: string;
  icon: string;
  sort_order: number;
  question_count: number;
}

export interface Tag {
  id: number;
  name: string;
  slug: string;
}

export interface Answer {
  id: number;
  question_id: number;
  content_md: string;
}

export interface Question {
  id: number;
  category_id: number;
  title: string;
  difficulty: 'easy' | 'medium' | 'hard';
  sort_order: number;
  category: Category;
  answer?: Answer;
  tags?: Tag[];
}

export interface ListResponse<T> {
  data: T;
  meta?: {
    page: number;
    size: number;
    total: number;
  };
}

export interface ErrorResponse {
  error: {
    code: string;
    message: string;
  };
}
