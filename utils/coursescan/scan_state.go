package coursescan

import (
	"context"
	"sync"
	"time"

	"github.com/geerew/off-course/utils/security"
	"github.com/geerew/off-course/utils/types"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ScanState represents the state of a course scan operation
type ScanState struct {
	// Immutable fields
	ID          string
	CourseID    string
	CoursePath  string
	CourseTitle string
	CreatedAt   time.Time

	// Mutable fields (protected by mu)
	mu        sync.RWMutex
	Status    types.ScanStatusType
	Message   string
	cancelled bool

	// Cancellation
	cancel context.CancelFunc
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewScanState creates a new scan state
func NewScanState(courseID, coursePath, courseTitle string) *ScanState {
	return &ScanState{
		ID:          generateScanID(),
		CourseID:    courseID,
		CoursePath:  coursePath,
		CourseTitle: courseTitle,
		Status:      types.ScanStatusWaiting,
		Message:     "",
		CreatedAt:   time.Now(),
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SetCancel stores the cancel function for this scan
func (s *ScanState) SetCancel(cancel context.CancelFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cancel = cancel
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateStatus updates the scan status
func (s *ScanState) UpdateStatus(status types.ScanStatusType) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Status = status
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateMessage updates the scan progress message
func (s *ScanState) UpdateMessage(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Message = message
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// UpdateStatusAndMessage updates both status and message atomically
func (s *ScanState) UpdateStatusAndMessage(status types.ScanStatusType, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Status = status
	s.Message = message
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Cancel cancels the scan's context if it exists and marks the scan as cancelled
func (s *ScanState) Cancel() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cancelled = true
	if s.cancel != nil {
		s.cancel()
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// IsCancelled returns whether the scan has been cancelled (thread-safe read)
func (s *ScanState) IsCancelled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cancelled
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetStatus returns the current scan status (thread-safe read)
func (s *ScanState) GetStatus() types.ScanStatusType {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Status
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GetMessage returns the current scan message (thread-safe read)
func (s *ScanState) GetMessage() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Message
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// generateScanID generates a unique scan ID
func generateScanID() string {
	return security.PseudorandomString(10)
}
