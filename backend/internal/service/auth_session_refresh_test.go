package service

import (
	"context"
	"testing"
	"time"

	"tutoring-platform/internal/models"
	"tutoring-platform/internal/repository"
	"tutoring-platform/pkg/auth"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// isContextType - helper to match any context in mock expectations
// (since goroutines create derived contexts with WithTimeout)
func isContextType() interface{} {
	return mock.MatchedBy(func(ctx context.Context) bool {
		return ctx != nil
	})
}

// MockSessionRepository mock для SessionRepositoryInterface
type MockSessionRepository struct {
	mock.Mock
}

func (m *MockSessionRepository) Create(ctx context.Context, session *models.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockSessionRepository) GetByID(ctx context.Context, sessionID uuid.UUID) (*models.Session, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockSessionRepository) GetWithUser(ctx context.Context, sessionID uuid.UUID) (*models.SessionWithUser, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SessionWithUser), args.Error(1)
}

func (m *MockSessionRepository) Delete(ctx context.Context, sessionID uuid.UUID) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockSessionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockSessionRepository) DeleteExpired(ctx context.Context) error {
	args := m.Called(ctx, ctx)
	return args.Error(0)
}

func (m *MockSessionRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Session, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Session), args.Error(1)
}

func (m *MockSessionRepository) UpdateExpiry(ctx context.Context, sessionID uuid.UUID, expiresAt time.Time) error {
	args := m.Called(ctx, sessionID, expiresAt)
	return args.Error(0)
}

// TestRefreshSessionToken тесты для RefreshSessionToken
func TestRefreshSessionToken(t *testing.T) {
	ctx := context.Background()
	sessionID := uuid.New()
	userID := uuid.New()
	secret := "test-secret-key"
	sessionMaxAge := 7 * 24 * time.Hour // 7 дней - соответствует новому default

	sessionMgr := auth.NewSessionManager(secret)

	tests := []struct {
		name              string
		sessionExpiresAt  time.Time
		expectedRefresh   bool
		setupMocks        func(*MockSessionRepository)
		expectError       bool
		expectTokenChange bool
		description       string
		shouldCallUpdate  bool
	}{
		{
			name:              "Сессия истекает менее чем через 72 часа - должна быть продлена",
			sessionExpiresAt:  time.Now().Add(48 * time.Hour),
			expectedRefresh:   true,
			expectError:       false,
			expectTokenChange: true,
			shouldCallUpdate:  true,
			description:       "Валидная сессия требует продления - она ещё не истекла",
			setupMocks: func(m *MockSessionRepository) {
				session := &models.Session{
					ID:        sessionID,
					UserID:    userID,
					ExpiresAt: time.Now().Add(48 * time.Hour),
					CreatedAt: time.Now().Add(-5 * 24 * time.Hour),
				}
				m.On("GetByID", ctx, sessionID).Return(session, nil)
				m.On("UpdateExpiry", ctx, sessionID, mock.MatchedBy(func(t time.Time) bool {
					return t.After(time.Now().Add(6 * 24 * time.Hour))
				})).Return(nil)
			},
		},
		{
			name:              "Сессия свежая (более 72 часов до истечения) - не нужно продлевать",
			sessionExpiresAt:  time.Now().Add(96 * time.Hour),
			expectedRefresh:   false,
			expectError:       false,
			expectTokenChange: false,
			shouldCallUpdate:  false,
			description:       "Свежая сессия не требует продления",
			setupMocks: func(m *MockSessionRepository) {
				session := &models.Session{
					ID:        sessionID,
					UserID:    userID,
					ExpiresAt: time.Now().Add(96 * time.Hour),
					CreatedAt: time.Now().Add(-3 * 24 * time.Hour),
				}
				m.On("GetByID", ctx, sessionID).Return(session, nil)
			},
		},
		{
			name:              "Сессия истекает ровно через 72 часа - нужно продлевать (граница)",
			sessionExpiresAt:  time.Now().Add(72 * time.Hour),
			expectedRefresh:   true,
			expectError:       false,
			expectTokenChange: true,
			shouldCallUpdate:  true,
			description:       "На границе порога продления - должна быть продлена",
			setupMocks: func(m *MockSessionRepository) {
				session := &models.Session{
					ID:        sessionID,
					UserID:    userID,
					ExpiresAt: time.Now().Add(72 * time.Hour),
					CreatedAt: time.Now().Add(-4 * 24 * time.Hour),
				}
				m.On("GetByID", ctx, sessionID).Return(session, nil)
				m.On("UpdateExpiry", ctx, sessionID, mock.MatchedBy(func(t time.Time) bool {
					return t.After(time.Now().Add(6 * 24 * time.Hour))
				})).Return(nil)
			},
		},
		{
			name:              "Ошибка при получении сессии из БД",
			sessionExpiresAt:  time.Now().Add(6 * time.Hour),
			expectedRefresh:   false,
			expectError:       true,
			expectTokenChange: false,
			shouldCallUpdate:  false,
			description:       "Ошибка БД должна быть передана",
			setupMocks: func(m *MockSessionRepository) {
				m.On("GetByID", ctx, sessionID).Return(nil, repository.ErrSessionNotFound)
			},
		},
		{
			name:              "Истекшая сессия - не может быть продлена",
			sessionExpiresAt:  time.Now().Add(-1 * time.Minute),
			expectedRefresh:   false,
			expectError:       true,
			expectTokenChange: false,
			shouldCallUpdate:  false,
			description:       "Истекшая сессия не может быть продлена - RACE CONDITION FIX",
			setupMocks: func(m *MockSessionRepository) {
				session := &models.Session{
					ID:        sessionID,
					UserID:    userID,
					ExpiresAt: time.Now().Add(-1 * time.Minute),
					CreatedAt: time.Now().Add(-24 * time.Hour),
				}
				m.On("GetByID", ctx, sessionID).Return(session, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Подготовка
			mockSessionRepo := new(MockSessionRepository)
			tt.setupMocks(mockSessionRepo)

			authService := &AuthService{
				userRepo:      nil, // Не используется в RefreshSessionToken
				sessionRepo:   mockSessionRepo,
				sessionMgr:    sessionMgr,
				sessionMaxAge: sessionMaxAge,
			}

			// Создаем исходный токен с текущим временем истечения
			originalToken, err := sessionMgr.CreateSessionToken(sessionID, userID, tt.sessionExpiresAt)
			assert.NoError(t, err)

			// Выполнение
			newToken, err := authService.RefreshSessionToken(ctx, originalToken)

			// Проверка
			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				if tt.expectTokenChange {
					assert.NotEqual(t, originalToken, newToken, "токен должен измениться при продлении")
				} else {
					assert.Equal(t, originalToken, newToken, "токен не должен измениться если продление не требуется")
				}
			}

			// Проверяем expectations только если был ожидаемый вызов UpdateExpiry
			if tt.shouldCallUpdate || !tt.expectError {
				mockSessionRepo.AssertExpectations(t)
			}
		})
	}
}

