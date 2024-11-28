package post

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type PostHandler struct {
	service  *PostService
	validate *validator.Validate
}

func NewPostHandler(service *PostService, logger *zap.Logger, validate *validator.Validate) *PostHandler {
	return &PostHandler{service: service, validate: validate}
}

type PostDto struct {
	Title       string  `json:"title" validate:"required"`
	Description string  `json:"description" validate:"required"`
	PricePerDay float64 `json:"pricePerDay" validate:"required"`
	Address     string  `json:"address" validate:"required"`
	Category    string  `json:"category" validate:"required"`
}

func (handler *PostHandler) CreatePost(c echo.Context) error {
	var newPost PostDto
	userId, ok := c.Get("userId").(uint)
	if !ok {
		zap.L().Error("failed to get userId from context")
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	if err := c.Bind(&newPost); err != nil {
		zap.L().Error("failed to bind request", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	if err := handler.validate.Struct(newPost); err != nil {
		zap.L().Error("provided data is invalid", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid data")
	}

	postId, err := handler.service.CreatePost(userId, newPost)
	if err != nil {
		zap.L().Error("Error creating post", zap.Error(err))
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
	pageStr := c.QueryParam("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	posts, err := handler.service.GetAllPosts(category, title, priceStr, page)
	if err != nil {
		zap.L().Error("error getting posts", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch posts")
	}

	return c.JSON(http.StatusOK, posts)

}

func (handler *PostHandler) GetPostByID(c echo.Context) error {
	postId := c.Param("postId")
	if postId == "" {
		zap.L().Error("missed postId")
		return echo.NewHTTPError(http.StatusBadRequest, "Post ID is required")
	}

	post, err := handler.service.GetPostByID(postId)
	if err != nil {
		zap.L().Error("error retrieving post", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "failed fo get post")
	}

	return c.JSON(http.StatusOK, post)
}

func (handler *PostHandler) GetPostsByOwnerId(c echo.Context) error {
	userId, ok := c.Get("userId").(uint)
	if !ok {
		zap.L().Error("failed to get userId from context")
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	pageStr := c.QueryParam("page")

	posts, err := handler.service.GetPostsByOwnerId(userId, pageStr)
	if err != nil {
		zap.L().Error("error retrieving post", zap.Error(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "faile to retrieve posts")
	}

	return c.JSON(http.StatusOK, posts)
}

func (handler *PostHandler) UpdatePost(c echo.Context) error {
	userId, ok := c.Get("userId").(uint)
	if !ok {
		zap.L().Error("failed to get userId from context")
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	postIdStr := c.Param("postId")
	if postIdStr == "" {
		zap.L().Error("missed postId")
		return echo.NewHTTPError(http.StatusBadRequest, "Post ID is required")
	}

	var updatedPost struct {
		Title       string  `json:"title" validate:"required"`
		Description string  `json:"description" validate:"required"`
		PricePerDay float64 `json:"pricePerDay" validate:"required"`
		Address     string  `json:"address" validate:"required"`
		Category    string  `json:"category" validate:"required"`
		IsActive    bool    `json:"isActive" validate:"required"`
	}
	if err := c.Bind(&updatedPost); err != nil {
		zap.L().Error("failed to bind request", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	err := handler.validate.Struct(updatedPost)
	if err != nil {
		zap.L().Error("provided data is invalid", zap.Error(err))
		return echo.NewHTTPError(http.StatusBadRequest, "invalid data")
	}

	err = handler.service.UpdatePost(userId, postIdStr, updatedPost)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "You are not authorized to update this post."})
		} else if errors.Is(err, ErrPostNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "this post not exist"})
		}
		zap.L().Error("error updating post")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update post")
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Post updated successfully"})
}

func (handler *PostHandler) DeletePost(c echo.Context) error {
	userId, ok := c.Get("userId").(uint)
	if !ok {
		zap.L().Error("failed to get userId from context")
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	postIdStr := c.Param("postId")
	if postIdStr == "" {
		zap.L().Error("missed postId")
		return echo.NewHTTPError(http.StatusBadRequest, "Post ID is required")
	}

	err := handler.service.DeletePost(postIdStr, userId)
	if err != nil {
		if errors.Is(err, echo.NewHTTPError(http.StatusForbidden)) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "You are not authorized to delete this post."})
		}
		zap.L().Error("error deleting post")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete post")
	}
	return c.JSON(http.StatusNoContent, map[string]string{"message": "Post deleted successfully"})
}
