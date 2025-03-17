package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/byuoitav/central-event-system/hub/base"
	"github.com/byuoitav/central-event-system/messenger"
	"github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/teams-help-request/goteamsnotification"
	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
)

// Function to set log level from string
func setLogLevel(levelStr string, level *slog.LevelVar) error {
	switch strings.ToLower(levelStr) {
	case "debug":
		level.Set(slog.LevelDebug)
	case "info":
		level.Set(slog.LevelInfo)
	case "warn", "warning":
		level.Set(slog.LevelWarn)
	case "error":
		level.Set(slog.LevelError)
	default:
		return fmt.Errorf("unknown log level: %s", levelStr)
	}
	return nil
}

// Returns the current log level as a string
func getLogLevelString(level slog.Level) string {
	switch level {
	case slog.LevelDebug:
		return "debug"
	case slog.LevelInfo:
		return "info"
	case slog.LevelWarn:
		return "warn"
	case slog.LevelError:
		return "error"
	default:
		return "unknown"
	}
}

// getEnvOrFlag returns the environment variable if set, otherwise the flag value
func getEnvOrFlag(envName string, flagValue string) string {
	if val, ok := os.LookupEnv(envName); ok && val != "" {
		return val
	}
	return flagValue
}

func main() {
	// Define the variables for command line flags
	var (
		logLevelFlag = pflag.StringP("log-level", "L", "", "Level at which the logger operates")
		hubAddress   = pflag.String("hub-address", "", "Address of the event hub")
		webhookUrl   = pflag.String("webhook-url", "", "URL of the webhook to send to")
		smeeUrl      = pflag.String("monitoring-url", "", "URL of the AV Monitoring Service")
		port         = pflag.StringP("port", "P", "8080", "Port to run the web server on")
	)

	// Parse flags
	pflag.Parse()

	// Check for environment variables if flags are not set
	*logLevelFlag = getEnvOrFlag("LOG_LEVEL", *logLevelFlag)
	if *logLevelFlag == "" {
		*logLevelFlag = "info" // default if neither flag nor env var is set
	}

	*hubAddress = getEnvOrFlag("EVENT_HUB_ADDRESS", *hubAddress)
	*webhookUrl = getEnvOrFlag("TEAMS_WEBHOOK_URL", *webhookUrl)
	*smeeUrl = getEnvOrFlag("TEAMS_MONITORING_URL", *smeeUrl)

	// Setup logger with level variable to enable dynamic changes
	logLevel := new(slog.LevelVar)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	// Default to info level
	logLevel.Set(slog.LevelInfo)

	// If running on Windows, set log level to debug by default for development
	// but only if not explicitly set through flag or env var
	if runtime.GOOS == "windows" && *logLevelFlag == "info" {
		*logLevelFlag = "debug"
		logger.Info("running from Windows, logging set to debug")
	}

	// Set the log level from command line argument or environment variable
	if err := setLogLevel(*logLevelFlag, logLevel); err != nil {
		logger.Error("cannot set log level", "error", err)
	}

	// Log configuration information
	logger.Info("starting service with configuration",
		"log_level", *logLevelFlag,
		"hub_address_set", *hubAddress != "",
		"webhook_url_set", *webhookUrl != "",
		"monitoring_url_set", *smeeUrl != "",
		"port", *port)

	// Check required configuration
	if *webhookUrl == "" {
		logger.Error("webhook URL not provided - notifications will not be sent",
			"usage", "Set TEAMS_WEBHOOK_URL environment variable or use --webhook-url flag")
	}

	if *hubAddress == "" {
		logger.Error("hub address not provided - event subscription will fail",
			"usage", "Set EVENT_HUB_ADDRESS environment variable or use --hub-address flag")
	}

	loggerMutex := &sync.Mutex{}

	// Initialize the request manager with our logger
	rm := goteamsnotification.RequestManager{
		Log:           logger,
		MonitoringURL: *smeeUrl,
		WebhookURL:    *webhookUrl,
	}

	// Setup Gin server
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// Add middleware for logging
	router.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		logger.Info("HTTP Request",
			"status", statusCode,
			"method", method,
			"path", path,
			"latency", latency,
		)
	})

	// Health check endpoints
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// New health check endpoint returning "healthy"
	router.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "healthy")
	})

	// Get current log level endpoint
	router.GET("/log_level", func(c *gin.Context) {
		loggerMutex.Lock()
		defer loggerMutex.Unlock()

		currentLevel := getLogLevelString(logLevel.Level())
		c.String(http.StatusOK, currentLevel)
	})

	// Set log level endpoint
	router.GET("/log_level/:log_level", func(c *gin.Context) {
		newLogLevel := c.Param("log_level")

		loggerMutex.Lock()
		defer loggerMutex.Unlock()

		if err := setLogLevel(newLogLevel, logLevel); err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("Invalid log level: %s. Valid options are debug, info, warn, error", newLogLevel))
			return
		}

		currentLevel := getLogLevelString(logLevel.Level())
		c.String(http.StatusOK, fmt.Sprintf("Log level set to %s", currentLevel))
	})

	router.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "ok",
			"service":     "teams-help-request",
			"webhook_url": *webhookUrl != "",
			"hub_address": *hubAddress,
			"monitor_url": *smeeUrl,
		})
	})

	// API endpoints under /api/v1
	api := router.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":         "healthy",
				"timestamp":      time.Now().Format(time.RFC3339),
				"event_listener": "active",
			})
		})

		// Modified endpoint to use path parameters instead of query parameters
		api.GET("/notify/building/:building/room/:room/device/:device", func(c *gin.Context) {
			// Get parameters from path
			building := c.Param("building")
			room := c.Param("room")
			device := c.Param("device")

			// Construct roomID from building and room
			roomID := fmt.Sprintf("%s-%s", building, room)

			// Format the generating system string
			generatingSystem := fmt.Sprintf("%s-%s-%s", building, room, device)

			logger.Info("Manual help request notification triggered",
				"generating_system", generatingSystem,
				"room_id", roomID)

			if err := rm.SendTheMessage(generatingSystem); err != nil {
				logger.Error("Failed to send Teams message", "error", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to send notification: " + err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message":  "Notification sent successfully",
				"room_id":  roomID,
				"building": building,
				"room":     room,
				"device":   device,
			})
		})

		// Keep the GET method with query parameters for backward compatibility
		api.GET("/notify", func(c *gin.Context) {
			// Get parameters from query string
			roomID := c.Query("room_id")
			building := c.Query("building")
			room := c.Query("room")
			device := c.Query("device")

			// Validate required parameters
			if building == "" || room == "" || device == "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Missing required parameters: building, room, and device are required",
				})
				return
			}

			// If roomID is not provided but we have building and room, construct it
			if roomID == "" {
				roomID = fmt.Sprintf("%s-%s", building, room)
			}

			// Format the generating system string
			generatingSystem := fmt.Sprintf("%s-%s-%s", building, room, device)

			logger.Info("Manual help request notification triggered",
				"generating_system", generatingSystem,
				"room_id", roomID)

			if err := rm.SendTheMessage(generatingSystem); err != nil {
				logger.Error("Failed to send Teams message", "error", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to send notification: " + err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message":  "Notification sent successfully",
				"room_id":  roomID,
				"building": building,
				"room":     room,
				"device":   device,
			})
		})

		// Keep the POST method for backward compatibility
		api.POST("/notify", func(c *gin.Context) {
			type NotifyRequest struct {
				RoomID   string `json:"room_id" binding:"required"`
				Building string `json:"building" binding:"required"`
				Room     string `json:"room" binding:"required"`
				Device   string `json:"device" binding:"required"`
			}

			var req NotifyRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid request format: " + err.Error(),
				})
				return
			}

			// Format the generating system string
			generatingSystem := fmt.Sprintf("%s-%s-%s", req.Building, req.Room, req.Device)

			logger.Info("Manual help request notification triggered",
				"generating_system", generatingSystem)

			if err := rm.SendTheMessage(generatingSystem); err != nil {
				logger.Error("Failed to send Teams message", "error", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to send notification: " + err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message": "Notification sent successfully",
				"room_id": req.RoomID,
			})
		})

		// Configuration endpoints
		api.GET("/config", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"webhook_url":    *webhookUrl != "", // Don't expose actual URL
				"monitoring_url": *smeeUrl != "",    // Don't expose actual URL
				"log_level":      getLogLevelString(logLevel.Level()),
			})
		})
	}

	// Start the HTTP server in a goroutine
	go func() {
		addr := fmt.Sprintf(":%s", *port)
		logger.Info("Starting HTTP server", "address", addr)
		if err := router.Run(addr); err != nil {
			logger.Error("Failed to start HTTP server", "error", err)
		}
	}()

	// Event hub connection behavior depends on log level
	isDebugMode := logLevel.Level() == slog.LevelDebug

	// Connecting to event hub
	logger.Info("Starting event hub messenger")
	eventMessenger, nerr := messenger.BuildMessenger(*hubAddress, base.Messenger, 5000)

	if nerr != nil {
		if isDebugMode {
			// In debug mode, log a warning but continue running the service
			logger.Warn("Failed to build event hub messenger - continuing in debug mode without event subscription",
				"error", nerr,
				"hub_address", *hubAddress)
		} else {
			// In non-debug mode, treat this as a fatal error
			logger.Error("Failed to build event hub messenger - exiting",
				"error", nerr,
				"hub_address", *hubAddress)
			return
		}
	}

	// Only run event listener if we have a messenger
	if eventMessenger != nil {
		// Subscribe to the event hub
		logger.Info("Listening for room events")
		eventMessenger.SubscribeToRooms("*")

		for {
			event := eventMessenger.ReceiveEvent()
			if checkEvent(event) {
				logger.Debug("Help request detected",
					"key", event.Key,
					"value", event.Value,
					"generating_system", event.GeneratingSystem)

				// Send Message Via Teams
				if err := rm.SendTheMessage(event.GeneratingSystem); err != nil {
					logger.Error("Failed to send Teams message", "error", err)
				}
			}
		}
	} else if isDebugMode {
		// In debug mode without an event messenger, just keep the service running
		logger.Info("Running in debug mode without event subscription - HTTP endpoints will remain available")

		// Keep the main goroutine alive
		select {}
	}
}

func checkEvent(event events.Event) bool {
	return event.Key == "help-request" && event.Value == "confirm" && contains(event.EventTags, "alert")
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
