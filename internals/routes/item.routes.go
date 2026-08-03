package routes

import (
	"github.com/gin-gonic/gin"
	handlers "github.com/othie12/scanner-api/internals/api/hanlders"
	"github.com/othie12/scanner-api/internals/api/middleware"
)

func ItemRoutes(superRouter *gin.RouterGroup) {
	itemHandlers := handlers.NewQrcodeHandler()

	itemRouter := superRouter.Group("/item")

	itemRouter.GET("/scan/:fdn", itemHandlers.Scan)
	itemRouter.GET("/find/:id", itemHandlers.Find)

	itemRouter.Use(middleware.UserLevelMiddleware())
	itemRouter.GET("/all/:page", itemHandlers.All)
	itemRouter.GET("/search", itemHandlers.Search)
	itemRouter.POST("/filter", itemHandlers.GetFiltered)
}
