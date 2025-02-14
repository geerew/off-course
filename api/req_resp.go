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

type courseProgressResponse struct {
	Started           bool           `json:"started"`
	StartedAt         types.DateTime `json:"started_at"`
	Percent           int            `json:"percent"`
	CompletedAt       types.DateTime `json:"completed_at"`
	ProgressUpdatedAt types.DateTime `json:"progress_updated_at"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type courseResponse struct {
	ID        string         `json:"id"`
	Title     string         `json:"title"`
	Path      string         `json:"path"`
	HasCard   bool           `json:"has_card"`
	Available bool           `json:"available"`
	CreatedAt types.DateTime `json:"created_at"`
	UpdatedAt types.DateTime `json:"updated_at"`

	// Scan status
	ScanStatus string `json:"scan_status"`

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
	ID       string `json:"id"`
	Tag      string `json:"tag,omitempty"`
	CourseID string `json:"course_id,omitempty"`
	Title    string `json:"title,omitempty"`
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
	VideoPos  int  `json:"video_position"`
	Completed bool `json:"completed"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type assetProgressResponse struct {
	VideoPos    int            `json:"video_position"`
	Completed   bool           `json:"completed"`
	CompletedAt types.DateTime `json:"completed_at"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func assetResponseHelper(assets []*models.Asset) []*assetResponse {
	responses := []*assetResponse{}
	for _, asset := range assets {

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

			Progress:    progress,
			Attachments: attachmentResponseHelper(asset.Attachments),
		})

	}

	return responses
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type attachmentResponse struct {
	ID        string         `json:"id"`
	AssetId   string         `json:"asset_id"`
	Title     string         `json:"title"`
	Path      string         `json:"path"`
	CreatedAt types.DateTime `json:"created_at"`
	UpdatedAt types.DateTime `json:"updated_at"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func attachmentResponseHelper(attachments []*models.Attachment) []*attachmentResponse {
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

type assetResponse struct {
	ID        string         `json:"id"`
	CourseID  string         `json:"course_id"`
	Title     string         `json:"title"`
	Prefix    int            `json:"prefix"`
	Chapter   string         `json:"chapter"`
	Path      string         `json:"path"`
	Type      types.Asset    `json:"asset_type"`
	CreatedAt types.DateTime `json:"created_at"`
	UpdatedAt types.DateTime `json:"updated_at"`

	// Relations
	Progress    *assetProgressResponse `json:"progress"`
	Attachments []*attachmentResponse  `json:"attachments,omitempty"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type scanResponse struct {
	ID        string           `json:"id"`
	CourseID  string           `json:"course_id"`
	Status    types.ScanStatus `json:"status"`
	CreatedAt types.DateTime   `json:"created_at"`
	UpdatedAt types.DateTime   `json:"updated_at"`
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
	ID          string               `json:"id"`
	Tag         string               `json:"tag"`
	CourseCount int                  `json:"course_count"`
	Courses     []*courseTagResponse `json:"courses,omitempty"`
	CreatedAt   types.DateTime       `json:"created_at"`
	UpdatedAt   types.DateTime       `json:"updated_at"`
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
			CourseCount: len(tag.CourseTags),
		}

		// Add the course tags
		if len(tag.CourseTags) > 0 {
			courses := []*courseTagResponse{}

			for _, ct := range tag.CourseTags {
				courses = append(courses, &courseTagResponse{
					ID:       ct.ID,
					CourseID: ct.CourseID,
					Title:    ct.Course,
				})
			}

			t.Courses = courses
		}

		responses = append(responses, t)
	}

	return responses
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type userRequest struct {
	Username        string `json:"username"`
	DisplayName     string `json:"display_name"`
	CurrentPassword string `json:"current_password"`
	Password        string `json:"password"`
	Role            string `json:"role"`
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type userResponse struct {
	ID          string         `json:"id"`
	Username    string         `json:"username"`
	DisplayName string         `json:"display_name"`
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
	CreatedAt types.DateTime `json:"created_at"`
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
