package handler

import (
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/jackc/pgx/v5"

	"github.com/clubepay/backend/internal/domain"
	"github.com/clubepay/backend/internal/repository"
)

// registerOwnerRequest holds the JSON body for RegisterOwner.
type registerOwnerRequest struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	Name         string `json:"name"`
	BusinessName string `json:"business_name"`
	Segment      string `json:"segment"`
	Phone        string `json:"phone"`
}

// loginRequest holds the JSON body for Login.
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// registerSubscriberRequest holds the JSON body for RegisterSubscriber.
type registerSubscriberRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
}

// RegisterOwner creates an owner user and a business, then returns a JWT.
// POST /api/auth/register
func (h *Handler) RegisterOwner(w http.ResponseWriter, r *http.Request) {
	var req registerOwnerRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate required fields
	if req.Email == "" {
		writeError(w, http.StatusBadRequest, "email is required")
		return
	}
	if req.Password == "" {
		writeError(w, http.StatusBadRequest, "password is required")
		return
	}
	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.BusinessName == "" {
		writeError(w, http.StatusBadRequest, "business_name is required")
		return
	}
	if req.Segment == "" {
		req.Segment = "cafeteria"
	}

	hash, err := domain.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	user, err := h.Queries.CreateUser(r.Context(), repository.CreateUserParams{
		Email:        req.Email,
		PasswordHash: hash,
		Name:         req.Name,
		Phone:        pgText(req.Phone),
		Role:         domain.RoleOwner,
	})
	if err != nil {
		if isDuplicateKeyError(err) {
			writeError(w, http.StatusConflict, "email already in use")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	slug := generateSlug(req.BusinessName)
	biz, err := h.Queries.CreateBusiness(r.Context(), repository.CreateBusinessParams{
		OwnerID: user.ID,
		Name:    req.BusinessName,
		Slug:    slug,
		Segment: req.Segment,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create business")
		return
	}

	token, err := domain.GenerateJWT(user.ID, domain.RoleOwner, h.Config.JWTSecret, 24*time.Hour)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
			"role":  user.Role,
		},
		"business": map[string]interface{}{
			"id":      biz.ID,
			"name":    biz.Name,
			"slug":    biz.Slug,
			"segment": biz.Segment,
		},
	})
}

// Login authenticates a user and returns a JWT.
// POST /api/auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" {
		writeError(w, http.StatusBadRequest, "email is required")
		return
	}
	if req.Password == "" {
		writeError(w, http.StatusBadRequest, "password is required")
		return
	}

	user, err := h.Queries.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to query user")
		return
	}

	if !domain.CheckPassword(req.Password, user.PasswordHash) {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	expiry := 24 * time.Hour
	if user.Role == domain.RoleSubscriber {
		expiry = 30 * 24 * time.Hour
	}

	token, err := domain.GenerateJWT(user.ID, user.Role, h.Config.JWTSecret, expiry)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
			"role":  user.Role,
		},
	})
}

// RegisterSubscriber creates a subscriber user and returns a long-lived JWT (30 days).
// POST /api/auth/register-subscriber
func (h *Handler) RegisterSubscriber(w http.ResponseWriter, r *http.Request) {
	var req registerSubscriberRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" {
		writeError(w, http.StatusBadRequest, "email is required")
		return
	}
	if req.Password == "" {
		writeError(w, http.StatusBadRequest, "password is required")
		return
	}
	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	hash, err := domain.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	user, err := h.Queries.CreateUser(r.Context(), repository.CreateUserParams{
		Email:        req.Email,
		PasswordHash: hash,
		Name:         req.Name,
		Phone:        pgText(req.Phone),
		Role:         domain.RoleSubscriber,
	})
	if err != nil {
		if isDuplicateKeyError(err) {
			writeError(w, http.StatusConflict, "email already in use")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	token, err := domain.GenerateJWT(user.ID, domain.RoleSubscriber, h.Config.JWTSecret, 30*24*time.Hour)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
			"role":  user.Role,
		},
	})
}

// generateSlug converts a business name to a URL-safe slug.
// E.g., "Café do João" → "cafe-do-joao"
func generateSlug(name string) string {
	// Normalize: remove accents / replace with ASCII equivalents
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
			// Replace accented chars with their ASCII base
			normalized := normalizeRune(r)
			return normalized
		}
	}, name)

	// Collapse multiple hyphens
	re := regexp.MustCompile(`-+`)
	slug = re.ReplaceAllString(slug, "-")

	// Trim leading/trailing hyphens
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
