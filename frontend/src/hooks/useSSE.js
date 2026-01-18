import { useState, useEffect, useCallback, useRef } from 'react';
import { logger } from '../utils/logger.js';

const SSE_URL = '/api/v1/events/chat';
const RECONNECT_DELAY_MS = 3000;

/**
 * React hook for SSE (Server-Sent Events) connection to chat events
 * Auto-reconnects on error after 3 seconds
 * Handles new_message and message_deleted events
 */
export const useSSE = () => {
  const [isConnected, setIsConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState(null);
  const [lastDeletedMessage, setLastDeletedMessage] = useState(null);

  const eventSourceRef = useRef(null);
  const reconnectTimeoutRef = useRef(null);
  const isMountedRef = useRef(true);

  const cleanup = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }
  }, []);

  const connect = useCallback(() => {
    cleanup();

    if (!isMountedRef.current) {
      return;
    }

    try {
      const eventSource = new EventSource(SSE_URL, {
        withCredentials: true,
      });

      eventSourceRef.current = eventSource;

      eventSource.onopen = () => {
        if (isMountedRef.current) {
          logger.debug('[SSE] Connection opened');
          setIsConnected(true);
        }
      };

      eventSource.onerror = (error) => {
        logger.error('[SSE] Connection error:', error);

        if (isMountedRef.current) {
          setIsConnected(false);

          eventSource.close();
          eventSourceRef.current = null;

          logger.debug(`[SSE] Reconnecting in ${RECONNECT_DELAY_MS}ms...`);
          reconnectTimeoutRef.current = setTimeout(() => {
            if (isMountedRef.current) {
              connect();
            }
          }, RECONNECT_DELAY_MS);
        }
      };

      eventSource.addEventListener('new_message', (event) => {
        if (!isMountedRef.current) return;

        try {
          const data = JSON.parse(event.data);
          logger.debug('[SSE] new_message event:', data);
          setLastMessage(data);
        } catch (err) {
          logger.error('[SSE] Failed to parse new_message event:', err);
        }
      });

      eventSource.addEventListener('message_deleted', (event) => {
        if (!isMountedRef.current) return;

        try {
          const data = JSON.parse(event.data);
          logger.debug('[SSE] message_deleted event:', data);
          setLastDeletedMessage(data);
        } catch (err) {
          logger.error('[SSE] Failed to parse message_deleted event:', err);
        }
      });
    } catch (err) {
      logger.error('[SSE] Failed to create EventSource:', err);

      if (isMountedRef.current) {
        reconnectTimeoutRef.current = setTimeout(() => {
          if (isMountedRef.current) {
            connect();
          }
        }, RECONNECT_DELAY_MS);
      }
    }
  }, [cleanup]);

  const reconnect = useCallback(() => {
    logger.debug('[SSE] Manual reconnect requested');
    connect();
  }, [connect]);

  useEffect(() => {
    isMountedRef.current = true;
    connect();

    return () => {
      isMountedRef.current = false;
      cleanup();
      setIsConnected(false);
    };
  }, [connect, cleanup]);

  return {
    isConnected,
    lastMessage,
    lastDeletedMessage,
    reconnect,
  };
};

export default useSSE;
