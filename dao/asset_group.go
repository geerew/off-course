package dao

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/geerew/off-course/database"
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// CreateAssetGroup inserts a new asset group record
func (dao *DAO) CreateAssetGroup(ctx context.Context, assetGroup *models.AssetGroup) error {
	if err := assetGroupValidation(assetGroup); err != nil {
		return err
	}

	if assetGroup.ID == "" {
		assetGroup.RefreshId()
	}

	assetGroup.RefreshCreatedAt()
	assetGroup.RefreshUpdatedAt()

	builderOptions := newBuilderOptions(models.ASSET_GROUP_TABLE).
		WithData(map[string]interface{}{
			models.BASE_ID:                      assetGroup.ID,
			models.ASSET_GROUP_COURSE_ID:        assetGroup.CourseID,
			models.ASSET_GROUP_TITLE:            assetGroup.Title,
			models.ASSET_GROUP_PREFIX:           assetGroup.Prefix,
			models.ASSET_GROUP_MODULE:           assetGroup.Module,
			models.ASSET_GROUP_DESCRIPTION_PATH: assetGroup.DescriptionPath,
			models.ASSET_GROUP_DESCRIPTION_TYPE: assetGroup.DescriptionType,
			models.BASE_CREATED_AT:              assetGroup.CreatedAt,
			models.BASE_UPDATED_AT:              assetGroup.UpdatedAt,
		})

	return createGeneric(ctx, dao, *builderOptions)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// GetAssetGroup gets a record from the asset groups table based upon the where clause in the options. If
// there is no where clause, it will return the first record in the table
func (dao *DAO) GetAssetGroup(ctx context.Context, dbOpts *database.Options) (*models.AssetGroup, error) {
	// Fetch asset group
	builderOpts := newBuilderOptions(models.ASSET_GROUP_TABLE).
		WithColumns(models.ASSET_GROUP_TABLE + ".*").
		SetDbOpts(dbOpts).
		WithLimit(1)

	assetGroup, err := getGeneric[models.AssetGroup](ctx, dao, *builderOpts)
	if err != nil {
		return nil, err
	}

	if assetGroup == nil {
		return nil, nil
	}

	// Fetch attachments (ordered by title)
	attachmentOpts := database.NewOptions().
		WithWhere(squirrel.Eq{models.ATTACHMENT_ASSET_GROUP_ID: assetGroup.ID}).
		WithOrderBy(models.ATTACHMENT_TABLE_TITLE + " ASC")

	attachments, err := dao.ListAttachments(ctx, attachmentOpts)
	if err != nil {
		return nil, err
	}
	assetGroup.Attachments = attachments

	// 3) Fetch assets (ordered by prefix + sub_prefix)
	assetDbOpts := database.NewOptions().
		WithWhere(squirrel.Eq{models.ASSET_ASSET_GROUP_ID: assetGroup.ID}).
		WithOrderBy(models.ASSET_TABLE_PREFIX + " ASC, " + models.ASSET_TABLE_SUB_PREFIX + " ASC")

	if dbOpts != nil {
		assetDbOpts.IncludeProgress = dbOpts.IncludeProgress
		assetDbOpts.IncludeAssetVideoMetadata = dbOpts.IncludeAssetVideoMetadata
	}

	assets, err := dao.ListAssets(ctx, assetDbOpts)
	if err != nil {
		return nil, err
	}
	assetGroup.Assets = assets

	return assetGroup, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ListAssetGroups gets all records from the asset groups table based upon the where clause and pagination
// in the options
func (dao *DAO) ListAssetGroups(ctx context.Context, dbOpts *database.Options) ([]*models.AssetGroup, error) {
	// Fetch asset groups
	builderOpts := newBuilderOptions(models.ASSET_GROUP_TABLE).
		WithColumns(models.ASSET_GROUP_TABLE + ".*").
		SetDbOpts(dbOpts)

	// Override order by
	builderOpts.DbOpts.WithOrderBy(
		models.ASSET_GROUP_TABLE_PREFIX+" ASC ",
		models.ASSET_GROUP_TABLE_MODULE+" ASC",
	)

	assetGroups, err := listGeneric[models.AssetGroup](ctx, dao, *builderOpts)
	if err != nil || len(assetGroups) == 0 {
		return assetGroups, err
	}

	// Gather IDs
	ids := make([]string, 0, len(assetGroups))
	for i := range assetGroups {
		ids = append(ids, assetGroups[i].ID)
	}

	// Fetch attachments for all asset groups, ordering by title
	attachmentDbOpts := database.NewOptions().
		WithWhere(squirrel.Eq{models.ATTACHMENT_ASSET_GROUP_ID: ids}).
		WithOrderBy(
			models.ATTACHMENT_TABLE_ASSET_GROUP_ID+" ASC",
			models.ATTACHMENT_TABLE_TITLE+" ASC",
		)

	attachments, err := dao.ListAttachments(ctx, attachmentDbOpts)
	if err != nil {
		return nil, err
	}

	attMap := make(map[string][]*models.Attachment)
	for _, a := range attachments {
		attMap[a.AssetGroupID] = append(attMap[a.AssetGroupID], a)
	}

	// Fetch assets for all groups
	assetDbOpts := database.NewOptions().
		WithWhere(squirrel.Eq{models.ASSET_ASSET_GROUP_ID: ids}).
		WithOrderBy(
			models.ASSET_TABLE_ASSET_GROUP_ID+" ASC ",
			models.ASSET_TABLE_PREFIX+" ASC ",
			models.ASSET_TABLE_SUB_PREFIX+" ASC",
		)

	if dbOpts != nil {
		assetDbOpts.IncludeProgress = dbOpts.IncludeProgress
		assetDbOpts.IncludeAssetVideoMetadata = dbOpts.IncludeAssetVideoMetadata
	}

	assets, err := dao.ListAssets(ctx, assetDbOpts)
	if err != nil {
		return nil, err
	}

	assetMap := make(map[string][]*models.Asset)
	for _, a := range assets {
		assetMap[a.AssetGroupID] = append(assetMap[a.AssetGroupID], a)
	}

	// Stitch children onto parents in order
	for _, assetGroup := range assetGroups {
		assetGroup.Attachments = attMap[assetGroup.ID]
		assetGroup.Assets = assetMap[assetGroup.ID]
	}

	return assetGroups, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateAssetGroup updates a single asset group record
func (dao *DAO) UpdateAssetGroup(ctx context.Context, assetGroup *models.AssetGroup) error {
	if err := assetGroupValidation(assetGroup); err != nil {
		return err
	}

	if assetGroup.ID == "" {
		return utils.ErrId
	}

	assetGroup.RefreshUpdatedAt()

	dbOpts := &database.Options{
		Where: squirrel.Eq{models.BASE_ID: assetGroup.ID},
	}

	builderOptions := newBuilderOptions(models.ASSET_GROUP_TABLE).
		WithData(map[string]interface{}{
			models.ASSET_GROUP_TITLE:            assetGroup.Title,
			models.ASSET_GROUP_PREFIX:           assetGroup.Prefix,
			models.ASSET_GROUP_MODULE:           assetGroup.Module,
			models.ASSET_GROUP_DESCRIPTION_PATH: assetGroup.DescriptionPath,
			models.ASSET_GROUP_DESCRIPTION_TYPE: assetGroup.DescriptionType,
			models.BASE_UPDATED_AT:              assetGroup.UpdatedAt,
		}).
		SetDbOpts(dbOpts)

	_, err := updateGeneric(ctx, dao, *builderOptions)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DeleteAssetGroups deletes records from the asset_groups table
//
// Errors when a WHERE clause is not provided.
func (dao *DAO) DeleteAssetGroups(ctx context.Context, dbOpts *database.Options) error {
	if dbOpts == nil || dbOpts.Where == nil {
		return utils.ErrWhere
	}

	builderOpts := newBuilderOptions(models.ASSET_GROUP_TABLE).SetDbOpts(dbOpts)
	sqlStr, args, _ := deleteBuilder(*builderOpts)

	q := database.QuerierFromContext(ctx, dao.db)
	_, err := q.ExecContext(ctx, sqlStr, args...)
	return err
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// assetGroupValidation validates the asset group fields
func assetGroupValidation(ag *models.AssetGroup) error {
	if ag == nil {
		return utils.ErrNilPtr
	}

	if ag.CourseID == "" {
		return utils.ErrCourseId
	}

	if ag.Title == "" {
		return utils.ErrTitle
	}

	if !ag.Prefix.Valid || ag.Prefix.Int16 < 0 {
		return utils.ErrPrefix
	}

	return nil
}
