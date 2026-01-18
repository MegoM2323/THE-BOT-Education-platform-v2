import { createContext, useReducer, useEffect, useRef } from 'react';
import { logger } from '../utils/logger.js';
import * as authAPI from '../api/auth.js';
import { setupCSRF, clearCSRFToken, setCSRFToken } from '../api/client.js';
import { queryClient } from '../lib/queryClient.js';

export const AuthContext = createContext(null);

// Reducer для управления auth state (избегаем race conditions)
const authReducer = (state, action) => {
  switch (action.type) {
    case 'AUTH_SUCCESS':
      return {
        ...state,
        user: action.payload.user,
        balance: action.payload.balance,
        isAuthenticated: true,
        loading: false,
      };
    case 'AUTH_FAILURE':
      return {
        ...state,
        user: null,
        balance: null,
        isAuthenticated: false,
        loading: false,
      };
    case 'SET_LOADING':
      return { ...state, loading: action.payload };
    case 'UPDATE_USER':
      return { ...state, user: { ...state.user, ...action.payload } };
    case 'LOGOUT':
      return {
        ...state,
        user: null,
        balance: null,
        isAuthenticated: false,
      };
    default:
      return state;
  }
};

const initialState = {
  user: null,
  balance: null,
  loading: true,
  isAuthenticated: false,
};

