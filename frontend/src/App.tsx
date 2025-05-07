import React, { useState, useEffect } from 'react';
import Layout from './components/layout/Layout';
import UrlForm from './components/UrlForm';
import UrlResult from './components/UrlResult';
import UrlHistory from './components/UrlHistory';
import { useUrlShortener } from './hooks/useUrlShortener';
import { useUrlHistory } from './hooks/useUrlHistory';
import { useUserUrls } from './hooks/useUserUrls';
import { UrlHistoryItem } from './types';
import { LinkIcon } from 'lucide-react';

function App() {
  const [shortUrl, setShortUrl] = useState<string | null>(null);
  const [originalUrl, setOriginalUrl] = useState<string>('');
  const { shortenUrl, isLoading } = useUrlShortener();
  const { history, addToHistory, removeFromHistory, clearHistory } = useUrlHistory();
  const { urls: serverUrls, refetch: refetchUrls } = useUserUrls();

  // Sync server URLs with local history when server URLs change
  useEffect(() => {
    if (serverUrls && serverUrls.length > 0) {
      // Add server URLs to local history if they don't exist
      serverUrls.forEach(item => {
        const exists = history.some(historyItem =>
          historyItem.shortUrl === item.short_url
        );

        if (!exists) {
          addToHistory(item.original_url, item.short_url);
        }
      });
    }
  }, [serverUrls, history, addToHistory]);

  const handleSubmit = async (url: string, alias?: string) => {
    try {
      const result = await shortenUrl(url, alias);
      setShortUrl(result);
      setOriginalUrl(url);
      addToHistory(url, result);

      // Refresh server URLs
      setTimeout(() => {
        refetchUrls();
      }, 500);
    } catch (error) {
      console.error('Error shortening URL:', error);
    }
  };

  const handleReset = () => {
    setShortUrl(null);
    setOriginalUrl('');
  };

  const handleSelectHistoryItem = (item: UrlHistoryItem) => {
    setShortUrl(item.shortUrl);
    setOriginalUrl(item.originalUrl);
  };

  return (
    <Layout>
      <div className="w-full animate-fadeIn">
        <div className="mb-8 text-center">
          <div className="bg-purple-600 inline-flex p-3 rounded-full mb-4 shadow-glow">
            <LinkIcon size={32} className="text-white" />
          </div>
          <h1 className="text-4xl font-bold text-white mb-2">Shortify</h1>
          <p className="text-purple-200 max-w-xs mx-auto">
            Transform your long URLs into short, shareable links in seconds
          </p>
        </div>

        <div className="bg-white/10 backdrop-blur-md p-6 rounded-xl shadow-xl border border-purple-300/20">
          {!shortUrl ? (
            <UrlForm onSubmit={handleSubmit} isSubmitting={isLoading} />
          ) : (
            <UrlResult
              shortUrl={shortUrl}
              originalUrl={originalUrl}
              onReset={handleReset}
            />
          )}
        </div>

        <UrlHistory
          items={history}
          onClear={clearHistory}
          onRemoveItem={removeFromHistory}
          onSelectItem={handleSelectHistoryItem}
        />
      </div>
    </Layout>
  );
}

export default App;