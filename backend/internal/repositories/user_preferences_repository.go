// internal/repositories/user_preferences_repository.go
// User preferences data access (MongoDB)

package repositories

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserPreferencesRepository handles user preferences in MongoDB
type UserPreferencesRepository struct {
	collection *mongo.Collection
}

// NewUserPreferencesRepository creates a new user preferences repository
func NewUserPreferencesRepository(db *mongo.Database) *UserPreferencesRepository {
	return &UserPreferencesRepository{
		collection: db.Collection("user_preferences"),
	}
}

// Get retrieves user preferences
func (r *UserPreferencesRepository) Get(ctx context.Context, userID string) (map[string]interface{}, error) {
	var prefs map[string]interface{}
	err := r.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&prefs)
	if err == mongo.ErrNoDocuments {
		return nil, nil // Return nil if no preferences found
	}
	return prefs, err
}

// Set creates or updates user preferences
func (r *UserPreferencesRepository) Set(ctx context.Context, userID string, preferences map[string]interface{}) error {
	preferences["user_id"] = userID
	preferences["updated_at"] = time.Now()

	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"user_id": userID},
		bson.M{"$set": preferences},
		opts,
	)
	return err
}

// Update partially updates user preferences
func (r *UserPreferencesRepository) Update(ctx context.Context, userID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"user_id": userID},
		bson.M{"$set": updates},
	)
	return err
}

// Delete removes user preferences
func (r *UserPreferencesRepository) Delete(ctx context.Context, userID string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"user_id": userID})
	return err
}
