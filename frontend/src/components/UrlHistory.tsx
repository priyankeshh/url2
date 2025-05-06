import React from 'react';
import { Clock, ExternalLink, Trash2 } from 'lucide-react';
import Button from './ui/Button';
import { UrlHistoryItem } from '../types';

interface UrlHistoryProps {
  items: UrlHistoryItem[];
  onClear: () => void;
  onRemoveItem: (id: string) => void;
  onSelectItem: (item: UrlHistoryItem) => void;
}

const UrlHistory: React.FC<UrlHistoryProps> = ({ 
  items, 
  onClear, 
  onRemoveItem,
  onSelectItem
}) => {
  if (items.length === 0) {
    return null;
  }
  
  // Format the timestamp to a readable format
  const formatDate = (timestamp: number) => {
    return new Date(timestamp).toLocaleString(undefined, {
      month: 'short',
      day: 'numeric',
      hour: 'numeric',
      minute: '2-digit'
    });
  };

  return (
    <div className="mt-8 animate-fadeIn">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-white font-semibold flex items-center">
          <Clock size={16} className="mr-2" />
          Recent URLs
        </h2>
        {items.length > 0 && (
          <Button 
            size="sm" 
            variant="outline" 
            onClick={onClear}
            className="text-xs"
          >
            Clear All
          </Button>
        )}
      </div>
      
      <div className="space-y-3 max-h-72 overflow-y-auto pr-1">
        {items.map((item) => (
          <div 
            key={item.id}
            className="bg-white bg-opacity-10 backdrop-blur-sm rounded-lg p-3 border border-purple-300 border-opacity-10 hover:border-opacity-20 transition-all duration-200"
          >
            <div className="flex justify-between items-start">
              <div 
                className="cursor-pointer flex-1 mr-2"
                onClick={() => onSelectItem(item)}
              >
                <p className="text-white text-sm font-medium truncate">
                  {item.shortUrl}
                </p>
                <p className="text-purple-200 text-xs truncate mt-1">
                  {item.originalUrl}
                </p>
                <p className="text-purple-300 text-xs mt-2 opacity-80">
                  {formatDate(item.createdAt)}
                </p>
              </div>
              <div className="flex space-x-2">
                <button 
                  className="text-purple-200 hover:text-white transition-colors p-1 rounded"
                  onClick={() => window.open(item.shortUrl, '_blank')}
                  title="Open shortened URL"
                >
                  <ExternalLink size={16} />
                </button>
                <button 
                  className="text-purple-200 hover:text-red-300 transition-colors p-1 rounded"
                  onClick={() => onRemoveItem(item.id)}
                  title="Remove from history"
                >
                  <Trash2 size={16} />
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default UrlHistory;