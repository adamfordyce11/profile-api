package profile

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"profile-api/auth"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var profilesCollection *mongo.Collection

type ErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

var imageStore ImageStore

func InitImageStore() error {
	storeType := os.Getenv("IMAGE_STORE")
	if storeType == "s3" {
		bucket := os.Getenv("S3_BUCKET")
		region := os.Getenv("AWS_REGION")
		accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
		secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
		endpoint := os.Getenv("AWS_S3_ENDPOINT") // For LocalStack, e.g. http://localstack:4566

		// Custom AWS config for LocalStack or real AWS
		cfg, err := config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		)
		if err != nil {
			return fmt.Errorf("unable to load AWS config: %w", err)
		}

		// If using LocalStack, override the endpoint
		var client *s3.Client
		if endpoint != "" {
			client = s3.NewFromConfig(cfg, func(o *s3.Options) {
				o.EndpointResolver = s3.EndpointResolverFromURL(endpoint)
				o.UsePathStyle = true // Required for LocalStack
			})
		} else {
			client = s3.NewFromConfig(cfg)
		}

		// Check if the bucket exists, create if not
		_, err = client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
			Bucket: &bucket,
		})
		if err != nil {
			_, createErr := client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
				Bucket: &bucket,
			})
			if createErr != nil {
				return fmt.Errorf("unable to create S3 bucket: %w", createErr)
			}
		}

		s3Store := &S3ImageStore{
			Client:     client,
			BucketName: bucket,
		}
		// Call the method on the concrete type
		if err := s3Store.InitBucketAndCORS(context.TODO()); err != nil {
			return err
		}
		// Now assign to the interface
		imageStore = s3Store
	} else {
		basePath := os.Getenv("LOCAL_PATH")
		imageStore = &LocalImageStore{BasePath: basePath}
	}
	return nil
}

// GetProfile retrieves the profile of the given user.
//
//	@Summary		Retrieve a user's profile.
//	@Description	Retrieves the profile of the user with the specified user ID.
//	@Tags			profile
//	@Security		BearerAuth
//	@ID				get-profile
//	@Param			userid	path		string			true	"The ID of the user whose profile to get"
//	@Success		200		{object}	Profile			"Profile retrieved successfully"
//	@Failure		401		{object}	ErrorResponse	"Not authenticated"
//	@Failure		500		{object}	ErrorResponse	"Could not retrieve profile"
//	@Router			/profile/{userid} [get]
func GetProfile(c *gin.Context) {
	userID := c.Param("userid")

	var profile Profile
	err := profilesCollection.FindOne(context.Background(), bson.M{"user_id": userID}).Decode(&profile)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve profile"})
		return
	}

	// If the user is not the owner of the profile, do not return the email address
	c.JSON(http.StatusOK, profile)
}

// PutImage updates the profile image of the given user.
//
//	@Summary		Update a user's profile image.
//	@Description	Updates the profile image of the user with the specified user ID.
//	@Tags			profile
//	@Security		BearerAuth
//	@ID				update-profile-image
//	@Param			userid			path		string			true	"The ID of the user whose profile image to update"
//	@Param			profileImage	formData	file			true	"Profile image to upload"
//	@Success		200				{string}	string			"Profile image updated"
//	@Failure		400				{object}	ErrorResponse	"Profile image not found"
//	@Failure		401				{object}	ErrorResponse	"Not authenticated"
//	@Failure		500				{object}	ErrorResponse	"Could not upload image"
//	@Router			/profile/{userid}/image [put]
func PutImage(c *gin.Context) {
	userID := c.Param("userid")

	fileHeader, err := c.FormFile("profileImage")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Profile image not found"})
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		log.Printf("Error opening file: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not open image"})
		return
	}
	defer file.Close()

	if imageStore == nil {
		log.Println("Image store not initialized")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Image store not initialized"})
		return
	}

	imageURL, err := imageStore.SaveImage(userID, fileHeader.Filename, file)
	if err != nil {
		log.Printf("Error saving image: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not upload image"})
		return
	}

	_, err = profilesCollection.UpdateOne(
		context.Background(),
		bson.M{"user_id": userID},
		bson.M{"$set": bson.M{"profile_img": imageURL}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		log.Printf("Error updating profile image in database: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not update profile image"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"profileImage": imageURL})
}

// PutProfile updates the profile of the given user.
//
//	@Summary		Update a user's profile.
//	@Description	Updates the profile of the user with the specified user ID using the provided profile data.
//	@Tags			profile
//	@Security		BearerAuth
//	@ID				update-profile
//	@Param			userid	path		string			true	"The ID of the user whose profile to update"
//	@Param			request	body		Profile			true	"Profile object that needs to be updated"
//	@Success		200		{string}	string			"Profile updated"
//	@Failure		400		{object}	ErrorResponse	"Invalid request body"
//	@Failure		401		{object}	ErrorResponse	"Not authenticated"
//	@Failure		500		{object}	ErrorResponse	"Could not update profile"
//	@Router			/profile/{userid} [put]
func PutProfile(c *gin.Context) {
	userID := c.Param("userid")

	fmt.Println("Put Profile")
	fmt.Println(userID)

	var profile Profile
	if err := c.BindJSON(&profile); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	profile.UserID = userID

	// Print out the profile json encoded
	profileJSON, err2 := json.Marshal(profile)
	if err2 != nil {
		log.Panicln("Error marshalling profile: ", err2)
	}
	fmt.Println(string(profileJSON))

	// Update the profile in the database
	_, err := profilesCollection.UpdateOne(context.Background(), bson.M{"user_id": userID}, bson.M{"$set": profile}, options.Update().SetUpsert(true))
	if err != nil {
		log.Panicln("Database Error: ", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated"})
}

// PostProfile creates a new profile for the given user.
//
//	@Summary		Create a new user profile.
//	@Description	Creates a new profile for the user with the specified user ID using the provided profile data.
//	@Tags			profile
//	@Security		BearerAuth
//	@ID				create-profile
//	@Param			userid	path		string			true	"The ID of the user for whom the profile is to be created"
//	@Param			request	body		Profile			true	"Profile object that needs to be created"
//	@Success		201		{string}	string			"Profile created"
//	@Failure		400		{object}	ErrorResponse	"Invalid request body"
//	@Failure		401		{object}	ErrorResponse	"Not authenticated"
//	@Failure		500		{object}	ErrorResponse	"Could not create profile"
//	@Router			/profile/{userid} [post]
func PostProfile(c *gin.Context) {
	userID := c.Param("userid")
	log.Printf("%v", userID)
	var req Profile
	if err := c.BindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	req.UserID = userID

	_, err := profilesCollection.InsertOne(context.Background(), req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not create profile"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Profile created"})
}

// InitializeRoutes initializes the profile routes.
func InitializeRoutes(router *gin.RouterGroup, db *mongo.Client, db_name string) {
	profilesCollection = db.Database(db_name).Collection("profiles")

	router.GET("/:userid", GetProfile)

	protected := router.Group("/")
	protected.Use(auth.AuthMiddleware(db, db_name, true))
	protected.PUT("/:userid", PutProfile)
	protected.PUT("/:userid/image", PutImage)
	protected.POST("/:userid", PostProfile)
}

func init() {
	if err := InitImageStore(); err != nil {
		log.Fatalf("Failed to initialize image store: %v", err)
	}
}
