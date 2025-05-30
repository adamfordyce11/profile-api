package certificates

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

var certificateCollection *mongo.Collection

// Certificate represents a user's certificate
type JSONResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

// GetCertificates retrieves all certificates for a given user.
//
//	@Summary		Get all certificates
//	@Description	Retrieves all certificates for a given user
//	@Tags			Certificates
//	@Accept			json
//	@Produce		json
//	@Param			userid	path		string	true	"User ID"
//	@Success		200		{array}		Certificate
//	@Failure		500		{object}	JSONResponse	"error":	"Could not retrieve certificates"
//	@Router			/certificates/{userid} [get]
func GetCertificates(c *gin.Context) {
	userID := c.Param("userid")

	var certificates []Certificate
	cursor, err := certificateCollection.Find(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve certificates"})
		return
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var certificate Certificate
		err := cursor.Decode(&certificate)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve certificate"})
			return
		}
		certificates = append(certificates, certificate)
	}

	c.JSON(http.StatusOK, certificates)
}

// GetCertificateEntry retrieves a specific certificate entry for a user.
//
//	@Summary		Get a certificate entry
//	@Description	Retrieves a specific certificate entry for a user
//	@Tags			Certificates
//	@Accept			json
//	@Produce		json
//	@Param			userid			path		string	true	"User ID"
//	@Param			certificateid	path		string	true	"Certificate ID"
//	@Success		200				{object}	Certificate
//	@Failure		404				{object}	JSONResponse	"error":	"Certificate not found"
//	@Failure		401				{object}	JSONResponse	"error":	"Unauthorized"
//	@Failure		403				{object}	JSONResponse	"error":	"Forbidden"
//	@Failure		400				{object}	JSONResponse	"error":	"Invalid request body"
//	@Failure		500				{object}	JSONResponse	"error":	"Could not retrieve certificate"
//	@Router			/certificates/{userid}/{certificateid} [get]
func GetCertificateEntry(c *gin.Context) {
	userID := c.Param("userid")
	certificateID := c.Param("certificateid")

	var certificate Certificate
	err := certificateCollection.FindOne(context.Background(), bson.M{"user_id": userID, "certificate_id": certificateID}).Decode(&certificate)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve certificate"})
		return
	}

	c.JSON(http.StatusOK, certificate)
}

// PutCertificateEntry updates or creates a specific certificate entry for a user.
//
//	@Summary		Update or create a certificate entry
//	@Description	Updates or creates a specific certificate entry for a user
//	@Tags			Certificates
//	@Accept			json
//	@Produce		json
//	@Param			userid			path		string		true	"User ID"
//	@Param			certificateid	path		string		true	"Certificate ID"
//	@Param			body			body		Certificate	true	"Certificate JSON object"
//	@Success		200				{object}	map[string]string
//	@Failure		400				{object}	JSONResponse	"error":	"Invalid request body"
//	@Failure		500				{object}	JSONResponse	"error":	"Could not update certificate"
//	@Failure		401				{object}	JSONResponse	"error":	"Unauthorized"
//	@Failure		403				{object}	JSONResponse	"error":	"Forbidden"
//	@Security		BearerAuth
//	@Router			/certificates/{userid}/{certificateid} [put]
func PutCertificateEntry(c *gin.Context) {
	userID := c.Param("userid")
	certificateID := c.Param("certificateid")

	var req Certificate
	if err := c.BindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	req.UserID = userID
	req.CertificateID = certificateID

	_, err := certificateCollection.UpdateOne(context.Background(), bson.M{"user_id": userID, "certificate_id": certificateID}, bson.M{"$set": req}, options.Update().SetUpsert(true))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not update certificate"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Certificate updated"})
}

// DeleteCertificateEntry deletes a specific certificate entry for a user.
//
//	@Summary		Delete a certificate entry
//	@Description	Deletes a specific certificate entry for a user
//	@Tags			Certificates
//	@Accept			json
//	@Produce		json
//	@Param			userid			path		string	true	"User ID"
//	@Param			certificateid	path		string	true	"Certificate ID"
//	@Success		200				{object}	map[string]string
//	@Router			/certificates/{userid}/{certificateid} [delete]
func DeleteCertificateEntry(c *gin.Context) {
	userID := c.Param("userid")
	certificateID := c.Param("certificateid")

	_, err := certificateCollection.DeleteOne(context.Background(), bson.M{"user_id": userID, "certificate_id": certificateID})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not delete certificate"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Certificate deleted"})
}

// PutCertificateImage uploads or updates the certificate image for a specific certificate entry.
//
//	@Summary		Upload or update certificate image
//	@Description	Uploads or updates the certificate image for a specific certificate entry
//	@Tags			Certificates
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			userid			path		string	true	"User ID"
//	@Param			certificateid	path		string	true	"Certificate ID"
//	@Param			file			formData	file	true	"Certificate Image"
//	@Success		200				{object}	map[string]string
//	@Router			/certificates/{userid}/{certificateid}/cert_image [put]
func PutCertificateImage(c *gin.Context) {
	userID := c.Param("userid")
	certificateID := c.Param("certificateid")

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

	_, err = certificateCollection.UpdateOne(context.Background(), bson.M{"user_id": userID, "certificate_id": certificateID}, bson.M{"$set": bson.M{"cert_image": FileBytes}}, options.Update().SetUpsert(true))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "could not update certification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cert image uploaded"})
}

// PostCertificate creates a new certificate entry for a user.
//
//	@Summary		Create a new certificate entry
//	@Description	Creates a new certificate entry for a user
//	@Tags			Certificates
//	@Accept			json
//	@Produce		json
//	@Param			userid	path		string		true	"User ID"
//	@Param			body	body		Certificate	true	"Certificate JSON object"
//	@Success		200		{object}	map[string]string
//	@Failure		400		{object}	JSONResponse	"error":	"Invalid request body"
//	@Failure		401		{object}	JSONResponse	"error":	"Unauthorized"
//	@Failure		403		{object}	JSONResponse	"error":	"Forbidden"
//	@Failure		404		{object}	JSONResponse	"error":	"User not found"
//	@Failure		409		{object}	JSONResponse	"error":	"Certificate already exists"
//	@Failure		500		{object}	JSONResponse	"error":	"Could not create certificate"
//	@Security		BearerAuth
//	@Router			/certificates/{userid} [post]
func PostCertificate(c *gin.Context) {
	userID := c.Param("userid")

	var req Certificate
	if err := c.BindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	req.UserID = userID
	req.CertificateID = primitive.NewObjectID().Hex()

	_, err := certificateCollection.InsertOne(context.Background(), req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not create certificate"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Certificate Added"})
}

// InitializeRoutes initializes the certificates routes
func InitializeRoutes(router *gin.RouterGroup, db *mongo.Client, db_name string) {
	certificateCollection = db.Database(db_name).Collection("certificates")

	authOptional := auth.AuthMiddleware(db, db_name, false)
	authRequired := auth.AuthMiddleware(db, db_name, true)

	router.GET("/:userid", authOptional, GetCertificates)
	router.GET("/:userid/:certificateid", authOptional, GetCertificateEntry)

	protected := router.Group("/")
	protected.Use(authRequired)
	protected.POST("/:userid", PostCertificate)
	protected.PUT("/:userid/:certificateid", PutCertificateEntry)
	protected.DELETE("/:userid/:certificateid", DeleteCertificateEntry)
	protected.PUT("/:userid/:certificateid/cert_image", PutCertificateImage)
}
