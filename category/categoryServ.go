package category

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CategoryService struct {
	catRepo *CategoryRepository
	logger  *zap.Logger
}

func NewCategoryService(catRepo *CategoryRepository, logger *zap.Logger) *CategoryService {
	return &CategoryService{catRepo: catRepo, logger: logger}
}

func (catService *CategoryService) CreateCategory(category Category) (uint, error) {
	category.CreatedAt = time.Now()
	err := catService.catRepo.AddCategory(&category)
	if err != nil {
		catService.logger.Error("error adding category", zap.Error(err))
		return 0, echo.NewHTTPError(http.StatusInternalServerError, "fail to add category")
	}
	return category.ID, nil
}

func (catService *CategoryService) GetAllCategories() ([]string, error) {
	categories, err := catService.catRepo.GetAllCategories()
	if err != nil {
		catService.logger.Error("error retrieving categories")
		return nil, echo.NewHTTPError(http.StatusInternalServerError)
	}

	var categoryList []string
	for _, category := range categories {
		categoryList = append(categoryList, category.Name)
	}

	return categoryList, nil
}

func (catService *CategoryService) GetCategoryById(categoryId uint) (string, error) {
	category, err := catService.catRepo.GetCategoryById(categoryId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			catService.logger.Error("category not found", zap.Error(err))
			return "", echo.NewHTTPError(http.StatusNotFound, "category not found")
		}
		catService.logger.Error("error retrieving category", zap.Error(err))
		return "", echo.NewHTTPError(http.StatusInternalServerError, "failed to retrive category")
	}

	categoryName := category.Name
	return categoryName, nil
}

func (catService *CategoryService) UpdateCategory(categoryId uint, categoryName string) error {
	category, err := catService.catRepo.GetCategoryById(categoryId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			catService.logger.Error("category not found", zap.Error(err))
			return echo.NewHTTPError(http.StatusNotFound, "category not found")
		}
		catService.logger.Error("error retrieving category", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to retrive category")
	}

	category.Name = categoryName
	err = catService.catRepo.UpdateCategory(category)
	if err != nil {
		catService.logger.Error("error updating category", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update category")
	}

	return nil
}
