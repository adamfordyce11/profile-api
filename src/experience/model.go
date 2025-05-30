package experience

// Experience represents a user's work experience
type Experience struct {
	UserID       string `bson:"user_id" json:"user_id"`
	ExperienceID string `bson:"experience_id" json:"experience_id"`
	Company      string `bson:"company" json:"company"`
	Position     string `bson:"position" json:"position"`
	Start        string `bson:"start" json:"start"`
	End          string `bson:"end" json:"end"`
	Description  string `bson:"description" json:"description"`
	Notes        string `bson:"notes" json:"notes"`
}
