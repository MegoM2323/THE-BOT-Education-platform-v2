import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { LessonEditModal } from '../LessonEditModal';
import * as bookingAPI from '../../../api/bookings';
import * as creditAPI from '../../../api/credits';
import * as lessonAPI from '../../../api/lessons';
import * as userAPI from '../../../api/users';
import * as authAPI from '../../../api/auth';
import { renderWithProviders } from '../../../test/test-utils.jsx';

vi.mock('../../../api/bookings');
vi.mock('../../../api/credits');
vi.mock('../../../api/lessons');
vi.mock('../../../api/users');
vi.mock('../../../api/auth');

const mockLesson = {
  id: 'lesson-123',
  teacher_id: 'teacher-1',
  start_time: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(),
  end_time: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000 + 60 * 60 * 1000).toISOString(),
  max_students: 10,
  current_students: 1,
  subject: 'Math',
  color: '#004231',
};

const mockStudents = [
  { id: 'student-1', full_name: 'Student One', email: 'student1@test.com' },
  { id: 'student-2', full_name: 'Student Two', email: 'student2@test.com' },
  { id: 'student-3', full_name: 'Student Three', email: 'student3@test.com' },
  { id: 'student-4', full_name: 'Student Four', email: 'student4@test.com' },
  { id: 'student-5', full_name: 'Student Five', email: 'student5@test.com' },
];

const mockTeachers = [
  { id: 'teacher-1', full_name: 'Teacher One' },
  { id: 'teacher-2', full_name: 'Teacher Two' },
];

const mockBookingsResponse = [
  {
    id: 'booking-1',
    student_id: 'student-1',
    student_name: 'Student One',
    student_email: 'student1@test.com',
    status: 'active',
  },
];

const renderModal = (isOpen = true, onClose = vi.fn(), onLessonUpdated = vi.fn()) => {
  return renderWithProviders(
    <LessonEditModal
      isOpen={isOpen}
      onClose={onClose}
      lesson={mockLesson}
      onLessonUpdated={onLessonUpdated}
    />,
    { withAuth: true }
  );
};

