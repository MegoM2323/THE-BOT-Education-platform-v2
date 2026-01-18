import { describe, it, expect } from 'vitest';
import { readFileSync } from 'fs';
import { resolve } from 'path';

describe('Russian Translations Verification - Code Inspection', () => {
  const readComponentFile = (filePath) => {
    return readFileSync(resolve(filePath), 'utf-8');
  };

  describe('SCENARIO 1: TemplateCalendarView - Russian Days of Week', () => {
    const componentPath = './src/components/admin/TemplateCalendarView.jsx';
    const content = readComponentFile(componentPath);

    it('SCENARIO 1.1: DAYS_OF_WEEK should have Понедельник (Monday)', () => {
      expect(content).toContain("'Понедельник'");
    });

    it('SCENARIO 1.2: DAYS_OF_WEEK should have Вторник (Tuesday)', () => {
      expect(content).toContain("'Вторник'");
    });

    it('SCENARIO 1.3: DAYS_OF_WEEK should have Среда (Wednesday)', () => {
      expect(content).toContain("'Среда'");
    });

    it('SCENARIO 1.4: DAYS_OF_WEEK should have Четверг (Thursday)', () => {
      expect(content).toContain("'Четверг'");
    });

    it('SCENARIO 1.5: DAYS_OF_WEEK should have Пятница (Friday)', () => {
      expect(content).toContain("'Пятница'");
    });

    it('SCENARIO 1.6: DAYS_OF_WEEK should have Суббота (Saturday)', () => {
      expect(content).toContain("'Суббота'");
    });

    it('SCENARIO 1.7: DAYS_OF_WEEK should have Воскресенье (Sunday)', () => {
      expect(content).toContain("'Воскресенье'");
    });

    it('SCENARIO 1.8: DAYS_OF_WEEK should have short names (Пн, Вт, etc)', () => {
      expect(content).toContain("short: 'Пн'");
      expect(content).toContain("short: 'Вт'");
      expect(content).toContain("short: 'Ср'");
      expect(content).toContain("short: 'Чт'");
      expect(content).toContain("short: 'Пт'");
      expect(content).toContain("short: 'Сб'");
      expect(content).toContain("short: 'Вс'");
    });
  });

  describe('SCENARIO 2: TemplateCalendarView - Calendar UI Text in Russian', () => {
    const componentPath = './src/components/admin/TemplateCalendarView.jsx';
    const content = readComponentFile(componentPath);

    it('SCENARIO 2.1: should have "Расписание на неделю" (Calendar Schedule)', () => {
      expect(content).toContain('Расписание на неделю');
    });

    it('SCENARIO 2.2: should have "Добавить занятие" (Add Lesson) button', () => {
      expect(content).toContain('Добавить занятие');
    });

    it('SCENARIO 2.3: should have "Нет занятий" (No Classes) text', () => {
      expect(content).toContain('Нет занятий');
    });

    it('SCENARIO 2.6: should have "Редактировать" (Edit) button', () => {
      expect(content).toContain('Редактировать');
    });

    it('SCENARIO 2.7: should have "Удалить" (Delete) button', () => {
      expect(content).toContain('Удалить');
    });
  });

  describe('SCENARIO 3: TemplateCalendarView - Modal Titles in Russian', () => {
    const componentPath = './src/components/admin/TemplateCalendarView.jsx';
    const content = readComponentFile(componentPath);

    it('SCENARIO 3.1: should have "Добавить занятие" modal title', () => {
      expect(content).toContain('title="Добавить занятие"');
    });

    it('SCENARIO 3.2: should have "Добавить урок" modal title', () => {
      expect(content).toContain('title="Добавить урок"');
    });

    it('SCENARIO 3.3: should have "Редактировать урок" modal title', () => {
      expect(content).toContain('title="Редактировать урок"');
    });

    it('SCENARIO 3.4: should have "Удалить урок" confirm modal title', () => {
      expect(content).toContain('title="Удалить урок"');
    });

  });

  describe('SCENARIO 5: LessonEditModal - Notification Messages in Russian', () => {
    const componentPath = './src/components/admin/LessonEditModal.jsx';
    const content = readComponentFile(componentPath);

    it('SCENARIO 5.1: should have "Изменений не обнаружено" (No changes detected)', () => {
      expect(content).toContain('Изменений не обнаружено');
    });

    it('SCENARIO 5.2: should have "Не удалось определить тип изменения" (Could not detect modification type)', () => {
      expect(content).toContain('Не удалось определить тип изменения');
    });

    it('SCENARIO 5.3: should have "Успешно применено к" (Successfully applied to)', () => {
      expect(content).toContain('Успешно применено к');
    });
  });

  describe('SCENARIO 6: LessonEditModal - Modal Titles in Russian', () => {
    const componentPath = './src/components/admin/LessonEditModal.jsx';
    const content = readComponentFile(componentPath);

    it('SCENARIO 6.1: should have "Редактирование занятия" (Edit Lesson) modal title', () => {
      expect(content).toContain('Редактирование занятия');
    });

    it('SCENARIO 6.2: should have "Основное" (Basic) tab', () => {
      expect(content).toContain('Основное');
    });

    it('SCENARIO 6.3: should have "Домашнее задание" (Homework) tab', () => {
      expect(content).toContain('Домашнее задание');
    });

    it('SCENARIO 6.4: should have "Рассылки" (Broadcasts) tab', () => {
      expect(content).toContain('Рассылки');
    });
  });

  describe('SCENARIO 7: LessonEditModal - Form Labels in Russian', () => {
    const componentPath = './src/components/admin/LessonEditModal.jsx';
    const content = readComponentFile(componentPath);

    it('SCENARIO 7.1: should have "Дата занятия" (Lesson Date) label', () => {
      expect(content).toContain('Дата занятия');
    });

    it('SCENARIO 7.2: should have "Время начала" (Start Time) label', () => {
      expect(content).toContain('Время начала');
    });

    it('SCENARIO 7.3: should have "Время окончания" (End Time) label', () => {
      expect(content).toContain('Время окончания');
    });

    it('SCENARIO 7.5: should have "Преподаватель" (Teacher) label', () => {
      expect(content).toContain('Преподаватель');
    });

    it('SCENARIO 7.7: should have "Тема занятия" (Lesson Subject) label', () => {
      expect(content).toContain('Тема занятия');
    });

  });

  describe('SCENARIO 8: LessonEditModal - Button Text in Russian', () => {
    const componentPath = './src/components/admin/LessonEditModal.jsx';
    const content = readComponentFile(componentPath);

    it('SCENARIO 8.1: should have "Удалить занятие" (Delete Lesson) button', () => {
      expect(content).toContain('Удалить занятие');
    });

    it('SCENARIO 8.2: should have "Сохранение..." (Saving...) indicator', () => {
      expect(content).toContain('Сохранение...');
    });

    it('SCENARIO 8.3: should have "Ошибка сохранения" (Save Error) indicator', () => {
      expect(content).toContain('Ошибка сохранения');
    });

    it('SCENARIO 8.4: should have "Сохранено" (Saved) indicator', () => {
      expect(content).toContain('Сохранено');
    });
  });

  describe('SCENARIO 10: LessonEditModal - Confirmation Messages in Russian', () => {
    const componentPath = './src/components/admin/LessonEditModal.jsx';
    const content = readComponentFile(componentPath);

    it('SCENARIO 10.1: should have "Отмена записи студента" (Cancel Student Registration) text', () => {
      expect(content).toContain('Отмена записи студента');
    });

    it('SCENARIO 10.2: should have "Удаление занятия" (Lesson Deletion) text', () => {
      expect(content).toContain('Удаление занятия');
    });

    it('SCENARIO 10.3: should have "Внимание: Вы редактируете занятие в прошлом" warning', () => {
      expect(content).toContain('Внимание: Вы редактируете занятие в прошлом');
    });
  });
});
