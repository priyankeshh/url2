import React, { useState } from 'react';
import { Link } from 'lucide-react';
import Input from './ui/Input';
import Button from './ui/Button';
import { ensureProtocol, getUrlValidationError } from '../utils/validation';
import { FormState } from '../types';

interface UrlFormProps {
  onSubmit: (url: string) => void;
  isSubmitting: boolean;
}

const UrlForm: React.FC<UrlFormProps> = ({ onSubmit, isSubmitting }) => {
  const [formState, setFormState] = useState<FormState>({
    url: '',
    error: '',
    isSubmitting: false
  });

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setFormState({
      ...formState,
      url: value,
      error: ''
    });
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    // Validate the URL
    const urlWithProtocol = ensureProtocol(formState.url);
    const error = getUrlValidationError(urlWithProtocol);
    
    if (error) {
      setFormState({
        ...formState,
        error
      });
      return;
    }
    
    // Call the parent's onSubmit with the validated URL
    onSubmit(urlWithProtocol);
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="flex flex-col space-y-2">
        <Input
          type="text"
          placeholder="Enter your long URL"
          value={formState.url}
          onChange={handleChange}
          error={formState.error}
          aria-label="URL to shorten"
          autoFocus
        />
      </div>
      <Button 
        type="submit" 
        className="w-full"
        isLoading={isSubmitting}
        icon={<Link size={18} />}
        size="lg"
      >
        Shorten URL
      </Button>
    </form>
  );
};

export default UrlForm;