package dao

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/appfs"
	"github.com/geerew/off-course/utils/pagination"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func setupLog(tb testing.TB) (*DAO, context.Context) {
	tb.Helper()

	// Filesystem
	appFs := appfs.New(afero.NewMemMapFs())

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

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_CreateLogsBatch(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dao, ctx := setupLog(t)

		logs := []*models.Log{}
		for i := range 5 {
			log := &models.Log{
				Data:    map[string]any{"test": i},
				Level:   i % 3,
				Message: fmt.Sprintf("batch log %d", i+1),
			}
			logs = append(logs, log)
		}

		require.NoError(t, dao.CreateLogsBatch(ctx, logs))

		// Verify all logs were inserted
		records, err := dao.ListLogs(ctx, nil)
		require.NoError(t, err)
		require.Len(t, records, 5)

		// Check that all messages are present
		messages := make(map[string]bool)
		for _, record := range records {
			messages[record.Message] = true
		}

		for _, log := range logs {
			require.True(t, messages[log.Message], "Message %s should be present", log.Message)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		dao, ctx := setupLog(t)

		logs := []*models.Log{}
		require.NoError(t, dao.CreateLogsBatch(ctx, logs))

		// Verify no logs were inserted
		records, err := dao.ListLogs(ctx, nil)
		require.NoError(t, err)
		require.Empty(t, records)
	})

	t.Run("nil pointer in slice", func(t *testing.T) {
		dao, ctx := setupLog(t)

		logs := []*models.Log{
			{Data: map[string]any{}, Level: 0, Message: "valid log"},
			nil,
			{Data: map[string]any{}, Level: 1, Message: "another valid log"},
		}

		require.ErrorIs(t, dao.CreateLogsBatch(ctx, logs), utils.ErrNilPtr)
	})

	t.Run("invalid message in slice", func(t *testing.T) {
		dao, ctx := setupLog(t)

		logs := []*models.Log{
			{Data: map[string]any{}, Level: 0, Message: "valid log"},
			{Data: map[string]any{}, Level: 0, Message: ""}, // Invalid
			{Data: map[string]any{}, Level: 1, Message: "another valid log"},
		}

		require.ErrorIs(t, dao.CreateLogsBatch(ctx, logs), utils.ErrLogMessage)
	})

	t.Run("large batch", func(t *testing.T) {
		dao, ctx := setupLog(t)

		logs := []*models.Log{}
		for i := range 100 {
			log := &models.Log{
				Data:    map[string]any{"index": i},
				Level:   i % 3,
				Message: fmt.Sprintf("large batch log %d", i+1),
			}
			logs = append(logs, log)
		}

		require.NoError(t, dao.CreateLogsBatch(ctx, logs))

		// Verify all logs were inserted
		records, err := dao.ListLogs(ctx, nil)
		require.NoError(t, err)
		require.Len(t, records, 100)
	})
}
