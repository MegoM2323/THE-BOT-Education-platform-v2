import { useState } from "react";
import "./StudentCheckboxList.css";

/**
 * Компонент для управления студентами через чекбоксы
 *
 * @param {Array} allStudents - Массив всех доступных студентов
 * @param {Array} enrolledStudentIds - ID записанных студентов
 * @param {Function} onToggle - Коллбэк при изменении чекбокса (studentId, isEnrolled)
 * @param {number} lessonId - ID занятия
 * @param {boolean} disabled - Блокировка всего компонента (например, для прошлых занятий)
 */
export const StudentCheckboxList = ({
  allStudents = [],
  enrolledStudentIds = [],
  onToggle,
  // eslint-disable-next-line no-unused-vars
  lessonId, // Зарезервировано для будущего использования (bulk операции)
  disabled = false,
}) => {
  const [searchQuery, setSearchQuery] = useState("");

  // Фильтрация студентов по поисковому запросу
  const filteredStudents = allStudents.filter(
    (student) =>
      student.full_name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      student.name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      student.email?.toLowerCase().includes(searchQuery.toLowerCase()),
  );

  // Обработчик клика по чекбоксу
  const handleToggle = (studentId, currentlyEnrolled, studentCredits) => {
    // Если пытаемся добавить студента без кредитов - логируем и позволяем родителю обработать
    if (!currentlyEnrolled && studentCredits < 1) {
      console.info(
        `[StudentCheckboxList] Попытка добавить студента ${studentId} без кредитов (баланс: ${studentCredits})`,
      );
    }

    if (onToggle) {
      onToggle(studentId, !currentlyEnrolled);
    }
  };

  // Получение имени студента (поддержка разных форматов)
  const getStudentName = (student) => {
    return student.full_name || student.name || "Неизвестный студент";
  };

  // Threshold для предупреждения о низких кредитах
  const LOW_CREDITS_THRESHOLD = 1;

  // Получение кредитов студента (всегда возвращает number)
  const getStudentCredits = (student) => {
    if (!student) {
      console.warn("[StudentCheckboxList] Student is null/undefined");
      return 0;
    }

    let credits = student.credits;

    // Проверка наличия credits
    if (credits === undefined || credits === null) {
      console.warn(
        `[StudentCheckboxList] Student ${student.id} has no credits field`,
        student,
      );
      return 0;
    }

    // Надежное преобразование в number (поддерживает строки, числа, null, undefined)
    const numCredits = Number(credits);

    // Проверка на NaN
    if (isNaN(numCredits)) {
      console.warn(
        `[StudentCheckboxList] Student ${student.id} has invalid credits value:`,
        credits,
      );
      return 0;
    }

    // Гарантируем что результат >= 0
    return Math.max(0, numCredits);
  };

  // Проверка низкого баланса кредитов
  const hasLowCredits = (student) => {
    return getStudentCredits(student) < LOW_CREDITS_THRESHOLD;
  };

  return (
    <div className={`student-checkbox-list ${disabled ? "disabled" : ""}`}>
      <div className="student-checkbox-list-header">
        <h4>Управление студентами</h4>
        <input
          type="text"
          className="student-search-input"
          placeholder="Поиск по имени или email..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          disabled={disabled}
        />
        <div className="student-count-info">
          Записано: {enrolledStudentIds.length} из {allStudents.length}
        </div>
      </div>

      <div className="student-list">
        {filteredStudents.length === 0 ? (
          <div className="student-list-empty">
            {searchQuery ? "Студенты не найдены" : "Нет доступных студентов"}
          </div>
        ) : (
          filteredStudents.map((student) => {
            const isEnrolled = enrolledStudentIds.includes(student.id);
            const studentCredits = getStudentCredits(student);
            const lowCredits = hasLowCredits(student);

            return (
              <label
                key={student.id}
                className={`student-item ${isEnrolled ? "enrolled" : ""} ${lowCredits ? "low-credits" : ""} ${(lowCredits && !isEnrolled) || disabled ? "disabled" : ""}`}
              >
                <input
                  type="checkbox"
                  checked={isEnrolled}
                  onChange={() =>
                    handleToggle(student.id, isEnrolled, studentCredits)
                  }
                  disabled={disabled || (lowCredits && !isEnrolled)}
                  title={
                    disabled
                      ? "Редактирование недоступно для прошедших занятий"
                      : lowCredits && !isEnrolled
                        ? "Недостаточно кредитов для добавления"
                        : ""
                  }
                />
                <div className="student-info">
                  <span className="student-name">
                    {getStudentName(student)} ({studentCredits} кр.)
                  </span>
                  <span className="student-email">
                    {student.email || "Нет email"}
                  </span>
                </div>
                <div className="student-credits-container">
                  <span
                    className={`student-credits ${lowCredits ? "credits-warning" : ""}`}
                  >
                    {studentCredits} кр.
                  </span>
                  {lowCredits && !isEnrolled && (
                    <span
                      className="credits-warning-icon"
                      title="Недостаточно кредитов для добавления"
                    >
                      ⚠️
                    </span>
                  )}
                </div>
              </label>
            );
          })
        )}
      </div>
    </div>
  );
};

export default StudentCheckboxList;
