package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/jackc/pgx/v5"

	"github.com/clubepay/backend/internal/analytics"
	"github.com/clubepay/backend/internal/config"
	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/email"
	"github.com/clubepay/backend/internal/repository"
)

// AuthService handles authentication and profile business logic.
type AuthService struct {
	Queries   *repository.Queries
	Config    *config.Config
	Email     email.Sender
	Analytics analytics.EventTracker // Optional analytics tracker
}

// NewAuthService creates a new AuthService.
func NewAuthService(q *repository.Queries, cfg *config.Config, e email.Sender) *AuthService {
	return &AuthService{Queries: q, Config: cfg, Email: e, Analytics: nil}
}

// NewAuthServiceWithAnalytics creates a new AuthService with analytics tracking.
func NewAuthServiceWithAnalytics(q *repository.Queries, cfg *config.Config, e email.Sender, tracker analytics.EventTracker) *AuthService {
	return &AuthService{Queries: q, Config: cfg, Email: e, Analytics: tracker}
}

// AuthResponse is returned by register/login methods (re-exported from domain).
type AuthResponse = domain.AuthResponse

// ProfileResponse wraps a user profile (re-exported from domain).
type ProfileResponse = domain.ProfileResponse

// RegisterOwner creates an owner user and a business, then returns a JWT.
func (s *AuthService) RegisterOwner(ctx context.Context, input domain.RegisterOwnerInput) (*domain.AuthResponse, error) {
	if input.Segment == "" {
		input.Segment = "cafeteria"
	}

	hash, err := domain.HashPassword(input.Password)
	if err != nil {
		return nil, domain.NewErrInternal("erro ao processar senha", err)
	}

	user, err := s.Queries.CreateUser(ctx, repository.CreateUserParams{
		Email:        input.Email,
		PasswordHash: hash,
		Name:         input.Name,
		Phone:        pgText(input.Phone),
		Role:         domain.RoleOwner,
	})
	if err != nil {
		if isDuplicateKeyError(err) {
			return nil, domain.NewErrConflict("email já está em uso")
		}
		return nil, domain.NewErrInternal("erro ao criar usuário", err)
	}

	slug := generateSlug(input.BusinessName)
	biz, err := s.Queries.CreateBusiness(ctx, repository.CreateBusinessParams{
		OwnerID: user.ID,
		Name:    input.BusinessName,
		Slug:    slug,
		Segment: input.Segment,
	})
	if err != nil {
		return nil, domain.NewErrInternal("erro ao criar negócio", err)
	}

	// Track analytics event if tracker is configured
	if s.Analytics != nil {
		userIDStr := formatUserID(user.ID)
		_ = s.Analytics.TrackBusinessCreated(userIDStr, input.BusinessName, input.Segment, "")
	}

	token, err := domain.GenerateJWT(user.ID, domain.RoleOwner, s.Config.JWTSecret, domain.OwnerJWTExpiry)
	if err != nil {
		return nil, domain.NewErrInternal("erro ao gerar token", err)
	}

	bizResp := &domain.BusinessResponse{
		ID:      biz.ID,
		Name:    biz.Name,
		Slug:    biz.Slug,
		Segment: biz.Segment,
	}

	return &domain.AuthResponse{
		Token:    token,
		User:     toUserResponse(&user),
		Business: bizResp,
	}, nil
}

// RegisterSubscriber creates a subscriber user and returns a long-lived JWT.
func (s *AuthService) RegisterSubscriber(ctx context.Context, input domain.RegisterSubscriberInput) (*domain.AuthResponse, error) {
	hash, err := domain.HashPassword(input.Password)
	if err != nil {
		return nil, domain.NewErrInternal("erro ao processar senha", err)
	}

	user, err := s.Queries.CreateUser(ctx, repository.CreateUserParams{
		Email:        input.Email,
		PasswordHash: hash,
		Name:         input.Name,
		Phone:        pgText(input.Phone),
		Role:         domain.RoleSubscriber,
	})
	if err != nil {
		if isDuplicateKeyError(err) {
			return nil, domain.NewErrConflict("email já está em uso")
		}
		return nil, domain.NewErrInternal("erro ao criar usuário", err)
	}

	token, err := domain.GenerateJWT(user.ID, domain.RoleSubscriber, s.Config.JWTSecret, domain.SubscriberJWTExpiry)
	if err != nil {
		return nil, domain.NewErrInternal("erro ao gerar token", err)
	}

	return &domain.AuthResponse{
		Token: token,
		User:  toUserResponse(&user),
	}, nil
}

