package routes

import (
	"github.com/gin-gonic/gin"
	handlers "github.com/othie12/scanner-api/internals/api/hanlders"
)

func UserRoutes(superRouter *gin.RouterGroup) {
	userHandlers := handlers.NewUserHandler()

	userRouter := superRouter.Group("/user")

	userRouter.GET("/all/:page", userHandlers.All)
	userRouter.GET("/search", userHandlers.Search)

	userRouter.GET("/authenticated", userHandlers.GetAuthenticated)
	userRouter.POST("/create", userHandlers.Create)

	userRouter.GET("/find/:id", userHandlers.Get)
	//userRouter.Use(middleware.RoleMiddleware())
}
