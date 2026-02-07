import { useState, useEffect, useCallback, useRef } from "react";
import { useQueryClient } from "@tanstack/react-query";
import Modal from "../common/Modal.jsx";
import Button from "../common/Button.jsx";
import Spinner from "../common/Spinner.jsx";
import ConfirmModal from "../common/ConfirmModal.jsx";
import ColorPicker from "../common/ColorPicker.jsx";
import BulkEditConfirmationDialog from "./BulkEditConfirmationDialog.jsx";
import { useNotification } from "../../hooks/useNotification.js";
import { useApplyToAllSubsequent } from "../../hooks/useBulkEdit.js";
import { useAutosave } from "../../hooks/useAutosave.js";
import {
  detectModificationType,
  getModificationDetails,
} from "../../utils/lessonModificationDetector.js";
import { DEFAULT_RECURRING_WEEKS } from "../../config/constants.js";
import { createRecurringSeries } from "../../api/lessons.js";
import * as bookingAPI from "../../api/bookings.js";
import * as userAPI from "../../api/users.js";
import * as creditAPI from "../../api/credits.js";
import {
  invalidateLessonData,
  invalidateBookingData,
  invalidateCreditData,
} from "../../utils/queryInvalidation.js";
import {
  allSettledWithLabels,
  withFallback,
} from "../../utils/promiseHelpers.js";
import HomeworkSection from "./HomeworkSection.jsx";
import BroadcastSection from "./BroadcastSection.jsx";
import StudentCheckboxList from "./StudentCheckboxList.jsx";
import { useAuth } from "../../hooks/useAuth.js";
import { ROLES } from "../../utils/constants.js";
import "./LessonEditModal.css";

/**
 * Modal для редактирования занятия
 * Позволяет просмотреть студентов, добавить/удалить их, удалить занятие
 */
