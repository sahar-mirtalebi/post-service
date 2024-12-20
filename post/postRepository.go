package post

import (
	"time"

	"gorm.io/gorm"
)

type Post struct {
	ID          uint
	Title       string
	Description string
	PricePerDay float64
	Address     string
	CategoryID  uint
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	OwnerId     uint
}

type PostRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (repo *PostRepository) AddPost(post *Post) error {
	return repo.db.Create(&post).Error
}

func (repo *PostRepository) UpdatePost(updatedpost *Post) error {
	return repo.db.Save(updatedpost).Error
}

func (repo *PostRepository) DeletePost(postId uint) error {
	return repo.db.Delete(&Post{}, postId).Error
}

func (repo *PostRepository) GetPostByID(postId uint) (*Post, error) {
	var post Post
	err := repo.db.First(&post, postId).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (repo *PostRepository) GetPostsByOwnerId(ownerId uint, offset, limit int) ([]Post, error) {
	var posts []Post
	err := repo.db.Model(&Post{}).Where("owner_id = ?", ownerId).Offset(offset).Limit(limit).Find(&posts).Error
	return posts, err
}

func (repo *PostRepository) GetAllPosts(minPrice, maxPrice *int, title string, categoryId *uint, offset, limit int) ([]Post, error) {
	var posts []Post

	query := repo.db.Model(&Post{})
	if categoryId != nil && *categoryId > 0 {
		query = query.Where("category_id = ?", categoryId)
	}

	if title != "" {
		query = query.Where("title ILIKE ?", "%"+title+"%")
	}

	if minPrice != nil && *minPrice > 0 {
		query = query.Where("price_per_day >= ?", minPrice)
	}

	if maxPrice != nil && *maxPrice > 0 {
		query = query.Where("price_per_day <= ?", maxPrice)
	}

	err := query.Offset(offset).Limit(limit).Find(&posts).Error
	return posts, err
}
