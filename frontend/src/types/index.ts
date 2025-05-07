export interface UrlHistoryItem {
  originalUrl: string;
  shortUrl: string;
  createdAt: number;
  id: string;
}

export interface FormState {
  url: string;
  alias?: string;
  error: string;
  aliasError?: string;
  isSubmitting: boolean;
}