// TestRefreshSessionTokenThreshold проверяет пороги продления
// SessionRefreshThreshold = 72 часа (3 дня)
func TestRefreshSessionTokenThreshold(t *testing.T) {
	ctx := context.Background()
	sessionID := uuid.New()
	userID := uuid.New()
	secret := "test-secret-key"
	sessionMaxAge := 7 * 24 * time.Hour // 7 дней

	sessionMgr := auth.NewSessionManager(secret)

	testCases := []struct {
		name             string
		hoursUntilExpiry float64
		shouldRefresh    bool
	}{
		{"71.99 часов до истечения (меньше порога)", 71.99, true},
		{"72 часов до истечения (на границе)", 72.0, true},
		{"72.01 часов до истечения (больше порога)", 72.01, false},
		{"24 часа до истечения", 24.0, true},
		{"120 часов до истечения (5 дней)", 120.0, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockSessionRepo := new(MockSessionRepository)

			expiresAt := time.Now().Add(time.Duration(tc.hoursUntilExpiry * float64(time.Hour)))
			session := &models.Session{
				ID:        sessionID,
				UserID:    userID,
				ExpiresAt: expiresAt,
				CreatedAt: time.Now(),
			}

			mockSessionRepo.On("GetByID", ctx, sessionID).Return(session, nil)

			if tc.shouldRefresh {
				mockSessionRepo.On("UpdateExpiry", ctx, sessionID, mock.MatchedBy(func(t time.Time) bool {
					return t.After(time.Now().Add(6 * 24 * time.Hour))
				})).Return(nil)
			}

			authService := &AuthService{
				userRepo:      nil,
				sessionRepo:   mockSessionRepo,
				sessionMgr:    sessionMgr,
				sessionMaxAge: sessionMaxAge,
			}

			originalToken, err := sessionMgr.CreateSessionToken(sessionID, userID, expiresAt)
			assert.NoError(t, err)

			newToken, err := authService.RefreshSessionToken(ctx, originalToken)
			assert.NoError(t, err)

			if tc.shouldRefresh {
				assert.NotEqual(t, originalToken, newToken, "токен должен быть продлен")
			} else {
				assert.Equal(t, originalToken, newToken, "токен не должен быть продлен")
			}

			mockSessionRepo.AssertExpectations(t)
		})
	}
}

