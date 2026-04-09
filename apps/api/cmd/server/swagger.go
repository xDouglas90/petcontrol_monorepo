package main

import (
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
}
