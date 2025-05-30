package skills

// Skill represents a user's skill
type Skill struct {
	UserID           string `bson:"user_id" json:"user_id"`
	SkillID          string `bson:"skill_id" json:"skill_id"`
	Name             string `bson:"name" json:"name"`
	ProficiencyLevel string `bson:"proficiency_level" json:"proficiency_level"`
	StartedAt        string `bson:"started_at" json:"started_at"`
	LastUsed         string `bson:"last_used" json:"last_used"`
	Description      string `bson:"description" json:"description"`
}