export const AuthProvider = ({ children }) => {
  const [state, dispatch] = useReducer(authReducer, initialState);
  const authCheckPromiseRef = useRef(null);
  const refreshUserPromiseRef = useRef(null);
  const loginPromiseRef = useRef(null);
  const requestIdRef = useRef(0); // Track request order to ignore stale responses
  const isMountedRef = useRef(true);

  // Проверка статуса аутентификации при монтировании
  useEffect(() => {
    isMountedRef.current = true;
    checkAuth();  // setupCSRF now called inside checkAuth

    // Cleanup: предотвращаем state updates после unmount
    return () => {
      isMountedRef.current = false;
      authCheckPromiseRef.current = null;
      refreshUserPromiseRef.current = null;
      loginPromiseRef.current = null;
    };
  }, []);

  const checkAuth = async () => {
    // Если запрос уже в процессе, используем существующее обещание
    if (authCheckPromiseRef.current) {
      return authCheckPromiseRef.current;
    }

    // Создаем новый запрос и сохраняем обещание
    const currentRequestId = ++requestIdRef.current;
    authCheckPromiseRef.current = (async () => {
      try {
        const data = await authAPI.getMe();

        // Проверяем mounted и что это последний запрос перед state updates
        if (!isMountedRef.current || requestIdRef.current !== currentRequestId) {
          logger.debug(`[AuthContext] Ignoring stale checkAuth response (request ${currentRequestId}, latest ${requestIdRef.current})`);
          return;
        }

        const userData = data.user || data;
        const balance = data.balance !== undefined ? data.balance : null;

        dispatch({
          type: 'AUTH_SUCCESS',
          payload: { user: userData, balance }
        });

        // Fetch CSRF token immediately after successful auth
        await setupCSRF();
        logger.debug('[AuthContext] CSRF token fetched after auth check');
      } catch (error) {
        // Проверяем mounted и что это последний запрос перед state updates
        if (!isMountedRef.current || requestIdRef.current !== currentRequestId) {
          logger.debug(`[AuthContext] Ignoring stale checkAuth error (request ${currentRequestId}, latest ${requestIdRef.current})`);
          return;
        }

        dispatch({ type: 'AUTH_FAILURE' });
        clearCSRFToken();  // Clear CSRF token on auth failure
        // Backend manages session cookie (httpOnly), no client-side token to clear
      } finally {
        // Очищаем кэш обещания после завершения
        authCheckPromiseRef.current = null;
      }
    })();

    return authCheckPromiseRef.current;
  };

  const login = async (email, password) => {
    // CRITICAL: Always clear previous login state before starting new login
    // This prevents returning stale cached user data from previous login
    // Even if a login is in progress, we abort it and start fresh
    // This fixes the issue where wrong user data was displayed after switching accounts
    loginPromiseRef.current = null;

    // Clear previous user state immediately to prevent showing old user during login
    // This ensures UI shows loading/empty state instead of stale user data
    if (isMountedRef.current) {
      dispatch({ type: 'SET_LOADING', payload: true });
      // Also clear React Query cache early to prevent old user's data from appearing
      queryClient.clear();
    }

    // Создаем новый запрос и сохраняем обещание
    const currentRequestId = ++requestIdRef.current;
    loginPromiseRef.current = (async () => {
      const maxRetries = 2; // First attempt + 1 retry
      let lastError = null;

      for (let attempt = 1; attempt <= maxRetries; attempt++) {
        try {
          logger.debug(`[AuthContext] Login attempt ${attempt}/${maxRetries}: ${email}`);

          const data = await authAPI.login(email, password);

          // Проверяем mounted и что это последний запрос перед state updates
          if (!isMountedRef.current || requestIdRef.current !== currentRequestId) {
            logger.debug(`[AuthContext] Ignoring stale login response (request ${currentRequestId}, latest ${requestIdRef.current})`);
            return null;
          }

          const userData = data.user || data;

          // Backend sets httpOnly session cookie automatically
          // No need to store token in localStorage (XSS protection)

          // CSRF token is already extracted from X-CSRF-Token header by client.js
          // No need to fetch separately or extract from body
          logger.debug('[AuthContext] CSRF token already set via response headers by client');

          // Финальная проверка mounted и что это последний запрос перед state updates
          if (!isMountedRef.current || requestIdRef.current !== currentRequestId) {
            logger.debug(`[AuthContext] Ignoring stale login response (request ${currentRequestId}, latest ${requestIdRef.current})`);
            return null;
          }

          const balance = data.balance !== undefined ? data.balance : null;

          // CRITICAL: Clear React Query cache BEFORE dispatching new user auth
          // This prevents previous user's cached data (balance, bookings, etc) from being visible
          logger.debug('[AuthContext] Clearing React Query cache for new login');
          queryClient.clear();

          dispatch({
            type: 'AUTH_SUCCESS',
            payload: { user: userData, balance }
          });

          logger.info('[AuthContext] Login successful');
          return userData;

        } catch (error) {
          lastError = error;
          logger.warn(`[AuthContext] Login attempt ${attempt} failed: ${error.message}`);

          if (attempt < maxRetries) {
            // Exponential backoff: 500ms, 1000ms
            const delayMs = 500 * Math.pow(2, attempt - 1);
            logger.debug(`[AuthContext] Retrying login in ${delayMs}ms...`);
            await new Promise(resolve => setTimeout(resolve, delayMs));
            continue; // Retry
          }
        }
      }

      // All retries exhausted - fail login
      if (isMountedRef.current && requestIdRef.current === currentRequestId) {
        dispatch({ type: 'AUTH_FAILURE' });
        clearCSRFToken();
        logger.error('[AuthContext] Login failed after all retries, AUTH_FAILURE dispatched');
      }
      throw lastError || new Error('Login failed after multiple attempts');
    })().finally(() => {
      // CRITICAL: Clear login promise ref after completion to prevent stale data
      // Without this, subsequent login calls would return the old resolved promise
      // with data from the previous user (causing wrong user to be displayed)
      loginPromiseRef.current = null;
    });

    return loginPromiseRef.current;
  };

  const logout = async () => {
    try {
      await authAPI.logout();
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      // CRITICAL: Clear all cached promises to prevent stale user data
      // This ensures next login will create fresh requests
      loginPromiseRef.current = null;
      authCheckPromiseRef.current = null;
      refreshUserPromiseRef.current = null;

      // CRITICAL: Clear React Query cache on logout
      // This ensures no cached data persists after user leaves
      logger.debug('[AuthContext] Clearing React Query cache for logout');
      queryClient.clear();

      // Проверяем mounted перед state updates
      if (isMountedRef.current) {
        dispatch({ type: 'LOGOUT' });
      }
      clearCSRFToken();
      // Backend clears session cookie (httpOnly) on logout
    }
  };

  const updateUser = (updates) => {
    // Проверяем mounted перед state updates
    if (!isMountedRef.current) return;
    dispatch({ type: 'UPDATE_USER', payload: updates });
  };

  const refreshUser = async () => {
    // Если запрос уже в процессе, используем существующее обещание
    if (refreshUserPromiseRef.current) {
      return refreshUserPromiseRef.current;
    }

    // Создаем новый запрос и сохраняем обещание
    const currentRequestId = ++requestIdRef.current;
    refreshUserPromiseRef.current = (async () => {
      try {
        const data = await authAPI.getMe();

        // Проверяем mounted и что это последний запрос перед state updates
        if (!isMountedRef.current || requestIdRef.current !== currentRequestId) {
          logger.debug(`[AuthContext] Ignoring stale refreshUser response (request ${currentRequestId}, latest ${requestIdRef.current})`);
          return null;
        }

        const userData = data.user || data;
        const balance = data.balance !== undefined ? data.balance : null;

        dispatch({
          type: 'AUTH_SUCCESS',
          payload: { user: userData, balance }
        });

        return userData;
      } catch (error) {
        console.error('Error refreshing user:', error);
        throw error;
      } finally {
        // Очищаем кэш обещания после завершения
        refreshUserPromiseRef.current = null;
      }
    })();

    return refreshUserPromiseRef.current;
  };

  const value = {
    user: state.user,
    balance: state.balance,
    loading: state.loading,
    isAuthenticated: state.isAuthenticated,
    login,
    logout,
    updateUser,
    refreshUser,
    checkAuth,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export default AuthContext;
