package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"profile-api/auth"
	"profile-api/certificates"
	"profile-api/experience"
	"profile-api/journal"
	"profile-api/profile"
	"profile-api/qualifications"
	"profile-api/skills"
	"profile-api/utils"

	_ "profile-api/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var templates *template.Template

// extractIdentifierMiddleware is a middleware that extracts the subdomain or email from the request and stores it in the Gin context.
func extractIdentifierMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		host := c.Request.Host
		subdomain := extractSubdomain(host)
		email := c.Param("email")

		if subdomain != "" {
			c.Set("identifier", subdomain)
		} else if email != "" {
			c.Set("identifier", email)
		}
		c.Next()
	}
}

// extractSubdomain takes the host and returns the subdomain part.
// It will only return a subdomain if the host contains a valid subdomain.
func extractSubdomain(host string) string {
	// Remove the port if present
	hostname, _, err := net.SplitHostPort(host)
	if err != nil {
		hostname = host // If no port, use the original host
	}

	// Split the hostname into parts
	parts := strings.Split(hostname, ".")

	// Ensure the hostname has at least three parts for a subdomain (e.g., sub.domain.com)
	if len(parts) > 2 {
		return parts[0] // Return the first part as the subdomain
	}

	// Return empty string if no subdomain is found
	return ""
}

// @title			Go Profile API
// @version		1
// @description	This is the Go Profile API documentation.
// @host			127.0.0.1:8080
// @basePath		/api/v1
// @schemes		http
// @produce		json
// @contact		adamfordyce@hotmail.com
// @license		MIT
func main() {

	db_name, ok := os.LookupEnv("MONGO_HOST")
	if ok {
		fmt.Printf("MONGO_HOST env is %s\n", db_name)
	}

	db_uri, ok := os.LookupEnv("MONGO_URI")
	if ok {
		fmt.Printf("MONGO_HOST env is %s\n", db_uri)
	}

	// Load config
	_, err := os.Stat("config-selfhosted.json")
	if os.IsNotExist(err) {
		log.Fatalf("Config file not found")
	}
	configData, err := os.ReadFile("config-selfhosted.json")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	var config map[string]interface{}
	err = json.Unmarshal(configData, &config)
	if err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}

	// Load from config and allow environment variables to override
	db_name = config["mongodb"].(map[string]interface{})["database"].(string)
	if os.Getenv("MONGO_DB_NAME") != "" {
		db_name = os.Getenv("MONGO_DB_NAME")
	}
	db_uri = config["mongodb"].(map[string]interface{})["uri"].(string)
	if os.Getenv("MONGO_URI") != "" {
		db_uri = os.Getenv("MONGO_URI")
	}

	listen_port := config["listen-port"].(float64)
	if os.Getenv("LISTEN_PORT") != "" {
		var err error
		listen_port, err = strconv.ParseFloat(os.Getenv("LISTEN_PORT"), 64)
		if err != nil {
			log.Fatalf("Error parsing LISTEN_PORT environment variable: %v", err)
		}
	}

	// Connect to the database
	db, err := utils.ConnectDB(db_uri)
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}

	router := gin.Default()
	router.Use(extractIdentifierMiddleware())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Initialize authentication routes
	authRouter := router.Group("/api/v1/auth")
	auth.InitializeRoutes(authRouter, db, db_name)

	// Initialize profile routes
	profileRouter := router.Group("/api/v1/profile")
	profile.InitializeRoutes(profileRouter, db, db_name)

	// Initialize experience routes
	experienceRouter := router.Group("/api/v1/experience")
	experience.InitializeRoutes(experienceRouter, db, db_name)

	// Initialize qualifications routes
	qualificationsRouter := router.Group("/api/v1/qualifications")
	qualifications.InitializeRoutes(qualificationsRouter, db, db_name)

	// Initialize qualifications routes
	certificatesRouter := router.Group("/api/v1/certificates")
	certificates.InitializeRoutes(certificatesRouter, db, db_name)

	// Initialize skills routes
	skillsRouter := router.Group("/api/v1/skills")
	skills.InitializeRoutes(skillsRouter, db, db_name)

	// Initialize journal routes
	journalRouter := router.Group("/api/v1/journal")
	journal.InitializeRoutes(journalRouter, db, db_name)

	router.NoRoute(func(c *gin.Context) {
		// Debugging the incoming path
		path := c.Request.URL.Path
		log.Printf("404: Incoming request path: %s", path)
		c.JSON(http.StatusNotFound, gin.H{"error": "NotFound"})
		return
	})

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", int(listen_port)),
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("Starting server on port %d", int(listen_port))

	//log.Fatal(s.ListenAndServeTLS(certPath, keyPath))
	// Start the server
	log.Fatal(s.ListenAndServe())
}
