import { useState } from 'react';

// This is a mock implementation that would be replaced when the backend is ready
export const useUrlShortener = () => {
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const shortenUrl = async (url: string): Promise<string> => {
    setIsLoading(true);
    setError(null);
    
    try {
      // Simulate API call with a timeout
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      // Mock response - this would be replaced with an actual API call
      // Generate a random short URL for demonstration purposes
      const randomId = Math.random().toString(36).substring(2, 8);
      const shortUrl = `https://short.url/${randomId}`;
      
      return shortUrl;
    } catch (err) {
      setError('Failed to shorten URL. Please try again later.');
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