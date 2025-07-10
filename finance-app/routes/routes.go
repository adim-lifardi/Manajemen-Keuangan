package routes

import (
	"finance-app/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// Endpoint transaksi
	r.GET("/transactions", controllers.GetTransactions)
	r.POST("/transactions", controllers.CreateTransaction)
	r.PUT("/transactions/:id", controllers.UpdateTransaction)
	r.DELETE("/transactions/:id", controllers.DeleteTransaction)
	r.GET("/transactions/:id", controllers.GetTransactionByID)

	r.GET("/export/pdf", controllers.ExportPDF)
	r.GET("/export/excel", controllers.ExportExcel)

	r.GET("/summary", controllers.GetSummary)
	r.GET("/chart", controllers.GetChartPerTipe)

	// Route to show login page
	r.GET("/login", controllers.ShowLoginPage)

	// Route to handle login
	r.POST("/login", controllers.HandleLogin)

	// Endpoint root untuk halaman HTML
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	// Opsional: redirect
	r.GET("/index", func(c *gin.Context) {
		c.Redirect(302, "/")
	})
}
