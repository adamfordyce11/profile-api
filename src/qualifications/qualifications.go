package qualifications

import (
	"context"
	"log"
	"net/http"

	"profile-api/auth"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var qualificationsCollection *mongo.Collection

// ErrorResponse is a struct that represents an error response.
//
// swagger:model ErrorResponse
type ErrorResponse struct {
	// Error message
	// example: Invalid request body
	Error string `json:"error"`
}

// GetQualifications retrieves all qualifications for a specific user.
//
//	@Summary		Get all qualifications for a user.
//	@Description	Retrieves all qualifications associated with the specified user ID.
//	@tags			Qualifications
//	@Security		BearerAuth
//	@ID				get-qualifications
//	@Param			userid	path		string	true	"The ID of the user whose qualifications are to be retrieved"
//	@Success		200		{array}		Qualification
//	@Failure		401		{object}	ErrorResponse	"Not authenticated"
//	@Failure		500		{object}	ErrorResponse	"Could not retrieve qualifications"
//	@Router			/qualifications/{userid} [get]
func GetQualifications(c *gin.Context) {
	userID := c.Param("userid")

	var qualifications []Qualification
	cursor, err := qualificationsCollection.Find(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve qualifications"})
		return
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var qualification Qualification
		err := cursor.Decode(&qualification)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve qualifications"})
			return
		}
		qualifications = append(qualifications, qualification)
	}

	c.JSON(http.StatusOK, qualifications)
}

// GetQualificationEntry retrieves a specific qualification for a user.
//
//	@Summary		Get a specific qualification for a user.
//	@Description	Retrieves the qualification entry associated with the specified user ID and qualification ID.
//	@tags			Qualifications
//	@Security		BearerAuth
//	@ID				get-qualification-entry
//	@Param			userid			path		string	true	"The ID of the user whose qualification is to be retrieved"
//	@Param			qualificationid	path		string	true	"The ID of the qualification to be retrieved"
//	@Success		200				{object}	Qualification
//	@Failure		401				{object}	ErrorResponse	"Not authenticated"
//	@Failure		500				{object}	ErrorResponse	"Could not retrieve qualification"
//	@Router			/qualifications/{userid}/{qualificationid} [get]
func GetQualificationEntry(c *gin.Context) {
	userID := c.Param("userid")
	qualificationID := c.Param("qualificationid")

	var qualification Qualification
	err := qualificationsCollection.FindOne(context.Background(), bson.M{"user_id": userID, "qualification_id": qualificationID}).Decode(&qualification)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve qualification"})
		return
	}

	c.JSON(http.StatusOK, qualification)
}

// PutQualificationEntry updates a specific qualification for a user.
//
//	@Summary		Update a specific qualification for a user.
//	@Description	Updates the qualification entry associated with the specified user ID and qualification ID using the provided qualification data.
//	@tags			Qualifications
//	@Security		BearerAuth
//	@ID				put-qualification-entry
//	@Param			userid			path		string			true	"The ID of the user whose qualification is to be updated"
//	@Param			qualificationid	path		string			true	"The ID of the qualification to be updated"
//	@Param			request			body		Qualification	true	"Qualification object that needs to be updated"
//	@Success		200				{string}	string			"Qualification updated"
//	@Failure		400				{object}	ErrorResponse	"Invalid request body"
//	@Failure		401				{object}	ErrorResponse	"Not authenticated"
//	@Failure		500				{object}	ErrorResponse	"Could not update qualification"
//	@Router			/qualifications/{userid}/{qualificationid} [put]
func PutQualificationEntry(c *gin.Context) {
	userID := c.Param("userid")
	qualificationID := c.Param("qualificationid")

	var req Qualification
	if err := c.BindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	req.UserID = userID
	req.QualificationID = qualificationID

	_, err := qualificationsCollection.UpdateOne(context.Background(), bson.M{"user_id": userID, "qualification_id": qualificationID}, bson.M{"$set": req}, options.Update().SetUpsert(true))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not update qualification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Qualification updated"})
}

