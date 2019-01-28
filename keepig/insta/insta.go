package insta

import "github.com/ahmdrz/goinsta"

// User represents a user
type User struct {
	ID         int64
	Username   string
	ProfileID  string
	ProfileURL string
}

// Instagram encapsulates goinsta
type Instagram struct {
	*goinsta.Instagram
}

// NewInstagram creates a new Instagram instance
func NewInstagram(user, pass string) *Instagram {
	return &Instagram{goinsta.New(user, pass)}
}

// Import deserializes into struct
func (ig *Instagram) Import() {

}

// Export serializes the underlying struct
func (ig *Instagram) Export() {

}

// GetFollowersAll gets all your followers
func (ig *Instagram) GetFollowersAll() ([]User, error) {
	res, err := ig.SelfTotalUserFollowing()
	if err != nil {
		return []User{}, err
	}

	users := make([]User, len(res.Users))
	for i, user := range res.Users {
		users[i] = User{
			ID:         user.ID,
			Username:   user.Username,
			ProfileID:  user.ProfilePictureID,
			ProfileURL: user.ProfilePictureURL,
		}
	}
	return users, nil
}

// CreateFeed creates a feed object
func (ig *Instagram) CreateFeed(userID int64) *Feed {
	return &Feed{
		ig: ig.Instagram,
		ID: userID,
	}
}
