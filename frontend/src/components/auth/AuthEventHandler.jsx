/**
 * Component that handles auth events (401 unauthorized)
 * Must be placed inside Router and AuthProvider
 * Listens for auth:unauthorized events and handles logout + navigation
 */
import { useAuthEventHandler } from '../../hooks/useAuthEventHandler.js';

export const AuthEventHandler = () => {
  useAuthEventHandler();
  return null; // This component doesn't render anything
};

export default AuthEventHandler;
