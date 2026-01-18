import { Profiler } from 'react';
import TemplateCalendarView from './TemplateCalendarView.jsx';
import { logger } from '../../utils/logger.js';

/**
 * Профилировщик для измерения производительности TemplateCalendarView компонента
 * Используется для тестирования оптимизаций (React.memo, useMemo, useCallback)
 */
export const TemplateCalendarProfiler = (props) => {
  const onRender = (id, phase, actualDuration, baseDuration, startTime, commitTime, interactions) => {
    if (import.meta.env.DEV) {
      console.group(`📊 TemplateCalendar Profiler - ${phase}`);
      logger.debug('Компонент:', id);
      logger.debug('Фаза:', phase); // mount или update
      logger.debug('Фактическое время рендера:', actualDuration.toFixed(2), 'мс');
      logger.debug('Базовое время (без мемоизации):', baseDuration.toFixed(2), 'мс');
      logger.debug('Старт рендера:', startTime.toFixed(2), 'мс');
      logger.debug('Коммит:', commitTime.toFixed(2), 'мс');
      logger.debug('Взаимодействия:', interactions);
      console.groupEnd();
    }

    // Сохраняем в window для автоматических тестов
    if (!window.templateCalendarProfilerData) {
      window.templateCalendarProfilerData = [];
    }
    window.templateCalendarProfilerData.push({
      id,
      phase,
      actualDuration,
      baseDuration,
      startTime,
      commitTime,
      interactions: Array.from(interactions),
    });
  };

  return (
    <Profiler id="TemplateCalendarView" onRender={onRender}>
      <TemplateCalendarView {...props} />
    </Profiler>
  );
};

export default TemplateCalendarProfiler;
