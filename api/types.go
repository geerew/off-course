package api

import (
	"github.com/geerew/off-course/models"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type fileSystemResponse struct {
	Count       int                 `json:"count"`
	Directories []*fileInfoResponse `json:"directories"`
	Files       []*fileInfoResponse `json:"files"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type fileInfoResponse struct {
	Title          string                   `json:"title"`
	Path           string                   `json:"path"`
	Classification types.PathClassification `json:"classification"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type courseRequest struct {
	Title string `json:"title"`
	Path  string `json:"path"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type courseProgressResponse struct {
	Started           bool           `json:"started"`
	StartedAt         types.DateTime `json:"startedAt"`
	Percent           int            `json:"percent"`
	CompletedAt       types.DateTime `json:"completedAt"`
	ProgressUpdatedAt types.DateTime `json:"progressUpdatedAt"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type courseResponse struct {
	ID        string         `json:"id"`
	Title     string         `json:"title"`
	Path      string         `json:"path"`
	HasCard   bool           `json:"hasCard"`
	Available bool           `json:"available"`
	Duration  int            `json:"duration"`
	CreatedAt types.DateTime `json:"createdAt"`
	UpdatedAt types.DateTime `json:"updatedAt"`

	// Scan status
	ScanStatus string `json:"scanStatus"`

	// Progress
	Progress courseProgressResponse `json:"progress"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func courseResponseHelper(courses []*models.Course) []*courseResponse {
	responses := []*courseResponse{}
	for _, course := range courses {
		c := &courseResponse{
			ID:        course.ID,
			Title:     course.Title,
			Path:      course.Path,
			HasCard:   course.CardPath != "",
			Available: course.Available,
			Duration:  course.Duration,
			CreatedAt: course.CreatedAt,
			UpdatedAt: course.UpdatedAt,

			// Scan status
			ScanStatus: course.ScanStatus.String(),

			// Progress
			Progress: courseProgressResponse{
				Started:           course.Progress.Started,
				StartedAt:         course.Progress.StartedAt,
				Percent:           course.Progress.Percent,
				CompletedAt:       course.Progress.CompletedAt,
				ProgressUpdatedAt: course.Progress.UpdatedAt,
			},
		}

		responses = append(responses, c)
	}

	return responses
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type courseTagResponse struct {
	ID  string `json:"id"`
	Tag string `json:"tag"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func courseTagResponseHelper(courseTags []*models.CourseTag) []*courseTagResponse {
	responses := []*courseTagResponse{}
	for _, tag := range courseTags {
		responses = append(responses, &courseTagResponse{
			ID:  tag.ID,
			Tag: tag.Tag,
		})
	}

	return responses
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type assetProgressRequest struct {
	VideoPos  int  `json:"videoPos"`
	Completed bool `json:"completed"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type assetProgressResponse struct {
	VideoPos    int            `json:"videoPos"`
	Completed   bool           `json:"completed"`
	CompletedAt types.DateTime `json:"completedAt"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type assetVideoMetadataResponse struct {
	Duration   int    `json:"duration"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Codec      string `json:"codec"`
	Resolution string `json:"resolution"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type assetResponse struct {
	ID        string         `json:"id"`
	CourseID  string         `json:"courseId"`
	Title     string         `json:"title"`
	Prefix    int            `json:"prefix"`
	Chapter   string         `json:"chapter"`
	Path      string         `json:"path"`
	Type      types.Asset    `json:"assetType"`
	CreatedAt types.DateTime `json:"createdAt"`
	UpdatedAt types.DateTime `json:"updatedAt"`

	// Relations
	VideoMetadata *assetVideoMetadataResponse `json:"videoMetadata,omitempty"`
	Progress      *assetProgressResponse      `json:"progress"`
	Attachments   []*attachmentResponse       `json:"attachments"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func assetResponseHelper(assets []*models.Asset) []*assetResponse {
	responses := []*assetResponse{}
	for _, asset := range assets {

		// Video metadata
		var videoMetadata *assetVideoMetadataResponse
		if asset.VideoMetadata != nil {
			videoMetadata = &assetVideoMetadataResponse{
				Duration:   asset.VideoMetadata.Duration,
				Width:      asset.VideoMetadata.Width,
				Height:     asset.VideoMetadata.Height,
				Codec:      asset.VideoMetadata.Codec,
				Resolution: asset.VideoMetadata.Resolution,
			}
		}

		// Asset progress
		progress := &assetProgressResponse{}
		if asset.Progress != nil {
			progress.VideoPos = asset.Progress.VideoPos
			progress.Completed = asset.Progress.Completed
			progress.CompletedAt = asset.Progress.CompletedAt
		}

		responses = append(responses, &assetResponse{
			ID:        asset.ID,
			CourseID:  asset.CourseID,
			Title:     asset.Title,
			Prefix:    int(asset.Prefix.Int16),
			Chapter:   asset.Chapter,
			Path:      asset.Path,
			Type:      asset.Type,
			CreatedAt: asset.CreatedAt,
			UpdatedAt: asset.UpdatedAt,

			VideoMetadata: videoMetadata,
			Progress:      progress,
			Attachments:   attachmentResponseHelper(asset.Attachments),
		})

	}

	return responses
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type attachmentResponse struct {
	ID        string         `json:"id"`
	AssetId   string         `json:"assetId"`
	Title     string         `json:"title"`
	Path      string         `json:"path"`
	CreatedAt types.DateTime `json:"createdAt"`
	UpdatedAt types.DateTime `json:"updatedAt"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func attachmentResponseHelper(attachments []*models.Attachment) []*attachmentResponse {
	if len(attachments) == 0 {
		return []*attachmentResponse{} // ← empty slice, not nil
	}

	responses := []*attachmentResponse{}
	for _, attachment := range attachments {
		responses = append(responses, &attachmentResponse{
			ID:        attachment.ID,
			AssetId:   attachment.AssetID,
			Title:     attachment.Title,
			Path:      attachment.Path,
			CreatedAt: attachment.CreatedAt,
			UpdatedAt: attachment.UpdatedAt,
		})
	}

	return responses
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type ScanRequest struct {
	CourseID string `json:"courseId"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type scanResponse struct {
	ID        string           `json:"id"`
	CourseID  string           `json:"courseId"`
	Status    types.ScanStatus `json:"status"`
	CreatedAt types.DateTime   `json:"createdAt"`
	UpdatedAt types.DateTime   `json:"updatedAt"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func scanResponseHelper(scans []*models.Scan) []*scanResponse {
	responses := []*scanResponse{}
	for _, scan := range scans {
		responses = append(responses, &scanResponse{
			ID:        scan.ID,
			CourseID:  scan.CourseID,
			Status:    scan.Status,
			CreatedAt: scan.CreatedAt,
			UpdatedAt: scan.UpdatedAt,
		})
	}

	return responses
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type tagRequest struct {
	Tag string `json:"tag"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type tagResponse struct {
	ID          string         `json:"id"`
	Tag         string         `json:"tag"`
	CourseCount int            `json:"courseCount"`
	CreatedAt   types.DateTime `json:"createdAt"`
	UpdatedAt   types.DateTime `json:"updatedAt"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func tagResponseHelper(tags []*models.Tag) []*tagResponse {
	responses := []*tagResponse{}

	for _, tag := range tags {
		t := &tagResponse{
			ID:          tag.ID,
			Tag:         tag.Tag,
			CreatedAt:   tag.CreatedAt,
			UpdatedAt:   tag.UpdatedAt,
			CourseCount: tag.CourseCount,
		}

		// Add basic course information
		responses = append(responses, t)
	}

	return responses
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type userRequest struct {
	Username        string `json:"username"`
	DisplayName     string `json:"displayName"`
	CurrentPassword string `json:"currentPassword"`
	Password        string `json:"password"`
	Role            string `json:"role"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type userResponse struct {
	ID          string         `json:"id"`
	Username    string         `json:"username"`
	DisplayName string         `json:"displayName"`
	Role        types.UserRole `json:"role"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func userResponseHelper(users []*models.User) []*userResponse {
	responses := []*userResponse{}

	for _, user := range users {
		responses = append(responses, &userResponse{
			ID:          user.ID,
			Username:    user.Username,
			DisplayName: user.DisplayName,
			Role:        user.Role,
		})
	}

	return responses
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type TokenResponse struct {
	Token string `json:"token"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type logResponse struct {
	ID        string         `json:"id"`
	Level     int            `json:"level"`
	Message   string         `json:"message"`
	Data      types.JsonMap  `json:"data"`
	CreatedAt types.DateTime `json:"createdAt"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func logsResponseHelper(logs []*models.Log) []*logResponse {
	responses := []*logResponse{}

	for _, log := range logs {
		responses = append(responses, &logResponse{
			ID:        log.ID,
			Level:     log.Level,
			Message:   log.Message,
			Data:      log.Data,
			CreatedAt: log.CreatedAt,
		})
	}

	return responses
}
