package post

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type PostHandler struct {
	service  *PostService
	logger   *zap.Logger
	validate *validator.Validate
}

func NewPostHandler(service *PostService, logger *zap.Logger, validate *validator.Validate) *PostHandler {
	return &PostHandler{service: service, logger: logger, validate: validate}
}

func (handler *PostHandler) CreatePost(c echo.Context) error {
	userId, ok := c.Get("userId").(uint)
	if !ok {
		handler.logger.Error("failed to get userId from context")
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var newPost struct {
		Title       string  `json:"title" validate:"required"`
		Description string  `json:"description" validate:"required"`
		PricePerDay float64 `json:"pricePerDay" validate:"required"`
		Address     string  `json:"address" validate:"required"`
		Category    string  `json:"category" validate:"required"`
	}
	if err := c.Bind(&newPost); err != nil {
		handler.logger.Error("failed to bind request", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	if err := handler.validate.Struct(newPost); err != nil {
		handler.logger.Error("provided data is invalid", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid data")
	}

	postId, err := handler.service.CreatePost(userId, newPost)
	if err != nil {
		handler.logger.Error("Error creating post", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create post")
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"post_id": postId,
	})
}

func (handler *PostHandler) GetAllPosts(c echo.Context) error {
	category := c.Param("category")
	title := c.QueryParam("title")
	priceStr := c.QueryParam("price")

	var minPrice, maxPrice *int

	if priceStr != "" {
		price := strings.Split(priceStr, "-")
		if price[0] != "" {
			min, err := strconv.Atoi(price[0])
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "Invalid maximum price")

			}
			minPrice = &min
		}
		if price[1] != "" {
			max, err := strconv.Atoi(price[1])
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "Invalid maximum price")
			}
			maxPrice = &max
		}
	}

	pageStr := c.QueryParam("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	size := 10

	posts, err := handler.service.GetAllPosts(category, title, minPrice, maxPrice, page, size)
	if err != nil {
		handler.logger.Error("error getting posts", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch posts")
	}

	return c.JSON(http.StatusOK, posts)

}

func (handler *PostHandler) GetPostByID(c echo.Context) error {
	postId := c.Param("postId")
	if postId == "" {
		handler.logger.Error("missed postId")
		return echo.NewHTTPError(http.StatusBadRequest, "Post ID is required")
	}

	id, err := strconv.ParseUint(postId, 10, 32)
	if err != nil {
		handler.logger.Error("invalid postId", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Post ID")
	}

	post, err := handler.service.GetPostByID(uint(id))
	if err != nil {
		handler.logger.Error("error retrieving post", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed fo get post")
	}
	return c.JSON(http.StatusOK, post)

}

func (handler *PostHandler) GetPostsByOwnerId(c echo.Context) error {
	userId, ok := c.Get("userId").(uint)
	if !ok {
		handler.logger.Error("failed to get userId from context")
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	pageStr := c.QueryParam("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	size := 10

	posts, err := handler.service.GetPostsByOwnerId(userId, page, size)
	if err != nil {
		handler.logger.Error("error retrieving post", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "faile to retrieve posts")
	}

	if len(posts) == 0 {
		return c.JSON(http.StatusOK, map[string]string{"message": "You have no posts"})
	}

	return c.JSON(http.StatusOK, posts)
}

func (handler *PostHandler) UpdatePost(c echo.Context) error {
	userId, ok := c.Get("userId").(uint)
	if !ok {
		handler.logger.Error("failed to get userId from context")
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	postIdStr := c.Param("postId")
	if postIdStr == "" {
		handler.logger.Error("missed postId")
		return echo.NewHTTPError(http.StatusBadRequest, "Post ID is required")
	}
	postId, err := strconv.ParseUint(postIdStr, 10, 32)
	if err != nil {
		handler.logger.Error("invalid postId", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Post ID")
	}

	var updatedPost struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		PricePerDay float64 `json:"pricePerDay"`
		Address     string  `json:"address"`
		Category    string  `json:"category"`
		IsActive    bool    `json:"isActive"`
	}
	if err := c.Bind(&updatedPost); err != nil {
		handler.logger.Error("failed to bind request", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	err = handler.validate.Struct(updatedPost)
	if err != nil {
		handler.logger.Error("provided data is invalid", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid data")
	}

	err = handler.service.UpdatePost(userId, uint(postId), updatedPost)
	if err != nil {
		if errors.Is(err, echo.NewHTTPError(http.StatusForbidden, "")) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "You are not authorized to update this post."})
		}
		handler.logger.Error("error updating post")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update post")
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Post updated successfully"})
}

func (handler *PostHandler) DeletePost(c echo.Context) error {
	userId, ok := c.Get("userId").(uint)
	if !ok {
		handler.logger.Error("failed to get userId from context")
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	postIdStr := c.Param("postId")
	if postIdStr == "" {
		handler.logger.Error("missed postId")
		return echo.NewHTTPError(http.StatusBadRequest, "Post ID is required")
	}
	postId, err := strconv.ParseUint(postIdStr, 10, 32)
	if err != nil {
		handler.logger.Error("invalid postId", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Post ID")
	}

	err = handler.service.DeletePost(uint(postId), userId)
	if err != nil {
		if errors.Is(err, echo.NewHTTPError(http.StatusForbidden)) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "You are not authorized to delete this post."})
		}
		handler.logger.Error("error deleting post")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete post")
	}
	return c.JSON(http.StatusNoContent, map[string]string{"message": "Post deleted successfully"})
}
