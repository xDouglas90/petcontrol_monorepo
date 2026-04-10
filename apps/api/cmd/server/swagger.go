package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	docs "github.com/xdouglas90/petcontrol_monorepo/docs"
)

func configureSwaggerInfo() {
	docs.SwaggerInfo.Title = "PetControl API"
	docs.SwaggerInfo.Description = "PetControl multi-tenant API"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.BasePath = "/api/v1"
}

func registerSwaggerRoute(router *gin.Engine) {
	configureSwaggerInfo()
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/api/v1/docs", redirectSwaggerAlias)
	router.GET("/api/v1/docs/*any", redirectSwaggerAlias)
}

func redirectSwaggerAlias(c *gin.Context) {
	target := "/swagger/index.html"
	if path := strings.TrimPrefix(c.Param("any"), "/"); path != "" {
		target = "/swagger/" + path
	}

	c.Redirect(http.StatusTemporaryRedirect, target)
}
