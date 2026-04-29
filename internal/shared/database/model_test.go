package database

import (
	"testing"
	"time"

	"go.uber.org/zap/zapcore"
)

func TestBaseModel_MarshalLogObject(t *testing.T) {
	tests := []struct {
		name  string
		model *BaseModel
	}{
		{
			name:  "with ID",
			model: &BaseModel{ID: 123},
		},
		{
			name:  "zero ID",
			model: &BaseModel{ID: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoder := zapcore.NewMapObjectEncoder()
			err := tt.model.MarshalLogObject(encoder)
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if encoder.Fields["id"] != tt.model.ID {
				t.Errorf("expected id %d, got %v", tt.model.ID, encoder.Fields["id"])
			}
		})
	}
}

func TestStandardModel_CreateSetTime(t *testing.T) {
	model := &StandardModel{}
	before := time.Now()

	model.CreateSetTime()

	after := time.Now()

	if model.CreatedAt.Before(before) {
		t.Error("created_at should be set to current time")
	}
	if model.CreatedAt.After(after) {
		t.Error("created_at should not be in the future")
	}
	if model.UpdatedAt.Before(before) {
		t.Error("updated_at should be set to current time")
	}
	if model.UpdatedAt.After(after) {
		t.Error("updated_at should not be in the future")
	}
	if model.CreatedAt != model.UpdatedAt {
		t.Error("created_at and updated_at should be equal when creating")
	}
}

func TestStandardModel_UpdateSetTime(t *testing.T) {
	model := &StandardModel{
		CreatedAt: time.Now().Add(-1 * time.Hour),
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}
	before := time.Now()

	model.UpdateSetTime()

	after := time.Now()

	if model.UpdatedAt.Before(before) {
		t.Error("updated_at should be updated to current time")
	}
	if model.UpdatedAt.After(after) {
		t.Error("updated_at should not be in the future")
	}
}

func TestStandardModel_MarshalLogObject(t *testing.T) {
	now := time.Now()
	model := &StandardModel{
		BaseModel: BaseModel{ID: 456},
		CreatedAt: now,
		UpdatedAt: now,
	}

	encoder := zapcore.NewMapObjectEncoder()
	err := model.MarshalLogObject(encoder)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if encoder.Fields["id"] != uint32(456) {
		t.Errorf("expected id 456, got %v", encoder.Fields["id"])
	}
	if encoder.Fields["created_at"] != now {
		t.Errorf("expected created_at %v, got %v", now, encoder.Fields["created_at"])
	}
	if encoder.Fields["updated_at"] != now {
		t.Errorf("expected updated_at %v, got %v", now, encoder.Fields["updated_at"])
	}
}

func TestStandardModel_MarshalLogObject_NilBaseModel(t *testing.T) {
	model := &StandardModel{}
	encoder := zapcore.NewMapObjectEncoder()
	err := model.MarshalLogObject(encoder)
	if err != nil {
		t.Errorf("expected no error for empty model, got %v", err)
	}
}

func TestBaseModel_ZeroValue(t *testing.T) {
	model := BaseModel{}
	if model.ID != 0 {
		t.Errorf("expected zero value ID, got %d", model.ID)
	}
}

func TestStandardModel_ZeroValue(t *testing.T) {
	model := StandardModel{}
	if model.ID != 0 {
		t.Errorf("expected zero value ID, got %d", model.ID)
	}
	if !model.CreatedAt.IsZero() {
		t.Errorf("expected zero value CreatedAt, got %v", model.CreatedAt)
	}
	if !model.UpdatedAt.IsZero() {
		t.Errorf("expected zero value UpdatedAt, got %v", model.UpdatedAt)
	}
}

func TestStandardModel_TimeConsistency(t *testing.T) {
	model := &StandardModel{}
	model.CreateSetTime()

	initialCreated := model.CreatedAt
	initialUpdated := model.UpdatedAt

	if initialCreated != initialUpdated {
		t.Error("created_at and updated_at should be equal after CreateSetTime")
	}

	time.Sleep(time.Millisecond * 10)
	model.UpdateSetTime()

	if model.CreatedAt != initialCreated {
		t.Error("created_at should not change after UpdateSetTime")
	}
	if model.UpdatedAt == initialUpdated {
		t.Error("updated_at should change after UpdateSetTime")
	}
	if model.UpdatedAt.Before(initialUpdated) {
		t.Error("updated_at should be after initial updated_at")
	}
}
