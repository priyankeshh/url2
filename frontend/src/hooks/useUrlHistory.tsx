import { useState, useEffect } from 'react';
import { UrlHistoryItem } from '../types';

const STORAGE_KEY = 'url_shortener_history';

export const useUrlHistory = () => {
  const [history, setHistory] = useState<UrlHistoryItem[]>([]);

  // Load history from localStorage on mount
  useEffect(() => {
    try {
      const storedHistory = localStorage.getItem(STORAGE_KEY);
      if (storedHistory) {
        setHistory(JSON.parse(storedHistory));
      }
    } catch (error) {
      console.error('Failed to load history:', error);
    }
  }, []);

  // Save history to localStorage when it changes
  useEffect(() => {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(history));
    } catch (error) {
      console.error('Failed to save history:', error);
    }
  }, [history]);

  // Add a new item to history
  const addToHistory = (originalUrl: string, shortUrl: string) => {
    const newItem: UrlHistoryItem = {
      originalUrl,
      shortUrl,
      createdAt: Date.now(),
      id: crypto.randomUUID()
    };
    
    setHistory(prev => [newItem, ...prev.slice(0, 9)]); // Keep last 10 items
  };

  // Remove an item from history
  const removeFromHistory = (id: string) => {
    setHistory(prev => prev.filter(item => item.id !== id));
  };

  // Clear all history
  const clearHistory = () => {
    setHistory([]);
  };

  return {
    history,
    addToHistory,
    removeFromHistory,
    clearHistory
  };
};