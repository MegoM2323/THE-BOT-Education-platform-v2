import { useState, useEffect, useCallback } from 'react';
import Modal from "../common/Modal.jsx";
import Button from "../common/Button.jsx";
import Spinner from '../common/Spinner.jsx';
import BroadcastModal from './BroadcastModal.jsx';
import { apiClient } from '../../api/client.js';
import { sendLessonBroadcast } from '../../api/teacher.js';
import { useNotification } from '../../hooks/useNotification.js';
import { formatDateWithDay, formatTime } from '../../utils/dateFormat.js';
import './StudentsList.css';

export const StudentsList = ({ isOpen, onClose, lesson }) => {
  const [students, setStudents] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [showBroadcastModal, setShowBroadcastModal] = useState(false);
  const { showNotification } = useNotification();

  const fetchStudents = useCallback(async () => {
    if (!lesson) return;

    try {
      setLoading(true);
      setError(null);
      const response = await apiClient.get(`/lessons/${lesson.id}/students`);
      // API возвращает { students: [...], count: ... }
      const studentsList = response?.students || response || [];

      // Нормализуем данные студентов - могут быть в разных форматах
      const normalizedStudents = Array.isArray(studentsList)
        ? studentsList.map(student => ({
            id: student.id,
            full_name: student.full_name || student.student_name || student.name || 'Неизвестный студент',
            email: student.email || student.student_email || 'Нет email',
            booking_created_at: student.booking_created_at || student.created_at || null
          }))
        : [];

      setStudents(normalizedStudents);
    } catch (err) {
      setError(err.message || 'Не удалось загрузить список учеников');
      setStudents([]);
    } finally {
      setLoading(false);
    }
  }, [lesson]);

  useEffect(() => {
    if (isOpen && lesson) {
      fetchStudents();
    }
  }, [isOpen, lesson, fetchStudents]);

  // Обработчик отправки рассылки
  const handleSendBroadcast = async (message) => {
    try {
      const response = await sendLessonBroadcast(lesson.id, message);
      const sentCount = response?.sent_count || students.length;
      showNotification(`Рассылка отправлена ${sentCount} студентам`, 'success');
    } catch (error) {
      throw error; // Пробрасываем ошибку в BroadcastModal для обработки
    }
  };

  if (!lesson) return null;

  return (
    <>
      <Modal
        isOpen={isOpen}
        onClose={onClose}
        title="Список учеников"
        data-testid="students-list"
        footer={
          students.length > 0 && (
            <div className="students-list-footer">
              <Button
                onClick={() => setShowBroadcastModal(true)}
                variant="primary"
                size="medium"
                data-testid="open-broadcast-modal"
              >
                📢 Отправить рассылку
              </Button>
            </div>
          )
        }
      >
      <div className="students-list-content">
        <div className="students-list-lesson-info">
          <div className="lesson-info-details">
            <span>{formatDateWithDay(lesson.start_time)}</span>
            <span>
              {formatTime(lesson.start_time)} - {formatTime(lesson.end_time)}
            </span>
          </div>
        </div>

        {loading ? (
          <div className="students-list-loading">
            <Spinner />
          </div>
        ) : error ? (
          <div className="students-list-error">
            <p>Ошибка загрузки списка учеников: {error instanceof Error ? error.message : (error?.message || String(error))}</p>
          </div>
        ) : students.length === 0 ? (
          <div className="students-list-empty">
            <p>На занятие пока никто не записан</p>
          </div>
        ) : (
          <div className="students-list-table-wrapper">
            <table className="students-list-table">
              <thead>
                <tr>
                  <th>Имя</th>
                  <th>Email</th>
                  <th>Дата записи</th>
                </tr>
              </thead>
              <tbody>
                {students.map((student) => (
                  <tr key={student.id} data-testid="student-row">
                    <td>
                      <div className="student-name">{student.full_name}</div>
                    </td>
                    <td>
                      <div className="student-email">{student.email}</div>
                    </td>
                    <td>
                      <div className="student-booking-date">
                        {student.booking_created_at
                          ? formatDateWithDay(student.booking_created_at)
                          : '—'}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        <div className="students-list-summary" data-testid="enrolled-count">
          <span className="summary-label">Всего записано:</span>
          <span className="summary-value">
            {students.length} из {lesson.max_students}
          </span>
        </div>
      </div>
      </Modal>

      {/* Модальное окно рассылки */}
      <BroadcastModal
        isOpen={showBroadcastModal}
        onClose={() => setShowBroadcastModal(false)}
        lesson={lesson}
        students={students}
        onSend={handleSendBroadcast}
      />
    </>
  );
};

export default StudentsList;
