package pkg

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lime666/spotify_api/config"
	"github.com/zmb3/spotify/v2"
	spotifyAuth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/exp/rand"
	"golang.org/x/oauth2"
)

type OAuthService struct {
	Config *oauth2.Config
	auth   *spotifyAuth.Authenticator
	state  string
}

func generateRandomState() string {
	rand.Seed(uint64(time.Now().UnixNano()))
	return fmt.Sprintf("%d", rand.Intn(100000))
}

func NewOAuthService(c config.Config) *OAuthService {
	state := generateRandomState()

	auth := spotifyAuth.New(
		spotifyAuth.WithClientID(c.SpotifyClientID),
		spotifyAuth.WithClientSecret(c.SpotifyClientSecret),
		spotifyAuth.WithRedirectURL(c.SpotifyRedirectURL),
		spotifyAuth.WithScopes(
			spotifyAuth.ScopeUserReadEmail,
			spotifyAuth.ScopePlaylistReadPrivate,
			spotifyAuth.ScopeUserLibraryRead,
			spotifyAuth.ScopeUserReadPrivate,
			spotifyAuth.ScopeUserTopRead,
		),
	)

	return &OAuthService{auth: auth, state: state}
}

// @Summary      Initiate Spotify login
// @Description  Redirects the user to the Spotify authorization page.
// @Tags         auth
// @Produce      json
// @Success      200  "Redirect to Spotify OAuth"
// @Router       /login/spotify [get]
func (o *OAuthService) LoginHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"link_to_redirect": o.auth.AuthURL(o.state),
	})
	// c.Redirect(http.StatusTemporaryRedirect, o.auth.AuthURL(o.state))
	// log.Println("listening on %s", o.auth.AuthURL(o.state))
}

// @Summary      Spotify OAuth2 callback
// @Description  Verifies state, exchanges code for an access token, fetches the current user and returns their email and token info.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        state    query     string  true  "Original state value for CSRF protection"
// @Param        code     query     string  true  "Authorization code returned by Spotify"
// @Success      200      {object}  map[string]interface{}  "Returns the user’s email, the access token, and its TTL in seconds"
// @Failure      400      {object}  map[string]string       "Invalid or missing state or code"
// @Failure      500      {object}  map[string]string       "Token exchange or user-info fetch failure"
// @Router       /callback/spotify [get]
func (o *OAuthService) CallbackHandler(c *gin.Context) {
	if c.Query("state") != o.state {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state"})
		c.JSON(http.StatusBadRequest, gin.H{"error": "state mismatch", "got": c.Query("state"), "want": o.state})
		return
	}

	token, err := o.auth.Exchange(c.Request.Context(), c.Query("code"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not exchange token", "details": err.Error()})
		return
	}
	log.Printf("▶️ granted scopes: %q", token.Extra("scope"))

	client := spotify.New(o.auth.Client(c.Request.Context(), token))
	c.Set("spotifyClient", client)

	user, err := client.CurrentUser(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch user"})
		return
	}

	c.JSON(200, gin.H{
		"email":        user.Email,
		"access_token": token.AccessToken,
		"expires_in":   token.Expiry.Sub(time.Now()).Seconds(),
	})
}

// @Summary      Analyze user music data
// @Description  Uses the provided Bearer token to fetch the user’s top tracks and audio features, then returns Archetype and top genres.
// @Tags         analyze
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer token (add 'Bearer ' prefix to token)"
// @Success      200            {object}  Summary  "Analysis metrics"
// @Failure      401            {object}  map[string]string  "Missing or invalid token"
// @Failure      403            {object}  map[string]string  "Forbidden"
// @Failure      500            {object}  map[string]string  "Analysis error"
// @Router       /analyze [get]
func (o *OAuthService) AnalyzeHandler(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		c.JSON(401, gin.H{"error": "missing token"})
		return
	}
	incoming := authHeader[len("Bearer "):]

	client := spotify.New(o.auth.Client(c.Request.Context(), &oauth2.Token{AccessToken: incoming}))
	summary, err := AnalyzeUser(c.Request.Context(), client)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error(), "token": incoming})
		return
	}
	c.JSON(200, summary)
}

// ---- SPOTIFY MIDDLEWARE ----

func (svc *OAuthService) SpotifyClientMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")
		if !strings.HasPrefix(strings.ToLower(authorization), "bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		tokenStr := strings.TrimPrefix(authorization, "Bearer ")

		token := &oauth2.Token{AccessToken: tokenStr}

		httpClient := svc.Config.Client(c.Request.Context(), token)
		spotifyClient := spotify.New(httpClient)

		c.Set("spotifyClient", spotifyClient)
		c.Next()
	}
}
