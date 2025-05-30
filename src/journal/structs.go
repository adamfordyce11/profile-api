package journal

import "time"

// JournalEntry represents a user's journal entry
type JournalEntry struct {
	JournalID string    `bson:"journal_id" json:"journalID"`
	UserID    string    `bson:"user_id" json:"userID"`
	Version   int       `bson:"version" json:"version"`
	Entries   []Entry   `bson:"entries" json:"entries"`
	Status    string    `bson:"status" json:"status"`
	Taxonomy  Taxonomy  `bson:"taxonomy" json:"taxonomy"`
	Summary   string    `bson:"summary" json:"summary"`
	CreatedAt time.Time `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time `bson:"updated_at" json:"updatedAt"`
}

// Entry represents a versioned entry in the journal
type Entry struct {
	Version     int       `bson:"version" json:"version"`
	Title       string    `bson:"title" json:"title"`
	Content     string    `bson:"content" json:"content"`
	Attachments []string  `bson:"attachments" json:"attachments"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updatedAt"`
}

// Taxonomy represents categories, subcategories, topics, and tags for the journal entry
type Taxonomy struct {
	Categories    []string `bson:"categories" json:"categories"`
	Subcategories []string `bson:"subcategories" json:"subcategories"`
	Topics        []string `bson:"topics" json:"topics"`
	Tags          []string `bson:"tags" json:"tags"`
}
