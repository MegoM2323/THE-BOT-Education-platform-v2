import { useState, useEffect } from 'react';
import Modal from "../common/Modal.jsx";
import Button from "../common/Button.jsx";
import Spinner from '../common/Spinner.jsx';
import { useTemplates } from '../../hooks/useTemplates.js';
import { apiClient } from '../../api/client.js';
import * as creditsAPI from '../../api/credits.js';
import './TemplateSelectionModal.css';

/**
 * Modal for selecting and applying a template to a week
 * @param {boolean} isOpen - Modal open state
 * @param {Function} onClose - Close callback
 * @param {Date} weekStartDate - Monday of the week
 * @param {Function} onApplied - Callback after template applied
 * @param {number|string|null} preselectedTemplateId - Pre-selected template ID (optional)
 */
export const TemplateSelectionModal = ({ isOpen, onClose, weekStartDate, onApplied, preselectedTemplateId = null }) => {
  const { templates, loading: templatesLoading, applyTemplate, isApplying } = useTemplates();
  const [selectedTemplateId, setSelectedTemplateId] = useState(() => preselectedTemplateId || '');
  const [preview, setPreview] = useState(null);
  const [loadingPreview, setLoadingPreview] = useState(false);
  const [error, setError] = useState('');
  const [creditValidationIssues, setCreditValidationIssues] = useState([]);
  const [validatingCredits, setValidatingCredits] = useState(false);

  useEffect(() => {
    if (preselectedTemplateId) {
      setSelectedTemplateId(preselectedTemplateId);
    }
  }, [preselectedTemplateId]);

  /**
   * Validate that all template students have sufficient credits
   * Fetches current credit balance for each student in the template
   */
  const validateCreditsBeforeTemplateApply = async (students) => {
    try {
      const issues = [];

      // Фильтруем студентов, у которых будут списаны кредиты
      const studentsWithDeductions = students.filter(s => s.creditsDeducted > 0);

      if (studentsWithDeductions.length === 0) {
        // Нет студентов с списаниями - валидация не нужна
        return [];
      }

      // Fetch all student credits in parallel
      const creditPromises = studentsWithDeductions.map(student =>
        creditsAPI.getUserCredits(student.student_id)
          .then(data => ({
            studentId: student.student_id,
            studentName: student.name,
            balance: data.balance ?? 0,
            required: student.creditsDeducted,
          }))
          .catch((err) => {
            console.warn('[validateCreditsBeforeTemplateApply] Failed to fetch credits for student:', student.student_id, err.message);
            return {
              studentId: student.student_id,
              studentName: student.name,
              balance: 0,
              required: student.creditsDeducted,
              fetchError: true,
            };
          })
      );

      const creditData = await Promise.all(creditPromises);

      // Check which students have insufficient credits
      creditData.forEach(data => {
        if (data.fetchError) {
          // If we can't fetch credits, assume insufficient for safety
          issues.push({
            studentId: data.studentId,
            studentName: data.studentName,
            balance: 0,
            required: data.required,
            fetchError: true,
          });
        } else if (data.balance < data.required) {
          issues.push(data);
        }
      });

      console.debug('[validateCreditsBeforeTemplateApply] Validation complete:', {
        totalStudents: students.length,
        studentsWithDeductions: studentsWithDeductions.length,
        issuesFound: issues.length,
      });

      return issues;
    } catch (err) {
      console.error('[validateCreditsBeforeTemplateApply] Error:', err);
      // On error, return empty issues to allow retry (user can see error in UI)
      return [];
    }
  };

  // Load preview when template selected
  useEffect(() => {
    const controller = new AbortController();

    const loadPreview = async () => {
      if (!selectedTemplateId) {
        setPreview(null);
        setCreditValidationIssues([]);
        return;
      }

      setLoadingPreview(true);
      setError('');
      setCreditValidationIssues([]);

      try {
        // Fetch template details for preview
        const template = await apiClient.get(`/templates/${selectedTemplateId}`);

        // Calculate preview data WITH CONFLICT DETECTION
        const lessonsCount = template.lesson_count || template.lessons?.length || 0;
        const studentsMap = {};

        // Parse week dates to calculate actual lesson dates and times
        const weekStart = new Date(weekStartDate);
        weekStart.setHours(0, 0, 0, 0);

        // Build list of template lesson times
        const templateLessonTimes = [];
        if (template.lessons) {
          template.lessons.forEach((lesson) => {
            // lesson.day_of_week: 0=Monday, 1=Tuesday, ..., 6=Sunday
            const lessonDate = new Date(weekStart);
            lessonDate.setDate(lessonDate.getDate() + lesson.day_of_week);

            // Parse time strings (HH:MM:SS format)
            const [startHour, startMin, startSec] = lesson.start_time.split(':').map(Number);
            const [endHour, endMin, endSec] = lesson.end_time.split(':').map(Number);

            const startTime = new Date(lessonDate);
            startTime.setHours(startHour, startMin, startSec, 0);

            const endTime = new Date(lessonDate);
            endTime.setHours(endHour, endMin, endSec, 0);

            templateLessonTimes.push({
              lessonId: lesson.id,
              startTime,
              endTime,
              students: lesson.students || [],
            });
          });
        }

        // Fetch existing bookings to check for conflicts
        // We need to check if students already have bookings at the same time
        const bookingsResponse = await apiClient.get('/bookings?status=active');

        // API возвращает paginated response после обработки apiClient:
        // { data: { bookings: [...], count: N }, meta: {...} }
        // или может быть просто массив bookings если pagination не используется
        let existingBookings = [];
        if (bookingsResponse?.data?.bookings && Array.isArray(bookingsResponse.data.bookings)) {
          // Paginated response: { data: { bookings: [...] }, meta: {...} }
          existingBookings = bookingsResponse.data.bookings;
        } else if (bookingsResponse?.bookings && Array.isArray(bookingsResponse.bookings)) {
          // Non-paginated response: { bookings: [...], count: N }
          existingBookings = bookingsResponse.bookings;
        } else if (Array.isArray(bookingsResponse)) {
          // Direct array response
          existingBookings = bookingsResponse;
        }

        // Group by student and check conflicts
        if (template.lessons) {
          template.lessons.forEach((lesson) => {
            if (lesson.students) {
              lesson.students.forEach((student) => {
                const studentId = student.student_id;

                if (!studentsMap[studentId]) {
                  // API возвращает student_name из JOIN с таблицей users (u.full_name as student_name)
                  const displayName =
                    student.student_name?.trim() ||
                    student.name?.trim() ||
                    student.full_name?.trim() ||
                    student.email?.trim() ||
                    `ID: ${studentId}`;

                  studentsMap[studentId] = {
                    student_id: studentId, // Добавляем ID для валидации кредитов
                    name: displayName,
                    creditsDeducted: 0,
                    existingBookings: 0,
                  };
                }

                // Find the template lesson time for this lesson
                const templateLesson = templateLessonTimes.find(
                  (tl) => tl.lessonId === lesson.id
                );

                if (templateLesson) {
                  // Check if this student has a conflicting booking at the same time
                  const hasConflict = existingBookings.some((booking) => {
                    if (booking.student_id !== studentId) return false;
                    if (booking.status !== 'active') return false;

                    // Booking has lesson details embedded
                    const bookingStart = booking.start_time ? new Date(booking.start_time) : null;
                    const bookingEnd = booking.end_time ? new Date(booking.end_time) : null;

                    if (!bookingStart || !bookingEnd) return false;

                    // Check for time conflict: overlapping intervals
                    // Conflict if: booking.start < template.end AND booking.end > template.start
                    return bookingStart < templateLesson.endTime && bookingEnd > templateLesson.startTime;
                  });

                  // Only count credit deduction if no conflict
                  if (!hasConflict) {
                    studentsMap[studentId].creditsDeducted += 1;
                  } else {
                    // Track that this student has conflicting booking
                    studentsMap[studentId].existingBookings += 1;
                  }
                }
              });
            }
          });
        }

        const studentsList = Object.values(studentsMap);
        setPreview({
          templateName: template.name,
          lessonsCount,
          students: studentsList,
        });

        // Validate credits after preview is loaded
        if (studentsList.length > 0) {
          setValidatingCredits(true);
          const issues = await validateCreditsBeforeTemplateApply(studentsList);
          setCreditValidationIssues(issues);
          setValidatingCredits(false);
        }
      } catch (err) {
        if (err.name !== 'AbortError') {
          console.error('[TemplateSelectionModal] Load preview error:', err);
          setError(err.message || 'Не удалось загрузить предпросмотр шаблона');
          setPreview(null);
          setCreditValidationIssues([]);
        }
      } finally {
        setLoadingPreview(false);
      }
    };

    loadPreview();
    return () => controller.abort();
  }, [selectedTemplateId, weekStartDate]);

  const handleApply = async () => {
    if (!selectedTemplateId) {
      setError('Пожалуйста, выберите шаблон');
      return;
    }

    // Check for credit validation issues
    if (creditValidationIssues.length > 0) {
      const studentsList = creditValidationIssues
        .map(issue => `${issue.studentName} (баланс: ${issue.balance}, требуется: ${issue.required})`)
        .join(', ');
      setError(`Недостаточно кредитов для студентов: ${studentsList}. Сначала пополните кредиты.`);
      return;
    }

    setError('');

    try {
      // Format week start date as ISO string (YYYY-MM-DD) using local timezone
      const weekStart = new Date(weekStartDate);
      const year = weekStart.getFullYear();
      const month = String(weekStart.getMonth() + 1).padStart(2, '0');
      const day = String(weekStart.getDate()).padStart(2, '0');
      const isoDate = `${year}-${month}-${day}`;

      const result = await applyTemplate(selectedTemplateId, isoDate);

      // Вызываем callback после успешного применения
      // useTemplates.applyTemplate уже инвалидирует кеши и показывает notification
      if (onApplied) {
        onApplied(result);
      }
    } catch (err) {
      // Логируем ошибку для отладки
      console.error('[TemplateSelectionModal] Apply template error:', err);
      setError(err.message || 'Не удалось применить шаблон');
    }
  };

  const formatWeekRange = (date) => {
    const start = new Date(date);
    const end = new Date(date);
    end.setDate(end.getDate() + 6);

    return `${start.toLocaleDateString('ru-RU', {
      day: 'numeric',
      month: 'short',
    })} - ${end.toLocaleDateString('ru-RU', {
      day: 'numeric',
      month: 'long',
      year: 'numeric',
    })}`;
  };

  // Determine if apply button should be disabled
  const isApplyDisabled = !selectedTemplateId || isApplying || loadingPreview || validatingCredits || creditValidationIssues.length > 0;

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Применить шаблон к неделе"
      loading={isApplying}
      footer={
        <div className="template-selection-footer">
          <Button variant="secondary" onClick={onClose} disabled={isApplying}>
            Отмена
          </Button>
          <Button
            variant={creditValidationIssues.length > 0 ? "danger" : "primary"}
            onClick={handleApply}
            disabled={isApplyDisabled}
            loading={isApplying}
            title={creditValidationIssues.length > 0 ? "Не удается применить: недостаточно кредитов у студентов" : ""}
          >
            {creditValidationIssues.length > 0 ? "Ошибка: недостаточно кредитов" : "Применить шаблон"}
          </Button>
        </div>
      }
    >
      <div className="template-selection-content">
        <div className="week-info">
          <strong>Неделя:</strong> {formatWeekRange(weekStartDate)}
        </div>

        {templatesLoading ? (
          <div className="template-selection-loading">
            <Spinner />
            <p>Загрузка шаблонов...</p>
          </div>
        ) : templates.length === 0 ? (
          <div className="template-selection-empty">
            <p>Нет доступных шаблонов. Сначала создайте шаблон.</p>
          </div>
        ) : (
          <>
            {!preselectedTemplateId ? (
              <div className="template-select-group">
                <label htmlFor="template-select">Выберите шаблон:</label>
                <select
                  id="template-select"
                  value={selectedTemplateId}
                  onChange={(e) => setSelectedTemplateId(e.target.value)}
                  disabled={isApplying}
                  className="template-select"
                >
                  <option value="">-- Выберите шаблон --</option>
                  {templates.map((template) => (
                    <option key={template.id} value={template.id}>
                      {template.name}
                    </option>
                  ))}
                </select>
              </div>
            ) : (
              <div className="preselected-template-info">
                <strong>Шаблон:</strong> {templates.find(t => t.id === Number(preselectedTemplateId))?.name || 'Загрузка...'}
              </div>
            )}

            {loadingPreview && (
              <div className="template-preview-loading">
                <Spinner />
                <p>Загрузка предпросмотра...</p>
              </div>
            )}

            {preview && !loadingPreview && (
              <div className="template-preview">
                <h4>Предпросмотр</h4>

                <div className="preview-section">
                  <strong>Шаблон:</strong> {preview.templateName}
                </div>

                <div className="preview-section">
                  <strong>Будет создано занятий:</strong> {preview.lessonsCount}
                </div>

                {preview.students.length > 0 && (
                  <div className="preview-section">
                    <strong>Влияние на кредиты:</strong>
                    {validatingCredits && (
                      <div className="validating-credits">
                        <Spinner size="small" />
                        <p>Проверка баланса кредитов...</p>
                      </div>
                    )}
                    {!validatingCredits && (
                      <div className="students-list">
                        {preview.students.map((student, index) => {
                          const creditIssue = creditValidationIssues.find(issue => issue.studentId === student.student_id);
                          const hasIssue = !!creditIssue;

                          return (
                            <div key={index} className={`student-credit-item ${hasIssue ? 'insufficient-credits' : ''}`}>
                              <span className={`student-name ${hasIssue ? 'insufficient' : ''}`}>
                                {hasIssue && '⚠️ '}
                                {student.name}
                              </span>
                              <div className="credit-details">
                                {hasIssue ? (
                                  <span className="credit-insufficient" title="Недостаточно кредитов для этого студента">
                                    Баланс: {creditIssue.balance}, требуется: {creditIssue.required}
                                  </span>
                                ) : (
                                  <>
                                    {student.creditsDeducted > 0 && (
                                      <span className="credit-deduction">
                                        -{student.creditsDeducted} {student.creditsDeducted !== 1 ? 'кредитов' : 'кредит'}
                                      </span>
                                    )}
                                    {student.existingBookings > 0 && (
                                      <span className="credit-skip" title="Уже забронирован на это время">
                                        (пропуск {student.existingBookings} — уже занято)
                                      </span>
                                    )}
                                    {student.creditsDeducted === 0 && student.existingBookings === 0 && (
                                      <span className="credit-zero">нет изменений</span>
                                    )}
                                  </>
                                )}
                              </div>
                            </div>
                          );
                        })}
                      </div>
                    )}
                  </div>
                )}

                {creditValidationIssues.length > 0 && !validatingCredits && (
                  <div className="preview-error-credits">
                    <strong>Ошибка: Недостаточно кредитов</strong>
                    <p>Следующим студентам не хватает кредитов для применения шаблона:</p>
                    <ul>
                      {creditValidationIssues.map((issue, index) => (
                        <li key={index}>
                          <strong>{issue.studentName}</strong>: баланс {issue.balance}, требуется {issue.required}
                          {issue.fetchError && ' (не удалось проверить баланс)'}
                        </li>
                      ))}
                    </ul>
                    <p>Пожалуйста, пополните кредиты перед применением шаблона.</p>
                  </div>
                )}

                {creditValidationIssues.length === 0 && !validatingCredits && (
                  <div className="preview-warning">
                    <strong>Внимание:</strong> Применение этого шаблона создаст занятия и спишет
                    кредиты у студентов. Это действие нельзя отменить (но можно откатить).
                  </div>
                )}
              </div>
            )}
          </>
        )}

        {error && <div className="template-selection-error">{error?.message || String(error) || 'Произошла ошибка'}</div>}
      </div>
    </Modal>
  );
};

export default TemplateSelectionModal;
