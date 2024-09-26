package category

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"unique" json:"name" validate:"required"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (catRepo *CategoryRepository) AddCategory(category *Category) error {
	if err := catRepo.db.Create(&category).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return errors.New("category already exists")
		}
		return err
	}
	return nil
}

func (catRepo *CategoryRepository) GetAllCategories() ([]Category, error) {
	var categories []Category
	result := catRepo.db.Find(&categories)
	if result.Error != nil {
		return nil, result.Error
	}
	return categories, nil
}

func (catRepo *CategoryRepository) GetCategoryById(categoryId uint) (*Category, error) {
	var category Category
	err := catRepo.db.First(&category, categoryId).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (catRepo *CategoryRepository) GetCategoryByName(categoryName string) (*Category, error) {
	var category Category
	if err := catRepo.db.Where("name = ?", categoryName).First(&category).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (catRepo *CategoryRepository) UpdateCategory(updatedCategory *Category) error {
	return catRepo.db.Save(updatedCategory).Error
}
