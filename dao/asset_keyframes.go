package dao

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateAssetKeyframes inserts a new asset keyframes record
func (dao *DAO) CreateAssetKeyframes(ctx context.Context, keyframes *models.AssetKeyframes) error {
	if keyframes == nil {
		return utils.ErrNilPtr
	}

	if keyframes.AssetID == "" {
		return utils.ErrAssetId
	}

	// Marshal keyframes to JSON before inserting
	if err := keyframes.MarshalKeyframes(); err != nil {
		return fmt.Errorf("failed to marshal keyframes: %w", err)
	}

	// Validate keyframes before storing
	if err := keyframes.ValidateKeyframes(); err != nil {
		return fmt.Errorf("invalid keyframes: %w", err)
	}

	if keyframes.ID == "" {
		keyframes.RefreshId()
	}

	keyframes.RefreshCreatedAt()
	keyframes.RefreshUpdatedAt()

	builderOpts := newBuilderOptions(models.ASSET_KEYFRAMES_TABLE).
		WithData(
			map[string]interface{}{
				models.BASE_ID:               keyframes.ID,
				models.KEYFRAMES_ASSET_ID:    keyframes.AssetID,
				models.KEYFRAMES_DATA:        keyframes.KeyframesJSON,
				models.KEYFRAMES_IS_COMPLETE: keyframes.IsComplete,
				models.BASE_CREATED_AT:       keyframes.CreatedAt,
				models.BASE_UPDATED_AT:       keyframes.UpdatedAt,
			})

	return createGeneric(ctx, dao, *builderOpts)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAssetKeyframes retrieves asset keyframes by asset ID
func (dao *DAO) GetAssetKeyframes(ctx context.Context, assetID string) (*models.AssetKeyframes, error) {
	if assetID == "" {
		return nil, utils.ErrAssetId
	}

	query := squirrel.Select(models.AssetKeyframesColumns()...).
		From(models.ASSET_KEYFRAMES_TABLE).
		Where(squirrel.Eq{models.KEYFRAMES_ASSET_ID: assetID})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	querier := database.QuerierFromContext(ctx, dao.db)
	row := querier.QueryRowContext(ctx, sql, args...)

	var keyframes models.AssetKeyframes
	err = row.Scan(
		&keyframes.ID,
		&keyframes.CreatedAt,
		&keyframes.UpdatedAt,
		&keyframes.AssetID,
		&keyframes.KeyframesJSON,
		&keyframes.IsComplete,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan keyframes: %w", err)
	}

	// Unmarshal the JSON keyframes
	if err := keyframes.UnmarshalKeyframes(); err != nil {
		return nil, fmt.Errorf("failed to unmarshal keyframes: %w", err)
	}

	return &keyframes, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateAssetKeyframes updates an existing asset keyframes record
func (dao *DAO) UpdateAssetKeyframes(ctx context.Context, keyframes *models.AssetKeyframes) error {
	if keyframes == nil {
		return utils.ErrNilPtr
	}

	if keyframes.ID == "" {
		return utils.ErrId
	}

	if keyframes.AssetID == "" {
		return utils.ErrAssetId
	}

	// Marshal keyframes to JSON before updating
	if err := keyframes.MarshalKeyframes(); err != nil {
		return fmt.Errorf("failed to marshal keyframes: %w", err)
	}

	// Validate keyframes before storing
	if err := keyframes.ValidateKeyframes(); err != nil {
		return fmt.Errorf("invalid keyframes: %w", err)
	}

	keyframes.RefreshUpdatedAt()

	builderOpts := newBuilderOptions(models.ASSET_KEYFRAMES_TABLE).
		WithData(
			map[string]interface{}{
				models.KEYFRAMES_DATA:        keyframes.KeyframesJSON,
				models.KEYFRAMES_IS_COMPLETE: keyframes.IsComplete,
				models.BASE_UPDATED_AT:       keyframes.UpdatedAt,
			}).
		SetDbOpts(database.NewOptions().WithWhere(squirrel.Eq{models.BASE_ID: keyframes.ID}))

	_, err := updateGeneric(ctx, dao, *builderOpts)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteAssetKeyframes deletes asset keyframes by asset ID
func (dao *DAO) DeleteAssetKeyframes(ctx context.Context, assetID string) error {
	if assetID == "" {
		return utils.ErrAssetId
	}

	builderOpts := newBuilderOptions(models.ASSET_KEYFRAMES_TABLE).
		SetDbOpts(database.NewOptions().WithWhere(squirrel.Eq{models.KEYFRAMES_ASSET_ID: assetID}))

	sqlStr, args, err := deleteBuilder(*builderOpts)
	if err != nil {
		return err
	}

	q := database.QuerierFromContext(ctx, dao.db)
	_, err = q.ExecContext(ctx, sqlStr, args...)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteAssetKeyframesById deletes asset keyframes by ID
func (dao *DAO) DeleteAssetKeyframesById(ctx context.Context, id string) error {
	if id == "" {
		return utils.ErrId
	}

	builderOpts := newBuilderOptions(models.ASSET_KEYFRAMES_TABLE).
		SetDbOpts(database.NewOptions().WithWhere(squirrel.Eq{models.BASE_ID: id}))

	sqlStr, args, err := deleteBuilder(*builderOpts)
	if err != nil {
		return err
	}

	q := database.QuerierFromContext(ctx, dao.db)
	_, err = q.ExecContext(ctx, sqlStr, args...)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ListAssetKeyframes retrieves asset keyframes with optional filtering
func (dao *DAO) ListAssetKeyframes(ctx context.Context, opts *database.Options) ([]*models.AssetKeyframes, error) {
	query := squirrel.Select(models.AssetKeyframesColumns()...).
		From(models.ASSET_KEYFRAMES_TABLE)

	if opts != nil {
		if opts.Where != nil {
			query = query.Where(opts.Where)
		}
		if opts.OrderBy != nil {
			query = query.OrderBy(opts.OrderBy...)
		}
		if opts.Pagination != nil {
			if opts.Pagination.Limit() > 0 {
				query = query.Limit(uint64(opts.Pagination.Limit()))
			}
			if opts.Pagination.Offset() > 0 {
				query = query.Offset(uint64(opts.Pagination.Offset()))
			}
		}
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	querier := database.QuerierFromContext(ctx, dao.db)
	rows, err := querier.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var keyframesList []*models.AssetKeyframes
	for rows.Next() {
		var keyframes models.AssetKeyframes
		err := rows.Scan(
			&keyframes.ID,
			&keyframes.CreatedAt,
			&keyframes.UpdatedAt,
			&keyframes.AssetID,
			&keyframes.KeyframesJSON,
			&keyframes.IsComplete,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan keyframes: %w", err)
		}

		// Unmarshal the JSON keyframes
		if err := keyframes.UnmarshalKeyframes(); err != nil {
			return nil, fmt.Errorf("failed to unmarshal keyframes: %w", err)
		}

		keyframesList = append(keyframesList, &keyframes)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return keyframesList, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ExistsAssetKeyframes checks if keyframes exist for the given asset ID
func (dao *DAO) ExistsAssetKeyframes(ctx context.Context, assetID string) (bool, error) {
	if assetID == "" {
		return false, utils.ErrAssetId
	}

	query := squirrel.Select("COUNT(*)").
		From(models.ASSET_KEYFRAMES_TABLE).
		Where(squirrel.Eq{models.KEYFRAMES_ASSET_ID: assetID})

	sql, args, err := query.ToSql()
	if err != nil {
		return false, fmt.Errorf("failed to build query: %w", err)
	}

	querier := database.QuerierFromContext(ctx, dao.db)
	var count int
	err = querier.QueryRowContext(ctx, sql, args...).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to count keyframes: %w", err)
	}

	return count > 0, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetAssetKeyframesCount returns the total number of asset keyframes records
func (dao *DAO) GetAssetKeyframesCount(ctx context.Context, opts *database.Options) (int, error) {
	query := squirrel.Select("COUNT(*)").
		From(models.ASSET_KEYFRAMES_TABLE)

	if opts != nil && opts.Where != nil {
		query = query.Where(opts.Where)
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build query: %w", err)
	}

	querier := database.QuerierFromContext(ctx, dao.db)
	var count int
	err = querier.QueryRowContext(ctx, sql, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count keyframes: %w", err)
	}

	return count, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpsertAssetKeyframes creates or updates asset keyframes
func (dao *DAO) UpsertAssetKeyframes(ctx context.Context, keyframes *models.AssetKeyframes) error {
	if keyframes == nil {
		return utils.ErrNilPtr
	}

	if keyframes.AssetID == "" {
		return utils.ErrAssetId
	}

	// Check if keyframes already exist
	exists, err := dao.ExistsAssetKeyframes(ctx, keyframes.AssetID)
	if err != nil {
		return fmt.Errorf("failed to check if keyframes exist: %w", err)
	}

	if exists {
		// Get existing record to preserve ID and timestamps
		existing, err := dao.GetAssetKeyframes(ctx, keyframes.AssetID)
		if err != nil {
			return fmt.Errorf("failed to get existing keyframes: %w", err)
		}

		// Preserve the existing ID and created_at
		keyframes.ID = existing.ID
		keyframes.CreatedAt = existing.CreatedAt

		return dao.UpdateAssetKeyframes(ctx, keyframes)
	} else {
		return dao.CreateAssetKeyframes(ctx, keyframes)
	}
}
