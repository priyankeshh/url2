import React, { useState } from 'react';
import { Check, Copy, ExternalLink } from 'lucide-react';
import Button from './ui/Button';

interface UrlResultProps {
  shortUrl: string;
  originalUrl: string;
  onReset: () => void;
}

const UrlResult: React.FC<UrlResultProps> = ({ shortUrl, originalUrl, onReset }) => {
  const [copied, setCopied] = useState(false);
  
  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(shortUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error('Failed to copy: ', err);
    }
  };
  
  return (
    <div className="animate-fadeIn">
      <div className="bg-white bg-opacity-10 backdrop-blur-lg p-4 rounded-lg border border-purple-300 border-opacity-20 shadow-lg mb-4">
        <div className="flex flex-col">
          <div className="text-xs text-purple-200 mb-1">Original URL:</div>
          <p className="text-white text-sm mb-4 truncate" title={originalUrl}>
            {originalUrl}
          </p>
          
          <div className="text-xs text-purple-200 mb-1">Shortened URL:</div>
          <div className="flex items-center justify-between bg-white bg-opacity-20 rounded-md px-3 py-2 mb-4">
            <a 
              href={shortUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="text-white font-medium hover:underline flex items-center"
            >
              {shortUrl}
              <ExternalLink size={14} className="ml-1 opacity-70" />
            </a>
            <Button
              variant="secondary"
              size="sm"
              onClick={handleCopy}
              icon={copied ? <Check size={16} /> : <Copy size={16} />}
              className={copied ? 'bg-green-600 text-white hover:bg-green-700' : ''}
            >
              {copied ? 'Copied!' : 'Copy'}
            </Button>
          </div>
        </div>
      </div>
      
      <Button 
        onClick={onReset} 
        variant="outline"
        className="w-full"
      >
        Shorten Another URL
      </Button>
    </div>
  );
};

export default UrlResult;