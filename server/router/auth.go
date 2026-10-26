package router

import (
	"github.com/gin-gonic/gin"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/auth"
)

func authRouter(r *gin.Engine) {
	service := auth.NewService()

	r.POST("/api/auth/login", service.Login)           // login
	r.POST("/api/auth/login/totp", service.LoginTOTP)  // login, second step
	r.POST("/api/auth/share", service.RedeemShareLink) // open a view-only session with a share link's token

	api := r.Group("/api").Use(middleware.CheckToken())

	api.GET("/auth/password", service.IsPasswordUpdated) // is password updated
	api.GET("/auth/account", service.GetAccount)         // get account
	api.POST("/auth/password", service.ChangePassword)   // change password
	api.POST("/auth/logout", service.Logout)             // logout

	admin := r.Group("/api").Use(
		middleware.CheckToken(),
		middleware.RequireRole(authn.RoleAdmin),
	)
	admin.GET("/auth/users", service.ListUsers)
	admin.POST("/auth/users", service.CreateUser)
	admin.PUT("/auth/users/:username", service.UpdateUser)
	admin.DELETE("/auth/users/:username", service.DeleteUser)
	admin.POST("/auth/users/:username/password", service.ChangeUserPassword)

	// Issuing and revoking keys is session work: an api key must not be able
	// to mint further keys.
	keys := r.Group("/api").Use(middleware.CheckSession())

	keys.GET("/auth/api-keys", service.GetAPIKeys)          // list api keys
	keys.POST("/auth/api-keys", service.CreateAPIKey)       // create an api key
	keys.DELETE("/auth/api-keys/:id", service.DeleteAPIKey) // revoke an api key

	// The second factor is session work too. API keys bypass it, as they
	// bypass the password, so a key must not be able to turn it off either.
	keys.GET("/auth/totp", service.GetTOTP)
	keys.POST("/auth/totp/setup", service.SetupTOTP)
	keys.POST("/auth/totp/enable", service.EnableTOTP)
	keys.POST("/auth/totp/disable", service.DisableTOTP)
	keys.POST("/auth/totp/recovery-codes", service.RegenerateRecoveryCodes)

	// Login sessions are session work too: a key must not be able to sign
	// its owner out, or see where they are signed in from.
	keys.GET("/auth/sessions", service.ListSessions)
	keys.DELETE("/auth/sessions/:id", service.RevokeSession)
	keys.POST("/auth/sessions/revoke-others", service.RevokeOtherSessions)

	adminSession := r.Group("/api").Use(
		middleware.CheckSession(),
		middleware.RequireRole(authn.RoleAdmin),
	)
	adminSession.DELETE("/auth/users/:username/totp", service.ResetUserTOTP)
	adminSession.GET("/auth/sessions/all", service.ListAllSessions)

	// Share links are session work like api keys: a key must not mint a
	// link that outlives it.
	adminSession.GET("/auth/share-links", service.ListShareLinks)
	adminSession.POST("/auth/share-links", service.CreateShareLink)
	adminSession.DELETE("/auth/share-links/:id", service.RevokeShareLink)
}
