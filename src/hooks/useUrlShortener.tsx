import { useState } from 'react';

// API URL - can be overridden with environment variables
const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export const useUrlShortener = () => {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const shortenUrl = async (url: string): Promise<string> => {
    setIsLoading(true);
    setError(null);

    try {
      // Make API request to the Go backend
      const response = await fetch(`${API_URL}/api/shorten`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ url }),
      });

      // Handle non-2xx responses
      if (!response.ok) {
        let errorMessage = 'Failed to shorten URL';

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

      // Return the full short URL
      return data.url || `${API_URL}/r/${data.code}`;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to shorten URL. Please try again later.';
      setError(errorMessage);
      throw err;
    } finally {
      setIsLoading(false);
    }
  };

  return {
    shortenUrl,
    isLoading,
    error
  };
};