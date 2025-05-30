package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var usersCollection *mongo.Collection

// ErrorResponse is a struct that represents an error response.
//
// swagger:model ErrorResponse
type ErrorResponse struct {
	// Error message
	// example: Invalid request body
	Error string `json:"error"`
}

// @Summary		Register
// @Description	Register a new user
// @Tags			Auth
// @Accept			json
// @Produce		json
// @Param			register	body		RegisterRequest	true	"Registration request object"
// @Success		201			{string}	string			"User created"
// @Failure		400			{object}	ErrorResponse
// @Failure		409			{object}	ErrorResponse
// @Failure		500			{object}	ErrorResponse
// @Router			/auth/register [post]
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.BindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not hash password"})
		return
	}

	// Check if the email is already registered
	var existingUser User
	err = usersCollection.FindOne(context.Background(), bson.M{"email": req.Email}).Decode(&existingUser)
	if err != nil && err != mongo.ErrNoDocuments {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not check email existence"})
		return
	}
	if existingUser.Email != "" {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	// Create the new user
	newUser := User{
		ID:       primitive.NewObjectID().Hex(),
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}
	_, err = usersCollection.InsertOne(context.Background(), newUser)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created"})
}

// @Summary		Login
// @Description	Login a user
// @Tags			Auth
// @Accept			json
// @Produce		json
// @Param			login	body		LoginRequest	true	"Login request object"
// @Success		200		{string}	string			"Token"
// @Failure		400		{object}	ErrorResponse "Invalid request body"
// @Failure		401		{object}	ErrorResponse "Invalid email or password"
// @Router			/auth/login [post]
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.BindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Find the user by email
	var user User
	err := usersCollection.FindOne(context.Background(), bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Check the password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Create a JWT token and return it to the client
	token := createToken(user.ID)
	c.SetCookie("token", token, 3600, "", "", false, true)
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// @Summary		Logout
// @Description	Logout the currently logged in user
// @Tags			Auth
// @Produce		json
// @Success		200	{string}	string	"Logged out"
// @Router			/auth/logout [post]
func Logout(c *gin.Context) {
	c.SetCookie("token", "", -1, "", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}

// InitializeRoutes initializes the authentication routes
func InitializeRoutes(router *gin.RouterGroup, db *mongo.Client, db_name string) {
	usersCollection = db.Database(db_name).Collection("users")
	router.POST("/register", Register)
	router.POST("/login", Login)
	router.POST("/logout", Logout)
}

// createToken creates a new JWT token for the given user ID
func createToken(userID string) string {
	claims := jwt.StandardClaims{
		Id:        userID,
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)
	signedToken, _ := token.SignedString([]byte("secret"))
	return signedToken
}
