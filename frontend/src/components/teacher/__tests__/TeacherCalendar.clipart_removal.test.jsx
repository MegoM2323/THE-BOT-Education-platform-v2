import { render, screen, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { vi } from "vitest";
import { TeacherCalendar } from "../TeacherCalendar";
import * as lessonsAPI from "../../../api/lessons";

// Mock API
vi.mock("../../../api/lessons", () => ({
  getTeacherSchedule: vi.fn(),
}));

// Mock TeacherLessonModal
vi.mock("../TeacherLessonModal", () => ({
  default: ({ isOpen, onClose, lesson }) =>
    isOpen ? (
      <div data-testid="teacher-lesson-modal">
        <button onClick={onClose}>Close</button>
        <div>{lesson?.subject}</div>
      </div>
    ) : null,
}));

const createQueryClient = () =>
  new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        cacheTime: 0,
      },
    },
  });

const renderWithClient = (component) => {
  const queryClient = createQueryClient();
  return render(
    <QueryClientProvider client={queryClient}>
      {component}
    </QueryClientProvider>,
  );
};

describe("TeacherCalendar - Clipart Removal (Скрепка удалена)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("Проверка что скрепка (📎) удалена", () => {
    test("НЕ должна отображаться скрепка для занятия с homework_count", async () => {
      const today = new Date();
      today.setHours(10, 0, 0, 0);

      const mockLesson = {
        id: 1,
        subject: "Математика",
        start_time: today.toISOString(),
        max_students: 5,
        current_students: 2,
        homework_count: 5, // есть домашки
        broadcasts_count: 0,
      };

      lessonsAPI.getTeacherSchedule.mockResolvedValue({
        lessons: [mockLesson],
        count: 1,
      });

      renderWithClient(<TeacherCalendar />);

      await waitFor(
        () => {
          expect(screen.getByText("Математика")).toBeInTheDocument();
        },
        { timeout: 3000 },
      );

      // Скрепка не должна быть в документе
      expect(screen.queryByText(/📎/)).not.toBeInTheDocument();
    });

    test("НЕ должна быть скрепка даже когда много домашек", async () => {
      const today = new Date();
      today.setHours(10, 0, 0, 0);

      const mockLesson = {
        id: 1,
        subject: "Физика",
        start_time: today.toISOString(),
        max_students: 1,
        current_students: 1,
        homework_count: 10, // много домашек
        broadcasts_count: 0,
      };

      lessonsAPI.getTeacherSchedule.mockResolvedValue({
        lessons: [mockLesson],
        count: 1,
      });

      renderWithClient(<TeacherCalendar />);

      await waitFor(
        () => {
          expect(screen.getByText("Физика")).toBeInTheDocument();
        },
        { timeout: 3000 },
      );

      // Скрепка вообще не отображается
      expect(screen.queryByText(/📎/)).not.toBeInTheDocument();
      // Но содержимое занятия должно быть
      expect(screen.getByText(/1\/1 студентов/)).toBeInTheDocument();
    });

    test("НЕ должна быть скрепка для нескольких занятий", async () => {
      const today = new Date();
      today.setHours(10, 0, 0, 0);

      const mockLessons = [
        {
          id: 1,
          subject: "Математика",
          start_time: today.toISOString(),
          max_students: 5,
          current_students: 2,
          homework_count: 3,
          broadcasts_count: 0,
        },
        {
          id: 2,
          subject: "Английский",
          start_time: new Date(today.getTime() + 3600000).toISOString(),
          max_students: 10,
          current_students: 7,
          homework_count: 0,
          broadcasts_count: 2,
        },
        {
          id: 3,
          subject: "История",
          start_time: new Date(today.getTime() + 7200000).toISOString(),
          max_students: 3,
          current_students: 3,
          homework_count: 15,
          broadcasts_count: 1,
        },
      ];

      lessonsAPI.getTeacherSchedule.mockResolvedValue({
        lessons: mockLessons,
        count: 3,
      });

      renderWithClient(<TeacherCalendar />);

      await waitFor(
        () => {
          expect(screen.getByText("Математика")).toBeInTheDocument();
          expect(screen.getByText("Английский")).toBeInTheDocument();
          expect(screen.getByText("История")).toBeInTheDocument();
        },
        { timeout: 3000 },
      );

      // Скрепка не должна быть нигде в документе
      expect(screen.queryByText(/📎/)).not.toBeInTheDocument();
    });
  });

  describe("Проверка что остальное содержимое отображается корректно", () => {
    test("должна отображаться рассылка (📢) когда есть broadcasts", async () => {
      const today = new Date();
      today.setHours(10, 0, 0, 0);

      const mockLesson = {
        id: 1,
        subject: "Тест",
        start_time: today.toISOString(),
        max_students: 5,
        current_students: 2,
        homework_count: 5,
        broadcasts_count: 2,
      };

      lessonsAPI.getTeacherSchedule.mockResolvedValue({
        lessons: [mockLesson],
        count: 1,
      });

      renderWithClient(<TeacherCalendar />);

      await waitFor(
        () => {
          expect(screen.getByText("Тест")).toBeInTheDocument();
        },
        { timeout: 3000 },
      );

      // Рассылка должна быть
      expect(screen.getByText(/📢 2/)).toBeInTheDocument();
    });

    test("должны отображаться: время, предмет, количество студентов", async () => {
      const today = new Date();
      today.setHours(14, 30, 0, 0);

      const mockLesson = {
        id: 1,
        subject: "Литература",
        start_time: today.toISOString(),
        max_students: 8,
        current_students: 5,
        homework_count: 2,
        broadcasts_count: 0,
      };

      lessonsAPI.getTeacherSchedule.mockResolvedValue({
        lessons: [mockLesson],
        count: 1,
      });

      renderWithClient(<TeacherCalendar />);

      await waitFor(
        () => {
          expect(screen.getByText("Литература")).toBeInTheDocument();
        },
        { timeout: 3000 },
      );

      // Проверяем что есть все нужные элементы
      // Время
      expect(screen.getByText(/14:30/)).toBeInTheDocument();
      // Предмет
      expect(screen.getByText("Литература")).toBeInTheDocument();
      // Количество студентов
      expect(screen.getByText(/5\/8 студентов/)).toBeInTheDocument();
    });

    test("должен отображаться статус 'Завершено' для прошедших занятий", async () => {
      const pastTime = new Date();
      pastTime.setHours(pastTime.getHours() - 2); // 2 часа назад

      const mockLesson = {
        id: 1,
        subject: "Прошедший урок",
        start_time: pastTime.toISOString(),
        max_students: 5,
        current_students: 3,
        homework_count: 1,
        broadcasts_count: 0,
      };

      lessonsAPI.getTeacherSchedule.mockResolvedValue({
        lessons: [mockLesson],
        count: 1,
      });

      renderWithClient(<TeacherCalendar />);

      await waitFor(
        () => {
          expect(screen.getByText("Прошедший урок")).toBeInTheDocument();
        },
        { timeout: 3000 },
      );

      // Проверяем статус
      expect(screen.getByText(/Завершено/)).toBeInTheDocument();
    });

    test("НЕ должен отображаться статус 'Завершено' для будущих занятий", async () => {
      const futureTime = new Date();
      futureTime.setHours(futureTime.getHours() + 2); // 2 часа в будущем

      const mockLesson = {
        id: 1,
        subject: "Будущий урок",
        start_time: futureTime.toISOString(),
        max_students: 5,
        current_students: 2,
        homework_count: 0,
        broadcasts_count: 0,
      };

      lessonsAPI.getTeacherSchedule.mockResolvedValue({
        lessons: [mockLesson],
        count: 1,
      });

      renderWithClient(<TeacherCalendar />);

      await waitFor(
        () => {
          expect(screen.getByText("Будущий урок")).toBeInTheDocument();
        },
        { timeout: 3000 },
      );

      // Статус "Завершено" НЕ должен быть
      expect(screen.queryByText(/Завершено/)).not.toBeInTheDocument();
    });
  });

  describe("Компилирование и отсутствие ошибок", () => {
    test("компонент должен рендериться без ошибок", async () => {
      lessonsAPI.getTeacherSchedule.mockResolvedValue({
        lessons: [],
        count: 0,
      });

      const consoleErrorSpy = vi
        .spyOn(console, "error")
        .mockImplementation(() => {});

      renderWithClient(<TeacherCalendar />);

      await waitFor(
        () => {
          expect(lessonsAPI.getTeacherSchedule).toHaveBeenCalled();
        },
        { timeout: 3000 },
      );

      // Ошибок в консоли быть не должно
      expect(consoleErrorSpy).not.toHaveBeenCalled();

      consoleErrorSpy.mockRestore();
    });

    test("не должно быть warning о homework_count в DOM", async () => {
      const today = new Date();
      today.setHours(10, 0, 0, 0);

      const mockLesson = {
        id: 1,
        subject: "Тест",
        start_time: today.toISOString(),
        max_students: 5,
        current_students: 2,
        homework_count: 10,
        broadcasts_count: 0,
      };

      lessonsAPI.getTeacherSchedule.mockResolvedValue({
        lessons: [mockLesson],
        count: 1,
      });

      renderWithClient(<TeacherCalendar />);

      await waitFor(
        () => {
          expect(screen.getByText("Тест")).toBeInTheDocument();
        },
        { timeout: 3000 },
      );

      // В DOM не должно быть никаких reference к homework badge
      const lessonElements = screen.getAllByText(/Тест/);
      for (const elem of lessonElements) {
        const container = elem.closest("div");
        // Не должно быть .homework-badge класса
        const homeworkBadge = container?.querySelector(".homework-badge");
        expect(homeworkBadge).not.toBeInTheDocument();
      }
    });
  });

  describe("Edge Cases - Граничные случаи", () => {
    test("должен корректно работать при homework_count = 0", async () => {
      const today = new Date();
      today.setHours(10, 0, 0, 0);

      const mockLesson = {
        id: 1,
        subject: "Без домашек",
        start_time: today.toISOString(),
        max_students: 5,
        current_students: 2,
        homework_count: 0,
        broadcasts_count: 0,
      };

      lessonsAPI.getTeacherSchedule.mockResolvedValue({
        lessons: [mockLesson],
        count: 1,
      });

      renderWithClient(<TeacherCalendar />);

      await waitFor(
        () => {
          expect(screen.getByText("Без домашек")).toBeInTheDocument();
        },
        { timeout: 3000 },
      );

      // Скрепка не должна быть
      expect(screen.queryByText(/📎/)).not.toBeInTheDocument();
    });

    test("должен работать при отсутствии поля homework_count", async () => {
      const today = new Date();
      today.setHours(10, 0, 0, 0);

      const mockLesson = {
        id: 1,
        subject: "Тест",
        start_time: today.toISOString(),
        max_students: 5,
        current_students: 2,
        // homework_count отсутствует
        broadcasts_count: 1,
      };

      lessonsAPI.getTeacherSchedule.mockResolvedValue({
        lessons: [mockLesson],
        count: 1,
      });

      renderWithClient(<TeacherCalendar />);

      await waitFor(
        () => {
          expect(screen.getByText("Тест")).toBeInTheDocument();
        },
        { timeout: 3000 },
      );

      // Скрепка не должна быть
      expect(screen.queryByText(/📎/)).not.toBeInTheDocument();
      // Но рассылка должна быть
      expect(screen.getByText(/📢 1/)).toBeInTheDocument();
    });

    test("должен работать при null/undefined в homework_count", async () => {
      const today = new Date();
      today.setHours(10, 0, 0, 0);

      const mockLessons = [
        {
          id: 1,
          subject: "Null",
          start_time: today.toISOString(),
          max_students: 5,
          current_students: 2,
          homework_count: null,
          broadcasts_count: 0,
        },
        {
          id: 2,
          subject: "Undefined",
          start_time: new Date(today.getTime() + 3600000).toISOString(),
          max_students: 5,
          current_students: 2,
          homework_count: undefined,
          broadcasts_count: 0,
        },
      ];

      lessonsAPI.getTeacherSchedule.mockResolvedValue({
        lessons: mockLessons,
        count: 2,
      });

      renderWithClient(<TeacherCalendar />);

      await waitFor(
        () => {
          expect(screen.getByText("Null")).toBeInTheDocument();
          expect(screen.getByText("Undefined")).toBeInTheDocument();
        },
        { timeout: 3000 },
      );

      // Скрепка не должна быть нигде
      expect(screen.queryByText(/📎/)).not.toBeInTheDocument();
    });
  });
});
