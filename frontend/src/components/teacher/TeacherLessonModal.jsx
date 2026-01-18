import { useState, useEffect } from 'react';
import Modal from "../common/Modal.jsx";
import Button from "../common/Button.jsx";
import Spinner from '../common/Spinner.jsx';
import HomeworkSection from '../admin/HomeworkSection.jsx';
import BroadcastSection from '../admin/BroadcastSection.jsx';
import { formatTime } from '../../utils/dateFormat.js';
import * as bookingAPI from '../../api/bookings.js';
import * as lessonsAPI from '../../api/lessons.js';
import './TeacherLessonModal.css';

/**
 * Модал для просмотра занятия преподавателем
 * READ-ONLY для основных параметров
 * Можно управлять: Homework, Broadcasts
 * Нельзя: редактировать время, студентов, удалять занятие
 */
export const TeacherLessonModal = ({ isOpen, onClose, lesson }) => {
  const [activeTab, setActiveTab] = useState('info');
  const [students, setStudents] = useState([]);
  const [loading, setLoading] = useState(false);
  const [homeworkCount, setHomeworkCount] = useState(0);
  const [lessonDetails, setLessonDetails] = useState(lesson);
  const [lessonDetailsLoading, setLessonDetailsLoading] = useState(false);

  const currentLesson = lessonDetails || lesson;

  // Проверка: прошло ли занятие
  const isPast = currentLesson ? new Date(currentLesson.start_time) < new Date() : false;

  // Загрузка студентов при открытии модала
  useEffect(() => {
    if (isOpen && currentLesson?.id) {
      loadStudents();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isOpen, currentLesson?.id]);

  useEffect(() => {
    if (!isOpen || !lesson?.id) {
      return;
    }

    setLessonDetails(lesson);

    const fetchLesson = async () => {
      try {
        setLessonDetailsLoading(true);
        const data = await lessonsAPI.getLessonById(lesson.id);
        const freshLesson = data?.lesson || data;
        if (freshLesson) {
          setLessonDetails(freshLesson);
        }
      } catch (error) {
        console.error('Ошибка загрузки занятия:', error);
      } finally {
        setLessonDetailsLoading(false);
      }
    };

    fetchLesson();
  }, [isOpen, lesson]);

  /**
   * Загрузить список студентов на занятии
   */
  const loadStudents = async () => {
    if (!currentLesson?.id) return;
    try {
      setLoading(true);

      const studentsResponse = await bookingAPI.getBookings({ lesson_id: currentLesson.id });

      // Фильтруем активные бронирования
      const activeBookings = Array.isArray(studentsResponse)
        ? studentsResponse.filter(b => b.status === 'active')
        : [];

      const studentsList = activeBookings.map(booking => ({
        id: booking.id,
        student_id: booking.student_id || booking.user_id,
        student_name: booking.student_name || booking.user_name || 'Неизвестный студент',
        student_email: booking.student_email || booking.user_email || 'Нет email'
      }));

      setStudents(studentsList);
    } catch (error) {
      console.error('Ошибка загрузки студентов:', error);
    } finally {
      setLoading(false);
    }
  };

  if (!currentLesson) return null;

  const lessonDate = new Date(currentLesson.start_time);

  const handleHomeworkSaved = (text) => {
    setLessonDetails((prev) => {
      if (prev) {
        return { ...prev, homework_text: text };
      }
      if (lesson) {
        return { ...lesson, homework_text: text };
      }
      return prev;
    });
  };

  const isAnyLoading = loading || lessonDetailsLoading;

  /**
   * Получить заголовок модального окна с бейджами
   */
  const getModalTitle = () => {
    return (
      <div className="teacher-lesson-modal-title">
        <span>Информация о занятии</span>
        {homeworkCount > 0 && (
          <span className="homework-badge" title={`${homeworkCount} файлов домашнего задания`}>
            📎 ДЗ ({homeworkCount})
          </span>
        )}
        {isPast && (
          <span className="past-lesson-badge" title="Занятие уже прошло">
            ⏰ Завершено
          </span>
        )}
      </div>
    );
  };

  // Вкладки
  const tabs = [
    { id: 'info', label: 'Информация' },
    { id: 'students', label: `Студенты (${students.length})` },
    { id: 'homework', label: 'Домашнее задание' },
    { id: 'broadcasts', label: 'Рассылки' },
  ];

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={getModalTitle()}
      footer={
        <div className="teacher-lesson-modal-footer">
          <Button variant="secondary" onClick={onClose}>
            Закрыть
          </Button>
        </div>
      }
    >
      {isPast && (
        <div className="teacher-lesson-warning">
          ⚠️ Это занятие уже прошло. Вы можете просматривать информацию, но не можете вносить изменения.
        </div>
      )}

      {/* Tab Navigation */}
      <div className="teacher-lesson-tabs">
        {tabs.map(tab => (
          <button
            key={tab.id}
            className={`teacher-lesson-tab ${activeTab === tab.id ? 'active' : ''}`}
            onClick={() => setActiveTab(tab.id)}
          >
            {tab.label}
            {tab.id === 'homework' && homeworkCount > 0 && (
              <span className="tab-badge">{homeworkCount}</span>
            )}
          </button>
        ))}
      </div>

      {/* Tab Content */}
      <div className="teacher-lesson-content">
        {isAnyLoading ? (
          <div className="teacher-lesson-loading">
            <Spinner />
          </div>
        ) : (
          <>
            {/* Информация о занятии (READ-ONLY) */}
            {activeTab === 'info' && (
              <section className="teacher-lesson-section">
                <h4 className="section-title">Основная информация</h4>
                <div className="lesson-info-grid">
                  <div className="info-item">
                    <label>Дата</label>
                    <span className="info-value">
                      {lessonDate.toLocaleDateString('ru-RU', {
                        weekday: 'long',
                        year: 'numeric',
                        month: 'long',
                        day: 'numeric'
                      })}
                    </span>
                  </div>
                  <div className="info-item">
                    <label>Время</label>
                    <span className="info-value">
                      {formatTime(currentLesson.start_time)} - {formatTime(currentLesson.end_time)}
                    </span>
                  </div>
                  <div className="info-item">
                    <label>Записано студентов</label>
                    <span className="info-value">
                      {currentLesson.current_students || 0} из {currentLesson.max_students}
                    </span>
                  </div>
                  {currentLesson.subject && (
                    <div className="info-item full-width">
                      <label>Тема занятия</label>
                      <span className="info-value">{currentLesson.subject}</span>
                    </div>
                  )}
                </div>
              </section>
            )}

            {/* Список студентов (READ-ONLY) */}
            {activeTab === 'students' && (
              <section className="teacher-lesson-section">
                <h4 className="section-title">Записанные студенты</h4>

                {students.length === 0 ? (
                  <div className="teacher-lesson-empty">
                    <p>Нет записанных студентов</p>
                  </div>
                ) : (
                  <div className="students-list">
                    {students.map((student, index) => (
                      <div key={student.id} className="student-item-readonly">
                        <div className="student-number">{index + 1}</div>
                        <div className="student-info">
                          <div className="student-name">{student.student_name}</div>
                          <div className="student-email">{student.student_email}</div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </section>
            )}

            {/* Домашнее задание (CAN EDIT, кроме прошедших занятий) */}
            {activeTab === 'homework' && (
              <section className="teacher-lesson-section">
                <HomeworkSection
                  lessonId={currentLesson.id}
                  lesson={currentLesson}
                  onHomeworkSaved={handleHomeworkSaved}
                  onHomeworkCountChange={setHomeworkCount}
                  readOnly={isPast}
                />
              </section>
            )}

            {/* Рассылки (CAN SEND, кроме прошедших занятий) */}
            {activeTab === 'broadcasts' && (
              <section className="teacher-lesson-section">
                <BroadcastSection
                  lessonId={currentLesson.id}
                  lesson={currentLesson}
                  readOnly={isPast}
                />
              </section>
            )}
          </>
        )}
      </div>
    </Modal>
  );
};

export default TeacherLessonModal;
