import { useState, useEffect } from 'react';

// API URL - can be overridden with environment variables
const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export interface UserUrl {
  code: string;
  short_url: string;
  original_url: string;
  created_at: string;
}

export const useUserUrls = () => {
  const [urls, setUrls] = useState<UserUrl[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchUrls = async () => {
    setIsLoading(true);
    setError(null);

    try {
      // Make API request to the Go backend
      const response = await fetch(`${API_URL}/api/urls`, {
        method: 'GET',
        credentials: 'include', // Include cookies for user identification
      });

      // Handle non-2xx responses
      if (!response.ok) {
        let errorMessage = 'Failed to fetch URLs';

        try {
          const errorData = await response.json();
          if (errorData.error) {
            errorMessage = errorData.error;
          }
        } catch (e) {
          // If we can't parse the error JSON, use the default message
        }

        throw new Error(errorMessage);
      }

      // Parse the response
      const data = await response.json();
      setUrls(data);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch URLs. Please try again later.';
      setError(errorMessage);
      console.error(errorMessage);
    } finally {
      setIsLoading(false);
    }
  };

  // Fetch URLs on mount
  useEffect(() => {
    fetchUrls();
  }, []);

  return {
    urls,
    isLoading,
    error,
    refetch: fetchUrls
  };
};