// TestRefreshSessionTokenInvalidToken проверяет обработку невалидных токенов
func TestRefreshSessionTokenInvalidToken(t *testing.T) {
	ctx := context.Background()
	secret := "test-secret-key"
	sessionMaxAge := 7 * 24 * time.Hour // 7 дней

	sessionMgr := auth.NewSessionManager(secret)
	mockSessionRepo := new(MockSessionRepository)

	authService := &AuthService{
		userRepo:      nil,
		sessionRepo:   mockSessionRepo,
		sessionMgr:    sessionMgr,
		sessionMaxAge: sessionMaxAge,
	}

	tests := []struct {
		name        string
		token       string
		expectError bool
	}{
		{"Пустой токен", "", true},
		{"Невалидный формат", "invalid.token.format", true},
		{"Поддельный токен", "ZXJyb3I=.Zm9vYmFy", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newToken, err := authService.RefreshSessionToken(ctx, tt.token)

			assert.Error(t, err)
			assert.Empty(t, newToken)
		})
	}
}

// TestValidateSessionWithBuffer проверяет валидацию сессии с буфером времени (race condition fix)
func TestValidateSessionWithBuffer(t *testing.T) {
	ctx := context.Background()
	sessionID := uuid.New()
	userID := uuid.New()
	secret := "test-secret-key"
	sessionMaxAge := 24 * time.Hour

	sessionMgr := auth.NewSessionManager(secret)

	tests := []struct {
		name           string
		sessionExpires time.Time
		setupMocks     func(*MockSessionRepository)
		expectError    bool
		errorType      error
		description    string
	}{
		{
			name:           "Валидная сессия - больше буфера до истечения",
			sessionExpires: time.Now().Add(10 * time.Minute),
			expectError:    false,
			description:    "Сессия действительна с достаточным запасом времени",
			setupMocks: func(m *MockSessionRepository) {
				session := &models.SessionWithUser{
					Session: models.Session{
						ID:        sessionID,
						UserID:    userID,
						ExpiresAt: time.Now().Add(10 * time.Minute),
						CreatedAt: time.Now().Add(-14 * time.Hour),
					},
					UserEmail: "user@example.com",
					UserName:  "Test User",
					UserRole:  models.RoleStudent,
				}
				m.On("GetWithUser", ctx, sessionID).Return(session, nil)
			},
		},
		{
			name:           "Сессия истекла в БД - должна быть отклонена (но еще валидна по JWT)",
			sessionExpires: time.Now().Add(5 * time.Second), // Valid for JWT, but will be expired when checked in DB
			expectError:    true,
			errorType:      auth.ErrExpiredSession,
			description:    "Истекшая сессия в БД должна быть отклонена и удалена",
			setupMocks: func(m *MockSessionRepository) {
				// Session in DB has already expired
				session := &models.SessionWithUser{
					Session: models.Session{
						ID:        sessionID,
						UserID:    userID,
						ExpiresAt: time.Now().Add(-1 * time.Second), // Expired in DB, but token still valid
						CreatedAt: time.Now().Add(-24 * time.Hour),
					},
					UserEmail: "user@example.com",
					UserName:  "Test User",
					UserRole:  models.RoleStudent,
				}
				m.On("GetWithUser", ctx, sessionID).Return(session, nil)
				// Delete is called in goroutine with WithTimeout context, so match any context
				m.On("Delete", isContextType(), sessionID).Return(nil)
			},
		},
		{
			name:           "Сессия в буфере (менее 30с до истечения) - должна быть принята (не истекла)",
			sessionExpires: time.Now().Add(15 * time.Second),
			expectError:    false,
			description:    "Сессия всё ещё валидна - не истекла реально",
			setupMocks: func(m *MockSessionRepository) {
				session := &models.SessionWithUser{
					Session: models.Session{
						ID:        sessionID,
						UserID:    userID,
						ExpiresAt: time.Now().Add(15 * time.Second),
						CreatedAt: time.Now().Add(-23 * time.Hour),
					},
					UserEmail: "user@example.com",
					UserName:  "Test User",
					UserRole:  models.RoleStudent,
				}
				m.On("GetWithUser", ctx, sessionID).Return(session, nil)
			},
		},
		{
			name:           "Сессия на границе буфера (ровно 30с) - должна быть принята",
			sessionExpires: time.Now().Add(30 * time.Second),
			expectError:    false,
			description:    "Сессия всё ещё валидна - не истекла реально",
			setupMocks: func(m *MockSessionRepository) {
				session := &models.SessionWithUser{
					Session: models.Session{
						ID:        sessionID,
						UserID:    userID,
						ExpiresAt: time.Now().Add(30 * time.Second),
						CreatedAt: time.Now().Add(-23 * time.Hour),
					},
					UserEmail: "user@example.com",
					UserName:  "Test User",
					UserRole:  models.RoleStudent,
				}
				m.On("GetWithUser", ctx, sessionID).Return(session, nil)
			},
		},
		{
			name:           "Сессия с запасом (31с) - должна быть принята",
			sessionExpires: time.Now().Add(31 * time.Second),
			expectError:    false,
			description:    "Сессия валидна",
			setupMocks: func(m *MockSessionRepository) {
				session := &models.SessionWithUser{
					Session: models.Session{
						ID:        sessionID,
						UserID:    userID,
						ExpiresAt: time.Now().Add(31 * time.Second),
						CreatedAt: time.Now().Add(-23 * time.Hour),
					},
					UserEmail: "user@example.com",
					UserName:  "Test User",
					UserRole:  models.RoleStudent,
				}
				m.On("GetWithUser", ctx, sessionID).Return(session, nil)
			},
		},
		{
			name:           "Ошибка при получении сессии из БД",
			sessionExpires: time.Now().Add(10 * time.Minute),
			expectError:    true,
			description:    "Ошибка БД должна быть передана",
			setupMocks: func(m *MockSessionRepository) {
				m.On("GetWithUser", ctx, sessionID).Return(nil, repository.ErrSessionNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Подготовка
			mockSessionRepo := new(MockSessionRepository)
			tt.setupMocks(mockSessionRepo)

			authService := &AuthService{
				userRepo:      nil,
				sessionRepo:   mockSessionRepo,
				sessionMgr:    sessionMgr,
				sessionMaxAge: sessionMaxAge,
			}

			// Создаем валидный токен
			token, err := sessionMgr.CreateSessionToken(sessionID, userID, tt.sessionExpires)
			assert.NoError(t, err)

			// Выполнение
			session, err := authService.ValidateSessionWithBuffer(ctx, token)

			// Проверка
			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Nil(t, session)
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, session)
				assert.Equal(t, sessionID, session.ID)
				assert.Equal(t, userID, session.UserID)
			}

			// Wait for goroutines that delete expired sessions
			time.Sleep(100 * time.Millisecond)
			mockSessionRepo.AssertExpectations(t)
		})
	}
}

