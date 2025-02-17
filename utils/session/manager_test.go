package session

import (
	"context"
	"testing"

	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/utils/appFs"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func setup(tb testing.TB) (database.Database, context.Context) {
	tb.Helper()

	dbManager, err := database.NewSqliteDBManager(&database.DatabaseConfig{
		DataDir:  "./oc_data",
		AppFs:    appFs.NewAppFs(afero.NewMemMapFs(), nil),
		InMemory: true,
	})

	require.NoError(tb, err)
	require.NotNil(tb, dbManager)

	return dbManager.DataDb, context.Background()
}
