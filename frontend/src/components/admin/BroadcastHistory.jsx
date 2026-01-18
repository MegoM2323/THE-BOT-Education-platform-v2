import { useState, useEffect } from 'react';
import Button from "../common/Button.jsx";
import Modal from "../common/Modal.jsx";
import Spinner from '../common/Spinner.jsx';
import { useBroadcast } from '../../hooks/useBroadcast.js';
import './BroadcastHistory.css';

const BROADCASTS_PER_PAGE = 20;

/**
 * Компонент истории рассылок
 */
export const BroadcastHistory = () => {
  const {
    broadcasts,
    totalBroadcasts,
    loading,
    fetchBroadcasts,
    getBroadcastDetailsById,
    cancelBroadcastById,
  } = useBroadcast();

  const [currentPage, setCurrentPage] = useState(0);
  const [selectedBroadcast, setSelectedBroadcast] = useState(null);
  const [broadcastDetails, setBroadcastDetails] = useState(null);
  const [showDetailsModal, setShowDetailsModal] = useState(false);
  const [loadingDetails, setLoadingDetails] = useState(false);
  const [detailsError, setDetailsError] = useState(null);
  const [cancelError, setCancelError] = useState(null);

  useEffect(() => {
    loadBroadcasts(0);
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const loadBroadcasts = async (page) => {
    const offset = page * BROADCASTS_PER_PAGE;
    await fetchBroadcasts(BROADCASTS_PER_PAGE, offset);
    setCurrentPage(page);
  };

  const handleOpenDetails = async (broadcast) => {
    setSelectedBroadcast(broadcast);
    setShowDetailsModal(true);
    setLoadingDetails(true);
    setDetailsError(null);

    try {
      const details = await getBroadcastDetailsById(broadcast.id);
      setBroadcastDetails(details);
    } catch (err) {
      setDetailsError(err.message || 'Ошибка при загрузке деталей рассылки');
    } finally {
      setLoadingDetails(false);
    }
  };

  const handleCloseDetails = () => {
    setShowDetailsModal(false);
    setSelectedBroadcast(null);
    setBroadcastDetails(null);
    setDetailsError(null);
  };

  const handleCancelBroadcast = async (broadcastId) => {
    if (window.confirm('Вы действительно хотите отменить эту рассылку?')) {
      setCancelError(null);
      try {
        await cancelBroadcastById(broadcastId);
        loadBroadcasts(currentPage);
        handleCloseDetails();
      } catch (err) {
        setCancelError(err.message || 'Ошибка при отмене рассылки');
      }
    }
  };

  const getStatusLabel = (status) => {
    const labels = {
      pending: 'Ожидание',
      in_progress: 'Отправка',
      processing: 'Отправка',
      completed: 'Завершена',
      cancelled: 'Отменена',
      failed: 'Ошибка',
    };
    return labels[status] || status;
  };

  const getStatusClass = (status) => {
    const classes = {
      pending: 'status-pending',
      in_progress: 'status-processing',
      processing: 'status-processing',
      completed: 'status-completed',
      cancelled: 'status-cancelled',
      failed: 'status-failed',
    };
    return classes[status] || '';
  };

  const formatDate = (dateString) => {
    if (!dateString) return '-';
    const date = new Date(dateString);
    if (isNaN(date.getTime())) return 'Неверная дата';
    return date.toLocaleString('ru-RU', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const truncateMessage = (message, maxLength = 100) => {
    if (message.length <= maxLength) return message;
    return message.substring(0, maxLength) + '...';
  };

  const totalPages = Math.ceil(totalBroadcasts / BROADCASTS_PER_PAGE);

  return (
    <div className="broadcast-history">
      <div className="broadcast-history-header">
        <h3 className="broadcast-history-title">История рассылок</h3>
        <div className="total-broadcasts">
          Всего рассылок: {totalBroadcasts}
        </div>
      </div>

      {cancelError && (
        <div className="broadcast-error-message">
          <svg className="error-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path
              d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"
              fill="currentColor"
            />
          </svg>
          <span>{cancelError}</span>
        </div>
      )}

      {loading && broadcasts.length === 0 ? (
        <div className="broadcast-history-loading">
          <Spinner />
        </div>
      ) : broadcasts.length === 0 ? (
        <div className="empty-state">
          <p>История рассылок пуста</p>
          <p className="empty-state-hint">Отправленные рассылки будут отображаться здесь</p>
        </div>
      ) : (
        <>
          <div className="broadcasts-table-wrapper">
            <table className="broadcasts-table">
              <thead>
                <tr>
                  <th>Дата/Время</th>
                  <th>Список</th>
                  <th>Сообщение</th>
                  <th>Статус</th>
                  <th>Успешно</th>
                  <th>Ошибок</th>
                  <th>Действия</th>
                </tr>
              </thead>
              <tbody>
                {broadcasts.map((broadcast) => (
                  <tr
                    key={broadcast.id}
                    onClick={() => handleOpenDetails(broadcast)}
                    className="broadcast-row"
                  >
                    <td>{formatDate(broadcast.created_at)}</td>
                    <td>{broadcast.list_name || (broadcast.list_id ? `ID: ${broadcast.list_id.slice(0, 8)}...` : 'Индивидуальная')}</td>
                    <td className="message-cell">
                      {truncateMessage(broadcast.message)}
                    </td>
                    <td>
                      <span className={`status-badge ${getStatusClass(broadcast.status)}`}>
                        {getStatusLabel(broadcast.status)}
                      </span>
                    </td>
                    <td className="success-count">{broadcast.sent_count || 0}</td>
                    <td className="error-count">{broadcast.failed_count || 0}</td>
                    <td onClick={(e) => e.stopPropagation()}>
                      <div className="row-actions">
                        <Button
                          size="small"
                          variant="outline"
                          onClick={() => handleOpenDetails(broadcast)}
                        >
                          Детали
                        </Button>
                        {(broadcast.status === 'pending' || broadcast.status === 'in_progress' || broadcast.status === 'processing') && (
                          <Button
                            size="small"
                            variant="danger"
                            onClick={() => handleCancelBroadcast(broadcast.id)}
                          >
                            Отменить
                          </Button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {totalPages > 1 && (
            <div className="pagination">
              <Button
                variant="outline"
                size="small"
                onClick={() => loadBroadcasts(currentPage - 1)}
                disabled={currentPage === 0}
              >
                Назад
              </Button>
              <span className="page-info">
                Страница {currentPage + 1} из {totalPages}
              </span>
              <Button
                variant="outline"
                size="small"
                onClick={() => loadBroadcasts(currentPage + 1)}
                disabled={currentPage >= totalPages - 1}
              >
                Вперед
              </Button>
            </div>
          )}
        </>
      )}

      {/* Модальное окно деталей рассылки */}
      <Modal
        isOpen={showDetailsModal}
        onClose={handleCloseDetails}
        title={`Детали рассылки #${selectedBroadcast?.id}`}
      >
        {loadingDetails ? (
          <div className="details-loading">
            <Spinner />
          </div>
        ) : detailsError ? (
          <div className="details-error">
            <svg className="error-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path
                d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"
                fill="currentColor"
              />
            </svg>
            <p>{detailsError}</p>
          </div>
        ) : broadcastDetails ? (
          <div className="broadcast-details">
            <div className="details-section">
              <h4 className="details-section-title">Информация о рассылке</h4>
              <div className="details-grid">
                <div className="detail-item">
                  <span className="detail-label">Дата создания:</span>
                  <span className="detail-value">{formatDate(broadcastDetails.created_at)}</span>
                </div>
                <div className="detail-item">
                  <span className="detail-label">Список:</span>
                  <span className="detail-value">{broadcastDetails.list_name || (broadcastDetails.list_id ? `ID: ${broadcastDetails.list_id.slice(0, 8)}...` : 'Индивидуальная')}</span>
                </div>
                <div className="detail-item">
                  <span className="detail-label">Статус:</span>
                  <span className={`status-badge ${getStatusClass(broadcastDetails.status)}`}>
                    {getStatusLabel(broadcastDetails.status)}
                  </span>
                </div>
                <div className="detail-item">
                  <span className="detail-label">Всего получателей:</span>
                  <span className="detail-value">{broadcastDetails.total_recipients || 0}</span>
                </div>
                <div className="detail-item">
                  <span className="detail-label">Успешно доставлено:</span>
                  <span className="detail-value success-count">{broadcastDetails.sent_count || 0}</span>
                </div>
                <div className="detail-item">
                  <span className="detail-label">Ошибок:</span>
                  <span className="detail-value error-count">{broadcastDetails.failed_count || 0}</span>
                </div>
              </div>
            </div>

            <div className="details-section">
              <h4 className="details-section-title">Текст сообщения</h4>
              <div className="message-box">
                {broadcastDetails.message}
              </div>
            </div>

            {broadcastDetails.deliveries && broadcastDetails.deliveries.length > 0 && (
              <div className="details-section">
                <h4 className="details-section-title">Статус доставки</h4>
                <div className="deliveries-table-wrapper">
                  <table className="deliveries-table">
                    <thead>
                      <tr>
                        <th>Получатель</th>
                        <th>Telegram ID</th>
                        <th>Статус</th>
                        <th>Время</th>
                        <th>Ошибка</th>
                      </tr>
                    </thead>
                    <tbody>
                      {broadcastDetails.deliveries.map((delivery, index) => (
                        <tr key={index}>
                          <td>{delivery.user_name || '-'}</td>
                          <td><code>{delivery.telegram_id}</code></td>
                          <td>
                            <span className={`delivery-status ${delivery.status === 'sent' ? 'status-success' : 'status-error'}`}>
                              {delivery.status === 'sent' ? 'Доставлено' : 'Ошибка'}
                            </span>
                          </td>
                          <td>{delivery.sent_at ? formatDate(delivery.sent_at) : '-'}</td>
                          <td className="error-cell">{delivery.error || '-'}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            )}
          </div>
        ) : (
          <div className="details-error">
            Не удалось загрузить детали рассылки
          </div>
        )}
      </Modal>
    </div>
  );
};

export default BroadcastHistory;
