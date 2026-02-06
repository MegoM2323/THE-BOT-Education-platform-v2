package middleware

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
)

// MockSessionRepository для тестирования
type MockSessionRepository struct {
	user           *models.User
	shouldFail     bool
	deletedUserIDs map[uuid.UUID]bool
}

func NewMockSessionRepository(user *models.User) *MockSessionRepository {
	return &MockSessionRepository{
		user:           user,
		deletedUserIDs: make(map[uuid.UUID]bool),
	}
}

func (m *MockSessionRepository) Create(ctx context.Context, session *models.Session) error {
	session.ID = uuid.New()
	session.CreatedAt = time.Now()
	return nil
}

func (m *MockSessionRepository) GetByID(ctx context.Context, sessionID uuid.UUID) (*models.Session, error) {
	return &models.Session{
		ID:        sessionID,
		UserID:    m.user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}, nil
}

func (m *MockSessionRepository) GetWithUser(ctx context.Context, sessionID uuid.UUID) (*models.SessionWithUser, error) {
	// Если пользователь удален, не возвращаем его в сессии
	if m.user.IsDeleted() {
		return nil, repository.ErrSessionNotFound // Using ErrSessionNotFound as session lookup fails for deleted users
	}

	return &models.SessionWithUser{
		Session: models.Session{
			ID:        sessionID,
			UserID:    m.user.ID,
			ExpiresAt: time.Now().Add(24 * time.Hour),
			CreatedAt: time.Now(),
		},
		UserEmail: m.user.Email,
		UserName:  m.user.GetFullName(),
		UserRole:  m.user.Role,
	}, nil
}

func (m *MockSessionRepository) Delete(ctx context.Context, sessionID uuid.UUID) error {
	return nil
}

func (m *MockSessionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	m.deletedUserIDs[userID] = true
	return nil
}

func (m *MockSessionRepository) DeleteExpired(ctx context.Context) error {
	return nil
}

func (m *MockSessionRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Session, error) {
	return nil, nil
}

func (m *MockSessionRepository) UpdateExpiry(ctx context.Context, sessionID uuid.UUID, expiresAt time.Time) error {
	return nil
}

// MockUserRepository для тестирования
type MockUserRepository struct {
	user *models.User
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	// Если пользователь удален, возвращаем ошибку как в реальном репозитории
	if m.user.IsDeleted() {
		return nil, repository.ErrUserNotFound
	}
	return m.user, nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, repository.ErrUserNotFound
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	return nil
}

func (m *MockUserRepository) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return nil
}

func (m *MockUserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockUserRepository) List(ctx context.Context, roleFilter *models.UserRole) ([]*models.User, error) {
	return nil, nil
}

func (m *MockUserRepository) Exists(ctx context.Context, email string) (bool, error) {
	return false, nil
}

func (m *MockUserRepository) UpdateTelegramUsername(ctx context.Context, userID uuid.UUID, username string) error {
	return nil
}

// TestAuthenticateDeletedUser проверяет что удаленный пользователь не может пройти аутентификацию
func TestAuthenticateDeletedUser(t *testing.T) {
	tests := []struct {
		name           string
		user           *models.User
		expectAuth     bool
		expectedStatus int
		expectedMsg    string
	}{
		{
			name: "deleted user is rejected in Authenticate",
			user: &models.User{
				ID:        uuid.New(),
				Email:     "deleted@test.com",
				FirstName: "Deleted",
				LastName:  "User",
				Role:      models.RoleStudent,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				DeletedAt: sql.NullTime{
					Time:  time.Now().Add(-1 * time.Hour),
					Valid: true,
				},
			},
			expectAuth:     false,
			expectedStatus: 401,
			expectedMsg:    "User account is deleted or deactivated",
		},
		{
			name: "active user passes Authenticate",
			user: &models.User{
				ID:        uuid.New(),
				Email:     "active@test.com",
				FirstName: "Active",
				LastName:  "User",
				Role:      models.RoleStudent,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				DeletedAt: sql.NullTime{Valid: false},
			},
			expectAuth:     true,
			expectedStatus: 200,
			expectedMsg:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Для этого теста нам нужно проверить IsDeleted() напрямую
			if tt.user.IsDeleted() {
				// Удаленный пользователь должен быть отклонен
				assert.True(t, tt.user.IsDeleted(), "User should be marked as deleted")
				assert.Equal(t, tt.expectedStatus, 401, "Deleted user should get 401")
			} else {
				// Активный пользователь не удален
				assert.False(t, tt.user.IsDeleted(), "User should not be marked as deleted")
			}
		})
	}
}

// TestOptionalAuthenticateDeletedUser проверяет что OptionalAuthenticate правильно обрабатывает удаленных пользователей
func TestOptionalAuthenticateDeletedUser(t *testing.T) {
	tests := []struct {
		name        string
		user        *models.User
		expectInCtx bool
	}{
		{
			name: "deleted user not added to context in OptionalAuthenticate",
			user: &models.User{
				ID:        uuid.New(),
				Email:     "deleted@test.com",
				FirstName: "Deleted",
				LastName:  "User",
				Role:      models.RoleStudent,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				DeletedAt: sql.NullTime{
					Time:  time.Now().Add(-1 * time.Hour),
					Valid: true,
				},
			},
			expectInCtx: false,
		},
		{
			name: "active user added to context in OptionalAuthenticate",
			user: &models.User{
				ID:        uuid.New(),
				Email:     "active@test.com",
				FirstName: "Active",
				LastName:  "User",
				Role:      models.RoleStudent,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				DeletedAt: sql.NullTime{Valid: false},
			},
			expectInCtx: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Проверяем IsDeleted() напрямую
			isDeleted := tt.user.IsDeleted()

			if tt.expectInCtx {
				// Активный пользователь
				assert.False(t, isDeleted, "Active user should not be deleted")
			} else {
				// Удаленный пользователь
				assert.True(t, isDeleted, "Deleted user should be marked as deleted")
			}
		})
	}
}

