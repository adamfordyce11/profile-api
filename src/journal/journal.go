package journal

import (
	"context"
	"net/http"
	"profile-api/auth"
	"profile-api/utils"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var journalCollection *mongo.Collection

type ErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

type ProcessingResponse struct {
	Message string `json:"message"`
	Body    string `json:"body"`
}

type DeleteResponse struct {
	Message string `json:"message"`
	Body    string `json:"body"`
}

type SuccessResponse struct {
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Version   int    `json:"version"`
	Status    string `json:"status"`
	UserID    string `json:"userID"`
}

// @Summary Create a new journal entry
// @Description Create a new journal entry
// @Tags journal
// @Accept json
// @Produce json
// @Param entry body Entry true "Journal Entry"
// @Success 201 {object} JournalEntry
// @Failure 400 {object} ErrorResponse "Error message"
// @Failure 500 {object} ErrorResponse "Error message"
// @Router /journal [post]
func CreateJournalEntry(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Type assert the user to the correct type
	userStruct, ok := user.(auth.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user information"})
		return
	}

	var newEntry Entry
	if err := c.ShouldBindJSON(&newEntry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	journalEntry := JournalEntry{
		JournalID: utils.GenerateID(),
		UserID:    userStruct.ID,
		Version:   1,
		Entries:   []Entry{newEntry},
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err := journalCollection.InsertOne(context.Background(), journalEntry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating journal entry"})
		return
	}

	c.JSON(http.StatusCreated, journalEntry)
}

// @Summary Update a journal entry
// @Description Update a journal entry by ID, increments the version
// @Tags journal
// @Accept json
// @Produce json
// @Param journalid path string true "Journal ID"
// @Param entry body Entry true "Updated Entry"
// @Success 200 {object} JournalEntry
// @Failure 400 {object} ErrorResponse "Error message"
// @Failure 404 {object} ErrorResponse "Error message"
// @Failure 500 {object} ErrorResponse "Error message"
// @Router /journal/{journalid} [put]
func UpdateJournalEntry(c *gin.Context) {
	journalID := c.Param("journalid")
	userID := c.MustGet("userID").(string)

	var updatedEntry Entry
	if err := c.ShouldBindJSON(&updatedEntry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var journal JournalEntry
	err := journalCollection.FindOne(context.Background(), bson.M{"journal_id": journalID, "user_id": userID}).Decode(&journal)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Journal entry not found"})
		return
	}

	updatedEntry.Version = journal.Version + 1
	updatedEntry.UpdatedAt = time.Now()
	journal.Entries = append(journal.Entries, updatedEntry)
	journal.Version = updatedEntry.Version
	journal.UpdatedAt = time.Now()

	_, err = journalCollection.UpdateOne(
		context.Background(),
		bson.M{"journal_id": journalID, "user_id": userID},
		bson.M{"$set": bson.M{"entries": journal.Entries, "version": journal.Version, "updated_at": journal.UpdatedAt}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating journal entry"})
		return
	}

	c.JSON(http.StatusOK, journal)
}

// @Summary Get journal metadata
// @Description Get metadata for a journal entry by ID
// @Tags journal
// @Produce json
// @Param journalid path string true "Journal ID"
// @Success 200 {object} SuccessResponse "createdAt", "updatedAt", "version", "status", "userID"
// @Failure 404 {object} ErrorResponse "Error message"
// @Router /journal/{journalid}/meta [get]
func GetJournalMeta(c *gin.Context) {
	journalID := c.Param("journalid")

	var journal JournalEntry
	err := journalCollection.FindOne(context.Background(), bson.M{"journal_id": journalID}).Decode(&journal)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Journal entry not found"})
		return
	}

	meta := gin.H{
		"createdAt": journal.CreatedAt,
		"updatedAt": journal.UpdatedAt,
		"version":   journal.Version,
		"status":    journal.Status,
		"userID":    journal.UserID,
	}

	c.JSON(http.StatusOK, meta)
}

// @Summary Process a journal entry
// @Description Trigger processing for a journal entry by ID
// @Tags journal
// @Accept json
// @Produce json
// @Param journalid path string true "Journal ID"
// @Success 200 {object} ProcessingResponse "Journal entry is being processed"
// @Failure 500 {object} ErrorResponse "Error message"
// @Router /journal/{journalid}/process [put]
func ProcessJournalEntry(c *gin.Context) {
	journalID := c.Param("journalid")
	userID := c.MustGet("userID").(string)

	_, err := journalCollection.UpdateOne(
		context.Background(),
		bson.M{"journal_id": journalID, "user_id": userID},
		bson.M{"$set": bson.M{"status": "processing"}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing journal entry"})
		return
	}

	// Trigger AI processing (to be implemented)
	c.JSON(http.StatusOK, gin.H{"message": "Journal entry is being processed"})
}

// @Summary Get journal versions
// @Description Get all versions of a journal entry by ID
// @Tags journal
// @Produce json
// @Param journalid path string true "Journal ID"
// @Success 200 {array} Entry
// @Failure 404 {object} ErrorResponse "Error message"
// @Router /journal/{journalid}/versions [get]
func GetJournalVersions(c *gin.Context) {
	journalID := c.Param("journalid")

	var journal JournalEntry
	err := journalCollection.FindOne(context.Background(), bson.M{"journal_id": journalID}).Decode(&journal)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Journal entry not found"})
		return
	}

	c.JSON(http.StatusOK, journal.Entries)
}

// @Summary Set the current version of a journal entry
// @Description Set the current version of a journal entry by ID
// @Tags journal
// @Accept json
// @Produce json
// @Param journalid path string true "Journal ID"
// @Param version body int true "Version"
// @Success 200 {object} JournalEntry
// @Failure 400 {object} ErrorResponse "Error message"
// @Failure 404 {object} ErrorResponse "Error message"
// @Failure 500 {object} ErrorResponse "Error message"
// @Router /journal/{journalid}/version [put]
func SetJournalVersion(c *gin.Context) {
	journalID := c.Param("journalid")
	userID := c.MustGet("userID").(string)

	var versionRequest struct {
		Version int `json:"version"`
	}
	if err := c.ShouldBindJSON(&versionRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var journal JournalEntry
	err := journalCollection.FindOne(context.Background(), bson.M{"journal_id": journalID, "user_id": userID}).Decode(&journal)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Journal entry not found"})
		return
	}

	for _, entry := range journal.Entries {
		if entry.Version == versionRequest.Version {
			journal.Version = versionRequest.Version
			journal.UpdatedAt = time.Now()

			_, err = journalCollection.UpdateOne(
				context.Background(),
				bson.M{"journal_id": journalID, "user_id": userID},
				bson.M{"$set": bson.M{"version": journal.Version, "updated_at": journal.UpdatedAt}},
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error setting journal version"})
				return
			}

			c.JSON(http.StatusOK, journal)
			return
		}
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "Version not found"})
}

// @Summary Set the status of a journal entry
// @Description Set the status of a journal entry by ID
// @Tags journal
// @Accept json
// @Produce json
// @Param journalid path string true "Journal ID"
// @Param status body string true "Status"
// @Success 200 {object} ProcessingResponse "Journal status updated"
// @Failure 400 {object} ErrorResponse "Error message"
// @Failure 500 {object} ErrorResponse "Error message"
// @Router /journal/{journalid}/status [put]
func SetJournalStatus(c *gin.Context) {
	journalID := c.Param("journalid")
	userID := c.MustGet("userID").(string)

	var statusRequest struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&statusRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := journalCollection.UpdateOne(
		context.Background(),
		bson.M{"journal_id": journalID, "user_id": userID},
		bson.M{"$set": bson.M{"status": statusRequest.Status, "updated_at": time.Now()}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error setting journal status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Journal status updated"})
}

// @Summary Get a single journal entry
// @Description Get a single journal entry by ID, returns metadata if the user is authenticated
// @Tags journal
// @Produce json
// @Param journalid path string true "Journal ID"
// @Success 200 {object} JournalEntry
// @Failure 404 {object} ErrorResponse "Error message"
// @Router /journal/{journalid} [get]
func GetJournalEntry(c *gin.Context) {
	journalID := c.Param("journalid")

	var journal JournalEntry
	err := journalCollection.FindOne(context.Background(), bson.M{"journal_id": journalID}).Decode(&journal)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Journal entry not found"})
		return
	}

	user, exists := c.Get("user")
	if exists && user != nil {
		meta := gin.H{
			"createdAt": journal.CreatedAt,
			"updatedAt": journal.UpdatedAt,
			"version":   journal.Version,
			"status":    journal.Status,
			"userID":    journal.UserID,
			"entries":   journal.Entries,
			"taxonomy":  journal.Taxonomy,
			"summary":   journal.Summary,
		}
		c.JSON(http.StatusOK, meta)
	} else {
		// Unauthenticated users get the latest entry as part of an array
		latestEntry := []Entry{}
		if len(journal.Entries) > 0 {
			latestEntry = append(latestEntry, journal.Entries[len(journal.Entries)-1])
		}

		c.JSON(http.StatusOK, gin.H{
			"journalID": journal.JournalID,
			"userID":    journal.UserID,
			"version":   journal.Version,
			"status":    journal.Status,
			"taxonomy":  journal.Taxonomy,
			"summary":   journal.Summary,
			"entries":   latestEntry, // Return only the latest version
		})
	}
}

// @Summary Get public journal entries
// @Description Get all public journal entries, supports filtering by date range, taxonomy, and users
// @Tags journal
// @Produce json
// @Param start query string false "Start date"
// @Param end query string false "End date"
// @Param category query string false "Category"
// @Param subcategory query string false "Subcategory"
// @Param topic query string false "Topic"
// @Param tag query string false "Tag"
// @Param user query string false "User ID"
// @Success 200 {array} JournalEntry
// @Failure 500 {object} ErrorResponse "Error message"
// @Router /journal [get]
func GetPublicJournals(c *gin.Context) {
	filter := bson.M{"status": "public"}

	startDate := c.Query("start")
	endDate := c.Query("end")
	category := c.Query("category")
	subcategory := c.Query("subcategory")
	topic := c.Query("topic")
	tag := c.Query("tag")
	user := c.Query("user")

	if startDate != "" && endDate != "" {
		filter["created_at"] = bson.M{
			"$gte": startDate,
			"$lte": endDate,
		}
	}

	if category != "" {
		filter["taxonomy.categories"] = category
	}

	if subcategory != "" {
		filter["taxonomy.subcategories"] = subcategory
	}

	if topic != "" {
		filter["taxonomy.topics"] = topic
	}

	if tag != "" {
		filter["taxonomy.tags"] = tag
	}

	if user != "" {
		filter["user_id"] = user
	}

	cursor, err := journalCollection.Find(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving journal entries"})
		return
	}
	defer cursor.Close(context.Background())

	var journals []JournalEntry
	if err := cursor.All(context.Background(), &journals); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing journal entries"})
		return
	}

	c.JSON(http.StatusOK, journals)
}

// @Summary Get user-specific journal entries
// @Description Get all journal entries for a specific user by ID
// @Tags journal
// @Produce json
// @Param userid path string true "User ID"
// @Success 200 {array} JournalEntry
// @Failure 500 {object} ErrorResponse "Error message"
// @Router /journal/u/{userid} [get]
func GetUserJournals(c *gin.Context) {
	userID := c.Param("userid")

	filter := bson.M{"user_id": userID}

	cursor, err := journalCollection.Find(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving journal entries"})
		return
	}
	defer cursor.Close(context.Background())

	var journals []JournalEntry
	if err := cursor.All(context.Background(), &journals); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing journal entries"})
		return
	}

	c.JSON(http.StatusOK, journals)
}

// @Summary Delete a journal entry
// @Description Delete a journal entry by ID
// @Tags journal
// @Produce json
// @Param journalid path string true "Journal ID"
// @Success 200 {object} DeleteResponse "Journal entry deleted"
// @Failure 500 {object} ErrorResponse "Error message"
// @Router /journal/{journalid} [delete]
func DeleteJournalEntry(c *gin.Context) {
	journalID := c.Param("journalid")
	userID := c.MustGet("userID").(string)

	_, err := journalCollection.DeleteOne(context.Background(), bson.M{"journal_id": journalID, "user_id": userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting journal entry"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Journal entry deleted"})
}

func InitializeRoutes(router *gin.RouterGroup, db *mongo.Client, db_name string) {
	journalCollection = db.Database(db_name).Collection("journal")

	router.GET("/", GetPublicJournals)
	router.GET("/u/:userid", GetUserJournals)
	router.GET("/:journalid", GetJournalEntry)
	router.GET("/:journalid/meta", GetJournalMeta)

	authRequired := auth.AuthMiddleware(db, db_name, true)
	protected := router.Group("/")
	protected.Use(authRequired)
	protected.POST("/", CreateJournalEntry)
	protected.PUT("/:journalid", UpdateJournalEntry)
	protected.PUT("/:journalid/process", ProcessJournalEntry)
	protected.GET("/:journalid/versions", GetJournalVersions)
	protected.PUT("/:journalid/version", SetJournalVersion)
	protected.PUT("/:journalid/status", SetJournalStatus)
	protected.DELETE("/:journalid", DeleteJournalEntry)
}
