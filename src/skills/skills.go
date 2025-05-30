package skills

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

var skillsCollection *mongo.Collection

// Skill represents a user's skill
type JSONResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

// GetSkills retrieves all skills for a specific user
//
//	@Summary		Retrieve all skills for a specific user
//	@Description	Retrieve all skills for a specific user
//	@Tags			Skills
//	@Produce		json
//	@Param			userid	path		string			true	"User ID"
//	@Success		200		{array}		Skill			"Skills retrieved"
//	@Failure		401		{object}	JSONResponse	"error":	"Unauthorized"
//	@Failure		403		{object}	JSONResponse	"error":	"Forbidden"
//	@Failure		404		{object}	JSONResponse	"error":	"Skill not found"
//	@Failure		500		{object}	JSONResponse	"error":	"Could not retrieve skills"
//	@Router			/skills/{userid} [get]
func GetSkills(c *gin.Context) {
	userID := c.Param("userid")

	var skills []Skill
	cursor, err := skillsCollection.Find(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve skills"})
		return
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var skill Skill
		err := cursor.Decode(&skill)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve skills"})
			return
		}
		skills = append(skills, skill)
	}

	c.JSON(http.StatusOK, skills)
}

// GetSkill retrieves a specific skill for a specific user
//
//	@Summary		Retrieve a specific skill for a specific user
//	@Description	Retrieve a specific skill for a specific user
//	@Tags			Skills
//	@Produce		json
//	@Param			userid		path		string			true	"User ID"
//	@Param			skillid	path		string			true	"Skill ID"
//	@Success		200			{object}	Skill			"Skill retrieved"
//	@Failure		404			{object}	JSONResponse	"error":	"Skill not found"
//	@Failure		401			{object}	JSONResponse	"error":	"Unauthorized"
//	@Failure		500			{object}	JSONResponse	"error":	"Could not retrieve skill"
//	@Router			/skills/{userid}/{skillid} [get]
func GetSkill(c *gin.Context) {
	userID := c.Param("userid")
	skillID := c.Param("skillid")

	var skill Skill
	err := skillsCollection.FindOne(context.Background(), bson.M{"user_id": userID, "skill_id": skillID}).Decode(&skill)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve skill"})
		return
	}

	c.JSON(http.StatusOK, skill)
}

// PostSkill creates a new skill for a specific user
//
//	@Summary		Create a new skill for a specific user
//	@Description	Create a new skill for a specific user
//	@Tags			Skills
//	@Accept			json
//	@Produce		json
//	@Param			userid	path		string			true	"User ID"
//	@Param			req		body		Skill			true	"Skill details"
//	@Success		200		{object}	JSONResponse	"Skill created"
//	@Failure		400		{object}	JSONResponse	"Invalid request body"
//	@Failure		401		{object}	JSONResponse	"Unauthorized"
//	@Failure		403		{object}	JSONResponse	"Forbidden"
//	@Failure		404		{object}	JSONResponse	"Skill not found"
//	@Failure		409		{object}	JSONResponse	"Skill already exists"
//	@Failure		500		{object}	JSONResponse	"Could not create skill"
//	@Router			/skills/{userid} [post]
func PostSkill(c *gin.Context) {
	userID := c.Param("userid")

	var req Skill
	if err := c.BindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	req.UserID = userID
	req.SkillID = primitive.NewObjectID().Hex()

	_, err := skillsCollection.InsertOne(context.Background(), req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not create skill"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Skill created"})
}

// PutSkill updates a specific skill for a specific user
//
//	@Summary		Update a specific skill for a specific user
//	@Description	Update a specific skill for a specific user
//	@Tags			Skills
//	@Accept			json
//	@Produce		json
//	@Param			userid		path		string			true	"User ID"
//	@Param			skillname	path		string			true	"Skill Name"
//	@Param			req			body		Skill			true	"Skill details"
//	@Success		200			{object}	JSONResponse	"Skill updated"
//	@Failure		400			{object}	JSONResponse	"Invalid request body"
//	@Failure		401			{object}	JSONResponse	"Unauthorized"
//	@Failure		403			{object}	JSONResponse	"Forbidden"
//	@Failure		404			{object}	JSONResponse	"Skill not found"
//	@Failure		500			{object}	JSONResponse	"Could not update skill"
//	@Router			/skills/{userid}/{skillId} [put]
func PutSkill(c *gin.Context) {
	userID := c.Param("userid")
	skillID := c.Param("skillid")

	var req Skill
	if err := c.BindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	req.UserID = userID
	req.SkillID = skillID

	_, err := skillsCollection.UpdateOne(context.Background(), bson.M{"user_id": userID, "skill_id": skillID}, bson.M{"$set": req}, options.Update().SetUpsert(true))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not update skill"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Skill updated"})
}

// DeleteSkill deletes a specific skill for a specific user
//
//	@Summary		Delete a specific skill for a specific user
//	@Description	Delete a specific skill for a specific user
//	@Tags			Skills
//	@Accept			json
//	@Produce		json
//	@Param			userid		path		string			true	"User ID"
//	@Param			skillid	path		string			true	"Skill ID"
//	@Success		200			{object}	JSONResponse	"Skill deleted"
//	@Failure		400			{object}	JSONResponse	"Invalid request body"
//	@Failure		404			{object}	JSONResponse	"Skill not found"
//	@Failure		401			{object}	JSONResponse	"Unauthorized"
//	@Failure		403			{object}	JSONResponse	"Forbidden"
//	@Failure		422			{object}	JSONResponse	"Invalid request body"
//	@Failure		429			{object}	JSONResponse	"Too many requests"
//	@Failure		500			{object}	JSONResponse	"Could not delete skill"
//	@Router			/skills/{userid}/{skillid} [delete]
func DeleteSkill(c *gin.Context) {
	userID := c.Param("userid")
	skillID := c.Param("skillid")

	_, err := skillsCollection.DeleteOne(context.Background(), bson.M{"user_id": userID, "skill_id": skillID})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not delete skill"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Skill deleted"})
}

// InitializeRoutes initializes the skills routes
func InitializeRoutes(router *gin.RouterGroup, db *mongo.Client, db_name string) {
	skillsCollection = db.Database(db_name).Collection("skills")
	router.GET("/:userid", GetSkills)
	router.GET("/:userid/:skillid", GetSkill)

	protected := router.Group("/")
	protected.Use(auth.AuthMiddleware(db, db_name, true))
	protected.POST("/:userid", PostSkill)
	protected.PUT("/:userid/:skillid", PutSkill)
	protected.DELETE("/:userid/:skillid", DeleteSkill)
}
