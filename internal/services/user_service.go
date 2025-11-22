package services

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/wozhdeleniye/avito-tech-internship/internal/config"
	"github.com/wozhdeleniye/avito-tech-internship/internal/repo/models"
	postgresrepository "github.com/wozhdeleniye/avito-tech-internship/internal/repo/repositories/postgres_repository"
	redisrepository "github.com/wozhdeleniye/avito-tech-internship/internal/repo/repositories/redis_repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenInvalid       = errors.New("token invalid")
)

type AuthService struct {
	userRepo  *postgresrepository.UserRepository
	tokenRepo *redisrepository.TokenRepository
	jwtConfig config.JWTConfig
}

type JWTConfig struct {
	AccessTokenSecret  string
	RefreshTokenSecret string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

func NewAuthService(userRepo *postgresrepository.UserRepository, tokenRepo *redisrepository.TokenRepository, jwtConfig config.JWTConfig) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		jwtConfig: jwtConfig,
	}
}

func (s *AuthService) Register(ctx context.Context, req *models.CreateUserRequest) (*models.AuthResponse, error) {

	existingUser, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("user already exists")
	}

	user := &models.User{
		Email:    req.Email,
		Nickname: req.Nickname,
	}

	if err := user.SetPassword(req.Password); err != nil {
		return nil, err
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	tokens, err := s.generateTokens(user)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		User:         user,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if !user.CheckPassword(req.Password) {
		return nil, ErrInvalidCredentials
	}

	tokens, err := s.generateTokens(user)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		User:         user,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (s *AuthService) RefreshTokens(ctx context.Context, refreshToken string) (*models.TokenPair, error) {

	if blacklisted, err := s.tokenRepo.IsTokenBlacklisted(ctx, refreshToken); err != nil || blacklisted {
		return nil, ErrTokenInvalid
	}

	claims, err := s.validateToken(refreshToken, s.jwtConfig.RefreshTokenSecret)
	if err != nil {
		return nil, ErrTokenInvalid
	}

	_, err = s.tokenRepo.GetRefreshToken(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	tokens, err := s.generateTokens(user)
	if err != nil {
		return nil, err
	}

	s.tokenRepo.StoreBlacklistedToken(ctx, refreshToken, s.jwtConfig.RefreshTokenExpiry)

	return tokens, nil
}

func (s *AuthService) Logout(ctx context.Context, accessToken, refreshToken string, userID uuid.UUID) error {

	if err := s.tokenRepo.StoreBlacklistedToken(ctx, accessToken, s.jwtConfig.AccessTokenExpiry); err != nil {
		return err
	}
	if err := s.tokenRepo.StoreBlacklistedToken(ctx, refreshToken, s.jwtConfig.RefreshTokenExpiry); err != nil {
		return err
	}

	return s.tokenRepo.DeleteRefreshToken(ctx, userID)
}

func (s *AuthService) ValidateToken(token string) (*Claims, error) {
	return s.validateToken(token, s.jwtConfig.AccessTokenSecret)
}

func (s *AuthService) generateTokens(user *models.User) (*models.TokenPair, error) {

	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, err
	}

	return &models.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) generateAccessToken(user *models.User) (string, error) {
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtConfig.AccessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtConfig.AccessTokenSecret))
}

func (s *AuthService) generateRefreshToken(user *models.User) (string, error) {
	tokenID := uuid.New().String()

	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtConfig.RefreshTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	refreshToken, err := token.SignedString([]byte(s.jwtConfig.RefreshTokenSecret))
	if err != nil {
		return "", err
	}

	ctx := context.Background()
	if err := s.tokenRepo.StoreRefreshToken(ctx, user.ID, tokenID, s.jwtConfig.RefreshTokenExpiry); err != nil {
		return "", err
	}

	return refreshToken, nil
}

func (s *AuthService) validateToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrTokenInvalid
}

func (s *AuthService) SetUserActive(ctx context.Context, userId string, isActive bool) (*models.User, error) {
	user, err := s.userRepo.GetUserByCustomID(ctx, userId)
	if err != nil {
		return nil, err
	}
	user.IsActive = isActive

	if err := s.userRepo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}
