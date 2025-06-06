package session

import (
	"context"
	"testing"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/appfs"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func setup(tb testing.TB) (database.Database, context.Context) {
	tb.Helper()

	dbManager, err := database.NewSQLiteManager(&database.DatabaseManagerConfig{
		DataDir: "./oc_data",
		AppFs:   appfs.New(afero.NewMemMapFs(), nil),
		Testing: true,
	})

	require.NoError(tb, err)
	require.NotNil(tb, dbManager)

	return dbManager.DataDb, context.Background()
}
