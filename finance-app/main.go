package main

import (
	"finance-app/config"
	"finance-app/controllers"
	"finance-app/models"
	"finance-app/routes"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.LoadHTMLGlob("template/*")
	r.GET("/index.html", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	r.POST("/register", controllers.Register)
	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", nil)
	})

	config.ConnectDB()

	config.DB.AutoMigrate(&models.User{}, &models.Category{}, &models.Transaction{})

	routes.SetupRoutes(r)

	r.Run(":8080")
}
