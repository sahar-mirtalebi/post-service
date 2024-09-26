package category

import (
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type CategoryHandler struct {
	catService *CategoryService
	logger     *zap.Logger
	validator  *validator.Validate
}

func NewCategoryHandler(catService *CategoryService, logger *zap.Logger, validator *validator.Validate) *CategoryHandler {
	return &CategoryHandler{catService: catService, logger: logger, validator: validator}
}

func (catHandler *CategoryHandler) CreateCategory(c echo.Context) error {
	var category Category

	if err := c.Bind(&category); err != nil {
		catHandler.logger.Error("error binding request", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "failed to bind request")
	}

	if err := catHandler.validator.Struct(category); err != nil {
		catHandler.logger.Error("provided data is invalid", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid data")
	}

	categoryId, err := catHandler.catService.CreateCategory(category)
	if err != nil {
		catHandler.logger.Error("error create category", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create category")
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"categoryId": categoryId,
	})
}

func (catHandler *CategoryHandler) GetAllCategories(c echo.Context) error {
	categories, err := catHandler.catService.GetAllCategories()
	if err != nil {
		catHandler.logger.Error("error retrieving categories")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to retrieve categories")
	}

	return c.JSON(http.StatusOK, categories)
}

func (catHandler *CategoryHandler) GetCategoryById(c echo.Context) error {
	categoryIdStr := c.Param("categoryId")
	if categoryIdStr == "" {
		catHandler.logger.Error("error missing category id")
		return echo.NewHTTPError(http.StatusBadRequest, "missing category id")
	}
	categoryId, err := strconv.ParseUint(categoryIdStr, 10, 32)
	if err != nil {
		catHandler.logger.Error("invalid categoryId", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid categoryID")
	}

	categoryname, err := catHandler.catService.GetCategoryById(uint(categoryId))
	if err != nil {
		catHandler.logger.Error("error retrieving category by id", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get category name")
	}

	return c.JSON(http.StatusOK, categoryname)
}

func (catHandler *CategoryHandler) UpdateCategory(c echo.Context) error {
	categoryIdStr := c.Param("categoryId")
	if categoryIdStr == "" {
		catHandler.logger.Error("error missing category id")
		return echo.NewHTTPError(http.StatusBadRequest, "missing category id")
	}
	categoryId, err := strconv.ParseUint(categoryIdStr, 10, 32)
	if err != nil {
		catHandler.logger.Error("invalid categoryId", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid categoryID")
	}

	var category struct {
		Name string `gorm:"unique" json:"name" validate:"required"`
	}
	if err := c.Bind(&category); err != nil {
		catHandler.logger.Error("error binding request")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	if err := catHandler.validator.Struct(category); err != nil {
		catHandler.logger.Error("invalid input data", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request data")
	}

	err = catHandler.catService.UpdateCategory(uint(categoryId), category.Name)
	if err != nil {
		catHandler.logger.Error("failed to update category", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update category")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "category has been updated seccusfully",
	})
}
