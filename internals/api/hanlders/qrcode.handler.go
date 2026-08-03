package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/othie12/scanner-api/internals/api/services"
	"github.com/othie12/scanner-api/utils"
)

type QrcodeHandler struct {
	service   *services.QrcodeService
	validator *validator.Validate
}

func NewQrcodeHandler() *QrcodeHandler {
	return &QrcodeHandler{
		validator: validator.New(),
		service:   services.NewQrcodeService(),
	}
}

func (h *QrcodeHandler) GetFiltered(c *gin.Context) {
	var dto services.FilterItemsDTO

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
		log.Println("Filter Validation Err: " + errs)
		return
	}

	items, status, err := h.service.GetFiltered(dto)
	if err != nil {
		c.JSON(status, utils.ResponseWrapper(false, err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, utils.ResponseWrapper(true, "Fetched successfully", items))
}

func (h *QrcodeHandler) Scan(c *gin.Context) {
	fdn := c.Param("fdn")
	id, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusBadRequest, utils.ResponseWrapper(false, "Please log in first", nil))
		return
	}

	item, status, err := h.service.Scan(fdn, id)
	if err != nil {
		c.JSON(status, utils.ResponseWrapper(false, err.Error(), nil))
		return
	}

	c.JSON(status, utils.ResponseWrapper(true, "ok", item))
}

func (h *QrcodeHandler) All(c *gin.Context) {
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

	c.JSON(http.StatusOK, utils.ResponseWrapper(true, "Fetched successfully", items))
}

func (h *QrcodeHandler) Find(c *gin.Context) {
	id := c.Param("id")

	item, status, err := h.service.Find(id)
	if err != nil {
		c.JSON(status, utils.ResponseWrapper(false, err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, utils.ResponseWrapper(true, "Retrieved succesfuly", item))
}

func (h *QrcodeHandler) Search(c *gin.Context) {
	query := c.Query("query")

	items, status, err := h.service.Search(query)
	if err != nil {
		c.JSON(status, utils.ResponseWrapper(false, err.Error(), nil))
		return
	}

	c.JSON(http.StatusOK, utils.ResponseWrapper(true, "Fetched successfully", items))
}
