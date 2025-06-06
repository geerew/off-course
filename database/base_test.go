package database

import (
	"testing"

	"github.com/geerew/off-course/utils/appfs"
	"github.com/geerew/off-course/utils/logger"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_NewSQLiteManager(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		logger, _, err := logger.InitLogger(&logger.BatchOptions{
			BatchSize: 1,
			WriteFn:   logger.NilWriteFn(),
		})
		require.NoError(t, err)

		appFs := appfs.New(afero.NewMemMapFs(), logger)

		dbManager, err := NewSQLiteManager(&DatabaseManagerConfig{
			DataDir: "./oc_data",
			AppFs:   appFs,
			Testing: true,
		})

		require.NoError(t, err)
		require.NotNil(t, dbManager)

	})

	t.Run("error creating data.db", func(t *testing.T) {
		logger, _, err := logger.InitLogger(&logger.BatchOptions{
			BatchSize: 1,
			WriteFn:   logger.NilWriteFn(),
		})
		require.NoError(t, err)

		appFs := appfs.New(afero.NewReadOnlyFs(afero.NewMemMapFs()), logger)

		dbManager, err := NewSQLiteManager(&DatabaseManagerConfig{
			DataDir: "./oc_data",
			AppFs:   appFs,
			Testing: true,
		})

		require.NotNil(t, err)
		require.EqualError(t, err, "operation not permitted")
		require.Nil(t, dbManager)
	})
}
