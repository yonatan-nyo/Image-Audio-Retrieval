package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetAllHandler[T any](db *gorm.DB, modelSlice *[]T, fields []string, relations ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := db.Order("id DESC")

		for _, relation := range relations {
			query = query.Preload(relation)
		}

		if len(fields) > 0 {
			query = query.Select(fields)
		}

		if err := query.Find(modelSlice).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve records"})
			return
		}
		c.JSON(http.StatusOK, modelSlice)
	}
}

func GetAllHandlerWithPagination[T any](db *gorm.DB, modelSlice *[]T, fields []string, relations ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var page, pageSize int
		var err error

		// Get pagination parameters from the request query, with defaults
		if page, err = strconv.Atoi(c.DefaultQuery("page", "1")); err != nil || page < 1 {
			page = 1
		}
		if pageSize, err = strconv.Atoi(c.DefaultQuery("page_size", "10")); err != nil || pageSize < 1 {
			pageSize = 10
		}

		// Enforce maximum page size of 10
		if pageSize > 10 {
			pageSize = 10
		}

		offset := (page - 1) * pageSize

		query := db.Order("id DESC")
		for _, relation := range relations {
			query = query.Preload(relation)
		}

		if len(fields) > 0 {
			query = query.Select(fields)
		}

		if err := query.Limit(pageSize).Offset(offset).Find(modelSlice).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve records"})
			return
		}

		// Optional: Add metadata for pagination
		var totalRecords int64
		db.Model(modelSlice).Count(&totalRecords)

		c.JSON(http.StatusOK, gin.H{
			"data":          modelSlice,
			"current_page":  page,
			"page_size":     pageSize,
			"total_records": totalRecords,
			"total_pages":   (totalRecords + int64(pageSize) - 1) / int64(pageSize),
		})
	}
}

func GetOneHandler[T any](db *gorm.DB, modelInstance *T, fields []string, relations ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		fmt.Println("Request ID:", id)

		newModelInstance := reflect.New(reflect.TypeOf(modelInstance).Elem()).Interface().(*T)

		query := db.Model(newModelInstance)
		for _, relation := range relations {
			query = query.Preload(relation)
		}

		if len(fields) > 0 {
			query = query.Select(fields)
		}

		if err := query.Take(newModelInstance, "id = ?", id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve record"})
			}
			return
		}

		c.JSON(http.StatusOK, newModelInstance)
	}
}

// DeleteOneHandler handles deleting a record and its associated files.
func DeleteOneHandler[T any](db *gorm.DB, fileFields ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")

		// Convert id to integer
		id, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		// Create a new instance of the model
		modelInstance := new(T)

		// Retrieve the record
		if err := db.First(modelInstance, "id = ?", id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve record"})
			}
			return
		}

		// Extract file paths and delete files
		for _, field := range fileFields {
			// Use reflection to get the field value
			fieldValue := reflect.ValueOf(modelInstance).Elem().FieldByName(field)

			var filePath string

			switch fieldValue.Kind() {
			case reflect.String:
				filePath = fieldValue.String()
			case reflect.Ptr:
				if fieldValue.IsNil() {
					continue
				}
				strValue := fieldValue.Elem().String()
				filePath = strValue
			default:
				continue
			}

			if filePath != "" {
				// Assuming the file paths are relative to "public/uploads/"
				fullPath := filepath.Join("", filePath)

				// Check if the file exists
				if _, err := os.Stat(fullPath); err == nil {
					// Attempt to delete the file
					if err := os.Remove(fullPath); err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete file %s", fullPath)})
						return
					}
				} else if !os.IsNotExist(err) {
					c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error checking file %s", fullPath)})
					return
				}
			}
		}

		// Delete the record
		if err := db.Delete(modelInstance, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete record"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Record and associated files deleted successfully"})
	}
}
