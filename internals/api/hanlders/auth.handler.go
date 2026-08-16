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
	database "github.com/othie12/scanner-api/internals/db"
	"github.com/othie12/scanner-api/internals/db/models"
	"github.com/othie12/scanner-api/utils"
)

type AuthHandler struct {
	userService *services.UserService
	validator   *validator.Validate
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		validator:   validator.New(),
		userService: services.NewUserService(),
	}
}

type LoginDTO struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var dto LoginDTO
	var user models.User

	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, utils.ResponseWrapper(false, "Please Check your request and try again", nil))
		log.Println("Error mashalling to json: " + err.Error())
		return
	}

	// validate input
	err := h.validator.Struct(dto)
	if err != nil {
		errs := ""
		for _, err := range err.(validator.ValidationErrors) {
			errs = fmt.Sprintf("%s, %s", errs, err.Error())
		}
		c.JSON(http.StatusBadRequest, utils.ResponseWrapper(false, errs, nil))
		log.Println("Login Validation Err: " + errs)
		return
	}

	if result := database.DB.Where("username = ?", dto.Username).First(&user); result.Error != nil {
		c.JSON(http.StatusNotFound, utils.ResponseWrapper(false, fmt.Sprintf("User with username: '%s' doesn't exist.", dto.Username), nil))
		return
	}

	if !utils.CheckPasswordHash(dto.Password, user.Password) {
		c.JSON(http.StatusNotFound, utils.ResponseWrapper(false, "Incorrect password", nil))
		return
	}

	token, err := utils.CreateToken(user.ID, user.UserLevel)
	if err != nil {
		log.Println("Error Creating token: " + err.Error())
		c.JSON(http.StatusNotFound, utils.ResponseWrapper(false, "An error occured on our server, please try again later", nil))
		return
	}

	user.RemoveSensitiveData()
	//set cookie for 7 days
	c.SetCookie("authToken", token, 3600*24*7, "/", "", false, true)
	c.JSON(http.StatusAccepted, utils.ResponseWrapper(true, "Logged in succesfuly", gin.H{"user": user, "token": token}))
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("authToken", "", -1, "/", "", false, true)
	c.JSON(http.StatusAccepted, utils.ResponseWrapper(true, "Logged out succesfuly", nil))
}

func (h *AuthHandler) ChangeUserLevel(c *gin.Context) {
	newLevel := c.Query("userLevel")
	id := c.Param("id")

	user, status, err := h.userService.ChangeUserLevel(id, newLevel)
	if err != nil {
		c.JSON(status, utils.ResponseWrapper(false, err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, utils.ResponseWrapper(true, "updated Role successfuly", user))
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	id := c.Param("id")
	var dto services.ChangePwdDTO
	authUserId, authUserLevel := middleware.GetAuthUser(c)

	// 1. Parse string to uint64 (base 10, up to 64-bit size)
	idUint, errsd := strconv.ParseUint(id, 10, 64)
	if errsd != nil {
		c.JSON(http.StatusBadRequest, utils.ResponseWrapper(false, "Please Check your request and try again", nil))
		fmt.Println("Error parsing string:", errsd)
		return
	}

	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, utils.ResponseWrapper(false, "Please Check your request and try again", nil))
		log.Println("Error mashalling to json: " + err.Error())
		return
	}

	// validate input
	err := h.validator.Struct(dto)
	if err != nil {
		errs := ""
		for _, err := range err.(validator.ValidationErrors) {
			errs = fmt.Sprintf("%s, %s", errs, err.Error())
		}
		c.JSON(http.StatusBadRequest, utils.ResponseWrapper(false, errs, nil))
		log.Println("Login Validation Err: " + errs)
		return
	}

	user, status, err := h.userService.ChangePwd(id, authUserId == uint(idUint) || authUserLevel != "admin", dto)
	if err != nil {
		c.JSON(status, utils.ResponseWrapper(false, err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, utils.ResponseWrapper(true, "updated Password successfuly", user))
}