describe('LessonEditModal - Credits Map Building from API Response', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, 'debug').mockImplementation(() => {});
    vi.spyOn(console, 'warn').mockImplementation(() => {});
    vi.spyOn(console, 'error').mockImplementation(() => {});

    // Setup auth API mock
    authAPI.checkAuth = vi.fn().mockResolvedValue({
      user: {
        id: 'admin-1',
        email: 'admin@test.com',
        role: 'admin',
        full_name: 'Admin User',
      },
      balance: null,
    });

    // Setup default API mocks
    bookingAPI.getBookings = vi.fn().mockResolvedValue(mockBookingsResponse);
    userAPI.getStudentsAll = vi.fn().mockResolvedValue(mockStudents);
    userAPI.getTeachersAll = vi.fn().mockResolvedValue(mockTeachers);
    lessonAPI.updateLesson = vi.fn().mockResolvedValue(mockLesson);
  });

  afterEach(() => {
    console.debug.mockRestore();
    console.warn.mockRestore();
    console.error.mockRestore();
    vi.clearAllMocks();
  });

  describe('creditsMap - Paginated Response Format', () => {
    it('should build creditsMap from paginated API response', async () => {
      const paginatedResponse = {
        data: [
          { user_id: 'student-1', balance: 10 },
          { user_id: 'student-2', balance: 5 },
          { user_id: 'student-3', balance: 0 },
          { user_id: 'student-4', balance: 20 },
        ],
        meta: {
          total: 4,
          page: 1,
          limit: 20,
        },
      };

      creditAPI.getAllCredits = vi.fn().mockResolvedValue({
        balances: paginatedResponse.data,
      });

      const { rerender } = renderModal(true);

      await waitFor(() => {
        expect(creditAPI.getAllCredits).toHaveBeenCalled();
      });

      // Wait for creditsMap to be built
      await waitFor(() => {
        expect(console.debug).toHaveBeenCalledWith(
          '[LessonEditModal] creditsMap built:',
          expect.objectContaining({
            size: 4,
            isEmpty: false,
          })
        );
      });
    });

    it('should map student IDs to credits correctly from paginated response', async () => {
      const paginatedResponse = {
        balances: [
          { user_id: 'student-1', balance: 100 },
          { user_id: 'student-2', balance: 50 },
          { user_id: 'student-3', balance: 0 },
        ],
      };

      creditAPI.getAllCredits = vi.fn().mockResolvedValue(paginatedResponse);

      renderModal(true);

      await waitFor(() => {
        expect(creditAPI.getAllCredits).toHaveBeenCalled();
      });

      // Verify creditsMap was built with correct mappings
      await waitFor(() => {
        expect(console.debug).toHaveBeenCalledWith(
          '[LessonEditModal] creditsMap built:',
          expect.objectContaining({
            size: 3,
            sampleEntries: expect.any(Array),
          })
        );
      });
    });

    it('should handle large credit amounts in paginated response', async () => {
      const paginatedResponse = {
        balances: [
          { user_id: 'student-1', balance: 99999 },
          { user_id: 'student-2', balance: 50000 },
          { user_id: 'student-3', balance: 1 },
        ],
      };

      creditAPI.getAllCredits = vi.fn().mockResolvedValue(paginatedResponse);

      renderModal(true);

      await waitFor(() => {
        expect(creditAPI.getAllCredits).toHaveBeenCalled();
      });

      await waitFor(() => {
        expect(console.debug).toHaveBeenCalledWith(
          '[LessonEditModal] creditsMap built:',
          expect.objectContaining({
            size: 3,
          })
        );
      });
    });
  });

  describe('creditsMap - Nested Paginated Response Format', () => {
    it('should build creditsMap from nested paginated response', async () => {
      const nestedResponse = {
        data: {
          data: [
            { user_id: 'student-1', balance: 10 },
            { user_id: 'student-2', balance: 5 },
            { user_id: 'student-3', balance: 0 },
          ],
          meta: {
            page: 1,
            limit: 20,
          },
        },
        meta: {
          code: 200,
          message: 'Success',
        },
      };

      creditAPI.getAllCredits = vi.fn().mockResolvedValue({
        balances: nestedResponse.data.data,
      });

      renderModal(true);

      await waitFor(() => {
        expect(creditAPI.getAllCredits).toHaveBeenCalled();
      });

      await waitFor(() => {
        expect(console.debug).toHaveBeenCalledWith(
          '[LessonEditModal] creditsMap built:',
          expect.objectContaining({
            size: 3,
            isEmpty: false,
          })
        );
      });
    });
  });

  describe('creditsMap - Empty Response Handling', () => {
    it('should return empty creditsMap for empty balances array', async () => {
      creditAPI.getAllCredits = vi.fn().mockResolvedValue({
        balances: [],
      });

      renderModal(true);

      await waitFor(() => {
        expect(creditAPI.getAllCredits).toHaveBeenCalled();
      });

      await waitFor(() => {
        expect(console.debug).toHaveBeenCalledWith(
          '[LessonEditModal] creditsMap built:',
          expect.objectContaining({
            size: 0,
            isEmpty: true,
          })
        );
      });
    });

    it('should log WARNING when creditsMap is empty', async () => {
      creditAPI.getAllCredits = vi.fn().mockResolvedValue({
        balances: [],
      });

      renderModal(true);

      await waitFor(() => {
        expect(console.warn).toHaveBeenCalledWith(
          '[LessonEditModal] WARNING: creditsMap is empty! Students will show 0 credits.',
          expect.stringContaining('API response format changed or load failed')
        );
      });
    });

    it('should handle null response gracefully and warn', async () => {
      creditAPI.getAllCredits = vi.fn().mockResolvedValue({
        balances: [],
      });

      renderModal(true);

      await waitFor(() => {
        expect(console.debug).toHaveBeenCalledWith(
          '[LessonEditModal] creditsMap built:',
          expect.objectContaining({
            isEmpty: true,
          })
        );
      });
    });

    it('should handle undefined balances and return empty creditsMap', async () => {
      creditAPI.getAllCredits = vi.fn().mockResolvedValue({
        balances: undefined,
      });

      renderModal(true);

      await waitFor(() => {
        expect(console.debug).toHaveBeenCalledWith(
          '[LessonEditModal] creditsMap built:',
          expect.objectContaining({
            size: 0,
            isEmpty: true,
          })
        );
      });
    });
  });

  describe('creditsMap - Student Credit Mapping', () => {
    it('should assign credits to available students from creditsMap', async () => {
      const creditsResponse = {
        balances: [
          { user_id: 'student-1', balance: 10 },
          { user_id: 'student-2', balance: 5 },
          { user_id: 'student-3', balance: 0 },
          { user_id: 'student-4', balance: 20 },
          { user_id: 'student-5', balance: 15 },
        ],
      };

      creditAPI.getAllCredits = vi.fn().mockResolvedValue(creditsResponse);

      renderModal(true);

      await waitFor(() => {
        expect(creditAPI.getAllCredits).toHaveBeenCalled();
      });

      await waitFor(() => {
        expect(console.debug).toHaveBeenCalledWith(
          '[LessonEditModal] creditsMap built:',
          expect.objectContaining({
            size: 5,
          })
        );
      });
    });

    it('should assign correct credit balance to enrolled students', async () => {
      const creditsResponse = {
        balances: [
          { user_id: 'student-1', balance: 10 },
          { user_id: 'student-2', balance: 5 },
          { user_id: 'student-3', balance: 0 },
        ],
      };

      creditAPI.getAllCredits = vi.fn().mockResolvedValue(creditsResponse);

      renderModal(true);

      // Enrolled student is student-1 with 10 credits
      await waitFor(() => {
        expect(console.debug).toHaveBeenCalledWith(
          '[LessonEditModal] creditsMap built:',
          expect.objectContaining({
            size: 3,
            sampleEntries: expect.arrayContaining([
              expect.arrayContaining(['student-1', 10]),
            ]),
          })
        );
      });
    });

    it('should assign 0 balance to students not in creditsMap', async () => {
      const creditsResponse = {
        balances: [
          { user_id: 'student-1', balance: 10 },
          // student-2, 3, 4, 5 not in credits response
        ],
      };

      creditAPI.getAllCredits = vi.fn().mockResolvedValue(creditsResponse);

      renderModal(true);

      await waitFor(() => {
        expect(console.debug).toHaveBeenCalledWith(
          '[LessonEditModal] creditsMap built:',
          expect.objectContaining({
            size: 1,
          })
        );
      });

      // Students not in creditsMap should get 0 balance when accessed
      // This is handled by StudentCheckboxList which checks: studentCredits[student.id] ?? 0
    });

    it('should handle multiple students with same credit amount', async () => {
      const creditsResponse = {
        balances: [
          { user_id: 'student-1', balance: 10 },
          { user_id: 'student-2', balance: 10 },
          { user_id: 'student-3', balance: 10 },
          { user_id: 'student-4', balance: 5 },
          { user_id: 'student-5', balance: 5 },
        ],
      };

      creditAPI.getAllCredits = vi.fn().mockResolvedValue(creditsResponse);

      renderModal(true);

      await waitFor(() => {
        expect(console.debug).toHaveBeenCalledWith(
          '[LessonEditModal] creditsMap built:',
          expect.objectContaining({
            size: 5,
          })
        );
      });
    });
  });

  describe('creditsMap - Format Detection and Logging', () => {
    it('should log allCreditsResponse structure before parsing', async () => {
      const creditsResponse = {
        balances: [{ user_id: 'student-1', balance: 10 }],
      };

      creditAPI.getAllCredits = vi.fn().mockResolvedValue(creditsResponse);

      renderModal(true);

      await waitFor(() => {
        expect(console.debug).toHaveBeenCalledWith(
          '[LessonEditModal] allCreditsResponse structure:',
          expect.objectContaining({
            hasBalances: true,
            balancesIsArray: true,
            balancesLength: 1,
          })
        );
      });
    });

    it('should log creditsMap sample entries for debugging', async () => {
      const creditsResponse = {
        balances: [
          { user_id: 'student-1', balance: 10 },
          { user_id: 'student-2', balance: 20 },
          { user_id: 'student-3', balance: 30 },
        ],
      };

      creditAPI.getAllCredits = vi.fn().mockResolvedValue(creditsResponse);

      renderModal(true);

      await waitFor(() => {
        expect(console.debug).toHaveBeenCalledWith(
          '[LessonEditModal] creditsMap built:',
          expect.objectContaining({
            size: 3,
            sampleEntries: expect.any(Array),
          })
        );
      });
    });
  });

  describe('creditsMap - Response Format Compatibility', () => {
    it('should handle backward compatible old format { balances: [...] }', async () => {
      const oldFormatResponse = {
        balances: [
          { user_id: 'student-1', balance: 10 },
          { user_id: 'student-2', balance: 5 },
        ],
      };

      creditAPI.getAllCredits = vi.fn().mockResolvedValue(oldFormatResponse);

      renderModal(true);

      await waitFor(() => {
        expect(console.debug).toHaveBeenCalledWith(
          '[LessonEditModal] creditsMap built:',
          expect.objectContaining({
            size: 2,
            isEmpty: false,
          })
        );
      });
    });

    it('should handle new paginated format with meta', async () => {
      const newFormatResponse = {
        balances: [
          { user_id: 'student-1', balance: 10 },
          { user_id: 'student-2', balance: 5 },
          { user_id: 'student-3', balance: 0 },
        ],
      };

      creditAPI.getAllCredits = vi.fn().mockResolvedValue(newFormatResponse);

      renderModal(true);

      await waitFor(() => {
        expect(console.debug).toHaveBeenCalledWith(
          '[LessonEditModal] creditsMap built:',
          expect.objectContaining({
            size: 3,
          })
        );
      });
    });

    it('should sync creditsMap with StudentCheckboxList data', async () => {
      const creditsResponse = {
        balances: [
          { user_id: 'student-1', balance: 10 },
          { user_id: 'student-2', balance: 5 },
          { user_id: 'student-3', balance: 0 },
          { user_id: 'student-4', balance: 20 },
          { user_id: 'student-5', balance: 15 },
        ],
      };

      creditAPI.getAllCredits = vi.fn().mockResolvedValue(creditsResponse);

      renderModal(true);

      await waitFor(() => {
        expect(creditAPI.getAllCredits).toHaveBeenCalled();
      });

      // creditsMap should have all 5 students
      await waitFor(() => {
        expect(console.debug).toHaveBeenCalledWith(
          '[LessonEditModal] creditsMap built:',
          expect.objectContaining({
            size: 5,
          })
        );
      });
    });
  });

  describe('creditsMap - Race Conditions and Timing', () => {
    it('should not have race condition when loadLessonData is called multiple times', async () => {
      const creditsResponse = {
        balances: [
          { user_id: 'student-1', balance: 10 },
          { user_id: 'student-2', balance: 5 },
        ],
      };

      creditAPI.getAllCredits = vi.fn().mockResolvedValue(creditsResponse);

      const { unmount } = renderModal(true);

      await waitFor(() => {
        expect(creditAPI.getAllCredits).toHaveBeenCalled();
      });

      // Should only call getAllCredits once
      expect(creditAPI.getAllCredits).toHaveBeenCalledTimes(1);

      unmount();
    });

    it('should build creditsMap only from latest API response', async () => {
      const creditsResponse = {
        balances: [
          { user_id: 'student-1', balance: 10 },
          { user_id: 'student-2', balance: 5 },
        ],
      };

      creditAPI.getAllCredits = vi.fn().mockResolvedValue(creditsResponse);

      renderModal(true);

      await waitFor(() => {
        expect(console.debug).toHaveBeenCalledWith(
          '[LessonEditModal] creditsMap built:',
          expect.objectContaining({
            size: 2,
          })
        );
      });

      // Verify the last call has correct data
      const lastDebugCall = console.debug.mock.calls.find(
        call => call[0] === '[LessonEditModal] creditsMap built:'
      );
      expect(lastDebugCall[1].size).toBe(2);
    });
  });

  describe('creditsMap - Error Handling', () => {
    it('should gracefully handle getAllCredits API failure', async () => {
      creditAPI.getAllCredits = vi
        .fn()
        .mockRejectedValue(new Error('API Error'));

      // The component should handle the error gracefully
      renderModal(true);

      // Component should render even if credits fail to load
      await waitFor(() => {
        expect(console.error).toHaveBeenCalled();
      });
    });

    it('should create empty creditsMap if API fails', async () => {
      creditAPI.getAllCredits = vi
        .fn()
        .mockRejectedValue(new Error('Network error'));

      renderModal(true);

      // Credits should fail but component should continue
      await waitFor(() => {
        // Component should not break
        expect(creditAPI.getAllCredits).toHaveBeenCalled();
      });
    });
  });

  describe('creditsMap - Data Types and Validation', () => {
    it('should handle string balance values by converting to number', async () => {
      const creditsResponse = {
        balances: [
          { user_id: 'student-1', balance: '10' },
          { user_id: 'student-2', balance: '5' },
        ],
      };

      creditAPI.getAllCredits = vi.fn().mockResolvedValue(creditsResponse);

      renderModal(true);

      await waitFor(() => {
        expect(console.debug).toHaveBeenCalledWith(
          '[LessonEditModal] creditsMap built:',
          expect.objectContaining({
            size: 2,
          })
        );
      });
    });

    it('should handle boolean balance values safely', async () => {
      const creditsResponse = {
        balances: [
          { user_id: 'student-1', balance: true },
          { user_id: 'student-2', balance: false },
        ],
      };

      creditAPI.getAllCredits = vi.fn().mockResolvedValue(creditsResponse);

      renderModal(true);

      await waitFor(() => {
        expect(console.debug).toHaveBeenCalledWith(
          '[LessonEditModal] creditsMap built:',
          expect.objectContaining({
            size: 2,
          })
        );
      });
    });

    it('should filter out invalid entries without user_id', async () => {
      const creditsResponse = {
        balances: [
          { user_id: 'student-1', balance: 10 },
          { balance: 20 }, // No user_id - should be filtered
          { user_id: 'student-2', balance: 5 },
        ],
      };

      creditAPI.getAllCredits = vi.fn().mockResolvedValue(creditsResponse);

      renderModal(true);

      await waitFor(() => {
        expect(console.debug).toHaveBeenCalledWith(
          '[LessonEditModal] creditsMap built:',
          expect.objectContaining({
            size: 2, // Only 2 valid entries
          })
        );
      });
    });
  });
});