export const LessonEditModal = ({
  isOpen,
  onClose,
  lesson,
  onLessonUpdated,
  onStudentSelected,
}) => {
  const { showNotification } = useNotification();
  const queryClient = useQueryClient();
  const applyToAllMutation = useApplyToAllSubsequent();
  const { user } = useAuth();

  // Состояния
  const [students, setStudents] = useState([]);
  const [originalStudents, setOriginalStudents] = useState([]); // Track original for change detection
  const [availableStudents, setAvailableStudents] = useState([]);
  const [selectedStudentId, setSelectedStudentId] = useState("");
  const [studentCredits, setStudentCredits] = useState({});

  const [teachers, setTeachers] = useState([]);
  const [selectedTeacherId, setSelectedTeacherId] = useState("");

  const [selectedSubject, setSelectedSubject] = useState("");

  const [selectedLink, setSelectedLink] = useState("");

  const [selectedColor, setSelectedColor] = useState("#2563eb");

  const [loading, setLoading] = useState(false);
  const [addingStudent, setAddingStudent] = useState(false);
  const [removingStudentId, setRemovingStudentId] = useState(null);
  const [deletingLesson, setDeletingLesson] = useState(false);

  const [showRemoveConfirm, setShowRemoveConfirm] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [confirmRemoveStudentId, setConfirmRemoveStudentId] = useState(null);

  // Локальное состояние занятия для управления данными
  const [lessonData, setLessonData] = useState(null);

  // Bulk edit states
  const [showBulkEditConfirmation, setShowBulkEditConfirmation] =
    useState(false);
  const [bulkEditDetails, setBulkEditDetails] = useState(null);
  const [hasChanges, setHasChanges] = useState(false);

  // Responsive state for button text
  const [isMobile, setIsMobile] = useState(window.innerWidth < 480);

  // Tab navigation state
  const [activeTab, setActiveTab] = useState("info"); // 'info', 'homework', 'broadcasts'

  // Homework count badge
  const [homeworkCount, setHomeworkCount] = useState(0);

  // Info tab form states
  const [lessonDate, setLessonDate] = useState("");
  const [startTime, setStartTime] = useState("");
  const [endTime, setEndTime] = useState("");
  const [maxStudents, setMaxStudents] = useState(1);
  const [creditsCost, setCreditsCost] = useState(1);

  // Recurring lesson states
  const [isRecurring, setIsRecurring] = useState(false);

  // Form validation errors
  const [formErrors, setFormErrors] = useState({});

  // Autosave state
  const [autosaveData, setAutosaveData] = useState(null);
  const closeModalRef = useRef(false);
  const [isInitialized, setIsInitialized] = useState(false); // Флаг завершения инициализации

  // Загрузка данных при открытии modal
  // Функция loadLessonData определена в теле компонента и безопасна для использования в зависимостях
  useEffect(() => {
    if (isOpen && lesson?.id) {
      setIsInitialized(false); // Сброс флага при открытии
      loadLessonData();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isOpen, lesson?.id, lesson]);

  // Отслеживание размера окна для адаптивного текста кнопки
  useEffect(() => {
    const handleResize = () => setIsMobile(window.innerWidth < 480);
    window.addEventListener("resize", handleResize);
    return () => window.removeEventListener("resize", handleResize);
  }, []);

  // Initialize form fields when lesson data loads
  useEffect(() => {
    if (lesson) {
      const startDate = new Date(lesson.start_time);
      const endDate = new Date(lesson.end_time);

      // Format date as YYYY-MM-DD for input[type="date"] using LOCAL timezone
      // ВАЖНО: НЕ использовать toISOString() - он конвертирует в UTC и сдвигает дату!
      const year = startDate.getFullYear();
      const month = String(startDate.getMonth() + 1).padStart(2, "0");
      const day = String(startDate.getDate()).padStart(2, "0");
      const dateStr = `${year}-${month}-${day}`;

      // Format time as HH:MM for input[type="time"]
      const startTimeStr = `${String(startDate.getHours()).padStart(2, "0")}:${String(startDate.getMinutes()).padStart(2, "0")}`;
      const endTimeStr = `${String(endDate.getHours()).padStart(2, "0")}:${String(endDate.getMinutes()).padStart(2, "0")}`;

      setLessonDate(dateStr);
      setStartTime(startTimeStr);
      setEndTime(endTimeStr);
      setMaxStudents(lesson.max_students || 1);
      setCreditsCost(lesson.credits_cost || 1);
    }
  }, [lesson]);

  // Notify parent when student is selected for adding
  useEffect(() => {
    if (selectedStudentId && onStudentSelected) {
      onStudentSelected(selectedStudentId);
    }
  }, [selectedStudentId, onStudentSelected]);

  /**
   * Валидация формы для отображения ошибок
   */
  const validateForm = useCallback(() => {
    const newErrors = {};

    // Преподаватель обязателен
    if (!selectedTeacherId) {
      newErrors.teacher_id = "Пожалуйста, выберите преподавателя";
    }

    // Время начала обязательно
    if (!startTime) {
      newErrors.start_time = "Время начала обязательно";
    }

    // Время окончания обязательно
    if (!endTime) {
      newErrors.end_time = "Время окончания обязательно";
    }

    // Проверить end_time > start_time
    if (startTime && endTime) {
      const [startH, startM] = startTime.split(":").map(Number);
      const [endH, endM] = endTime.split(":").map(Number);
      const startMinutes = startH * 60 + startM;
      const endMinutes = endH * 60 + endM;

      if (endMinutes <= startMinutes) {
        newErrors.end_time = "Время окончания должно быть после времени начала";
      }
    }

    // Max students 1-20
    if (maxStudents < 1 || maxStudents > 20) {
      newErrors.max_students = "Максимум студентов должен быть от 1 до 20";
    }

    setFormErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  }, [selectedTeacherId, startTime, endTime, maxStudents]);

  /**
   * Валидация данных перед сохранением
   */
  const validateData = useCallback((data) => {
    // Валидация teacher_id - должен быть валидный UUID или пустая строка
    if (
      data.teacher_id !== undefined &&
      data.teacher_id !== "" &&
      data.teacher_id !== null
    ) {
      const uuidRegex =
        /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
      if (!uuidRegex.test(data.teacher_id)) {
        console.warn("Invalid teacher_id format:", data.teacher_id);
        return false;
      }
    }

    // Валидация maxStudents
    if (data.maxStudents !== undefined && data.maxStudents !== null) {
      if (typeof data.maxStudents !== "number" || data.maxStudents < 1) {
        console.warn("Invalid maxStudents:", data.maxStudents);
        return false;
      }
    }

    // Валидация creditsCost
    if (data.creditsCost !== undefined && data.creditsCost !== null) {
      if (typeof data.creditsCost !== "number" || data.creditsCost < 0) {
        console.warn("Invalid creditsCost:", data.creditsCost);
        return false;
      }
    }

    // Валидация даты и времени
    if (data.lessonDate && data.startTime && data.endTime) {
      try {
        const startDate = new Date(`${data.lessonDate}T${data.startTime}`);
        const endDate = new Date(`${data.lessonDate}T${data.endTime}`);

        if (isNaN(startDate.getTime()) || isNaN(endDate.getTime())) {
          console.warn("Invalid date/time format");
          return false;
        }

        if (endDate <= startDate) {
          console.warn("End time must be after start time");
          return false;
        }
      } catch (e) {
        console.warn("Date validation error:", e);
        return false;
      }
    }

    // Валидация color (hex формат)
    if (data.color !== undefined && data.color !== null && data.color !== "") {
      const hexRegex = /^#[0-9A-F]{6}$/i;
      if (!hexRegex.test(data.color)) {
        console.warn("Invalid color format:", data.color);
        return false;
      }
    }

    // Валидация subject (макс 200 символов)
    if (data.subject !== undefined && data.subject !== null) {
      if (typeof data.subject !== "string" || data.subject.length > 200) {
        console.warn("Invalid subject length:", data.subject?.length);
        return false;
      }
    }

    return true;
  }, []);

  /**
   * Функция автосохранения изменений занятия
   */
  const saveChanges = useCallback(
    async (data) => {
      if (!lesson?.id || !data) return;

      // Валидация данных
      if (!validateData(data)) {
        console.error("Data validation failed, skipping autosave");
        return;
      }

      try {
        const updates = {};

        // Построить обновления из data
        // teacher_id - конвертировать пустую строку в null
        if (data.teacher_id !== undefined) {
          const newTeacherId = data.teacher_id === "" ? null : data.teacher_id;
          const currentTeacherId = lesson.teacher_id || null;
          if (newTeacherId !== currentTeacherId) {
            updates.teacher_id = newTeacherId;
          }
        }

        // subject - trim и конвертировать пустую строку в null
        if (data.subject !== undefined) {
          const trimmedSubject = (data.subject || "").trim();
          const newSubject =
            trimmedSubject === "" ? null : trimmedSubject.slice(0, 200);
          const currentSubject = (lesson.subject || "").trim() || null;
          if (newSubject !== currentSubject) {
            updates.subject = newSubject;
          }
        }

        // link - trim и конвертировать пустую строку в null
        if (data.link !== undefined) {
          const trimmedLink = (data.link || "").trim();
          const newLink = trimmedLink === "" ? null : trimmedLink;
          const currentLink = (lesson.link || "").trim() || null;
          if (newLink !== currentLink) {
            updates.link = newLink;
          }
        }

        // color - валидация hex
        if (data.color !== undefined) {
          const newColor = data.color || "#2563eb";
          const currentColor = lesson.color || "#2563eb";
          if (newColor !== currentColor) {
            updates.color = newColor;
          }
        }

        // Проверить изменения даты/времени с ISO формат
        if (data.lessonDate && data.startTime && data.endTime) {
          // Создать Date объекты и конвертировать в ISO 8601
          const startDate = new Date(`${data.lessonDate}T${data.startTime}`);
          const endDate = new Date(`${data.lessonDate}T${data.endTime}`);

          const newStartTime = startDate.toISOString();
          const newEndTime = endDate.toISOString();

          // Сравнить с текущими значениями
          const currentStartTime = new Date(lesson.start_time).toISOString();
          const currentEndTime = new Date(lesson.end_time).toISOString();

          if (newStartTime !== currentStartTime) {
            updates.start_time = newStartTime;
          }
          if (newEndTime !== currentEndTime) {
            updates.end_time = newEndTime;
          }
        }

        // maxStudents - обработка null/undefined
        if (data.maxStudents !== undefined && data.maxStudents !== null) {
          const newMaxStudents = parseInt(data.maxStudents, 10);
          const currentMaxStudents = lesson.max_students || 1;
          if (newMaxStudents !== currentMaxStudents && newMaxStudents >= 1) {
            updates.max_students = newMaxStudents;
          }
        }

        // creditsCost - обработка null/undefined
        if (data.creditsCost !== undefined && data.creditsCost !== null) {
          const newCreditsCost = parseInt(data.creditsCost, 10);
          const currentCreditsCost = lesson.credits_cost || 0;
          if (newCreditsCost !== currentCreditsCost && newCreditsCost >= 0) {
            updates.credits_cost = newCreditsCost;
          }
        }

        // Если есть изменения, отправить на сервер
        if (Object.keys(updates).length > 0) {
          console.log("[Autosave] Sending updates:", updates);
          await lessonAPI.updateLesson(lesson.id, updates);

          // Инвалидация кэша
          invalidateLessonData(queryClient);

          // НЕ вызываем onLessonUpdated здесь - это закроет модалку
          // onLessonUpdated вызывается только при явном закрытии модалки
        } else {
          console.log("[Autosave] No changes detected, skipping save");
        }
      } catch (error) {
        console.error("Autosave error:", error);
        const errorMsg =
          error.response?.data?.message ||
          error.response?.data?.error ||
          "Не удалось сохранить изменения";
        showNotification(errorMsg, "error");
        throw error;
      }
    },
    [lesson, queryClient, onLessonUpdated, showNotification, validateData],
  );

  /**
   * Обновить данные для автосохранения при изменении полей
   * НЕ обновлять пока не завершена инициализация (чтобы не триггерить автосохранение при загрузке)
   */
  useEffect(() => {
    if (!lesson || !isInitialized) return;

    setAutosaveData({
      teacher_id: selectedTeacherId,
      subject: selectedSubject,
      link: selectedLink,
      color: selectedColor,
      lessonDate,
      startTime,
      endTime,
      maxStudents,
      creditsCost,
    });
  }, [
    lesson,
    isInitialized,
    selectedTeacherId,
    selectedSubject,
    selectedLink,
    selectedColor,
    lessonDate,
    startTime,
    endTime,
    maxStudents,
    creditsCost,
  ]);

  /**
   * Интеграция хука автосохранения
   */
  const {
    isSaving: isAutosaving,
    saveNow,
    error: autosaveError,
    lastSaved,
  } = useAutosave(
    autosaveData,
    saveChanges,
    500, // debounce 500ms
  );

  /**
   * Сохранить изменения при закрытии модалки
   */
  const handleClose = useCallback(async () => {
    if (closeModalRef.current) return; // Уже закрываем
    closeModalRef.current = true;

    try {
      // Сохранить все несохраненные изменения
      if (autosaveData && !isAutosaving) {
        await saveNow();
      }
    } catch (error) {
      console.error("Error saving on close:", error);
    } finally {
      closeModalRef.current = false;
      // Уведомить родителя об обновлении при закрытии
      onLessonUpdated?.(lessonData);
      onClose();
    }
  }, [
    autosaveData,
    isAutosaving,
    saveNow,
    onClose,
    onLessonUpdated,
    lessonData,
  ]);

  // Detect changes for "Apply to all subsequent" button
  useEffect(() => {
    if (!lesson || !lessonData) {
      setHasChanges(false);
      return;
    }

    // Check for any changes
    const modificationType = detectModificationType(
      lesson,
      {
        teacher_id: selectedTeacherId,
        max_students: lessonData.max_students,
        start_time: lessonData.start_time,
      },
      originalStudents,
      students,
    );

    setHasChanges(!!modificationType);
  }, [lesson, lessonData, selectedTeacherId, originalStudents, students]);

  /**
   * Загрузить все необходимые данные для modal
   */
  const loadLessonData = async () => {
    try {
      setLoading(true);

      // Use Promise.allSettled to handle partial failures
      // NOTE: getBookings with lesson_id filter is now more reliable with client-side fallback
      const { results, failures } = await allSettledWithLabels(
        [
          bookingAPI.getBookings({ lesson_id: lesson.id }),
          userAPI.getStudentsAll(),
          creditAPI.getAllCredits(),
          userAPI.getAssignableTeachersAll(),
        ],
        ["Бронирования", "Студенты", "Кредиты", "Преподаватели"],
      );

      // Extract data with fallbacks
      const bookingsResponse = withFallback(results[0], []);
      const allStudentsResponse = withFallback(results[1], []);
      const allCreditsResponse = withFallback(results[2], { balances: [] });
      const teachersResponse = withFallback(results[3], []);

      // Фильтруем студентов с активными бронированиями
      const activeBookings = Array.isArray(bookingsResponse)
        ? bookingsResponse.filter((b) => b.status === "active")
        : [];

      const enrolledStudentIds = new Set(
        activeBookings.map((b) => b.student_id || b.user_id),
      );

      const studentsList = activeBookings.map((booking) => ({
        id: booking.id, // booking ID для отмены бронирования
        booking_id: booking.id, // дублируем для ясности
        student_id: booking.student_id || booking.user_id,
        student_name:
          booking.student_name || booking.user_name || "Неизвестный студент",
        student_email:
          booking.student_email || booking.user_email || "Нет email",
      }));

      setStudents(studentsList);
      setOriginalStudents(studentsList); // Save original for change detection

      // Создать карту кредитов
      // API возвращает { balances: [{user_id, balance}, ...] }
      const creditsMap = {};

      // Логирование ДО парсинга (для диагностики)
      console.debug("[LessonEditModal] allCreditsResponse structure:", {
        hasBalances: !!allCreditsResponse?.balances,
        balancesIsArray: Array.isArray(allCreditsResponse?.balances),
        balancesLength: allCreditsResponse?.balances?.length,
        fullResponse: JSON.stringify(allCreditsResponse).substring(0, 200),
      });

      // Парсинг creditsMap с лучшей валидацией
      if (allCreditsResponse) {
        const balances = allCreditsResponse.balances || [];

        if (Array.isArray(balances) && balances.length > 0) {
          balances.forEach(({ user_id, balance }) => {
            if (user_id && typeof balance === "number") {
              creditsMap[user_id] = balance;
            } else if (user_id) {
              creditsMap[user_id] = Number(balance) || 0;
            }
          });
        }
      }

      // Логирование ПОСЛЕ парсинга
      console.debug("[LessonEditModal] creditsMap built:", {
        size: Object.keys(creditsMap).length,
        isEmpty: Object.keys(creditsMap).length === 0,
        sampleEntries: Object.entries(creditsMap).slice(0, 3),
      });

      // WARNING если creditsMap пуста
      if (Object.keys(creditsMap).length === 0) {
        console.warn(
          "[LessonEditModal] WARNING: creditsMap is empty! Students will show 0 credits.",
          "This may indicate API response format changed or load failed.",
        );
      }

      setStudentCredits(creditsMap);

      // Доступные студенты (не записанные на это занятие)
      const available = Array.isArray(allStudentsResponse)
        ? allStudentsResponse.filter((s) => !enrolledStudentIds.has(s.id))
        : [];

      setAvailableStudents(available);
      setSelectedStudentId("");

      // Установить список преподавателей и текущего преподавателя
      // Для методиста показываем только его самого
      const isUserTeacher = user?.role === ROLES.TEACHER;
      let teachersList;

      if (isUserTeacher && user?.id) {
        // Методист видит только себя в списке
        teachersList = [
          {
            id: user.id,
            full_name: user.full_name || user.name || user.email,
          },
        ];
      } else {
        teachersList = Array.isArray(teachersResponse) ? teachersResponse : [];
      }

      setTeachers(teachersList);
      setSelectedTeacherId(lesson.teacher_id || "");

      // Установить subject
      setSelectedSubject(lesson.subject || "");

      // Установить link
      setSelectedLink(lesson.link || "");

      // Установить color
      setSelectedColor(lesson.color || "#2563eb");

      // Инициализировать локальное состояние занятия для управления данными
      setLessonData({
        ...lesson,
        current_students: studentsList.length,
      });

      // Отметить завершение инициализации (для предотвращения автосохранения при загрузке)
      // Задержка чтобы дать useEffect с lesson обработать данные
      setTimeout(() => setIsInitialized(true), 100);

      // Show warnings for failed items, but only for critical data
      // Filter out non-critical failures (Бронирования can be derived from StudentCheckboxList)
      const criticalFailures = failures.filter(
        (f) => f.label === "Студенты" || f.label === "Преподаватели",
      );

      if (criticalFailures.length > 0) {
        criticalFailures.forEach(({ label, error }) => {
          console.error(`Критический сбой загрузки: ${label}`, error);
        });
        showNotification(
          `Не удалось загрузить: ${criticalFailures.map((f) => f.label).join(", ")}`,
          "error",
        );
      }

      // Log non-critical failures for debugging
      const nonCriticalFailures = failures.filter(
        (f) => !criticalFailures.includes(f),
      );
      if (nonCriticalFailures.length > 0) {
        nonCriticalFailures.forEach(({ label, error }) => {
          console.warn(
            `Частичный сбой загрузки (некритичный): ${label}`,
            error,
          );
        });
      }
    } catch (error) {
      console.error("Ошибка загрузки данных занятия:", error);
      showNotification("Ошибка загрузки данных занятия", "error");
    } finally {
      setLoading(false);
    }
  };

  /**
   * Добавить студента на занятие
   * Админ добавляет студента - API вызывается с lesson_id и student_id
   * Использует оптимистичные обновления
   * Администратор может добавлять студентов без проверки кредитов
   */
  const handleAddStudent = async () => {
    if (!selectedStudentId) {
      showNotification("Выберите студента", "error");
      return;
    }

    const student = availableStudents.find((s) => s.id === selectedStudentId);
    const credits = studentCredits[selectedStudentId] || 0;

    // DEBUG: проверка ID студента и кредитов
    console.debug("[handleAddStudent] Проверка студента:", {
      selectedStudentId,
      studentExists: !!student,
      studentData: {
        id: student?.id,
        full_name: student?.full_name,
      },
      credits,
      creditsMapHasKey: selectedStudentId in studentCredits,
      creditsMapSample: Object.entries(studentCredits).slice(0, 3),
    });

    // Никто не может добавлять студентов без кредитов (включая администратора)
    if (credits < 1) {
      showNotification(
        `У студента ${student?.full_name || "этого"} недостаточно кредитов`,
        "error",
      );
      return;
    }

    try {
      setAddingStudent(true);

      // Оптимистичное обновление: мгновенно обновить UI
      const newStudent = {
        id: `temp_${selectedStudentId}`,
        student_id: selectedStudentId,
        student_name: student?.full_name || "Новый студент",
        student_email: student?.email || "",
      };
      setStudents((prev) => [...prev, newStudent]);
      setLessonData((prev) => ({
        ...prev,
        current_students: (prev?.current_students || 0) + 1,
      }));

      // Удалить из доступных студентов
      setAvailableStudents((prev) =>
        prev.filter((s) => s.id !== selectedStudentId),
      );

      // Создать бронирование с lesson_id и student_id (админ добавляет студента)
      await bookingAPI.createBooking(lesson.id, selectedStudentId);

      showNotification(
        `${student?.full_name} успешно добавлен на занятие`,
        "success",
      );
      setSelectedStudentId("");

      // Перезагрузить данные занятия чтобы получить реальные ID бронирования
      await loadLessonData();

      // Инвалидировать связанные кеши для обновления данных по всему приложению
      invalidateLessonData(queryClient);
      invalidateBookingData(queryClient);
      invalidateCreditData(queryClient);

      // НЕ вызываем onLessonUpdated - это закроет модалку
    } catch (error) {
      console.error("Ошибка добавления студента:", error);
      const errorCode = error?.response?.data?.error_code;
      let errorMsg = "Не удалось добавить студента";

      if (errorCode === "LESSON_FULL") {
        errorMsg = "На занятии нет свободных мест";
      } else if (errorCode === "INSUFFICIENT_CREDITS") {
        errorMsg = "У студента недостаточно кредитов";
      } else if (errorCode === "STUDENT_ALREADY_BOOKED") {
        errorMsg = "Студент уже зарегистрирован на это занятие";
      } else if (errorCode === "CONFLICT") {
        errorMsg = "Студент уже записан на это занятие";
      } else if (error?.response?.data?.message) {
        errorMsg = error.response.data.message;
      } else if (error?.message) {
        errorMsg = error.message;
      }

      showNotification(errorMsg, "error");
      // Откатить оптимистичные изменения
      await loadLessonData();
    } finally {
      setAddingStudent(false);
    }
  };

  /**
   * Отменить запись студента
   */
  const handleRemoveStudent = async (bookingId) => {
    setConfirmRemoveStudentId(bookingId);
    setShowRemoveConfirm(true);
  };

  /**
   * Подтверждение удаления студента с оптимистичным скрытием
   * Отменяет бронирование и возвращает кредиты студенту
   */
  const confirmRemoveStudent = async () => {
    const studentToRemove = students.find(
      (s) => s.id === confirmRemoveStudentId,
    );

    try {
      setRemovingStudentId(confirmRemoveStudentId);

      // Оптимистичное обновление: скрыть студента (затемнить строку)
      setStudents((prev) =>
        prev.map((s) =>
          s.id === confirmRemoveStudentId ? { ...s, removing: true } : s,
        ),
      );

      // Отменить бронирование через API (backend автоматически вернёт кредиты)
      await bookingAPI.cancelBooking(confirmRemoveStudentId);

      showNotification("Студент успешно удалён с занятия", "success");

      // Удалить студента из списка после успешного ответа
      setStudents((prev) =>
        prev.filter((s) => s.id !== confirmRemoveStudentId),
      );

      // Обновить данные занятия
      setLessonData((prev) => ({
        ...prev,
        current_students: Math.max(0, (prev?.current_students || 0) - 1),
      }));

      // Добавить студента обратно в доступных
      if (studentToRemove) {
        const studentData = { ...studentToRemove };
        delete studentData.removing;
        const allStudentsResponse = await userAPI.getStudentsAll();
        const originalStudent = allStudentsResponse.find(
          (s) => s.id === studentToRemove.student_id,
        );
        if (originalStudent) {
          setAvailableStudents((prev) => [...prev, originalStudent]);
        }
      }

      // Перезагрузить данные занятия для синхронизации
      await loadLessonData();

      // Инвалидировать связанные кеши для обновления данных по всему приложению
      invalidateLessonData(queryClient);
      invalidateBookingData(queryClient);
      invalidateCreditData(queryClient);

      // НЕ вызываем onLessonUpdated - это закроет модалку

      // Закрыть модаль подтверждения только при успехе
      setShowRemoveConfirm(false);
      setConfirmRemoveStudentId(null);
    } catch (error) {
      console.error("Ошибка удаления студента:", error);
      const errorMsg =
        error?.response?.data?.message ||
        error?.message ||
        "Не удалось удалить студента";

      // Откатить оптимистичное обновление при ошибке
      setStudents((prev) =>
        prev.map((s) =>
          s.id === confirmRemoveStudentId ? { ...s, removing: undefined } : s,
        ),
      );

      showNotification(errorMsg, "error");
      // При ошибке НЕ закрываем модаль и НЕ очищаем данные
      // Пользователь видит статус и может повторить попытку
    } finally {
      setRemovingStudentId(null);
    }
  };

  /**
   * Универсальный обработчик добавления/удаления студентов для StudentCheckboxList
   * @param {string} studentId - ID студента
   * @param {boolean} shouldEnroll - true = добавить, false = удалить
   */
  const handleStudentToggle = async (studentId, shouldEnroll) => {
    try {
      if (shouldEnroll) {
        // Добавить студента - валидировать кредиты ПЕРЕД API вызовом
        const credits = studentCredits[studentId] ?? 0;
        const requiredCredits = lesson?.credits_cost || 1;

        // DEBUG: проверка ID и кредитов при добавлении
        console.debug("[handleStudentToggle] Добавление студента:", {
          studentId,
          credits,
          requiredCredits,
          creditsMapHasKey: studentId in studentCredits,
          creditsValidation: credits >= requiredCredits,
        });

        if (credits < requiredCredits) {
          console.warn(
            "[handleStudentToggle] Блокировка добавления: недостаточно кредитов",
            {
              studentId,
              credits,
              required: requiredCredits,
            },
          );
          showNotification(
            `Студент имеет недостаточно кредитов (${credits}, требуется ${requiredCredits})`,
            "error",
          );
          return; // НЕ запускать API вызов и НЕ менять UI
        }

        // Кредиты достаточны - добавить студента
        setAddingStudent(true);
        await bookingAPI.createBooking(lesson.id, studentId);
        showNotification("Студент успешно добавлен", "success");
      } else {
        // Удалить студента
        const booking = students.find((s) => s.student_id === studentId);

        // DEBUG: проверка ID при удалении
        const bookingId = booking?.booking_id || booking?.id;
        console.debug("[handleStudentToggle] Удаление студента:", {
          studentId,
          bookingFound: !!booking,
          bookingData: booking
            ? {
                booking_id: bookingId,
                student_id: booking.student_id,
                student_name: booking.student_name,
              }
            : null,
        });

        if (booking && bookingId) {
          setRemovingStudentId(studentId);
          await bookingAPI.cancelBooking(bookingId);
          showNotification("Студент успешно удалён", "success");
        } else {
          console.error(
            "Не найден booking_id для студента:",
            studentId,
            booking,
          );
          showNotification("Ошибка: не найден ID бронирования", "error");
          return;
        }
      }

      // Перезагрузить данные занятия
      await loadLessonData();

      // Инвалидировать связанные кеши
      invalidateLessonData(queryClient);
      invalidateBookingData(queryClient);
      invalidateCreditData(queryClient);
    } catch (error) {
      console.error("Ошибка при изменении записи:", error);
      const errorMsg =
        error?.response?.data?.message ||
        error?.message ||
        "Ошибка при изменении записи";
      showNotification(errorMsg, "error");
    } finally {
      setAddingStudent(false);
      setRemovingStudentId(null);
    }
  };

  /**
   * Удалить занятие
   */
  const handleDeleteLesson = async () => {
    setShowDeleteConfirm(true);
  };

  /**
   * Подтверждение удаления занятия
   */
  const confirmDeleteLesson = async () => {
    try {
      setDeletingLesson(true);

      // Удалить занятие (backend автоматически вернёт кредиты всем студентам)
      await lessonAPI.deleteLesson(lesson.id);

      showNotification("Занятие успешно удалено", "success");
      setShowDeleteConfirm(false);
      onClose();
      onLessonUpdated?.(null);
    } catch (error) {
      console.error("Ошибка удаления занятия:", error);
      const errorMsg =
        error?.response?.data?.message ||
        error?.message ||
        "Не удалось удалить занятие";
      showNotification(errorMsg, "error");
    } finally {
      setDeletingLesson(false);
    }
  };

  /**
   * Handle "Apply to all subsequent" button click
   * Detects modification type and shows confirmation dialog
   */
  const handleApplyToAll = () => {
    console.log("[LessonEditModal] Apply to All button clicked");
    console.log("[LessonEditModal] hasChanges:", hasChanges);

    if (!hasChanges) {
      showNotification("Изменений не обнаружено", "error");
      return;
    }

    // Detect what changed
    const modificationType = detectModificationType(
      lesson,
      {
        teacher_id: selectedTeacherId,
        max_students: lessonData?.max_students,
        start_time: lessonData?.start_time,
      },
      originalStudents,
      students,
    );

    console.log(
      "[LessonEditModal] Detected modification type:",
      modificationType,
    );

    if (!modificationType) {
      showNotification("Не удалось определить тип изменения", "error");
      return;
    }

    // Get modification details
    const details = getModificationDetails(
      modificationType,
      lesson,
      {
        teacher_id: selectedTeacherId,
        max_students: lessonData?.max_students,
        start_time: lessonData?.start_time,
      },
      originalStudents,
      students,
      teachers,
    );

    console.log("[LessonEditModal] Modification details:", details);

    // Store details and show confirmation
    setBulkEditDetails({
      modificationType,
      ...details,
    });
    setShowBulkEditConfirmation(true);
  };

  /**
   * Обработчик создания серии повторяющихся занятий
   */
  const handleCreateRecurringSeries = async () => {
    if (!lesson?.id) return;
    const confirmed = window.confirm(
      `Создать серию из ${DEFAULT_RECURRING_WEEKS} повторяющихся занятий?`,
    );
    if (!confirmed) return;
    try {
      const result = await createRecurringSeries(
        lesson.id,
        DEFAULT_RECURRING_WEEKS,
      );
      showNotification(`Создано ${result.data?.count || 0} занятий`, "success");
      await loadLessonData();
      invalidateLessonData(queryClient);
      onClose();
      onLessonUpdated?.(lesson);
    } catch (error) {
      showNotification(error.response?.data?.message || error.message, "error");
    }
  };

  /**
   * Confirm bulk edit and apply to all subsequent lessons
   */
  const handleConfirmBulkEdit = async () => {
    if (!bulkEditDetails) return;

    console.log(
      "[LessonEditModal] Confirming bulk edit with details:",
      bulkEditDetails,
    );

    try {
      const { modificationType, ...params } = bulkEditDetails;

      console.log("[LessonEditModal] Calling mutation with:", {
        lessonId: lesson.id,
        modificationType,
        params,
      });

      const result = await applyToAllMutation.mutateAsync({
        lessonId: lesson.id,
        modificationType,
        ...params,
      });

      console.log("[LessonEditModal] Mutation result:", result);

      const affectedCount = result?.affected_lessons_count || 0;
      console.log("[LessonEditModal] Affected count:", affectedCount);

      showNotification(
        `Успешно применено к ${affectedCount} занятию(ям)`,
        "success",
      );

      setShowBulkEditConfirmation(false);
      setBulkEditDetails(null);

      // Reload lesson data (НЕ закрываем модалку)
      await loadLessonData();
      invalidateLessonData(queryClient);
      invalidateBookingData(queryClient);
      invalidateCreditData(queryClient);

      // НЕ вызываем onLessonUpdated - это закроет модалку
    } catch (error) {
      console.error("[LessonEditModal] Error applying bulk edit:", error);
      const errorMsg =
        error?.response?.data?.message ||
        error?.message ||
        "Failed to apply bulk edit";
      showNotification(errorMsg, "error");
    }
  };

  if (!lesson) return null;

  // Использовать локальные данные занятия для отражения изменений в UI
  const currentStudents = lessonData?.current_students ?? students.length;
  const spotsLeft = (lesson.max_students || 0) - currentStudents;

  // Проверка: занятие в прошлом?
  // Используем текущие значения из формы (lessonDate, startTime), если они изменены
  // Сравниваем с текущим временем в том же формате
  let checkStartTime;
  if (lessonDate && startTime) {
    // Построить дату из текущих значений формы в локальном часовом поясе
    const [year, month, day] = lessonDate.split("-").map(Number);
    const [hours, minutes] = startTime.split(":").map(Number);
    checkStartTime = new Date(year, month - 1, day, hours, minutes);
  } else {
    // Использовать исходное время из lesson
    checkStartTime = new Date(lesson.start_time);
  }
  const isPastLesson = checkStartTime < new Date();
  const isTeacher = user?.role === ROLES.TEACHER;

  // Проверка: методист может редактировать только свои занятия
  const isOwnLesson = !isTeacher || lesson.teacher_id === user?.id;
  const canEditLesson = isOwnLesson;

  // Заморозка редактирования: прошедшее занятие для методиста ИЛИ чужое занятие для методиста
  const shouldFreezeInfoTab =
    (isPastLesson && isTeacher) || (isTeacher && !isOwnLesson);

  // Скрыть recurring UI для занятий в серии
  const showRecurringControls = !lesson?.recurring_group_id;

  /**
   * Получить заголовок модального окна с бейджами
   */
  const getModalTitle = () => {
    return (
      <div className="lesson-edit-modal-title">
        <span>Редактирование занятия</span>
        {homeworkCount > 0 && (
          <span
            className="homework-badge"
            title={`${homeworkCount} файлов домашнего задания`}
          >
            📎 ДЗ ({homeworkCount})
          </span>
        )}
      </div>
    );
  };

  return (
    <>
      <Modal
        isOpen={isOpen}
        onClose={handleClose}
        title={getModalTitle()}
        footer={
          <div className="lesson-edit-footer">
            <div className="lesson-edit-footer-left">
              {/* Кнопка удаления доступна только для своих занятий или админа */}
              {canEditLesson && (
                <Button
                  variant="danger"
                  onClick={handleDeleteLesson}
                  loading={deletingLesson}
                >
                  Удалить занятие
                </Button>
              )}
            </div>
            <div className="lesson-edit-footer-right">
              {isAutosaving && (
                <span className="autosave-indicator saving">
                  <Spinner size="small" /> Сохранение...
                </span>
              )}
              {!isAutosaving && autosaveError && (
                <span
                  className="autosave-indicator error"
                  title={autosaveError.message}
                >
                  ❌ Ошибка сохранения
                </span>
              )}
              {!isAutosaving && !autosaveError && lastSaved && (
                <span
                  className="autosave-indicator success"
                  title={new Date(lastSaved).toLocaleString("ru-RU")}
                >
                  ✓ Сохранено
                </span>
              )}
            </div>
          </div>
        }
      >
        {loading ? (
          <div className="lesson-edit-loading">
            <Spinner />
          </div>
        ) : (
          <div className="lesson-edit-content">
            {/* Tab Navigation */}
            <div className="lesson-edit-tabs">
              <button
                className={`lesson-edit-tab ${activeTab === "info" ? "active" : ""}`}
                onClick={() => setActiveTab("info")}
              >
                Основное
              </button>
              <button
                className={`lesson-edit-tab ${activeTab === "homework" ? "active" : ""}`}
                onClick={() => setActiveTab("homework")}
              >
                Домашнее задание
                {homeworkCount > 0 && (
                  <span className="tab-badge">{homeworkCount}</span>
                )}
              </button>
              <button
                className={`lesson-edit-tab ${activeTab === "broadcasts" ? "active" : ""}`}
                onClick={() => setActiveTab("broadcasts")}
              >
                Рассылки
              </button>
            </div>

            {/* Tab Content */}
            {activeTab === "info" && (
              <>
                {/* Предупреждение о чужом занятии для методиста */}
                {isTeacher && !isOwnLesson && (
                  <div className="lesson-edit-warning not-own-lesson-warning">
                    <span className="warning-icon">
                      <svg
                        width="20"
                        height="20"
                        viewBox="0 0 20 20"
                        fill="none"
                        xmlns="http://www.w3.org/2000/svg"
                      >
                        <path
                          d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
                          stroke="currentColor"
                          strokeWidth="2"
                          strokeLinecap="round"
                          strokeLinejoin="round"
                        />
                      </svg>
                    </span>
                    <span className="warning-text">
                      Это занятие другого преподавателя. Редактирование
                      недоступно.
                    </span>
                  </div>
                )}

                {/* Предупреждение о редактировании прошедшего занятия */}
                {isPastLesson && isOwnLesson && (
                  <div className="lesson-edit-warning past-lesson-warning">
                    <span className="warning-icon">
                      <svg
                        width="20"
                        height="20"
                        viewBox="0 0 20 20"
                        fill="none"
                        xmlns="http://www.w3.org/2000/svg"
                      >
                        <path
                          d="M10 6V10M10 14H10.01M19 10C19 14.9706 14.9706 19 10 19C5.02944 19 1 14.9706 1 10C1 5.02944 5.02944 1 10 1C14.9706 1 19 5.02944 19 10Z"
                          stroke="currentColor"
                          strokeWidth="2"
                          strokeLinecap="round"
                          strokeLinejoin="round"
                        />
                      </svg>
                    </span>
                    <span className="warning-text">
                      Внимание: Вы редактируете занятие в прошлом.
                    </span>
                  </div>
                )}

                {/* Информация о занятии - унифицировано с TemplateLessonForm */}
                <section className="lesson-edit-section">
                  <h4 className="section-title">Основное</h4>
                  <form data-testid="lesson-edit-form">
                    <div className="form-row">
                      <div className="form-group">
                        <label className="form-label">Преподаватель *</label>
                        <select
                          className={`form-select ${formErrors.teacher_id ? "error" : ""}`}
                          value={selectedTeacherId}
                          onChange={(e) => {
                            setSelectedTeacherId(e.target.value);
                            setFormErrors((prev) => ({
                              ...prev,
                              teacher_id: "",
                            }));
                          }}
                          disabled={
                            shouldFreezeInfoTab ||
                            teachers.length === 0 ||
                            isTeacher
                          }
                          title={
                            isTeacher
                              ? "Вы можете назначать только себя"
                              : ""
                          }
                        >
                          <option value="">Выберите преподавателя</option>
                          {teachers.map((teacher) => (
                            <option key={teacher.id} value={teacher.id}>
                              {teacher.full_name}
                            </option>
                          ))}
                        </select>
                        {formErrors.teacher_id && (
                          <span className="form-error">
                            {formErrors.teacher_id}
                          </span>
                        )}
                        {isTeacher && isOwnLesson && (
                          <small className="form-hint">
                            Вы можете назначать только себя
                          </small>
                        )}
                      </div>

                      <div className="form-group">
                        <label className="form-label">Дата занятия</label>
                        <input
                          type="date"
                          className="form-input"
                          value={lessonDate}
                          onChange={(e) => setLessonDate(e.target.value)}
                          disabled={shouldFreezeInfoTab}
                        />
                      </div>
                    </div>

                    <div className="form-row">
                      <div className="form-group">
                        <label className="form-label">Время начала *</label>
                        <input
                          type="time"
                          className={`form-input ${formErrors.start_time ? "error" : ""}`}
                          value={startTime}
                          onChange={(e) => {
                            const newStartTime = e.target.value;
                            setStartTime(newStartTime);
                            setFormErrors((prev) => ({
                              ...prev,
                              start_time: "",
                            }));
                            if (newStartTime) {
                              const [hours, minutes] = newStartTime
                                .split(":")
                                .map(Number);
                              const endHours = (hours + 2) % 24;
                              const newEndTime = `${String(endHours).padStart(2, "0")}:${String(minutes).padStart(2, "0")}`;
                              setEndTime(newEndTime);
                              setFormErrors((prev) => ({
                                ...prev,
                                end_time: "",
                              }));
                            }
                          }}
                          disabled={shouldFreezeInfoTab}
                        />
                        {formErrors.start_time && (
                          <span className="form-error">
                            {formErrors.start_time}
                          </span>
                        )}
                      </div>

                      <div className="form-group">
                        <label className="form-label">Время окончания *</label>
                        <input
                          type="time"
                          className={`form-input ${formErrors.end_time ? "error" : ""}`}
                          value={endTime}
                          onChange={(e) => {
                            setEndTime(e.target.value);
                            setFormErrors((prev) => ({
                              ...prev,
                              end_time: "",
                            }));
                          }}
                          disabled={shouldFreezeInfoTab}
                        />
                        {formErrors.end_time && (
                          <span className="form-error">
                            {formErrors.end_time}
                          </span>
                        )}
                      </div>
                    </div>

                    {showRecurringControls && (
                      <div className="form-row">
                        <div className="form-group">
                          <label className="recurring-label">
                            <input
                              type="checkbox"
                              checked={isRecurring}
                              onChange={(e) => setIsRecurring(e.target.checked)}
                              disabled={shouldFreezeInfoTab}
                            />
                            <span>Повторять еженедельно</span>
                          </label>
                        </div>

                        {isRecurring && (
                          <div
                            className="form-group"
                            style={{ gridColumn: "1 / -1" }}
                          >
                            <Button
                              variant="secondary"
                              onClick={handleCreateRecurringSeries}
                              disabled={!isRecurring || shouldFreezeInfoTab}
                              className="create-recurring-series-btn"
                            >
                              Создать серию
                            </Button>
                          </div>
                        )}
                      </div>
                    )}

                    <div className="form-row">
                      <div className="form-group">
                        <label className="form-label">
                          Максимум студентов *
                        </label>
                        <input
                          type="number"
                          className={`form-input ${formErrors.max_students ? "error" : ""}`}
                          value={maxStudents}
                          onChange={(e) => {
                            setMaxStudents(parseInt(e.target.value, 10) || 1);
                            setFormErrors((prev) => ({
                              ...prev,
                              max_students: "",
                            }));
                          }}
                          min="1"
                          max="20"
                          disabled={shouldFreezeInfoTab}
                        />
                        {formErrors.max_students && (
                          <span className="form-error">
                            {formErrors.max_students}
                          </span>
                        )}
                      </div>

                      <div className="form-group">
                        <label className="form-label">
                          Стоимость в кредитах *
                        </label>
                        <input
                          type="number"
                          className="form-input disabled-field"
                          name="creditsCost"
                          value={creditsCost}
                          min="0"
                          disabled
                          title="Стоимость нельзя изменить после создания занятия"
                        />
                        <small className="form-hint form-hint-disabled">
                          Нельзя изменить после создания
                        </small>
                      </div>
                    </div>

                    <div className="form-group">
                      <label className="form-label">Тема занятия</label>
                      <input
                        type="text"
                        className="form-input"
                        value={selectedSubject}
                        onChange={(e) => setSelectedSubject(e.target.value)}
                        placeholder="Например: Математика 10 класс"
                        maxLength="200"
                        disabled={shouldFreezeInfoTab}
                      />
                      <small className="form-hint">Максимум 200 символов</small>
                    </div>

                    <div className="form-group">
                      <label className="form-label">Ссылка</label>
                      <input
                        type="text"
                        className="form-input"
                        value={selectedLink}
                        onChange={(e) => setSelectedLink(e.target.value)}
                        placeholder="Например: https://meet.google.com/..."
                        disabled={shouldFreezeInfoTab}
                      />
                      <small className="form-hint">
                        Ссылка на видеоконференцию или материалы
                      </small>
                    </div>

                    <div className="form-group">
                      <ColorPicker
                        value={selectedColor}
                        onChange={setSelectedColor}
                        disabled={shouldFreezeInfoTab}
                      />
                    </div>
                  </form>
                </section>

                {/* Управление студентами через чекбоксы */}
                <section className="lesson-edit-section">
                  {(() => {
                    // DEBUG: синхронизация ID при построении allStudents для StudentCheckboxList
                    const allStudentsForCheckbox = [
                      ...students.map((s) => {
                        const studentId = s.student_id || s.id;
                        return {
                          id: studentId,
                          name: s.student_name,
                          full_name: s.student_name,
                          email: s.student_email,
                          credits: studentCredits[studentId] || 0,
                        };
                      }),
                      ...availableStudents.map((s) => {
                        const studentId = s.id;
                        return {
                          id: studentId,
                          name: s.full_name || s.name,
                          full_name: s.full_name || s.name,
                          email: s.email,
                          credits: studentCredits[studentId] || 0,
                        };
                      }),
                    ];

                    // DEBUG: логирование при построении allStudents
                    const studentsWithoutCredits =
                      allStudentsForCheckbox.filter(
                        (s) => studentCredits[s.id] === undefined,
                      );

                    if (studentsWithoutCredits.length > 0) {
                      console.warn(
                        "[StudentCheckboxList] WARNING: Some students have no credits in map:",
                        {
                          count: studentsWithoutCredits.length,
                          examples: studentsWithoutCredits
                            .slice(0, 3)
                            .map((s) => ({
                              id: s.id,
                              name: s.full_name,
                            })),
                        },
                      );
                    }

                    console.debug(
                      "[StudentCheckboxList] Построен список студентов:",
                      {
                        totalStudents: allStudentsForCheckbox.length,
                        enrolledCount: students.length,
                        availableCount: availableStudents.length,
                        creditsMapSize: Object.keys(studentCredits).length,
                        studentsWithoutCredits: studentsWithoutCredits.length,
                        sampleEnrolled: allStudentsForCheckbox
                          .slice(0, 2)
                          .map((s) => ({
                            id: s.id,
                            fullName: s.full_name,
                            credits: s.credits,
                            creditsInMap: studentCredits[s.id],
                          })),
                        sampleAvailable: allStudentsForCheckbox
                          .slice(-2)
                          .map((s) => ({
                            id: s.id,
                            fullName: s.full_name,
                            credits: s.credits,
                            creditsInMap: studentCredits[s.id],
                          })),
                      },
                    );

                    return (
                      <StudentCheckboxList
                        allStudents={allStudentsForCheckbox}
                        enrolledStudentIds={students.map(
                          (s) => s.student_id || s.id,
                        )}
                        onToggle={handleStudentToggle}
                        lessonId={lesson?.id}
                        disabled={shouldFreezeInfoTab}
                      />
                    );
                  })()}
                </section>
              </>
            )}

            {/* Homework Tab */}
            {activeTab === "homework" && (
              <section className="lesson-edit-section">
                <HomeworkSection
                  lessonId={lesson.id}
                  lesson={lesson}
                  onHomeworkCountChange={setHomeworkCount}
                  students={students}
                />
              </section>
            )}

            {/* Broadcasts Tab */}
            {activeTab === "broadcasts" && (
              <section className="lesson-edit-section">
                <BroadcastSection lessonId={lesson.id} lesson={lesson} />
              </section>
            )}
          </div>
        )}
      </Modal>

      {/* Confirm dialogs */}
      <ConfirmModal
        isOpen={showRemoveConfirm}
        onClose={() => {
          setShowRemoveConfirm(false);
          setConfirmRemoveStudentId(null);
        }}
        onConfirm={confirmRemoveStudent}
        title="Отмена записи студента"
        message="Вы уверены? Кредиты будут возвращены студенту."
        confirmText="Отменить запись"
        confirmVariant="danger"
        loading={removingStudentId === confirmRemoveStudentId}
      />

      <ConfirmModal
        isOpen={showDeleteConfirm}
        onClose={() => setShowDeleteConfirm(false)}
        onConfirm={confirmDeleteLesson}
        title="Удаление занятия"
        message={`Вы уверены? Это удалит занятие от ${new Date(lesson.start_time).toLocaleDateString("ru-RU")} и вернёт кредиты всем ${students.length} студентам.`}
        confirmText="Удалить"
        confirmVariant="danger"
        loading={deletingLesson}
      />

      {/* Bulk Edit Confirmation Dialog */}
      <BulkEditConfirmationDialog
        isOpen={showBulkEditConfirmation}
        onClose={() => {
          setShowBulkEditConfirmation(false);
          setBulkEditDetails(null);
        }}
        onConfirm={handleConfirmBulkEdit}
        modificationType={bulkEditDetails?.modificationType}
        modificationDetails={bulkEditDetails}
        affectedLessons={[]} // Preview not implemented yet
        estimatedCount={0} // Will be calculated by backend
        isLoading={applyToAllMutation.isPending}
        error={applyToAllMutation.error?.message}
      />
    </>
  );
};

export default LessonEditModal;
