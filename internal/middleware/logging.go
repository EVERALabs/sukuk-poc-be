package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kadzu/sukuk-poc-be/internal/logger"
	"github.com/sirupsen/logrus"
)

// RequestLogger logs all HTTP requests with timing and response information
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Log request
		requestLogger := logger.WithFields(logrus.Fields{
			"method":     method,
			"path":       path,
			"query":      c.Request.URL.RawQuery,
			"user_agent": c.Request.UserAgent(),
			"client_ip":  c.ClientIP(),
			"referer":    c.Request.Referer(),
		})

		// Read and restore request body for POST/PUT requests (for debugging)
		var requestBody string
		if method == "POST" || method == "PUT" || method == "PATCH" {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				requestBody = string(bodyBytes)
				// Restore the body for the actual handler
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				
				// Only log body for non-file uploads and keep it reasonably short
				if len(requestBody) < 1000 && !isFileUpload(c) {
					requestLogger = requestLogger.WithField("request_body", requestBody)
				} else if isFileUpload(c) {
					requestLogger = requestLogger.WithField("request_type", "file_upload")
				}
			}
		}

		requestLogger.Info("Request started")

		// Process request
		c.Next()

		// Calculate response time
		duration := time.Since(start)
		status := c.Writer.Status()

		// Determine log level based on status code
		responseLogger := logger.WithFields(logrus.Fields{
			"method":        method,
			"path":          path,
			"status":        status,
			"duration_ms":   duration.Milliseconds(),
			"response_size": c.Writer.Size(),
		})

		// Add error information if present
		if len(c.Errors) > 0 {
			responseLogger = responseLogger.WithField("errors", c.Errors.String())
		}

		// Log based on status code
		switch {
		case status >= 500:
			responseLogger.Error("Request completed with server error")
		case status >= 400:
			responseLogger.Warn("Request completed with client error")
		case status >= 300:
			responseLogger.Info("Request completed with redirect")
		default:
			responseLogger.Info("Request completed successfully")
		}

		// Log slow requests
		if duration > 1*time.Second {
			logger.WithFields(logrus.Fields{
				"method":      method,
				"path":        path,
				"duration_ms": duration.Milliseconds(),
			}).Warn("Slow request detected")
		}
	}
}

// isFileUpload checks if the request is a file upload
func isFileUpload(c *gin.Context) bool {
	contentType := c.Request.Header.Get("Content-Type")
	return contentType != "" && (contentType == "multipart/form-data" || 
		contentType[:19] == "multipart/form-data")
}

// ErrorLogger logs errors that occur during request processing
func ErrorLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Log any errors that occurred
		for _, err := range c.Errors {
			logger.WithFields(logrus.Fields{
				"method": c.Request.Method,
				"path":   c.Request.URL.Path,
				"error":  err.Error(),
				"type":   err.Type,
			}).Error("Request error occurred")
		}
	}
}

// DatabaseLogger logs database-related operations
func DatabaseLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		
		// Check if this was a database-heavy operation
		duration := time.Since(start)
		if duration > 500*time.Millisecond {
			logger.WithFields(logrus.Fields{
				"method":      c.Request.Method,
				"path":        c.Request.URL.Path,
				"duration_ms": duration.Milliseconds(),
			}).Warn("Database operation took longer than expected")
		}
	}
}