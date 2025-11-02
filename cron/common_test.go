package cron

import (
	"context"
	"testing"

	"github.com/geerew/off-course/app"
	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/types"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func setup(t *testing.T) (*app.App, context.Context) {
	t.Helper()

	application := app.NewTestApp(t)

	// Create DAO to create a user
	appDao := dao.New(application.DbManager.DataDb)

	// User
	user := &models.User{
		Username:     "test-user",
		DisplayName:  "Test User",
		PasswordHash: "test-password",
		Role:         types.UserRoleAdmin,
	}
	require.NoError(t, appDao.CreateUser(context.Background(), user))

	principal := types.Principal{
		UserID: user.ID,
		Role:   user.Role,
	}

	ctx := context.WithValue(context.Background(), types.PrincipalContextKey, principal)

	return application, ctx
}