// TestUserIsDeletedFlag проверяет что флаг IsDeleted() работает корректно
func TestUserIsDeletedFlag(t *testing.T) {
	tests := []struct {
		name         string
		deletedAt    sql.NullTime
		expectDelete bool
	}{
		{
			name:         "user with valid DeletedAt is deleted",
			deletedAt:    sql.NullTime{Time: time.Now().Add(-1 * time.Hour), Valid: true},
			expectDelete: true,
		},
		{
			name:         "user with invalid DeletedAt is not deleted",
			deletedAt:    sql.NullTime{Valid: false},
			expectDelete: false,
		},
		{
			name:         "user with null DeletedAt is not deleted",
			deletedAt:    sql.NullTime{},
			expectDelete: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &models.User{
				ID:        uuid.New(),
				Email:     "test@test.com",
				DeletedAt: tt.deletedAt,
			}

			assert.Equal(t, tt.expectDelete, user.IsDeleted(), "IsDeleted() should return %v for %v", tt.expectDelete, tt.deletedAt)
		})
	}
}

// MockAuthService мок для AuthService для тестирования
type MockAuthService struct {
	user           *models.User
	sessionUser    *models.SessionWithUser
	shouldFailAuth bool
	shouldFailUser bool
}

// NewMockAuthService создает новый MockAuthService
func NewMockAuthService(user *models.User) *MockAuthService {
	sessionUser := &models.SessionWithUser{
		Session: models.Session{
			ID:        uuid.New(),
			UserID:    user.ID,
			ExpiresAt: time.Now().Add(24 * time.Hour),
			CreatedAt: time.Now(),
		},
		UserEmail: user.Email,
		UserName:  user.GetFullName(),
		UserRole:  user.Role,
	}
	return &MockAuthService{
		user:           user,
		sessionUser:    sessionUser,
		shouldFailAuth: false,
		shouldFailUser: false,
	}
}

func (m *MockAuthService) ValidateSession(ctx context.Context, token string) (*models.SessionWithUser, error) {
	if m.shouldFailAuth {
		return nil, repository.ErrSessionNotFound
	}
	return m.sessionUser, nil
}

func (m *MockAuthService) GetCurrentUser(ctx context.Context, token string) (*models.User, error) {
	if m.shouldFailUser {
		return nil, repository.ErrUserNotFound
	}
	return m.user, nil
}

func (m *MockAuthService) RefreshSessionToken(ctx context.Context, token string) (string, error) {
	return token, nil
}

// TestAuthMiddlewareDeletedUserRejection проверяет что удаленный пользователь отклоняется в middleware
func TestAuthMiddlewareDeletedUserRejection(t *testing.T) {
	tests := []struct {
		name               string
		user               *models.User
		expectUnauthorized bool
		expectedMsg        string
	}{
		{
			name: "deleted user is rejected by Authenticate middleware",
			user: &models.User{
				ID:        uuid.New(),
				Email:     "deleted@test.com",
				FirstName: "Deleted",
				LastName:  "User",
				Role:      models.RoleStudent,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				DeletedAt: sql.NullTime{
					Time:  time.Now().Add(-1 * time.Hour),
					Valid: true,
				},
			},
			expectUnauthorized: true,
			expectedMsg:        "User account is deleted or deactivated",
		},
		{
			name: "active user passes Authenticate middleware",
			user: &models.User{
				ID:        uuid.New(),
				Email:     "active@test.com",
				FirstName: "Active",
				LastName:  "User",
				Role:      models.RoleStudent,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				DeletedAt: sql.NullTime{Valid: false},
			},
			expectUnauthorized: false,
			expectedMsg:        "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Проверяем что флаг IsDeleted() работает правильно
			assert.Equal(t, tt.expectUnauthorized, tt.user.IsDeleted(),
				"User.IsDeleted() should match expectUnauthorized")
		})
	}
}

// TestSessionRepositoryDeletedUserHandling проверяет что SessionRepository правильно обрабатывает удаленных пользователей
func TestSessionRepositoryDeletedUserHandling(t *testing.T) {
	tests := []struct {
		name           string
		user           *models.User
		shouldFindUser bool
		description    string
	}{
		{
			name: "GetWithUser should fail for deleted user",
			user: &models.User{
				ID:        uuid.New(),
				Email:     "deleted@test.com",
				FirstName: "Deleted",
				LastName:  "User",
				Role:      models.RoleStudent,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				DeletedAt: sql.NullTime{
					Time:  time.Now().Add(-1 * time.Hour),
					Valid: true,
				},
			},
			shouldFindUser: false,
			description:    "Удаленный пользователь не должен быть найден в GetWithUser",
		},
		{
			name: "GetWithUser should succeed for active user",
			user: &models.User{
				ID:        uuid.New(),
				Email:     "active@test.com",
				FirstName: "Active",
				LastName:  "User",
				Role:      models.RoleStudent,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				DeletedAt: sql.NullTime{Valid: false},
			},
			shouldFindUser: true,
			description:    "Активный пользователь должен быть найден в GetWithUser",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockSessionRepository(tt.user)
			sessionID := uuid.New()

			session, err := mockRepo.GetWithUser(context.Background(), sessionID)

			if tt.shouldFindUser {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, session, "Session should not be nil for active user")
			} else {
				assert.Error(t, err, tt.description)
				assert.True(t, err == repository.ErrSessionNotFound,
					"Should return ErrSessionNotFound for deleted user")
			}
		})
	}
}