// TestValidateSessionWithBufferDBTruth проверяет что БД является источником истины
// Даже если токен указывает на истечение, проверка по БД определяет валидность
func TestValidateSessionWithBufferDBTruth(t *testing.T) {
	ctx := context.Background()
	sessionID := uuid.New()
	userID := uuid.New()
	secret := "test-secret-key"
	sessionMaxAge := 24 * time.Hour

	sessionMgr := auth.NewSessionManager(secret)
	mockSessionRepo := new(MockSessionRepository)

	// Сценарий: токен указывает на истечение через 25 секунд
	// Но в БД сессия была продлена и истекает через час
	tokenExpiresAt := time.Now().Add(25 * time.Second)
	dbExpiresAt := time.Now().Add(1 * time.Hour)

	session := &models.SessionWithUser{
		Session: models.Session{
			ID:        sessionID,
			UserID:    userID,
			ExpiresAt: dbExpiresAt, // БД говорит что сессия продлена
			CreatedAt: time.Now().Add(-23 * time.Hour),
		},
		UserEmail: "user@example.com",
		UserName:  "Test User",
		UserRole:  models.RoleStudent,
	}

	mockSessionRepo.On("GetWithUser", ctx, sessionID).Return(session, nil)

	authService := &AuthService{
		userRepo:      nil,
		sessionRepo:   mockSessionRepo,
		sessionMgr:    sessionMgr,
		sessionMaxAge: sessionMaxAge,
	}

	// Создаем токен с коротким временем истечения
	token, err := sessionMgr.CreateSessionToken(sessionID, userID, tokenExpiresAt)
	assert.NoError(t, err)

	// Валидируем - должна быть принята потому что БД говорит что сессия валидна
	result, err := authService.ValidateSessionWithBuffer(ctx, token)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, sessionID, result.ID)

	mockSessionRepo.AssertExpectations(t)
}

