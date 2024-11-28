package post

import (
	"errors"
	"post-service/category"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type PostService struct {
	catRepo *category.CategoryRepository
	repo    *PostRepository
}

func NewPostService(catRepo *category.CategoryRepository, repo *PostRepository) *PostService {
	return &PostService{catRepo: catRepo, repo: repo}
}

var ErrPostNotFound = errors.New("post not found")
var ErrForbidden = errors.New("not allowed to update post")

func (service *PostService) CreatePost(userId uint, newPost struct {
	Title       string  `json:"title" validate:"required"`
	Description string  `json:"description" validate:"required"`
	PricePerDay float64 `json:"pricePerDay" validate:"required"`
	Address     string  `json:"address" validate:"required"`
	Category    string  `json:"category" validate:"required"`
}) (*uint, error) {
	category, err := service.catRepo.GetCategoryByName(newPost.Category)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	return &post.ID, nil
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

func (service *PostService) GetAllPosts(categoryName, title, priceStr string, page int) (*[]PostResponse, error) {
	var postResponseList []PostResponse
	var categoryId *uint
	if categoryName != "" {
		category, err := service.catRepo.GetCategoryByName(categoryName)
		if err != nil {
			return nil, err
		}
		categoryId = &category.ID
	}

	var minPrice, maxPrice *int

	if priceStr != "" {
		price := strings.Split(priceStr, "-")
		if price[0] != "" {
			min, err := strconv.Atoi(price[0])
			if err != nil {
				return nil, err

			}
			minPrice = &min
		}
		if price[1] != "" {
			max, err := strconv.Atoi(price[1])
			if err != nil {
				return nil, err
			}
			maxPrice = &max
		}
		if minPrice != nil && maxPrice != nil && *minPrice > *maxPrice {

			return nil, errors.New("minimum price cannot be greater than maximum price")
		}
	}

	size := 10
	offset := (page - 1) * size

	posts, err := service.repo.GetAllPosts(minPrice, maxPrice, title, categoryId, offset, size)
	if err != nil {
		return nil, err
	}

	for _, post := range posts {
		category, err := service.catRepo.GetCategoryById(post.CategoryID)
		if err != nil {
			return nil, err
		}
		postResponseList = append(postResponseList, PostResponse{
			Title:       post.Title,
			Description: post.Description,
			PricePerDay: post.PricePerDay,
			Address:     post.Address,
			Category:    category.Name,
		})

	}
	return &postResponseList, nil
}

func (service *PostService) GetPostByID(postId string) (*PostResponseWithOwner, error) {
	id, err := strconv.ParseUint(postId, 10, 32)
	if err != nil {
		return nil, err
	}
	retrieveedPost, err := service.repo.GetPostByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	category, err := service.catRepo.GetCategoryById(retrieveedPost.CategoryID)
	if err != nil {
		return nil, err
	}

	return &PostResponseWithOwner{
		Title:       retrieveedPost.Title,
		Description: retrieveedPost.Description,
		PricePerDay: retrieveedPost.PricePerDay,
		Address:     retrieveedPost.Address,
		Category:    category.Name,
		OwnerId:     retrieveedPost.OwnerId,
	}, nil
}

func (service *PostService) GetPostsByOwnerId(userId uint, pageStr string) ([]PostResponse, error) {
	var postResponseList []PostResponse
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	size := 10

	offset := (page - 1) * size
	posts, err := service.repo.GetPostsByOwnerId(userId, offset, size)
	if err != nil {
		return nil, err
	}

	for _, post := range posts {
		category, err := service.catRepo.GetCategoryById(post.CategoryID)
		if err != nil {
			return nil, err
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

func (service *PostService) UpdatePost(userId uint, postIdStr string, updatedPost struct {
	Title       string  `json:"title" validate:"required"`
	Description string  `json:"description" validate:"required"`
	PricePerDay float64 `json:"pricePerDay" validate:"required"`
	Address     string  `json:"address" validate:"required"`
	Category    string  `json:"category" validate:"required"`
	IsActive    bool    `json:"isActive" validate:"required"`
}) error {
	postId, err := strconv.ParseUint(postIdStr, 10, 32)
	if err != nil {
		return err
	}
	post, err := service.repo.GetPostByID(uint(postId))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPostNotFound
		}
		return err
	}

	if userId != post.OwnerId {
		return ErrForbidden
	}

	var categoryId *uint
	if updatedPost.Category != "" {
		category, err := service.catRepo.GetCategoryByName(updatedPost.Category)
		if err != nil {
			return err
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
		return err
	}
	return nil
}

func (service *PostService) DeletePost(postIdStr string, userId uint) error {
	postId, err := strconv.ParseUint(postIdStr, 10, 32)
	if err != nil {
		return err
	}
	post, err := service.repo.GetPostByID(uint(postId))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPostNotFound
		}
		return err

	}

	if post.OwnerId != userId {
		return ErrForbidden
	}

	err = service.repo.DeletePost(uint(postId))
	if err != nil {
		return err
	}
	return nil
}
