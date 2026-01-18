import { useState } from 'react';
import Button from "../common/Button.jsx";
import ConfirmModal from '../common/ConfirmModal.jsx';
import Spinner from '../common/Spinner.jsx';
import { useNotification } from '../../hooks/useNotification.js';
import { useBookings } from '../../hooks/useBookings.js';
import { formatDate, formatTime } from '../../utils/dateFormat.js';
import './BookingManagement.css';

export const BookingManagement = () => {
  const [statusFilter, setStatusFilter] = useState('');
  const { bookings, loading, fetchBookings, cancelBooking } = useBookings(
    statusFilter ? { status: statusFilter } : {}
  );
  const [cancelBookingId, setCancelBookingId] = useState(null);
  const notification = useNotification();

  const handleCancelBooking = async () => {
    if (!cancelBookingId) return;

    try {
      await cancelBooking(cancelBookingId);
      notification.success('Запись отменена, кредит возвращен');
      setCancelBookingId(null);
      // fetchBookings вызывается из cancelBooking hook'а, но добавим для синхронизации
      await fetchBookings();
    } catch (error) {
      notification.error(error.message || 'Ошибка отмены записи');
      setCancelBookingId(null);
    }
  };

  // Фильтрация уже происходит в useBookings через параметр filters (уже на русском)
  const filteredBookings = bookings;

  if (loading && bookings.length === 0) {
    return (
      <div className="booking-management-loading">
        <Spinner />
      </div>
    );
  }

  return (
    <div className="booking-management" data-testid="booking-management">
      <div className="booking-management-header">
        <h2 className="booking-management-title">Управление записями</h2>
      </div>

      <div className="booking-management-filters">
        <label className="filter-label">Статус:</label>
        <select
          className="filter-select"
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          data-testid="status-filter"
        >
          <option value="">Все</option>
          <option value="active">Активные</option>
          <option value="cancelled">Отмененные</option>
          <option value="completed">Завершенные</option>
        </select>
      </div>

      <div className="booking-management-table-wrapper">
        <table className="booking-management-table">
          <thead>
            <tr>
              <th>Ученик</th>
              <th>Занятие</th>
              <th>Дата и время</th>
              <th>Статус</th>
              <th>Действия</th>
            </tr>
          </thead>
          <tbody>
            {Array.isArray(filteredBookings) && filteredBookings.length > 0 ? (
              filteredBookings.map((booking) => {
                // API может возвращать либо вложенную структуру (booking.lesson.*), либо плоскую (booking.*) (уже на русском)
                const teacherName = booking.lesson?.teacher_name || booking.teacher_name || 'N/A';
                const startTime = booking.lesson?.start_time || booking.start_time;
                const endTime = booking.lesson?.end_time || booking.end_time;

                return (
                <tr key={booking.id} data-testid="booking-row">
                  <td>{booking.student_name}</td>
                  <td>
                    <div className="booking-lesson-info">
                      <div className="booking-teacher">
                        Препод.: {teacherName}
                      </div>
                    </div>
                  </td>
                  <td>
                    <div className="booking-datetime">
                      <div>{startTime ? formatDate(startTime) : 'N/A'}</div>
                      <div className="booking-time">
                        {startTime && endTime ? `${formatTime(startTime)} - ${formatTime(endTime)}` : 'N/A'}
                      </div>
                    </div>
                  </td>
                  <td>
                    <span className={`status-badge status-${booking.status}`}>
                      {booking.status === 'active' ? 'Активна' :
                       booking.status === 'cancelled' ? 'Отменена' : 'Завершена'}
                    </span>
                  </td>
                  <td>
                    {booking.status === 'active' && (
                      <div className="booking-actions">
                        <Button
                          size="small"
                          variant="danger"
                          onClick={() => setCancelBookingId(booking.id)}
                          data-testid="cancel-booking-button"
                        >
                          Отменить
                        </Button>
                      </div>
                    )}
                  </td>
                </tr>
                );
              })
            ) : (
              <tr>
                <td colSpan="5" className="empty-message">
                  Записи не найдены
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <ConfirmModal
        isOpen={!!cancelBookingId}
        onClose={() => setCancelBookingId(null)}
        onConfirm={handleCancelBooking}
        title="Отмена записи"
        message="Вы действительно хотите отменить эту запись? Кредит будет возвращен ученику."
        confirmText="Отменить"
        confirmVariant="danger"
      />
    </div>
  );
};

export default BookingManagement;