// TestRefreshSessionTokenRaceConditionFix проверяет исправление race condition
// между ValidateSessionWithBuffer и RefreshSessionToken
// SessionRefreshThreshold = 72 часа (3 дня)
func TestRefreshSessionTokenRaceConditionFix(t *testing.T) {
	ctx := context.Background()
	sessionID := uuid.New()
	userID := uuid.New()
	secret := "test-secret-key"
	sessionMaxAge := 7 * 24 * time.Hour // 7 дней

	sessionMgr := auth.NewSessionManager(secret)

	tests := []struct {
		name             string
		sessionExpiresAt time.Time
		description      string
		setupMocks       func(*MockSessionRepository)
		expectError      bool
		expectRefreshed  bool
	}{
		{
			name:             "Сессия близка к истечению (20с) - должна быть продлена (буфер)",
			sessionExpiresAt: time.Now().Add(20 * time.Second),
			description:      "Сессия в буфере (20с < 30с) должна быть продлена",
			setupMocks: func(m *MockSessionRepository) {
				session := &models.Session{
					ID:        sessionID,
					UserID:    userID,
					ExpiresAt: time.Now().Add(20 * time.Second),
					CreatedAt: time.Now().Add(-6 * 24 * time.Hour),
				}
				m.On("GetByID", ctx, sessionID).Return(session, nil)
				m.On("UpdateExpiry", ctx, sessionID, mock.MatchedBy(func(t time.Time) bool {
					return t.After(time.Now().Add(6 * 24 * time.Hour))
				})).Return(nil)
			},
			expectError:     false,
			expectRefreshed: true,
		},
		{
			name:             "Сессия истекла (-1м) - RefreshSessionToken отклонит",
			sessionExpiresAt: time.Now().Add(-1 * time.Minute),
			description:      "Не может продлить уже истекшую сессию",
			setupMocks: func(m *MockSessionRepository) {
				session := &models.Session{
					ID:        sessionID,
					UserID:    userID,
					ExpiresAt: time.Now().Add(-1 * time.Minute),
					CreatedAt: time.Now().Add(-7 * 24 * time.Hour),
				}
				m.On("GetByID", ctx, sessionID).Return(session, nil)
			},
			expectError:     true,
			expectRefreshed: false,
		},
		{
			name:             "Сессия валидна (48ч) - RefreshSessionToken продлит",
			sessionExpiresAt: time.Now().Add(48 * time.Hour),
			description:      "Валидная сессия требует продления (менее 72ч)",
			setupMocks: func(m *MockSessionRepository) {
				session := &models.Session{
					ID:        sessionID,
					UserID:    userID,
					ExpiresAt: time.Now().Add(48 * time.Hour),
					CreatedAt: time.Now().Add(-5 * 24 * time.Hour),
				}
				m.On("GetByID", ctx, sessionID).Return(session, nil)
				m.On("UpdateExpiry", ctx, sessionID, mock.MatchedBy(func(t time.Time) bool {
					return t.After(time.Now().Add(6 * 24 * time.Hour))
				})).Return(nil)
			},
			expectError:     false,
			expectRefreshed: true,
		},
		{
			name:             "Сессия свежая (96ч) - не требует продления",
			sessionExpiresAt: time.Now().Add(96 * time.Hour),
			description:      "Свежая сессия (>72ч) не требует продления",
			setupMocks: func(m *MockSessionRepository) {
				session := &models.Session{
					ID:        sessionID,
					UserID:    userID,
					ExpiresAt: time.Now().Add(96 * time.Hour),
					CreatedAt: time.Now().Add(-3 * 24 * time.Hour),
				}
				m.On("GetByID", ctx, sessionID).Return(session, nil)
			},
			expectError:     false,
			expectRefreshed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSessionRepo := new(MockSessionRepository)
			tt.setupMocks(mockSessionRepo)

			authService := &AuthService{
				userRepo:      nil,
				sessionRepo:   mockSessionRepo,
				sessionMgr:    sessionMgr,
				sessionMaxAge: sessionMaxAge,
			}

			token, err := sessionMgr.CreateSessionToken(sessionID, userID, tt.sessionExpiresAt)
			assert.NoError(t, err)

			// Вызываем RefreshSessionToken напрямую (как middleware это делает)
			newToken, err := authService.RefreshSessionToken(ctx, token)

			if tt.expectError {
				assert.Error(t, err, tt.description)
				assert.Empty(t, newToken)
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotEmpty(t, newToken)
				if tt.expectRefreshed {
					assert.NotEqual(t, token, newToken, "токен должен быть обновлён")
				} else {
					assert.Equal(t, token, newToken, "токен не должен быть обновлён")
				}
			}

			// Проверяем expectations только для успешных тестов или если ожидается UpdateExpiry
			if !tt.expectError || tt.expectRefreshed {
				mockSessionRepo.AssertExpectations(t)
			}
		})
	}
}
