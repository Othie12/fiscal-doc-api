package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/othie12/scanner-api/internals/api/middleware"
)

func ApiRouter(mainRouter *gin.Engine) {
	apiRouter := mainRouter.Group("/api")

	AuthRoutes(apiRouter)

	apiRouter.Use(middleware.AuthMiddleware())
	UserRoutes(apiRouter)
	ItemRoutes(apiRouter)
}
