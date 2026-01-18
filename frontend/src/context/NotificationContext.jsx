import { createContext, useState, useCallback, useRef, useEffect, useMemo } from 'react';

export const NotificationContext = createContext(null);

let notificationId = 0;

export const NotificationProvider = ({ children }) => {
  const [notifications, setNotifications] = useState([]);
  const [exitingIds, setExitingIds] = useState(new Set());
  const [collapsingIds, setCollapsingIds] = useState(new Set());
  const timersRef = useRef(new Map());
  const animationTimersRef = useRef(new Map());

  const removeNotification = useCallback((id) => {
    if (timersRef.current.has(id)) {
      clearTimeout(timersRef.current.get(id));
      timersRef.current.delete(id);
    }
    setNotifications((prev) => prev.filter((n) => n.id !== id));
  }, []);

  const animatedRemove = useCallback((id) => {
    if (exitingIds.has(id) || collapsingIds.has(id)) return;

    setExitingIds((prev) => new Set(prev).add(id));

    const bounceTimer = setTimeout(() => {
      setExitingIds((prev) => {
        const newSet = new Set(prev);
        newSet.delete(id);
        return newSet;
      });
      setCollapsingIds((prev) => new Set(prev).add(id));

      const collapseTimer = setTimeout(() => {
        removeNotification(id);
        setCollapsingIds((prev) => {
          const newSet = new Set(prev);
          newSet.delete(id);
          return newSet;
        });
        animationTimersRef.current.delete(`${id}-collapse`);
      }, 300);

      animationTimersRef.current.set(`${id}-collapse`, collapseTimer);
      animationTimersRef.current.delete(id);
    }, 500);

    animationTimersRef.current.set(id, bounceTimer);
  }, [removeNotification, exitingIds, collapsingIds]);

  const dismissNotification = useCallback((id) => {
    if (timersRef.current.has(id)) {
      clearTimeout(timersRef.current.get(id));
      timersRef.current.delete(id);
    }
    animatedRemove(id);
  }, [animatedRemove]);

  const addNotification = useCallback((message, type = 'info', duration = 5000) => {
    const id = ++notificationId;
    const notification = { id, message, type };

    setNotifications((prev) => [...prev, notification]);

    if (duration > 0) {
      const timerId = setTimeout(() => {
        animatedRemove(id);
      }, duration);

      timersRef.current.set(id, timerId);
    }

    return id;
  }, [animatedRemove]);

  const clearAll = useCallback(() => {
    timersRef.current.forEach((timerId) => {
      clearTimeout(timerId);
    });
    timersRef.current.clear();

    animationTimersRef.current.forEach((timerId) => {
      clearTimeout(timerId);
    });
    animationTimersRef.current.clear();

    setExitingIds(new Set());
    setCollapsingIds(new Set());
    setNotifications([]);
  }, []);

  const success = useCallback(
    (message, duration) => addNotification(message, 'success', duration ?? 5000),
    [addNotification]
  );

  const error = useCallback(
    (message, duration) => addNotification(message, 'error', duration ?? 5000),
    [addNotification]
  );

  const warning = useCallback(
    (message, duration) => addNotification(message, 'warning', duration ?? 5000),
    [addNotification]
  );

  const info = useCallback(
    (message, duration) => addNotification(message, 'info', duration ?? 5000),
    [addNotification]
  );

  const showNotification = useCallback(
    (message, type = 'info', duration = 5000) => addNotification(message, type, duration),
    [addNotification]
  );

  useEffect(() => {
    const timers = timersRef.current;
    const animationTimers = animationTimersRef.current;
    return () => {
      timers.forEach((timerId) => clearTimeout(timerId));
      timers.clear();
      animationTimers.forEach((timerId) => clearTimeout(timerId));
      animationTimers.clear();
    };
  }, []);

  const api = useMemo(() => ({
    addNotification,
    removeNotification,
    dismissNotification,
    clearAll,
    success,
    error,
    warning,
    info,
    showNotification,
  }), [addNotification, removeNotification, dismissNotification, clearAll, success, error, warning, info, showNotification]);

  const value = useMemo(() => ({
    ...api,
    notifications,
    exitingIds,
    collapsingIds,
  }), [api, notifications, exitingIds, collapsingIds]);

  return (
    <NotificationContext.Provider value={value}>
      {children}
    </NotificationContext.Provider>
  );
};

export default NotificationContext;
