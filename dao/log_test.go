package dao

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/appfs"
	"github.com/geerew/off-course/utils/logger"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func setupLog(tb testing.TB) (*DAO, context.Context) {
	tb.Helper()

	// Logger
	var logs []*logger.Log
	var logsMux sync.Mutex
	logger, _, err := logger.InitLogger(&logger.BatchOptions{
		BatchSize: 1,
		WriteFn:   logger.TestWriteFn(&logs, &logsMux),
	})
	require.NoError(tb, err, "Failed to initialize logger")

	// Filesystem
	appFs := appfs.New(afero.NewMemMapFs(), logger)

	// DB
	dbManager, err := database.NewSQLiteManager(&database.DatabaseManagerConfig{
		DataDir: "./oc_data",
		AppFs:   appFs,
		Testing: true,
	})

	require.NoError(tb, err)
	require.NotNil(tb, dbManager)

	dao := &DAO{db: dbManager.LogsDb}

	return dao, context.Background()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CreateLog(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setupLog(t)

		log := &models.Log{Data: map[string]any{}, Level: 0, Message: fmt.Sprintf("log %d", 1)}
		require.NoError(t, dao.CreateLog(ctx, log))
	})

	t.Run("nil pointer", func(t *testing.T) {
		dao, ctx := setupLog(t)

		require.ErrorIs(t, dao.CreateLog(ctx, nil), utils.ErrNilPtr)
	})

	t.Run("invalid message", func(t *testing.T) {
		dao, ctx := setupLog(t)

		log := &models.Log{Data: map[string]any{}, Level: 0, Message: ""}
		require.ErrorIs(t, dao.CreateLog(ctx, log), utils.ErrLogMessage)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_GetLog(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setupLog(t)

		log := &models.Log{Data: map[string]any{}, Level: 0, Message: fmt.Sprintf("log %d", 1)}
		require.NoError(t, dao.CreateLog(ctx, log))

		dbOpts := database.NewOptions().WithWhere(squirrel.Eq{models.LOG_TABLE_ID: log.ID})
		record, err := dao.GetLog(ctx, dbOpts)
		require.Nil(t, err)
		require.Equal(t, log.ID, record.ID)
	})

	t.Run("not found", func(t *testing.T) {
		dao, ctx := setupLog(t)

		record, err := dao.GetLog(ctx, nil)
		require.Nil(t, err)
		require.Nil(t, record)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_ListLogs(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setupLog(t)

		logs := []*models.Log{}

		for i := range 3 {
			log := &models.Log{Data: map[string]any{}, Level: 0, Message: fmt.Sprintf("log %d", i+1)}
			logs = append(logs, log)
			require.NoError(t, dao.CreateLog(ctx, log))
			time.Sleep(1 * time.Millisecond)
		}

		records, err := dao.ListLogs(ctx, nil)
		require.Nil(t, err)
		require.Len(t, records, 3)

		for i, record := range records {
			require.Equal(t, logs[i].ID, record.ID)
		}
	})

	t.Run("empty", func(t *testing.T) {
		dao, ctx := setupLog(t)

		records, err := dao.ListLogs(ctx, nil)
		require.Nil(t, err)
		require.Empty(t, records)
	})

	t.Run("order by", func(t *testing.T) {
		dao, ctx := setupLog(t)

		logs := []*models.Log{}
		for i := range 3 {
			log := &models.Log{Data: map[string]any{}, Level: 0, Message: fmt.Sprintf("log %d", i+1)}
			logs = append(logs, log)
			require.NoError(t, dao.CreateLog(ctx, log))
			time.Sleep(1 * time.Millisecond)
		}

		// Descending order by created_at
		opts := database.NewOptions().WithOrderBy(models.LOG_TABLE_CREATED_AT + " DESC")

		records, err := dao.ListLogs(ctx, opts)
		require.Nil(t, err)
		require.Len(t, records, 3)

		for i, record := range records {
			require.Equal(t, logs[2-i].ID, record.ID)
		}

		// Ascending order by created_at
		opts = database.NewOptions().WithOrderBy(models.LOG_TABLE_CREATED_AT + " ASC")

		records, err = dao.ListLogs(ctx, opts)
		require.Nil(t, err)
		require.Len(t, records, 3)

		for i, record := range records {
			require.Equal(t, logs[i].ID, record.ID)
		}
	})

	t.Run("where", func(t *testing.T) {
		dao, ctx := setupLog(t)

		log := &models.Log{Data: map[string]any{}, Level: 0, Message: fmt.Sprintf("log %d", 1)}
		require.NoError(t, dao.CreateLog(ctx, log))

		opts := database.NewOptions().WithWhere(squirrel.Eq{models.LOG_TABLE_ID: log.ID})
		records, err := dao.ListLogs(ctx, opts)
		require.Nil(t, err)
		require.Len(t, records, 1)
		require.Equal(t, log.ID, records[0].ID)
	})

	t.Run("pagination", func(t *testing.T) {
		dao, ctx := setupLog(t)

		logs := []*models.Log{}
		for i := range 17 {
			log := &models.Log{Data: map[string]any{}, Level: 0, Message: fmt.Sprintf("log %d", i+1)}
			logs = append(logs, log)
			require.NoError(t, dao.CreateLog(ctx, log))
			time.Sleep(1 * time.Millisecond)
		}

		// First page with 10 records
		p := database.NewOptions().WithPagination(pagination.New(1, 10))
		records, err := dao.ListLogs(ctx, p)
		require.Nil(t, err)
		require.Len(t, records, 10)
		require.Equal(t, logs[0].ID, records[0].ID)
		require.Equal(t, logs[9].ID, records[9].ID)

		// Second page with remaining 7 records
		p = database.NewOptions().WithPagination(pagination.New(2, 10))
		records, err = dao.ListLogs(ctx, p)
		require.Nil(t, err)
		require.Len(t, records, 7)
		require.Equal(t, logs[10].ID, records[0].ID)
		require.Equal(t, logs[16].ID, records[6].ID)
	})
}
