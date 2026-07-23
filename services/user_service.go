package services

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	dto "supermarket-backend/dtos"
	"supermarket-backend/entities"
	"supermarket-backend/exceptions"
	"supermarket-backend/repositories"
)

// AuthConfig is the minimal configuration surface user service needs for JWT
// issuance. Defined locally (rather than depending on the concrete config
// package) to avoid a config -> controllers -> services -> config import cycle.
type AuthConfig interface {
	GetJWTExpiration() time.Duration
	GetJWTSecret() string
}

// TokenBlacklist is the minimal surface needed to revoke tokens on logout.
// Satisfied structurally by config.RistrettoTokenBlacklist.
type TokenBlacklist interface {
	Blacklist(token string, ttl time.Duration)
}

type userServiceImpl struct {
	userRepo  repositories.UserRepository
	db        *gorm.DB
	cfg       AuthConfig
	blacklist TokenBlacklist
}

// NewUserService constructs a new UserService implementation.
func NewUserService(
	userRepo repositories.UserRepository,
	db *gorm.DB,
	cfg AuthConfig,
	blacklist TokenBlacklist,
) UserService {
	return &userServiceImpl{
		userRepo:  userRepo,
		db:        db,
		cfg:       cfg,
		blacklist: blacklist,
	}
}

func (s *userServiceImpl) RegisterUser(ctx context.Context, req *dto.RegisterRequest) (*dto.UserResponse, error) {
	// 1. Verify username uniqueness
	existing, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, exceptions.NewInternalError("Database error during username uniqueness check: " + err.Error())
	}
	if existing != nil {
		return nil, exceptions.NewConflictError("Este nombre de usuario se encuentra en uso")
	}

	// 2. Hash password with BCrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, exceptions.NewInternalError("Failed to hash password: " + err.Error())
	}

	user := &entities.User{
		FullName: req.FullName,
		Username: req.Username,
		Password: string(hashedPassword),
		Role:     req.Role,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, exceptions.NewInternalError("Failed to register user: " + err.Error())
	}

	return &dto.UserResponse{
		ID:        user.ID,
		FullName:  user.FullName,
		Username:  user.Username,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *userServiceImpl) LoginUser(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// 1. Fetch active user
	user, err := s.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, exceptions.NewInternalError("Database error during login: " + err.Error())
	}
	if user == nil {
		return nil, exceptions.NewUnauthorizedError("Credenciales invalidas. Por favor, intente nuevamente.")
	}

	// 2. Compare passwords
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, exceptions.NewUnauthorizedError("Credenciales invalidas. Por favor, intente nuevamente.")
	}

	// 3. Generate JWT
	expiration := s.cfg.GetJWTExpiration()
	expiresAt := time.Now().Add(expiration)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"user": user.Username,
		"exp":  expiresAt.Unix(),
		"iat":  time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.cfg.GetJWTSecret()))
	if err != nil {
		return nil, exceptions.NewInternalError("Failed to sign authentication token: " + err.Error())
	}

	return &dto.LoginResponse{
		Token:     tokenString,
		ExpiresAt: expiresAt,
		User: dto.UserResponse{
			ID:        user.ID,
			FullName:  user.FullName,
			Username:  user.Username,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
		},
	}, nil
}

func (s *userServiceImpl) LogoutUser(ctx context.Context, token string) error {
	if s.blacklist != nil && token != "" {
		s.blacklist.Blacklist(token, s.cfg.GetJWTExpiration())
	}
	return nil
}

func (s *userServiceImpl) ListUsers(ctx context.Context, page, size int) (*dto.PaginatedUsers, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	offset := (page - 1) * size

	users, total, err := s.userRepo.FindAll(ctx, offset, size)
	if err != nil {
		return nil, exceptions.NewInternalError("Failed to list users: " + err.Error())
	}

	items := make([]dto.UserResponse, len(users))
	for i, user := range users {
		items[i] = dto.UserResponse{
			ID:        user.ID,
			FullName:  user.FullName,
			Username:  user.Username,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
		}
	}

	return &dto.PaginatedUsers{
		Items:      items,
		TotalCount: total,
	}, nil
}

func (s *userServiceImpl) GetUserByID(ctx context.Context, id uint64, reqUserID uint64, reqRole string) (*dto.UserResponse, error) {
	// 1. Enforce ownership: only ADMIN or the user themselves can retrieve
	if reqRole != "ADMIN" && id != reqUserID {
		return nil, exceptions.NewForbiddenError("Access denied: non-admin operators are prohibited from reading or modifying another user's profile")
	}

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, exceptions.NewInternalError("Database error: " + err.Error())
	}
	if user == nil {
		return nil, exceptions.NewNotFoundError("User not found")
	}

	return &dto.UserResponse{
		ID:        user.ID,
		FullName:  user.FullName,
		Username:  user.Username,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *userServiceImpl) UpdateUser(ctx context.Context, id uint64, reqUserID uint64, reqRole string, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	// 1. Enforce ownership: only ADMIN or the user themselves can update
	if reqRole != "ADMIN" && id != reqUserID {
		return nil, exceptions.NewForbiddenError("Access denied: non-admin operators are prohibited from reading or modifying another user's profile")
	}

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, exceptions.NewInternalError("Database error during user update lookup: " + err.Error())
	}
	if user == nil {
		return nil, exceptions.NewNotFoundError("User not found")
	}

	// 2. Apply modifications
	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, exceptions.NewInternalError("Failed to hash password: " + err.Error())
		}
		user.Password = string(hashed)
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, exceptions.NewInternalError("Failed to update user profile: " + err.Error())
	}

	return &dto.UserResponse{
		ID:        user.ID,
		FullName:  user.FullName,
		Username:  user.Username,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *userServiceImpl) DeleteUser(ctx context.Context, id uint64) error {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return exceptions.NewInternalError("Database error: " + err.Error())
	}
	if user == nil {
		return exceptions.NewNotFoundError("User not found")
	}

	if err := s.userRepo.Delete(ctx, id); err != nil {
		return exceptions.NewInternalError("Failed to delete user: " + err.Error())
	}

	return nil
}