package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Handler struct {
	service      Service
	googleConfig *oauth2.Config
}

func NewHandler(service Service, googleClientID, googleClientSecret, baseURL string) *Handler {
	return &Handler{
		service: service,
		googleConfig: &oauth2.Config{
			ClientID:     googleClientID,
			ClientSecret: googleClientSecret,
			RedirectURL:  baseURL + "/auth/google/callback",
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
	}
}

type registerRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Username string `json:"username" binding:"required,min=3,max=30,alphanum"`
	Password string `json:"password" binding:"required,min=8"`
}

type loginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, tokens, err := h.service.Register(c.Request.Context(), req.Email, req.Username, req.Password)
	if errors.Is(err, ErrEmailTaken) {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}
	if errors.Is(err, ErrUsernameTaken) {
		c.JSON(http.StatusConflict, gin.H{"error": "username already taken"})
		return
	}
	if errors.Is(err, ErrUsernameReserved) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "username is reserved and cannot be used"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "registered",
		"data": gin.H{
			"user":  user,
			"token": tokens.Token,
		},
	})
}

func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, tokens, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if errors.Is(err, ErrInvalidCreds) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"data": gin.H{
			"user":         user,
			"token":        tokens.Token,
			"refreshToken": tokens.RefreshToken,
		},
	})
}

func (h *Handler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := h.service.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"data": gin.H{
			"token":        tokens.Token,
			"refreshToken": tokens.RefreshToken,
		},
	})
}

func (h *Handler) GoogleLogin(c *gin.Context) {
	state := "random-state" // in production use crypto/rand
	url := h.googleConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *Handler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
		return
	}

	oauthToken, err := h.googleConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to exchange token"})
		return
	}

	info, err := fetchGoogleUserInfo(oauthToken.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user info"})
		return
	}

	user, tokens, err := h.service.LoginOrRegisterGoogle(
		c.Request.Context(),
		info.ID, info.Email, info.Name, info.Picture,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"data": gin.H{
			"user":         user,
			"token":        tokens.Token,
			"refreshToken": tokens.RefreshToken,
		},
	})
}
