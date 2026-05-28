package services

import (
	"context"
	"testing"

	"fanfare-backend/internal/models"
	"fanfare-backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockUserRepository implements repository.UserRepository for unit testing.
type MockUserRepository struct {
	Users          map[primitive.ObjectID]*models.User
	Usernames      map[string]*models.User
	Emails         map[string]*models.User
	CreateError    error
	GetByIDError   error
	GetByUserError error
	GetByEmailErr  error
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		Users:     make(map[primitive.ObjectID]*models.User),
		Usernames: make(map[string]*models.User),
		Emails:    make(map[string]*models.User),
	}
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	if m.CreateError != nil {
		return m.CreateError
	}
	if user.ID.IsZero() {
		user.ID = primitive.NewObjectID()
	}
	m.Users[user.ID] = user
	m.Usernames[user.Username] = user
	m.Emails[user.Email] = user
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	if m.GetByIDError != nil {
		return nil, m.GetByIDError
	}
	user, exists := m.Users[id]
	if !exists {
		return nil, repository.ErrNotFound
	}
	return user, nil
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	if m.GetByUserError != nil {
		return nil, m.GetByUserError
	}
	user, exists := m.Usernames[username]
	if !exists {
		return nil, repository.ErrNotFound
	}
	return user, nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.GetByEmailErr != nil {
		return nil, m.GetByEmailErr
	}
	user, exists := m.Emails[email]
	if !exists {
		return nil, repository.ErrNotFound
	}
	return user, nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	m.Users[user.ID] = user
	m.Usernames[user.Username] = user
	m.Emails[user.Email] = user
	return nil
}

func (m *MockUserRepository) FollowUniverse(ctx context.Context, userID, universeID primitive.ObjectID) error {
	return nil
}

func (m *MockUserRepository) GetLeaderboard(ctx context.Context, limit int) ([]models.LeaderboardEntry, error) {
	return nil, nil
}

func (m *MockUserRepository) GetFandomLeaderboard(ctx context.Context, universeID primitive.ObjectID, limit int) ([]models.LeaderboardEntry, error) {
	return nil, nil
}

func TestRegister(t *testing.T) {
	repo := NewMockUserRepository()
	service := NewAuthService(repo, "secretkey")

	req := models.RegisterRequest{
		Username: "tester",
		Email:    "test@example.com",
		Password: "password123",
	}

	ctx := context.Background()

	// 1. Success
	user, err := service.Register(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if user.Username != "tester" || user.Email != "test@example.com" {
		t.Errorf("Returned user details incorrect")
	}

	// 2. Duplicate Username
	_, err = service.Register(ctx, req)
	if err == nil || err.Error() != "username is already taken" {
		t.Errorf("Expected duplicate username error, got %v", err)
	}

	// 3. Duplicate Email
	req2 := models.RegisterRequest{
		Username: "newtester",
		Email:    "test@example.com",
		Password: "password321",
	}
	_, err = service.Register(ctx, req2)
	if err == nil || err.Error() != "email is already registered" {
		t.Errorf("Expected duplicate email error, got %v", err)
	}
}

func TestLoginAndJWT(t *testing.T) {
	repo := NewMockUserRepository()
	service := NewAuthService(repo, "yoursecretkeyhere")

	req := models.RegisterRequest{
		Username: "user1",
		Email:    "user1@example.com",
		Password: "securepassword",
	}

	ctx := context.Background()
	_, _ = service.Register(ctx, req)

	// 1. Invalid Login Username
	_, err := service.Login(ctx, models.LoginRequest{Username: "unknown", Password: "securepassword"})
	if err == nil || err.Error() != "invalid username or password" {
		t.Errorf("Expected invalid credentials error, got %v", err)
	}

	// 2. Invalid Login Password
	_, err = service.Login(ctx, models.LoginRequest{Username: "user1", Password: "wrongpassword"})
	if err == nil || err.Error() != "invalid username or password" {
		t.Errorf("Expected invalid credentials error, got %v", err)
	}

	// 3. Successful Login
	resp, err := service.Login(ctx, models.LoginRequest{Username: "user1", Password: "securepassword"})
	if err != nil {
		t.Fatalf("Expected successful login, got %v", err)
	}
	if resp.Token == "" {
		t.Error("Expected signed JWT token inside auth response")
	}

	// 4. Token Validation
	claims, err := service.ValidateToken(resp.Token)
	if err != nil {
		t.Fatalf("Expected valid token parse, got %v", err)
	}
	if claims.Username != "user1" {
		t.Errorf("Expected claims username to be 'user1', got %s", claims.Username)
	}

	// 5. Insecure/Tampered Token Check
	_, err = service.ValidateToken("invalid.token.string")
	if err == nil {
		t.Error("Expected validation error for arbitrary token string")
	}
}
