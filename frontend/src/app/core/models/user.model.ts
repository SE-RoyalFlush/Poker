/**
 * User model interface representing the user data returned from backend API.
 * Matches the backend User model structure.
 */
export interface User {
  id: number;
  username: string;
  createdAt?: string;
  updatedAt?: string;
}

/**
 * Login credentials for authentication.
 */
export interface LoginCredentials {
  username: string;
  password: string;
}

/**
 * Registration data for creating a new user account.
 */
export interface RegisterData {
  username: string;
  password: string;
  confirmPassword: string;
}

