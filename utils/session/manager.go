package session

import (
	"context"
	"fmt"

	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/types"
	"github.com/gofiber/fiber/v2"
	fs "github.com/gofiber/fiber/v2/middleware/session"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Storage is an interface that is implemented by storage providers. It extends the fiber.Storage
// interface
type Storage interface {
	fiber.Storage
	DeleteUser(userId string) error
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type SessionManager struct {
	dao        *dao.DAO
	fiberStore *fs.Store
	storage    Storage
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// New creates a new session manager. It is essentially a wrapper around the fiber session store
func New(db database.Database, config fs.Config, storage Storage) *SessionManager {
	if storage != nil {
		config.Storage = storage
	}

	sessionManager := &SessionManager{
		dao:        dao.New(db),
		fiberStore: fs.New(config),
		storage:    storage,
	}

	return sessionManager
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get gets the session for a user
func (s *SessionManager) Get(c *fiber.Ctx) (*fs.Session, error) {
	return s.fiberStore.Get(c)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SetSession sets the session for a user
func (s *SessionManager) SetSession(c *fiber.Ctx, userId string, userRole types.UserRole) error {
	session, err := s.Get(c)
	if err != nil {
		return err
	}

	// Prevent session fixation on login
	if err := session.Regenerate(); err != nil {
		return err
	}

	fmt.Println("userid", userId, "role", userRole.String())
	session.Set("id", userId)
	session.Set("role", userRole.String())

	return session.Save()

	// // Update the user_id in the session. This is an extra field that makes it easier to look up
	// // sessions by user ID
	// //
	// // It must be done AFTER the session is saved, otherwise the session will not exist
	// return s.storage.SetUser(sessionId, userId)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteAllSessions deletes all sessions from the storage
func (s *SessionManager) DeleteAllSessions() error {
	return s.fiberStore.Storage.Reset()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteSession deletes a single session based on the session ID
func (s *SessionManager) DeleteSession(c *fiber.Ctx) error {
	session, err := s.Get(c)
	if err != nil {
		return err
	}

	return session.Destroy()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteUserSessions deletes all sessions for a user
func (s *SessionManager) DeleteUserSessions(id string) error {
	return s.storage.DeleteUser(id)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateSessionRoleForUser updates the role for all sessions belonging to a user
func (s *SessionManager) UpdateSessionRoleForUser(userID string, newRole types.UserRole) error {
	return s.dao.UpdateSessionRoleForUser(context.Background(), userID, newRole)
}
