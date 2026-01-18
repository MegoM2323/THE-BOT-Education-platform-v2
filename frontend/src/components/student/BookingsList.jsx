import { useState } from 'react';
import { logger } from '../../utils/logger.js';
import ConfirmModal from '../common/ConfirmModal.jsx';
import Button from "../common/Button.jsx";
import './BookingsList.css';

/**
 * Format date and time for display
 */
const formatDateTime = (dateString) => {
  const date = new Date(dateString);
  const dateStr = date.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  });
  const timeStr = date.toLocaleTimeString('en-US', {
    hour: '2-digit',
    minute: '2-digit',
  });
  return `${dateStr} at ${timeStr}`;
};

/**
 * Check if booking can be cancelled
 */
const canCancelBooking = (booking) => {
  if (booking.status !== 'active') return false;

  const now = new Date();
  const lessonTime = new Date(booking.lesson?.start_time || booking.start_time);
  const hoursUntilLesson = (lessonTime - now) / (1000 * 60 * 60);

  // Must be at least 24 hours before lesson
  return hoursUntilLesson >= 24;
};

/**
 * BookingsList - Display list of bookings with cancel functionality
 */
export const BookingsList = ({ bookings, onCancel, readonly = false }) => {
  const [showCancelModal, setShowCancelModal] = useState(false);
  const [selectedBooking, setSelectedBooking] = useState(null);
  const [isCancelling, setIsCancelling] = useState(false);

  const handleCancelClick = (booking) => {
    setSelectedBooking(booking);
    setShowCancelModal(true);
  };

  const handleConfirmCancel = async () => {
    if (!selectedBooking) return;

    setIsCancelling(true);
    try {
      await onCancel(selectedBooking.id);
      setShowCancelModal(false);
      setSelectedBooking(null);
    } catch (error) {
      console.error('Cancel failed:', error);
      // Error handled by mutation
    } finally {
      setIsCancelling(false);
    }
  };

  const handleCloseModal = () => {
    setShowCancelModal(false);
    setSelectedBooking(null);
  };

  if (!bookings || bookings.length === 0) {
    return (
      <div className="bookings-list-empty">
        <p>No bookings found</p>
      </div>
    );
  }

  return (
    <>
      <div className="bookings-list" data-testid="bookings-list">
        <table className="bookings-table">
          <thead>
            <tr>
              <th>Lesson</th>
              <th>Teacher</th>
              <th>Time</th>
              <th>Status</th>
              {!readonly && <th>Actions</th>}
            </tr>
          </thead>
          <tbody>
            {bookings.map(booking => {
              const lesson = booking.lesson || {};
              const lessonTime = lesson.start_time || booking.start_time;
              const canCancel = !readonly && canCancelBooking(booking);
              const status = booking.status || 'active';

              const lessonColor = lesson.color || '#3b82f6';
              const rowStyle = {
                borderLeft: `4px solid ${lessonColor}`,
                background: `${lessonColor}08`,
              };

              return (
                <tr
                  key={booking.id}
                  data-testid="booking-row"
                  style={rowStyle}
                >
                  <td>{lesson.subject || 'Lesson'}</td>
                  <td>{lesson.teacher_name || 'Teacher'}</td>
                  <td>{formatDateTime(lessonTime)}</td>
                  <td>
                    <span className={`status-badge ${status}`}>
                      {status === 'active' ? 'Active' : 'Cancelled'}
                    </span>
                  </td>
                  {!readonly && (
                    <td>
                      {canCancel ? (
                        <Button
                          onClick={() => handleCancelClick(booking)}
                          variant="danger"
                          size="small"
                          data-testid="cancel-button"
                        >
                          Cancel
                        </Button>
                      ) : (
                        <span className="cancel-disabled">
                          {status === 'cancelled' ? 'Cancelled' : 'Cannot cancel'}
                        </span>
                      )}
                    </td>
                  )}
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      <ConfirmModal
        isOpen={showCancelModal}
        onClose={handleCloseModal}
        onConfirm={handleConfirmCancel}
        title="Cancel Booking"
        message={
          selectedBooking
            ? `Are you sure you want to cancel your booking for "${
                selectedBooking.lesson?.subject || 'this lesson'
              }" with ${
                selectedBooking.lesson?.teacher_name || 'the teacher'
              }? Your credit will be refunded.`
            : 'Loading...'
        }
        confirmText="Yes, Cancel Booking"
        cancelText="No, Keep Booking"
        confirmVariant="danger"
        loading={isCancelling}
      />
    </>
  );
};

export default BookingsList;
