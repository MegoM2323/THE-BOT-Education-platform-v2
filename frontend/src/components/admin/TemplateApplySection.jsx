import { useState, useMemo } from "react";
import Button from "../common/Button.jsx";
import Input from "../common/Input.jsx";
import TemplateSelectionModal from "./TemplateSelectionModal.jsx";
import "./TemplateApplySection.css";

/**
 * Вычисляет понедельник недели из выбранной даты
 * @param {string} dateStr - Дата в формате YYYY-MM-DD
 * @returns {Date} - Понедельник недели (ISO week, Пн=0)
 */
const getWeekStartDate = (dateStr) => {
  const date = new Date(dateStr);
  const day = date.getDay();
  const daysSinceMonday = (day + 6) % 7;
  date.setDate(date.getDate() - daysSinceMonday);
  date.setHours(0, 0, 0, 0);
  return date;
};

/**
 * Проверяет, является ли неделя прошедшей
 * @param {Date} weekStart - Понедельник недели
 * @returns {boolean} - true если неделя в прошлом
 */
const isPastWeek = (weekStart) => {
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  return weekStart < today;
};

/**
 * Форматирует диапазон недели (например: "20 янв - 26 января 2026")
 * @param {Date} weekStart - Понедельник недели
 * @returns {string} - Форматированный диапазон
 */
const formatWeekRange = (weekStart) => {
  const weekEnd = new Date(weekStart);
  weekEnd.setDate(weekEnd.getDate() + 6);

  return `${weekStart.toLocaleDateString("ru-RU", {
    day: "numeric",
    month: "short",
  })} - ${weekEnd.toLocaleDateString("ru-RU", {
    day: "numeric",
    month: "long",
    year: "numeric",
  })}`;
};

/**
 * Компонент для применения шаблона к выбранной неделе
 * @param {number} templateId - ID выбранного шаблона (обязательный)
 * @param {Function} onApplied - Callback после применения (опциональный)
 */
export const TemplateApplySection = ({ templateId, onApplied }) => {
  const [selectedDate, setSelectedDate] = useState("");
  const [showModal, setShowModal] = useState(false);

  // Вычисляем понедельник и проверяем валидность
  const weekInfo = useMemo(() => {
    if (!selectedDate) {
      return { weekStart: null, weekRange: "", isPast: false };
    }

    const weekStart = getWeekStartDate(selectedDate);
    const weekRange = formatWeekRange(weekStart);
    const isPast = isPastWeek(weekStart);

    return { weekStart, weekRange, isPast };
  }, [selectedDate]);

  // Минимальная дата (сегодня)
  const now = new Date();
  const year = now.getFullYear();
  const month = String(now.getMonth() + 1).padStart(2, "0");
  const day = String(now.getDate()).padStart(2, "0");
  const today = `${year}-${month}-${day}`;

  const handleApplyClick = () => {
    if (!templateId) {
      return;
    }

    if (weekInfo.isPast) {
      return;
    }

    setShowModal(true);
  };

  const handleModalClose = () => {
    setShowModal(false);
  };

  const handleApplied = (result) => {
    setShowModal(false);
    setSelectedDate("");
    if (onApplied) {
      onApplied(result);
    }
  };

  const isApplyDisabled = !templateId || !selectedDate || weekInfo.isPast;

  return (
    <div className="template-apply-section">
      <h3>Применить к выбранной неделе</h3>

      {!templateId && (
        <div className="template-apply-info">
          <strong>Информация:</strong> Сначала выберите шаблон на предыдущем
          шаге.
        </div>
      )}

      {templateId && (
        <>
          <div className="template-apply-controls">
            <div className="template-apply-date-picker">
              <Input
                type="date"
                label="Выберите дату"
                name="template-apply-date"
                value={selectedDate}
                onChange={(e) => setSelectedDate(e.target.value)}
                min={today}
              />
            </div>

            <Button
              variant="primary"
              onClick={handleApplyClick}
              disabled={isApplyDisabled}
            >
              Применить к выбранной неделе
            </Button>
          </div>

          {selectedDate && weekInfo.weekRange && (
            <div className="template-apply-week-info">
              <strong>Неделя:</strong> {weekInfo.weekRange}
            </div>
          )}

          {weekInfo.isPast && (
            <div className="template-apply-warning">
              <strong>Предупреждение:</strong> Выбранная неделя уже прошла.
              Выберите текущую или будущую неделю.
            </div>
          )}
        </>
      )}

      {showModal && weekInfo.weekStart && (
        <TemplateSelectionModal
          isOpen={showModal}
          onClose={handleModalClose}
          weekStartDate={weekInfo.weekStart}
          onApplied={handleApplied}
          preselectedTemplateId={templateId}
        />
      )}
    </div>
  );
};

export default TemplateApplySection;
