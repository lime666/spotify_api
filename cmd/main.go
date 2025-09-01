package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/lime666/spotify_api/config"
	"github.com/lime666/spotify_api/pkg"

	"github.com/gin-gonic/gin"
	_ "github.com/lime666/spotify_api/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//	@title		Spotify API
//	@version	0.0.1

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath	/api
// @host      localhost:50000
// @schemes	http

// @externalDocs.description	OpenAPI
// @externalDocs.url			https://swagger.io/resources/open-api/
func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	oauth := pkg.NewOAuthService(cfg)

	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Authorization", "Content-Type"}
	r.Use(cors.New(config))

	api := r.Group("/api")
	{
		api.GET("/login/spotify", oauth.LoginHandler)
		api.GET("/callback/spotify", oauth.CallbackHandler)
		api.GET("/analyze", oauth.SpotifyClientMiddleware(), oauth.AnalyzeHandler)
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Println("listening on %s", cfg.Port)
	if err := r.Run(cfg.Port); err != nil {
		log.Fatal(err)
	}
}