// Login authenticates a user and returns a JWT.
func (s *AuthService) Login(ctx context.Context, input domain.LoginInput) (*domain.AuthResponse, error) {
	user, err := s.Queries.GetUserByEmail(ctx, input.Email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.NewErrUnauthorized("credenciais inválidas")
		}
		return nil, domain.NewErrInternal("erro ao buscar usuário", err)
	}

	if !domain.CheckPassword(input.Password, user.PasswordHash) {
		return nil, domain.NewErrUnauthorized("credenciais inválidas")
	}

	expiry := domain.OwnerJWTExpiry
	if user.Role == domain.RoleSubscriber {
		expiry = domain.SubscriberJWTExpiry
	}

	token, err := domain.GenerateJWT(user.ID, user.Role, s.Config.JWTSecret, expiry)
	if err != nil {
		return nil, domain.NewErrInternal("erro ao gerar token", err)
	}

	return &domain.AuthResponse{
		Token: token,
		User:  toUserResponse(&user),
	}, nil
}

// RequestPasswordReset generates a password reset token and sends an email.
// Always succeeds from the caller's perspective (to avoid leaking user existence).
func (s *AuthService) RequestPasswordReset(ctx context.Context, emailAddr string) {
	user, err := s.Queries.GetUserByEmail(ctx, emailAddr)
	if err != nil {
		return // user not found or DB error — silently return
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return
	}
	token := hex.EncodeToString(tokenBytes)

	expiresAt := pgTimestamptz(time.Now().Add(time.Hour))
	if _, err := s.Queries.CreatePasswordReset(ctx, repository.CreatePasswordResetParams{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: expiresAt,
	}); err != nil {
		return
	}

	go s.sendPasswordResetEmail(user.Name, user.Email, token)
}

// ConfirmPasswordReset resets the user's password using a valid token.
func (s *AuthService) ConfirmPasswordReset(ctx context.Context, input domain.ConfirmPasswordResetInput) error {
	reset, err := s.Queries.GetPasswordResetByToken(ctx, input.Token)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.NewErrBadRequest("token inválido ou expirado")
		}
		return domain.NewErrInternal("erro ao validar token", err)
	}

	hash, err := domain.HashPassword(input.Password)
	if err != nil {
		return domain.NewErrInternal("erro ao processar senha", err)
	}

	if err := s.Queries.UpdateUserPassword(ctx, repository.UpdateUserPasswordParams{
		ID:           reset.UserID,
		PasswordHash: hash,
	}); err != nil {
		return domain.NewErrInternal("erro ao atualizar senha", err)
	}

	if err := s.Queries.MarkPasswordResetUsed(ctx, reset.ID); err != nil {
		return domain.NewErrInternal("erro ao marcar token como usado", err)
	}

	return nil
}

// GetProfile returns a user's profile by ID.
func (s *AuthService) GetProfile(ctx context.Context, userID int64) (*domain.ProfileResponse, error) {
	user, err := s.Queries.GetUserByID(ctx, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.NewErrNotFound("usuário não encontrado")
		}
		return nil, domain.NewErrInternal("erro ao buscar usuário", err)
	}

	resp := toUserResponse(&user)
	return &domain.ProfileResponse{User: resp}, nil
}

// UpdateProfile updates a user's name and phone.
func (s *AuthService) UpdateProfile(ctx context.Context, userID int64, input domain.UpdateProfileInput) (*domain.ProfileResponse, error) {
	updated, err := s.Queries.UpdateUserProfile(ctx, repository.UpdateUserProfileParams{
		ID:    userID,
		Name:  input.Name,
		Phone: pgText(input.Phone),
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.NewErrNotFound("usuário não encontrado")
		}
		return nil, domain.NewErrInternal("erro ao atualizar perfil", err)
	}

	resp := toUserResponse(&updated)
	return &domain.ProfileResponse{User: resp}, nil
}

