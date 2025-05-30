package qualifications

// Qualification represents a user's qualification
type Qualification struct {
	UserID          string `bson:"user_id" json:"user_id"`
	QualificationID string `bson:"qualification_id" json:"qualification_id"`
	Title           string `bson:"title" json:"title"`
	Institution     string `bson:"institution" json:"institution"`
	Start           string `bson:"start" json:"start"`
	End             string `bson:"end" json:"end"`
	Description     string `bson:"description" json:"description"`
}
