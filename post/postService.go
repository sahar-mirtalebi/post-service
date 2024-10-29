package post

import (
	"errors"
	"net/http"
	"post-service/category"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PostService struct {
	catRepo *category.CategoryRepository
	repo    *PostRepository
	logger  *zap.Logger
}

func NewPostService(catRepo *category.CategoryRepository, repo *PostRepository, logger *zap.Logger) *PostService {
	return &PostService{catRepo: catRepo, repo: repo, logger: logger}
}

func (service *PostService) CreatePost(userId uint, newPost struct {
	Title       string  `json:"title" validate:"required"`
	Description string  `json:"description" validate:"required"`
	PricePerDay float64 `json:"pricePerDay" validate:"required"`
	Address     string  `json:"address" validate:"required"`
	Category    string  `json:"category" validate:"required"`
}) (uint, error) {
	category, err := service.catRepo.GetCategoryByName(newPost.Category)
	if err != nil {
		service.logger.Error("failed to retrieve category", zap.Error(err))
		return 0, echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	post := Post{
		Title:       newPost.Title,
		Description: newPost.Description,
		PricePerDay: newPost.PricePerDay,
		Address:     newPost.Address,
		CategoryID:  category.ID,
		IsActive:    true,
		CreatedAt:   time.Now(),
		OwnerId:     userId,
	}

	if err := service.repo.AddPost(&post); err != nil {
		service.logger.Error("Error creating post", zap.Error(err))
		return 0, echo.NewHTTPError(http.StatusInternalServerError, "Failed to create post")
	}

	return post.ID, nil
}

type PostResponse struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	PricePerDay float64 `json:"pricePerDay"`
	Address     string  `json:"address"`
	Category    string  `json:"category"`
}

type PostResponseWithOwner struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	PricePerDay float64 `json:"pricePerDay"`
	Address     string  `json:"address"`
	Category    string  `json:"category"`
	OwnerId     uint    `json:"ownerId"`
}

func (service *PostService) GetAllPosts(categoryName, title string, minPrice, maxPrice *int, page, size int) ([]PostResponse, error) {
	var postResponseList []PostResponse
	var categoryId *uint
	if categoryName != "" {
		category, err := service.catRepo.GetCategoryByName(categoryName)
		if err != nil {
			service.logger.Error("failed to retrieve category", zap.Error(err))
			return nil, echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
		}
		categoryId = &category.ID
	}

	offset := (page - 1) * size

	posts, err := service.repo.GetAllPosts(minPrice, maxPrice, title, categoryId, offset, size)
	if err != nil {
		service.logger.Error("error getting posts", zap.Error(err))
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch posts")
	}

	for _, post := range posts {
		category, err := service.catRepo.GetCategoryById(post.CategoryID)
		if err != nil {
			service.logger.Error("error retrieving category by id", zap.Error(err))
			return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to get category")
		}
		postResponseList = append(postResponseList, PostResponse{
			Title:       post.Title,
			Description: post.Description,
			PricePerDay: post.PricePerDay,
			Address:     post.Address,
			Category:    category.Name,
		})

	}
	return postResponseList, nil
}

func (service *PostService) GetPostByID(postId uint) (PostResponseWithOwner, error) {
	retrieveedPost, err := service.repo.GetPostByID(postId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			service.logger.Error("error finding post", zap.Error(err))
			return PostResponseWithOwner{}, echo.NewHTTPError(http.StatusNotFound, "Post not found")
		}
		service.logger.Error("error retrieving post", zap.Error(err))
		return PostResponseWithOwner{}, echo.NewHTTPError(http.StatusInternalServerError, "failed fo get post")
	}
	category, err := service.catRepo.GetCategoryById(retrieveedPost.CategoryID)
	if err != nil {
		service.logger.Error("error retrieve category by id", zap.Error(err))
		return PostResponseWithOwner{}, echo.NewHTTPError(http.StatusInternalServerError, "failed to get category")
	}

	return PostResponseWithOwner{
		Title:       retrieveedPost.Title,
		Description: retrieveedPost.Description,
		PricePerDay: retrieveedPost.PricePerDay,
		Address:     retrieveedPost.Address,
		Category:    category.Name,
		OwnerId:     retrieveedPost.OwnerId,
	}, nil
}

func (service *PostService) GetPostsByOwnerId(userId uint, page, size int) ([]PostResponse, error) {
	var postResponseList []PostResponse
	offset := (page - 1) * size
	posts, err := service.repo.GetPostsByOwnerId(userId, offset, size)
	if err != nil {
		service.logger.Error("failed to retrieve posts by owner ID", zap.Error(err))
		return nil, err
	}

	for _, post := range posts {
		category, err := service.catRepo.GetCategoryById(post.CategoryID)
		if err != nil {
			service.logger.Error("error retrieving category by id", zap.Error(err))
			return nil, echo.NewHTTPError(http.StatusInternalServerError, "failed to get category")
		}

		postResponseList = append(postResponseList, PostResponse{
			Title:       post.Title,
			Description: post.Description,
			PricePerDay: post.PricePerDay,
			Address:     post.Address,
			Category:    category.Name,
		})
	}
	return postResponseList, nil
}

func (service *PostService) UpdatePost(userId, postId uint, updatedPost struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	PricePerDay float64 `json:"pricePerDay"`
	Address     string  `json:"address"`
	Category    string  `json:"category"`
	IsActive    bool    `json:"isActive"`
}) error {
	post, err := service.repo.GetPostByID(postId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			service.logger.Error("Post not found", zap.Error(err))
			return echo.NewHTTPError(http.StatusNotFound, "Post not found")
		}
		service.logger.Error("error retrieving post", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to retrive post")
	}

	if userId != post.OwnerId {
		service.logger.Error("error not allowed to update", zap.Error(err))
		return echo.NewHTTPError(http.StatusForbidden, "not authorised to update post")
	}

	var categoryId *uint
	if updatedPost.Category != "" {
		category, err := service.catRepo.GetCategoryByName(updatedPost.Category)
		if err != nil {
			service.logger.Error("error retrieving category", zap.Error(err))
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve category")
		}
		categoryId = &category.ID
	}

	if updatedPost.Title != "" {
		post.Title = updatedPost.Title
	}
	if updatedPost.Description != "" {
		post.Description = updatedPost.Description
	}
	if updatedPost.PricePerDay != 0 {
		post.PricePerDay = updatedPost.PricePerDay
	}
	if updatedPost.Address != "" {
		post.Address = updatedPost.Address
	}
	if categoryId != nil {
		post.CategoryID = *categoryId
	}
	post.IsActive = updatedPost.IsActive
	post.UpdatedAt = time.Now()

	err = service.repo.UpdatePost(post)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update post")
	}
	return nil
}

func (service *PostService) DeletePost(postId, userId uint) error {
	post, err := service.repo.GetPostByID(postId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			service.logger.Error("Post not found", zap.Error(err))
			return echo.NewHTTPError(http.StatusNotFound, "Post not found")
		}
		service.logger.Error("error retrieving post", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to retrive post")

	}

	if post.OwnerId != userId {
		service.logger.Error("error not allowed to delete", zap.Error(err))
		return echo.NewHTTPError(http.StatusForbidden, "not authorised to delete post")
	}

	err = service.repo.DeletePost(postId)
	if err != nil {
		service.logger.Error("error deleting post", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete post")
	}
	return nil
}
