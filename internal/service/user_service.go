package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/muyue/comic-harmony-backend/internal/model"
	"github.com/muyue/comic-harmony-backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, req model.RegisterRequest) (*model.LoginResponse, error)
	Login(ctx context.Context, req model.LoginRequest) (*model.LoginResponse, error)
	GetProfile(ctx context.Context, userID int64) (*model.User, error)
}

type userService struct {
	repo       repository.UserRepository
	jwtSecret  string
	jwtExpiry  time.Duration
}

func NewUserService(repo repository.UserRepository, jwtSecret string) UserService {
	return &userService{
		repo:      repo,
		jwtSecret: jwtSecret,
		jwtExpiry: 7 * 24 * time.Hour,
	}
}

func (s *userService) Register(ctx context.Context, req model.RegisterRequest) (*model.LoginResponse, error) {
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return nil, errors.New("username, email and password are required")
	}
	if len(req.Password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	existing, _ := s.repo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	existing, _ = s.repo.GetByUsername(ctx, req.Username)
	if existing != nil {
		return nil, errors.New("username already taken")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hashed),
		AvatarURL: "",
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{Token: token, User: *user}, nil
}

func (s *userService) Login(ctx context.Context, req model.LoginRequest) (*model.LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("email and password are required")
	}

	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	if user.Status == 0 {
		return nil, errors.New("account is disabled")
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	user.Password = ""
	return &model.LoginResponse{Token: token, User: *user}, nil
}

func (s *userService) GetProfile(ctx context.Context, userID int64) (*model.User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	user.Password = ""
	return user, nil
}

func (s *userService) generateToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(s.jwtExpiry).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
