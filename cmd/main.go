package main

import (
	"log"
	"post-service/auth"
	"post-service/category"
	"post-service/post"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDB() (*gorm.DB, error) {
	dsn := "host=localhost user=admin password=sahar223010 dbname=rental_service_db search_path=post-service port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

// func NewLogger() (*zap.Logger, error) {
// 	logger, err := zap.NewProduction()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return logger, nil
// }

func NewValidator() *validator.Validate {
	return validator.New()
}

func RegisterRoutes(e *echo.Echo, postHandler *post.PostHandler, categoryHandler *category.CategoryHandler) {
	e.GET("/posts", postHandler.GetAllPosts)
	e.GET("/posts/category/:category", postHandler.GetAllPosts)
	e.GET("/posts/:postId", postHandler.GetPostByID)

	e.GET("/categories", categoryHandler.GetAllCategories)
	e.GET("/categories/:categoryId", categoryHandler.GetCategoryById)

	e.GET("/my-posts", postHandler.GetPostsByOwnerId, auth.AuthMiddleware)

	postGroup := e.Group("/posts")
	postGroup.Use(auth.AuthMiddleware)
	postGroup.POST("", postHandler.CreatePost)
	postGroup.PUT("/:postId", postHandler.UpdatePost)
	postGroup.DELETE("/:postId", postHandler.DeletePost)

	categoryGroup := e.Group("/categories")
	categoryGroup.Use(auth.AuthMiddleware)
	categoryGroup.POST("", categoryHandler.CreateCategory)
	categoryGroup.PUT("/:categoryId", categoryHandler.UpdateCategory)

}

func main() {
	e := echo.New()

	app := fx.New(
		fx.Provide(
			NewDB,
			//NewLogger,
			NewValidator,
			category.NewCategoryRepository,
			category.NewCategoryService,
			category.NewCategoryHandler,
			post.NewPostRepository,
			post.NewPostService,
			post.NewPostHandler,
			func() *echo.Echo { return e },
		),
		fx.Invoke(
			func(e *echo.Echo, postHandler *post.PostHandler, categoryHandler *category.CategoryHandler) {
				RegisterRoutes(e, postHandler, categoryHandler)
			},
			func() {
				if err := e.Start(":8081"); err != nil {
					log.Fatal("Echo server failed to start", zap.Error(err))
				}
			},
		),
	)
	app.Run()
}
