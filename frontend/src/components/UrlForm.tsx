import React, { useState } from 'react';
import { Link } from 'lucide-react';
import Input from './ui/Input';
import Button from './ui/Button';
import { ensureProtocol, getUrlValidationError } from '../utils/validation';
import { FormState } from '../types';

interface UrlFormProps {
  onSubmit: (url: string, alias?: string) => void;
  isSubmitting: boolean;
}

const UrlForm: React.FC<UrlFormProps> = ({ onSubmit, isSubmitting }) => {
  const [formState, setFormState] = useState<FormState>({
    url: '',
    alias: '',
    error: '',
    aliasError: '',
    isSubmitting: false
  });

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target;

    setFormState({
      ...formState,
      [name]: value,
      error: name === 'url' ? '' : formState.error,
      aliasError: name === 'alias' ? '' : formState.aliasError
    });
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    // Validate the URL
    const urlWithProtocol = ensureProtocol(formState.url);
    const error = getUrlValidationError(urlWithProtocol);

    // Validate the alias if provided
    let aliasError = '';
    if (formState.alias) {
      if (formState.alias.length < 3) {
        aliasError = 'Alias must be at least 3 characters';
      } else if (formState.alias.length > 20) {
        aliasError = 'Alias must be at most 20 characters';
      } else if (!/^[a-zA-Z0-9]+$/.test(formState.alias)) {
        aliasError = 'Alias must contain only letters and numbers';
      }
    }

    if (error || aliasError) {
      setFormState({
        ...formState,
        error,
        aliasError
      });
      return;
    }

    // Call the parent's onSubmit with the validated URL and optional alias
    onSubmit(urlWithProtocol, formState.alias);
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="flex flex-col space-y-4">
        <Input
          type="text"
          name="url"
          placeholder="Enter your long URL"
          value={formState.url}
          onChange={handleChange}
          error={formState.error}
          aria-label="URL to shorten"
          autoFocus
        />

        <div className="relative">
          <Input
            type="text"
            name="alias"
            placeholder="Custom alias (optional)"
            value={formState.alias}
            onChange={handleChange}
            error={formState.aliasError}
            aria-label="Custom alias"
          />
          <div className="mt-1 text-xs text-purple-200">
            Choose a custom name for your short URL (letters and numbers only)
          </div>
        </div>
      </div>

      <Button
        type="submit"
        className="w-full mt-6"
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