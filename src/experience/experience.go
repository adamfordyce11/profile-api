package experience

import (
	"context"
	"net/http"
	"profile-api/auth"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var experienceCollection *mongo.Collection

type JSONResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

// GetExperience retrieves all work experience records for the specified user.
//
//	@Summary		Get all user experiences
//	@Description	Retrieves all work experience records for the specified user
//	@Tags			experience
//	@Accept			json
//	@Produce		json
//	@Param			userid	path		string	true	"User ID"
//	@Success		200		{array}		Experience
//	@Failure		500		{object}	JSONResponse	"error":	"Could not retrieve experience"
//	@Router			/experience/{userid} [get]
func GetExperience(c *gin.Context) {
	userID := c.Param("userid")
	var experience []Experience
	cursor, err := experienceCollection.Find(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve experience"})
		return
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var exp Experience
		err := cursor.Decode(&exp)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve experience"})
			return
		}
		experience = append(experience, exp)
	}

	c.JSON(http.StatusOK, experience)
}

// GetExperienceItem retrieves a specific work experience record for the specified user and experience ID.
//
//	@Summary		Get specific experience item
//	@Description	Retrieves a specific work experience record for the specified user and experience ID
//	@Tags			experience
//	@Accept			json
//	@Produce		json
//	@Param			userid			path		string	true	"User ID"
//	@Param			experienceid	path		string	true	"Experience ID"
//	@Success		200				{object}	Experience
//	@Failure		500				{object}	JSONResponse	"error":	"Could not retrieve experience"
//	@Router			/experience/{userid}/{experienceid} [get]
func GetExperienceItem(c *gin.Context) {
	userID := c.Param("userid")
	experienceID := c.Param("experienceid")
	var exp Experience
	err := experienceCollection.FindOne(context.Background(), bson.M{"user_id": userID, "experience_id": experienceID}).Decode(&exp)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve experience"})
		return
	}

	c.JSON(http.StatusOK, exp)
}

// PutExperienceItem updates a specific work experience record for the specified user and experience ID.
//
//	@Summary		Update specific experience item
//	@Description	Updates a specific work experience record for the specified user and experience ID
//	@Tags			experience
//	@Accept			json
//	@Produce		json
//	@Param			userid			path		string			true		"User ID"
//	@Param			experienceid	path		string			true		"Experience ID"
//	@Param			Experience		body		Experience		true		"Experience Object"
//	@Success		200				{object}	JSONResponse	"message":	"Experience updated"
//	@Failure		400				{object}	JSONResponse	"error":	"Invalid request body"
//	@Failure		401				{object}	JSONResponse	"error":	"Unauthorized"
//	@Failure		403				{object}	JSONResponse	"error":	"Forbidden"
//	@Failure		500				{object}	JSONResponse	"error":	"Could not update experience"
//	@Security		BearerAuth
//	@Router			/experience/{userid}/{experienceid} [put]
func PutExperienceItem(c *gin.Context) {
	userID := c.Param("userid")
	experienceID := c.Param("experienceid")

	var req Experience
	if err := c.BindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	req.UserID = userID
	req.ExperienceID = experienceID

	_, err := experienceCollection.UpdateOne(context.Background(), bson.M{"user_id": userID, "experience_id": experienceID}, bson.M{"$set": req}, options.Update().SetUpsert(true))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not update experience"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Experience updated"})
}

// PostExperience creates a new work experience record for the specified user.
//
//	@Summary		Create a new experience item
//	@Description	Creates a new work experience record for the specified user
//	@Tags			experience
//	@Accept			json
//	@Produce		json
//	@Param			userid		path		string		true	"User ID"
//	@Param			Experience	body		Experience	true	"Experience Object"
//	@Success		200			{object}	Experience
//	@Failure		400			{object}	JSONResponse	"error":	"Invalid request body"
//	@Failure		401			{object}	JSONResponse	"error":	"Unauthorized"
//	@Failure		403			{object}	JSONResponse	"error":	"Forbidden"
//	@Failure		404			{object}	JSONResponse	"error":	"User not found"
//	@Failure		409			{object}	JSONResponse	"error":	"Experience already exists"
//	@Failure		422			{object}	JSONResponse	"error":	"Invalid experience type"
//	@Failure		500			{object}	JSONResponse	"error":	"Could not insert experience"
//	@Router			/experience/{userid} [post]
func PostExperience(c *gin.Context) {
	userID := c.Param("userid")

	var req Experience
	if err := c.BindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	req.UserID = userID
	req.ExperienceID = primitive.NewObjectID().Hex()

	_, err := experienceCollection.InsertOne(context.Background(), req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not insert experience"})
		return
	}

	c.JSON(http.StatusOK, req)
}

// DeleteExperienceItem deletes a specific work experience record for the specified user and experience ID.
//
//	@Summary		Delete specific experience item
//	@Description	Deletes a specific work experience record for the specified user and experience ID
//	@Tags			experience
//	@Accept			json
//	@Produce		json
//	@Param			userid			path		string			true		"User ID"
//	@Param			experienceid	path		string			true		"Experience ID"
//	@Success		200				{object}	JSONResponse	"message":	"Experience deleted"
//	@Failure		500				{object}	JSONResponse	"error":	"Could not delete experience"
//	@Router			/experience/{userid}/{experienceid} [delete]
func DeleteExperienceItem(c *gin.Context) {
	userID := c.Param("userid")
	experienceID := c.Param("experienceid")

	_, err := experienceCollection.DeleteOne(context.Background(), bson.M{"user_id": userID, "experience_id": experienceID})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not delete experience"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Experience deleted"})
}

// InitializeRoutes initializes the experience routes
func InitializeRoutes(router *gin.RouterGroup, db *mongo.Client, db_name string) {
	experienceCollection = db.Database(db_name).Collection("experience")

	router.GET("/:userid", GetExperience)
	router.GET("/:userid/:experienceid", GetExperienceItem)

	authRequired := auth.AuthMiddleware(db, db_name, true)
	protected := router.Group("/")
	protected.Use(authRequired)
	protected.POST("/:userid", PostExperience)
	protected.PUT("/:userid/:experienceid", PutExperienceItem)
	protected.DELETE("/:userid/:experienceid", DeleteExperienceItem)
}
