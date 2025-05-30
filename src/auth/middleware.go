package auth

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func AuthMiddleware(db *mongo.Client, dbName string, required bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("token")
		if err != nil {
			if required {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
				return
			}
			c.Next()
			return
		}

		claims := &Claims{}
		t, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("secret"), nil
		})
		if err != nil || !t.Valid {
			if required {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
				return
			}
			c.Next()
			return
		}

		// Check if the user exists
		usersCollection := db.Database(dbName).Collection("users")
		var user User
		err = usersCollection.FindOne(context.Background(), bson.M{"_id": claims.Id}).Decode(&user)
		if err != nil {
			if required {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
				return
			}
			c.Next()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}
