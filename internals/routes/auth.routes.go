package routes

import (
	"github.com/gin-gonic/gin"
	handlers "github.com/othie12/scanner-api/internals/api/hanlders"
	"github.com/othie12/scanner-api/internals/api/middleware"
)

func AuthRoutes(superRouter *gin.RouterGroup) {
	authHandlers := handlers.NewAuthHandler()

	authRouter := superRouter.Group("/auth")
	authRouter.POST("/login", authHandlers.Login)
	authRouter.GET("/logout", authHandlers.Logout)

	authRouter.Use(middleware.AuthMiddleware())
	authRouter.PATCH("/password/:id", authHandlers.ChangePassword)

	authRouter.Use(middleware.UserLevelMiddleware())
	authRouter.PATCH("/userlevel/:id", authHandlers.ChangeUserLevel)
}