// ChangePassword updates a user's password after verifying the current one.
func (s *AuthService) ChangePassword(ctx context.Context, userID int64, input domain.ChangePasswordInput) error {
	user, err := s.Queries.GetUserByID(ctx, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return domain.NewErrNotFound("usuário não encontrado")
		}
		return domain.NewErrInternal("erro ao buscar usuário", err)
	}

	if !domain.CheckPassword(input.CurrentPassword, user.PasswordHash) {
		return domain.NewErrBadRequest("senha atual incorreta")
	}

	hash, err := domain.HashPassword(input.NewPassword)
	if err != nil {
		return domain.NewErrInternal("erro ao processar nova senha", err)
	}

	if err := s.Queries.UpdateUserPassword(ctx, repository.UpdateUserPasswordParams{
		ID:           userID,
		PasswordHash: hash,
	}); err != nil {
		return domain.NewErrInternal("erro ao atualizar senha", err)
	}

	return nil
}

// sendPasswordResetEmail sends a password reset email asynchronously.
func (s *AuthService) sendPasswordResetEmail(userName, userEmail, token string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = ctx

	resetURL := s.Config.FrontendURL + "/resetar-senha?token=" + token
	subject, body := email.PasswordResetEmail(userName, resetURL)
	if s.Email != nil {
		if err := s.Email.Send(userEmail, subject, body); err != nil {
			slog.Error("failed to send password reset email", "error", err, "to", userEmail)
		}
	}
}

// toUserResponse converts a repository.User to a domain.UserResponse.
func toUserResponse(u *repository.User) domain.UserResponse {
	return domain.UserResponse{
		ID:    u.ID,
		Email: u.Email,
		Name:  u.Name,
		Phone: u.Phone.String,
		Role:  u.Role,
	}
}

// generateSlug converts a business name to a URL-safe slug.
// E.g., "Café do João" → "cafe-do-joao"
func generateSlug(name string) string {
	slug := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return unicode.ToLower(r)
		case r >= '0' && r <= '9':
			return r
		case r == ' ' || r == '-' || r == '_':
			return '-'
		default:
			return normalizeRune(r)
		}
	}, name)

	re := regexp.MustCompile(`-+`)
	slug = re.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	return slug
}

// normalizeRune maps accented/special characters to their ASCII equivalents.
func normalizeRune(r rune) rune {
	replacements := map[rune]rune{
		'á': 'a', 'à': 'a', 'ã': 'a', 'â': 'a', 'ä': 'a',
		'é': 'e', 'è': 'e', 'ê': 'e', 'ë': 'e',
		'í': 'i', 'ì': 'i', 'î': 'i', 'ï': 'i',
		'ó': 'o', 'ò': 'o', 'õ': 'o', 'ô': 'o', 'ö': 'o',
		'ú': 'u', 'ù': 'u', 'û': 'u', 'ü': 'u',
		'ç': 'c', 'ñ': 'n',
		'Á': 'a', 'À': 'a', 'Ã': 'a', 'Â': 'a', 'Ä': 'a',
		'É': 'e', 'È': 'e', 'Ê': 'e', 'Ë': 'e',
		'Í': 'i', 'Ì': 'i', 'Î': 'i', 'Ï': 'i',
		'Ó': 'o', 'Ò': 'o', 'Õ': 'o', 'Ô': 'o', 'Ö': 'o',
		'Ú': 'u', 'Ù': 'u', 'Û': 'u', 'Ü': 'u',
		'Ç': 'c', 'Ñ': 'n',
	}
	if v, ok := replacements[r]; ok {
		return v
	}
	return -1 // drop unknown characters
}

// isDuplicateKeyError returns true if the error is a PostgreSQL unique constraint violation.
func isDuplicateKeyError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "duplicate key")
}

// formatUserID converts a user ID to a string for analytics tracking.
func formatUserID(id int64) string {
	return fmt.Sprintf("%d", id)
}
