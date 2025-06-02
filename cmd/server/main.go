package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Zipklas/anime-site-backend/internal/comment"
	"github.com/Zipklas/anime-site-backend/internal/kodik"
	"github.com/Zipklas/anime-site-backend/internal/user"
	"github.com/Zipklas/anime-site-backend/pkg/database"

	"github.com/Zipklas/anime-site-backend/internal/shikimori"

	"github.com/joho/godotenv"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	_ = godotenv.Load()

	db := database.InitPostgres()

	shikimoriService := shikimori.NewService()
	shikimoriHandler := shikimori.NewHandler(shikimoriService)
	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo, shikimoriService)
	userHandler := user.NewHandler(userService)

	e := echo.New()

	e.GET("/kodik.txt", func(c echo.Context) error {
		return c.File("kodik.txt")
	})
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.POST("/register", userHandler.Register)
	e.POST("/login", userHandler.Login)
	e.POST("/api/shikimori/search", shikimoriHandler.SearchAnime)
	e.GET("/api/shikimori/top", shikimoriHandler.GetTopAnime)
	e.GET("/api/shikimori/anime/:id", shikimoriHandler.GetAnimeByID)
	e.GET("/api/shikimori/new", shikimoriHandler.GetNewReleases)

	commentRepo := comment.NewRepository(db)
	commentService := comment.NewService(commentRepo)
	commentHandler := comment.NewHandler(commentService)

	commentGroup := e.Group("/api/comments")
	commentGroup.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey: []byte(os.Getenv("JWT_SECRET")),
	}))

	commentGroup.POST("/:anime_id", commentHandler.CreateComment)
	commentGroup.GET("/:anime_id", commentHandler.GetComments)
	commentGroup.DELETE("/:comment_id", commentHandler.DeleteComment)
	commentGroup.PUT("/:comment_id", commentHandler.UpdateComment)

	commentGroup.PUT("/:comment_id/vote", commentHandler.VoteComment)
	commentGroup.DELETE("/:comment_id/vote", commentHandler.RemoveVote)

	kodikService := kodik.NewService("None")
	kodikHandler := kodik.NewHandler(kodikService)

	e.GET("/api/kodik/search", kodikHandler.SearchVideos)
	e.GET("/api/kodik/videos/:shikimori_id", kodikHandler.GetVideoOptions)

	playerGroup := e.Group("/player")
	playerGroup.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey: []byte(os.Getenv("JWT_SECRET")),
	}))
	playerGroup.GET("/:video_id", func(c echo.Context) error {

		return c.JSON(http.StatusOK, echo.Map{"status": "under construction"})
	})

	r := e.Group("/profile")
	r.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey: []byte(os.Getenv("JWT_SECRET")),
		ErrorHandler: func(c echo.Context, err error) error {
			log.Printf("Error validating JWT token: %v", err)
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		},
	}))

	r.GET("", userHandler.Profile)
	r.POST("/watched/:anime_id", userHandler.AddWatched)
	r.POST("/favorite/:anime_id", userHandler.AddFavorite)
	r.GET("/watched", userHandler.GetWatchedAnime)
	r.GET("/favorite", userHandler.GetFavouriteAnime)

	log.Fatal(e.Start(":8080"))
}
