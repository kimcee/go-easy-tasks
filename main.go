package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

const (
	DBUser     = "root"
	DBPassword = ""
	DBHost     = "127.0.0.1"
	DBPort     = "3306"
	DBName     = "vereaze"
	ParseTime  = "true"
)

type EasyTask struct {
	ID          uint   `gorm:"primary_key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	TaskAction  string `json:"task_action" gorm:"column:task_action"`
}

type EasyTaskRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	TaskAction  string `json:"task_action" binding:"required"`
}

func main() {
	dbConnectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=%s", DBUser, DBPassword, DBHost, DBPort, DBName, ParseTime)
	db, err := gorm.Open("mysql", dbConnectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.AutoMigrate(&EasyTask{})

	r := gin.Default()

	// static home
	r.GET("/", loadPage())

	// api endpoints
	r.GET("/tasks", getTasks(db))
	r.GET("/tasks/:id", getTaskByID(db))
	r.POST("/tasks", createTask(db))
	r.PUT("/tasks/:id", updateTask(db))
	r.DELETE("/tasks/:id", deleteTask(db))

	// server
	r.Run(":8080")
}

func loadPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.File("frontend/index.html")
	}
}

func getTasks(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tasks []EasyTask

		if err := db.Table("easy_tasks AS tasks").
			Select("*").
			Where("tasks.name IS NOT NULL AND tasks.name <> ''").
			Order("tasks.id DESC").
			Scan(&tasks).Error; err != nil {

			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, tasks)
	}
}

func getTaskByID(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var easyTask EasyTask
		id := c.Param("id")

		if err := db.Where("easy_tasks.id = ?", id).First(&easyTask).Error; err != nil {

			c.JSON(404, gin.H{"error": fmt.Sprintf("Error: Task %v not found", id)})
			return
		}

		c.JSON(200, easyTask)
	}
}

func createTask(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req EasyTaskRequest
		if err := c.ShouldBindJSON(&req); err != nil {

			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		easyTask := EasyTask{
			Name:        req.Name,
			Description: req.Description,
			TaskAction:  req.TaskAction,
		}

		db = db.Debug()

		if err := db.Create(&easyTask).Error; err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
			return
		}

		c.JSON(http.StatusCreated, easyTask)
	}
}

func updateTask(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		db = db.Debug()

		var easyTask EasyTask
		if err := db.First(&easyTask, id).Error; err != nil {

			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}

		var req EasyTaskRequest
		if err := c.ShouldBindJSON(&req); err != nil {

			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		easyTask.Name = req.Name
		easyTask.Description = req.Description
		easyTask.TaskAction = req.TaskAction

		if err := db.Save(&easyTask).Error; err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
			return
		}

		c.JSON(http.StatusOK, easyTask)
	}
}

func deleteTask(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var easyTask EasyTask
		id := c.Param("id")

		if err := db.Where("id = ?", id).Delete(&easyTask).Error; err != nil {

			c.JSON(404, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(204, nil)
	}
}
