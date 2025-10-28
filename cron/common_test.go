package cron

import (
	"context"
	"testing"

	"github.com/geerew/off-course/dao"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/appfs"
	"github.com/geerew/off-course/utils/logger"
	"github.com/geerew/off-course/utils/types"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func setup(t *testing.T) (database.Database, *appfs.AppFs, context.Context, *logger.Logger) {
	t.Helper()

	// Create a test logger
	testLogger := logger.New(&logger.Config{
		Level:         logger.LevelInfo,
		ConsoleOutput: false, // Disable console output for tests
	})

	appFs := appfs.New(afero.NewMemMapFs())

	dbManager, err := database.NewSQLiteManager(&database.DatabaseManagerConfig{
		DataDir: "./oc_data",
		AppFs:   appFs,
		Testing: true,
	})

	require.NoError(t, err)
	require.NotNil(t, dbManager)

	dao := dao.New(dbManager.DataDb)

	// User
	user := &models.User{
		Username:     "test-user",
		DisplayName:  "Test User",
		PasswordHash: "test-password",
		Role:         types.UserRoleAdmin,
	}
	require.NoError(t, dao.CreateUser(context.Background(), user))

	principal := types.Principal{
		UserID: user.ID,
		Role:   user.Role,
	}

	ctx := context.WithValue(context.Background(), types.PrincipalContextKey, principal)

	return dbManager.DataDb, appFs, ctx, testLogger
}
