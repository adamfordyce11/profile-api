package certificates

// Certificate represents a user's certification
type Certificate struct {
	UserID        string `bson:"user_id" json:"user_id"`
	CertificateID string `bson:"certificate_id" json:"certificate_id"`
	Title         string `bson:"title" json:"title"`
	Institution   string `bson:"institution" json:"institution"`
	Start         string `bson:"start" json:"start"`
	End           string `bson:"end" json:"end"`
	Description   string `bson:"description" json:"description"`
}
