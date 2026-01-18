import { useState, useEffect, useCallback, useMemo } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { FiInfo } from 'react-icons/fi';
import Button from "../common/Button.jsx";
import Input from '../common/Input.jsx';
import Modal from "../common/Modal.jsx";
import Spinner from '../common/Spinner.jsx';
import StudentCreditsHistoryModal from '../common/StudentCreditsHistoryModal.jsx';
import { useNotification } from '../../hooks/useNotification.js';
import * as usersAPI from '../../api/users.js';
import * as creditsAPI from '../../api/credits.js';
import './CreditsManagement.css';

export const CreditsManagement = () => {
  const [students, setStudents] = useState([]);
  const [loading, setLoading] = useState(true);
  const [selectedStudent, setSelectedStudent] = useState(null);
  const [showCreditsModal, setShowCreditsModal] = useState(false);
  const [operation, setOperation] = useState('add');
  const [amount, setAmount] = useState('');
  const [reason, setReason] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [sortColumn, setSortColumn] = useState('balance');
  const [sortDirection, setSortDirection] = useState('asc');
  const [showHistoryModal, setShowHistoryModal] = useState(false);
  const [selectedStudentForHistory, setSelectedStudentForHistory] = useState(null);
  const notification = useNotification();
  const queryClient = useQueryClient();

  const fetchStudents = useCallback(async () => {
    try {
      setLoading(true);
      const data = await usersAPI.getStudentsAll();
      // Получить кредиты для каждого студента
      const studentsWithCredits = await Promise.all(
        data.map(async (student) => {
          try {
            const credits = await creditsAPI.getUserCredits(student.id);
            return { ...student, credits: credits.balance };
          } catch (error) {
            return { ...student, credits: 0 };
          }
        })
      );
      setStudents(studentsWithCredits);
    } catch (error) {
      notification.error('Ошибка загрузки студентов');
    } finally {
      setLoading(false);
    }
  }, [notification]);

  useEffect(() => {
    fetchStudents();
  }, [fetchStudents]);

  const handleOpenCreditsModal = (student, op) => {
    setSelectedStudent(student);
    setOperation(op);
    setAmount('');
    setReason('');
    setShowCreditsModal(true);
  };

  const handleCloseModal = () => {
    setShowCreditsModal(false);
    setSelectedStudent(null);
    setAmount('');
    setReason('');
  };

  const handleViewHistory = (student) => {
    setSelectedStudentForHistory(student);
    setShowHistoryModal(true);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!selectedStudent || !amount) return;

    setSubmitting(true);
    try {
      const numAmount = parseInt(amount, 10);
      if (isNaN(numAmount) || numAmount <= 0) {
        throw new Error('Сумма должна быть положительным числом');
      }

      // Проверка на недостаток кредитов при снятии
      if (operation === 'deduct' && numAmount > selectedStudent.credits) {
        throw new Error(`Недостаточно кредитов. Текущий баланс: ${selectedStudent.credits}`);
      }

      // Reason может быть пустым, backend будет использовать значение по умолчанию
      const reasonText = reason.trim() || `${operation === 'add' ? 'Начисление' : 'Списание'} админом`;

      if (operation === 'add') {
        await creditsAPI.addCredits(selectedStudent.id, numAmount, reasonText);
        notification.success(`Начислено ${numAmount} кредитов студенту ${selectedStudent.full_name}`);
      } else {
        await creditsAPI.deductCredits(selectedStudent.id, numAmount, reasonText);
        notification.success(`Списано ${numAmount} кредитов у студента ${selectedStudent.full_name}`);
      }

      // Инвалидируем кэш React Query для обновления sidebar, истории и других компонентов
      queryClient.invalidateQueries({ queryKey: ['credits'] });
      queryClient.invalidateQueries({ queryKey: ['credits', 'history'] });

      handleCloseModal();
      // Обновляем локальный список студентов
      await fetchStudents();
    } catch (error) {
      // Обработка специфичных ошибок от backend
      let errorMessage = error.message || 'Ошибка операции с кредитами';

      // Проверка на ошибку недостатка кредитов от backend
      if (error.message?.toLowerCase().includes('insufficient') ||
          error.data?.error?.code === 'INSUFFICIENT_CREDITS') {
        errorMessage = `Недостаточно кредитов. Текущий баланс: ${selectedStudent.credits}`;
      }

      notification.error(errorMessage);
    } finally {
      setSubmitting(false);
    }
  };

  const handleSort = (column) => {
    if (sortColumn === column) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortColumn(column);
      setSortDirection('asc');
    }
  };

  const sortedStudents = useMemo(() => {
    const sorted = [...students].sort((a, b) => {
      let comparison = 0;

      if (sortColumn === 'balance') {
        comparison = a.credits - b.credits;
      } else if (sortColumn === 'name') {
        comparison = a.full_name.localeCompare(b.full_name, 'ru-RU');
      }

      return sortDirection === 'asc' ? comparison : -comparison;
    });

    return sorted;
  }, [students, sortColumn, sortDirection]);

  if (loading) {
    return (
      <div className="credits-management-loading">
        <Spinner />
      </div>
    );
  }

  return (
    <div className="credits-management" data-testid="credits-management">
      <div className="credits-management-header">
        <h2 className="credits-management-title">Управление кредитами</h2>
      </div>

      <div className="credits-management-table-wrapper">
        <table className="credits-management-table">
          <thead>
            <tr>
              <th
                onClick={() => handleSort('name')}
                style={{ cursor: 'pointer' }}
              >
                Имя{sortColumn === 'name' && (sortDirection === 'asc' ? ' ↑' : ' ↓')}
              </th>
              <th>Email</th>
              <th
                onClick={() => handleSort('balance')}
                style={{ cursor: 'pointer' }}
              >
                Баланс{sortColumn === 'balance' && (sortDirection === 'asc' ? ' ↑' : ' ↓')}
              </th>
              <th>Действия</th>
            </tr>
          </thead>
          <tbody>
            {students.length > 0 ? (
              sortedStudents.map((student) => (
                <tr key={student.id}>
                  <td>{student.full_name}</td>
                  <td>{student.email}</td>
                  <td>
                    <span className="balance-badge">
                      {student.credits} кредитов
                    </span>
                  </td>
                  <td>
                    <div className="credits-actions">
                      <button
                        type="button"
                        className="credits-action-btn credits-action-btn-add"
                        onClick={() => handleOpenCreditsModal(student, 'add')}
                        data-testid="add-credits-button"
                        aria-label="Добавить кредиты"
                      >
                        +
                      </button>
                      <button
                        type="button"
                        className="credits-action-btn credits-action-btn-deduct"
                        onClick={() => handleOpenCreditsModal(student, 'deduct')}
                        aria-label="Списать кредиты"
                      >
                        -
                      </button>
                      <button
                        type="button"
                        className="credits-action-btn credits-action-btn-history"
                        onClick={() => handleViewHistory(student)}
                        data-testid="view-history-button"
                        aria-label="Просмотреть историю кредитов"
                      >
                        <FiInfo size={20} />
                      </button>
                    </div>
                  </td>
                </tr>
              ))
            ) : (
              <tr>
                <td colSpan="4" className="empty-message">
                  Студенты не найдены
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      <Modal
        isOpen={showCreditsModal}
        onClose={handleCloseModal}
        title={operation === 'add' ? 'Начислить кредиты' : 'Списать кредиты'}
      >
        {selectedStudent && (
          <form onSubmit={handleSubmit} className="credits-form" data-testid="credits-form">
            <div className="student-info">
              <strong>{selectedStudent.full_name}</strong>
              <span className="current-balance">
                Текущий баланс: {selectedStudent.credits} кредитов
              </span>
            </div>
            <div data-testid="amount-input">
              <Input
                label="Количество кредитов"
                name="amount"
                type="number"
                min="1"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                required
              />
            </div>
            <div className="form-group">
              <label className="form-label" htmlFor="reason">Причина (необязательно)</label>
              <textarea
                id="reason"
                name="reason"
                className="form-textarea"
                value={reason}
                onChange={(e) => setReason(e.target.value)}
                rows="3"
                placeholder="Укажите причину операции..."
              />
            </div>
            <div className="credits-form-actions">
              <Button type="button" variant="secondary" onClick={handleCloseModal}>
                Отмена
              </Button>
              <Button
                type="submit"
                variant={operation === 'add' ? 'primary' : 'danger'}
                loading={submitting}
              >
                {operation === 'add' ? 'Начислить' : 'Списать'}
              </Button>
            </div>
          </form>
        )}
      </Modal>

      <StudentCreditsHistoryModal
        isOpen={showHistoryModal}
        onClose={() => {
          setShowHistoryModal(false);
          setSelectedStudentForHistory(null);
        }}
        student={selectedStudentForHistory}
      />
    </div>
  );
};

export default CreditsManagement;
