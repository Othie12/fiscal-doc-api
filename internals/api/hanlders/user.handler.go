package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/othie12/scanner-api/internals/api/middleware"
	"github.com/othie12/scanner-api/internals/api/services"
	"github.com/othie12/scanner-api/internals/db/models"
	"github.com/othie12/scanner-api/utils"
)

type UserHandler struct {
	service   *services.UserService
	validator *validator.Validate
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		validator: validator.New(),
		service:   services.NewUserService(),
	}
}

func (h *UserHandler) Create(c *gin.Context) {
	var dto models.User

	err := c.ShouldBindJSON(&dto)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ResponseWrapper(false, "An error occured on our side, Pleas try again later", nil))
		log.Println("Error mashalling to json: " + err.Error())
		return
	}

	// validate input
	err = h.validator.Struct(dto)
	if err != nil {
		errs := ""
		for _, err := range err.(validator.ValidationErrors) {
			errs = fmt.Sprintf("%s, %s", errs, err.Error())
		}
		c.JSON(http.StatusBadRequest, utils.ResponseWrapper(false, errs, nil))
		log.Println("Signup Validation Err: " + errs)
		return
	}

	// prevent non-approvers from creating entrants and approvers
	authUserID, authUserRole := middleware.GetAuthUser(c)
	if authUserRole != "admin" {
		errMsg := "You have no permission to create users"
		log.Printf("User with ID: %d and role: %s tried to create user with role: %s", authUserID, authUserRole, dto.UserLevel)
		c.JSON(http.StatusBadRequest, utils.ResponseWrapper(false, errMsg, nil))
		return
	}

	user, status, err := h.service.Create(dto)
	if err != nil {
		c.JSON(status, utils.ResponseWrapper(false, err.Error(), nil))
		return
	}

	user.RemoveSensitiveData()
	c.JSON(http.StatusAccepted, utils.ResponseWrapper(true, "Created succesfuly", user))
}

func (h *UserHandler) Get(c *gin.Context) {

	id := c.Param("id")

	user, status, err := h.service.Get(id)
	if err != nil {
		c.JSON(status, utils.ResponseWrapper(false, err.Error(), nil))
		return
	}

	user.RemoveSensitiveData()
	c.JSON(http.StatusAccepted, utils.ResponseWrapper(true, "Retrieved succesfuly", user))
}

// Get authenticated user
func (h *UserHandler) GetAuthenticated(c *gin.Context) {

	id, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusBadRequest, utils.ResponseWrapper(false, "User not found in request", nil))
		return
	}

	user, status, err := h.service.Get(id)
	if err != nil {
		c.JSON(status, utils.ResponseWrapper(false, err.Error(), nil))
		return
	}

	user.RemoveSensitiveData()
	c.JSON(http.StatusAccepted, utils.ResponseWrapper(true, "Retrieved succesfuly", user))
}

func (h *UserHandler) Search(c *gin.Context) {
	query := c.Query("query")

	items, status, err := h.service.Search(query)
	if err != nil {
		c.JSON(status, utils.ResponseWrapper(false, err.Error(), nil))
		return
	}

	c.JSON(http.StatusAccepted, utils.ResponseWrapper(true, "Fetched successfully", items))
}

func (h *UserHandler) All(c *gin.Context) {
	page := c.Param("page")

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		pageInt = 1
	}
	items, status, err := h.service.All(pageInt)
	if err != nil {
		c.JSON(status, utils.ResponseWrapper(false, err.Error(), nil))
		return
	}

	c.JSON(http.StatusAccepted, utils.ResponseWrapper(true, "Fetched successfully", items))
}
