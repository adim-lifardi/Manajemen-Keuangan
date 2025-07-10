package controllers

import (
	"finance-app/config"
	"finance-app/models"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
)

func GetTransactions(c *gin.Context) {
	month := c.Query("month")
	tipe := c.Query("type")

	var transactions []models.Transaction
	query := config.DB.Preload("Category")

	if month != "" {
		query = query.Where("DATE_FORMAT(date, '%Y-%m') = ?", month)
	}

	if tipe != "" {
		query = query.Joins("JOIN categories ON categories.id = transactions.category_id").
			Where("categories.type = ?", tipe)
	}

	query.Find(&transactions)
	c.JSON(http.StatusOK, transactions)
}

func CreateTransaction(c *gin.Context) {
	var transaction models.Transaction
	if err := c.ShouldBindJSON(&transaction); err != nil {
		fmt.Println("Transaksi Error:", err)
		fmt.Printf("Body: %+v\n", transaction)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.DB.Create(&transaction)
	c.JSON(http.StatusOK, transaction)
}

func UpdateTransaction(c *gin.Context) {
	id := c.Param("id")
	var transaction models.Transaction

	if err := config.DB.First(&transaction, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaksi tidak ditemukan"})
		return
	}

	var input models.Transaction
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transaction.Date = input.Date
	transaction.Description = input.Description
	transaction.Amount = input.Amount
	transaction.CategoryID = input.CategoryID
	transaction.UserID = input.UserID

	config.DB.Save(&transaction)
	c.JSON(http.StatusOK, transaction)
}

func DeleteTransaction(c *gin.Context) {
	id := c.Param("id")
	config.DB.Delete(&models.Transaction{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "Berhasil dihapus"})
}

func GetTransactionByID(c *gin.Context) {
	id := c.Param("id")
	var transaction models.Transaction

	if err := config.DB.Preload("Category").First(&transaction, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaksi tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, transaction)
}

func ExportPDF(c *gin.Context) {
	month := c.Query("month")
	tipe := c.Query("type")

	var transactions []models.Transaction
	db := config.DB.Preload("Category")

	if month != "" {
		db = db.Where("DATE_FORMAT(date, '%Y-%m') = ?", month)
	}
	if tipe != "" {
		db = db.Joins("JOIN categories ON categories.id = transactions.category_id").
			Where("categories.type = ?", tipe)
	}

	db.Find(&transactions)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, "Laporan Transaksi")

	for _, tx := range transactions {
		pdf.Ln(10)
		pdf.Cell(40, 10, tx.Date)
		pdf.Cell(40, 10, tx.Category.Name)
		pdf.Cell(40, 10, tx.Description)
		pdf.Cell(40, 10, fmt.Sprintf("Rp %.2f", tx.Amount))
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=laporan.pdf")
	err := pdf.Output(c.Writer)
	if err != nil {
		c.JSON(500, gin.H{"error": "Gagal generate PDF"})
	}
}

func ExportExcel(c *gin.Context) {
	month := c.Query("month")
	tipe := c.Query("type")

	var transactions []models.Transaction
	db := config.DB.Preload("Category")

	if month != "" {
		db = db.Where("DATE_FORMAT(date, '%Y-%m') = ?", month)
	}
	if tipe != "" {
		db = db.Joins("JOIN categories ON categories.id = transactions.category_id").
			Where("categories.type = ?", tipe)
	}

	db.Find(&transactions)

	f := excelize.NewFile()
	sheet := "Laporan"
	f.NewSheet(sheet)
	sheetIdx, err := f.GetSheetIndex(sheet)
	if err == nil {
		f.SetActiveSheet(sheetIdx)
	}

	headers := map[string]string{
		"A1": "Tanggal",
		"B1": "Kategori",
		"C1": "Deskripsi",
		"D1": "Jumlah",
	}
	for cell, text := range headers {
		f.SetCellValue(sheet, cell, text)
	}

	for i, tx := range transactions {
		row := strconv.Itoa(i + 2)
		f.SetCellValue(sheet, "A"+row, tx.Date)
		f.SetCellValue(sheet, "B"+row, tx.Category.Name)
		f.SetCellValue(sheet, "C"+row, tx.Description)
		f.SetCellValue(sheet, "D"+row, tx.Amount)
	}

	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=laporan.xlsx")
	if err := f.Write(c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal generate Excel"})
	}
}

func GetSummary(c *gin.Context) {
	var totalIncome float64
	var totalExpense float64

	// Hitung total pemasukan
	config.DB.Table("transactions").
		Joins("JOIN categories ON categories.id = transactions.category_id").
		Where("categories.type = ?", "income").
		Select("SUM(transactions.amount)").Scan(&totalIncome)

	// Hitung total pengeluaran
	config.DB.Table("transactions").
		Joins("JOIN categories ON categories.id = transactions.category_id").
		Where("categories.type = ?", "expense").
		Select("SUM(transactions.amount)").Scan(&totalExpense)

	c.JSON(http.StatusOK, gin.H{
		"income":  totalIncome,
		"expense": totalExpense,
	})
}

func GetChartPerTipe(c *gin.Context) {
	type Row struct {
		Month string
		Type  string
		Total float64
	}
	var rows []Row

	config.DB.Raw(`
		SELECT DATE_FORMAT(t.date, '%Y-%m') as month, c.type, SUM(t.amount) as total
		FROM transactions t
		JOIN categories c ON c.id = t.category_id
		GROUP BY month, c.type
		ORDER BY month
	`).Scan(&rows)

	// map untuk menyimpan hasil
	result := map[string]map[string]float64{}
	for _, row := range rows {
		if _, ok := result[row.Month]; !ok {
			result[row.Month] = map[string]float64{}
		}
		result[row.Month][row.Type] = row.Total
	}

	// format hasil
	var final []gin.H
	for month, data := range result {
		final = append(final, gin.H{
			"month":   month,
			"income":  data["income"],
			"expense": data["expense"],
		})
	}

	c.JSON(http.StatusOK, final)
}

func ShowLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

func HandleLogin(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format login tidak valid"})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email tidak ditemukan"})
		return
	}

	if input.Password != user.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Password salah"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login berhasil",
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
		},
	})
}

func Register(c *gin.Context) {
	var input models.User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid"})
		return
	}

	var existing models.User
	if err := config.DB.Where("email = ?", input.Email).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email sudah digunakan"})
		return
	}

	if err := config.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan pengguna"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Registrasi berhasil",
		"user": gin.H{
			"id":    input.ID,
			"name":  input.Name,
			"email": input.Email,
		},
	})
}
