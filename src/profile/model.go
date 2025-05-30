package profile

// Profile represents a user's profile information
type Profile struct {
	UserID     string  `bson:"user_id"	json:"userid"`
	Name       *string `bson:"name" json:"name"`
	Email      *string `bson:"email" json:"email"`
	Number     *string `bson:"number" json:"number"`
	Bio        *string `bson:"bio" json:"bio"`
	ProfileImg *string `bson:"profile_img" json:"profile_img"`
	Interests  *string `bson:"interests" json:"interests"`
	Domain     *string `bson:"domain" json:"domain"`
}
