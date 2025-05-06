import React, { forwardRef } from 'react';

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
  fullWidth?: boolean;
}

const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ label, error, fullWidth = true, className = '', ...props }, ref) => {
    const baseClasses = 'block rounded-lg border bg-white px-4 py-3 text-gray-700 focus:outline-none focus:ring-2 transition-all duration-200';
    const widthClass = fullWidth ? 'w-full' : '';
    const errorClasses = error
      ? 'border-red-300 focus:border-red-500 focus:ring-red-200'
      : 'border-gray-200 focus:border-purple-500 focus:ring-purple-200';
    
    const inputClasses = [
      baseClasses,
      errorClasses,
      widthClass,
      className
    ].join(' ');

    return (
      <div className={fullWidth ? 'w-full' : ''}>
        {label && (
          <label className="mb-2 block font-medium text-gray-700">
            {label}
          </label>
        )}
        <input ref={ref} className={inputClasses} {...props} />
        {error && (
          <p className="mt-1 text-sm text-red-600 animate-fadeIn">{error}</p>
        )}
      </div>
    );
  }
);

Input.displayName = 'Input';

export default Input;