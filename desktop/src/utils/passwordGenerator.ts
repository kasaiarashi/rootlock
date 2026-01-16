export interface PasswordOptions {
  length: number;
  uppercase: boolean;
  lowercase: boolean;
  numbers: boolean;
  symbols: boolean;
}

const UPPERCASE = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
const LOWERCASE = 'abcdefghijklmnopqrstuvwxyz';
const NUMBERS = '0123456789';
const SYMBOLS = '!@#$%^&*()_+-=[]{}|;:,.<>?';

export function generatePassword(options: PasswordOptions): string {
  let charset = '';
  
  if (options.uppercase) charset += UPPERCASE;
  if (options.lowercase) charset += LOWERCASE;
  if (options.numbers) charset += NUMBERS;
  if (options.symbols) charset += SYMBOLS;
  
  if (charset === '') {
    throw new Error('At least one character type must be selected');
  }
  
  // Use crypto.getRandomValues for cryptographically secure random
  const array = new Uint32Array(options.length);
  crypto.getRandomValues(array);
  
  let password = '';
  for (let i = 0; i < options.length; i++) {
    password += charset[array[i] % charset.length];
  }
  
  // Ensure at least one character from each selected type
  if (options.uppercase && !hasUppercase(password)) {
    password = replaceRandomChar(password, UPPERCASE);
  }
  if (options.lowercase && !hasLowercase(password)) {
    password = replaceRandomChar(password, LOWERCASE);
  }
  if (options.numbers && !hasNumber(password)) {
    password = replaceRandomChar(password, NUMBERS);
  }
  if (options.symbols && !hasSymbol(password)) {
    password = replaceRandomChar(password, SYMBOLS);
  }
  
  return password;
}

function replaceRandomChar(password: string, charset: string): string {
  const array = new Uint32Array(2);
  crypto.getRandomValues(array);
  
  const pos = array[0] % password.length;
  const char = charset[array[1] % charset.length];
  
  return password.substring(0, pos) + char + password.substring(pos + 1);
}

function hasUppercase(str: string): boolean {
  return /[A-Z]/.test(str);
}

function hasLowercase(str: string): boolean {
  return /[a-z]/.test(str);
}

function hasNumber(str: string): boolean {
  return /[0-9]/.test(str);
}

function hasSymbol(str: string): boolean {
  return /[!@#$%^&*()_+\-=[\]{}|;:,.<>?]/.test(str);
}

export function calculatePasswordStrength(password: string): {
  score: number;
  label: string;
  color: string;
} {
  let score = 0;
  
  // Length scoring
  if (password.length >= 8) score += 1;
  if (password.length >= 12) score += 1;
  if (password.length >= 16) score += 1;
  
  // Character variety
  if (hasUppercase(password)) score += 1;
  if (hasLowercase(password)) score += 1;
  if (hasNumber(password)) score += 1;
  if (hasSymbol(password)) score += 1;
  
  // Normalize to 0-4 scale
  const normalizedScore = Math.min(Math.floor(score / 2), 4);
  
  const labels = ['Very Weak', 'Weak', 'Fair', 'Strong', 'Very Strong'];
  const colors = ['#d32f2f', '#f57c00', '#fbc02d', '#689f38', '#388e3c'];
  
  return {
    score: normalizedScore,
    label: labels[normalizedScore],
    color: colors[normalizedScore]
  };
}
