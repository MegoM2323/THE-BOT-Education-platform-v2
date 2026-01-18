import { useRef, useEffect, useCallback } from 'react';
import { useNotification } from '../../hooks/useNotification.js';
import './Notification.css';

const NotificationItem = ({ notification, isExiting, isCollapsing, onDismiss }) => {
  const elementRef = useRef(null);
  const heightRef = useRef(null);

  useEffect(() => {
    if (elementRef.current && !heightRef.current) {
      heightRef.current = elementRef.current.offsetHeight;
    }
  }, []);

  useEffect(() => {
    if (isCollapsing && elementRef.current && heightRef.current) {
      elementRef.current.style.height = `${heightRef.current}px`;
      requestAnimationFrame(() => {
        if (elementRef.current) {
          elementRef.current.style.height = '0px';
          elementRef.current.style.padding = '0';
          elementRef.current.style.margin = '0';
        }
      });
    }
  }, [isCollapsing]);

  const handleDismiss = useCallback(() => {
    onDismiss(notification.id);
  }, [notification.id, onDismiss]);

  const className = [
    'notification',
    `notification-${notification.type}`,
    isExiting ? 'notification-exit' : '',
    isCollapsing ? 'notification-collapsing' : '',
  ].filter(Boolean).join(' ');

  return (
    <div ref={elementRef} className={className}>
      <div className="notification-content">
        <span className="notification-message">{notification.message}</span>
      </div>
      <button
        className="notification-close"
        onClick={handleDismiss}
        aria-label="Закрыть"
      >
        &times;
      </button>
    </div>
  );
};

export const Notification = () => {
  const { notifications, exitingIds, collapsingIds, dismissNotification } = useNotification();

  if (notifications.length === 0) return null;

  return (
    <div className="notifications-container">
      {notifications.map((notification) => (
        <NotificationItem
          key={notification.id}
          notification={notification}
          isExiting={exitingIds.has(notification.id)}
          isCollapsing={collapsingIds.has(notification.id)}
          onDismiss={dismissNotification}
        />
      ))}
    </div>
  );
};

export default Notification;