// DeleteQualificationEntry deletes a specific qualification for a user.
//
//	@Summary		Delete a specific qualification for a user.
//	@Description	Deletes the qualification entry associated with the specified user ID and qualification ID.
//	@tags			Qualifications
//	@Security		BearerAuth
//	@ID				delete-qualification-entry
//	@Param			userid			path		string			true	"The ID of the user whose qualification is to be deleted"
//	@Param			qualificationid	path		string			true	"The ID of the qualification to be deleted"
//	@Success		200				{string}	string			"Qualification deleted"
//	@Failure		401				{object}	ErrorResponse	"Not authenticated"
//	@Failure		500				{object}	ErrorResponse	"Could not delete qualification"
//	@Router			/qualifications/{userid}/{qualificationid} [delete]
func DeleteQualificationEntry(c *gin.Context) {
	userID := c.Param("userid")
	qualificationID := c.Param("qualificationid")

	_, err := qualificationsCollection.DeleteOne(context.Background(), bson.M{"user_id": userID, "qualification_id": qualificationID})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not delete qualification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Qualification deleted"})
}

// PutQualificationImage uploads a certificate image for a specific qualification.
//
//	@Summary		Upload a certificate image for a qualification.
//	@Description	Updates the certificate image for the qualification associated with the specified user ID and qualification ID using the provided image file.
//	@tags			Qualifications
//	@Security		BearerAuth
//	@ID				put-qualification-image
//	@Accept			mpfd
//	@Param			userid			path		string			true	"The ID of the user whose qualification certificate image is to be updated"
//	@Param			qualificationid	path		string			true	"The ID of the qualification whose certificate image is to be updated"
//	@Param			file			formData	file			true	"Certificate image file to upload"
//	@Success		200				{string}	string			"cert image uploaded"
//	@Failure		400				{object}	ErrorResponse	"invalid request body"
//	@Failure		401				{object}	ErrorResponse	"Not authenticated"
//	@Failure		500				{object}	ErrorResponse	"could not update qualification"
//	@Router			/qualifications/{userid}/{qualificationid}/cert_image [put]
func PutQualificationImage(c *gin.Context) {
	userID := c.Param("userid")
	qualificationID := c.Param("qualificationid")

	file, err := c.FormFile("file")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	FileBytes, err := file.Open()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	defer FileBytes.Close()

	_, err = qualificationsCollection.UpdateOne(context.Background(), bson.M{"user_id": userID, "qualification_id": qualificationID}, bson.M{"$set": bson.M{"cert_image": FileBytes}}, options.Update().SetUpsert(true))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "could not update qualification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cert image uploaded"})
}

// PostQualification creates a new qualification for a user.
//
//	@Summary		Create a new qualification for a user.
//	@Description	Creates a new qualification entry associated with the specified user ID using the provided qualification data.
//	@tags			Qualifications
//	@Security		BearerAuth
//	@ID				post-qualification
//	@Param			userid	path		string			true	"The ID of the user for whom the qualification is to be created"
//	@Param			request	body		Qualification	true	"Qualification object to be created"
//	@Success		200		{string}	string			"Qualification Created"
//	@Failure		400		{object}	ErrorResponse	"Invalid request body"
//	@Failure		401		{object}	ErrorResponse	"Not authenticated"
//	@Failure		500		{object}	ErrorResponse	"Could not update qualification"
//	@Router			/qualifications/{userid} [post]
func PostQualification(c *gin.Context) {
	userID := c.Param("userid")

	var req Qualification
	if err := c.BindJSON(&req); err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	req.UserID = userID
	req.QualificationID = primitive.NewObjectID().Hex()

	_, err := qualificationsCollection.InsertOne(context.Background(), req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not update qualification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Qualification Created"})
}

// InitializeRoutes initializes the qualifications routes
func InitializeRoutes(router *gin.RouterGroup, db *mongo.Client, db_name string) {
	qualificationsCollection = db.Database(db_name).Collection("qualifications")

	router.GET("/:userid", GetQualifications)
	router.GET("/:userid/:qualificationid", GetQualificationEntry)

	protected := router.Group("/")
	protected.Use(auth.AuthMiddleware(db, db_name, true))
	protected.POST("/:userid", PostQualification)
	protected.PUT("/:userid/:qualificationid", PutQualificationEntry)
	protected.DELETE("/:userid/:qualificationid", DeleteQualificationEntry)
	protected.PUT("/:userid/:qualificationid/cert_image", PutQualificationImage)
}
