/**
 * Validates if the provided string is a valid URL
 * @param url - The URL string to validate
 * @returns boolean indicating if the URL is valid
 */
export const isValidUrl = (url: string): boolean => {
  try {
    // Try to construct a URL object
    new URL(url);
    return true;
  } catch (error) {
    return false;
  }
};

/**
 * Validates if the URL has a protocol (http/https)
 * @param url - The URL string to check
 * @returns The URL with protocol if missing, or the original URL
 */
export const ensureProtocol = (url: string): string => {
  if (url && !url.startsWith('http://') && !url.startsWith('https://')) {
    return `https://${url}`;
  }
  return url;
};

/**
 * Get validation error message for URL input
 * @param url - The URL string to validate
 * @returns Error message or empty string if valid
 */
export const getUrlValidationError = (url: string): string => {
  if (!url) {
    return 'URL is required';
  }
  
  const urlWithProtocol = ensureProtocol(url);
  
  if (!isValidUrl(urlWithProtocol)) {
    return 'Please enter a valid URL';
  }
  
  return '';
};