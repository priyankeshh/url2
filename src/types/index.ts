export interface UrlHistoryItem {
  originalUrl: string;
  shortUrl: string;
  createdAt: number;
  id: string;
}

export interface FormState {
  url: string;
  error: string;
  isSubmitting: boolean;
}