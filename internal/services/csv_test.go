package services

import (
	"encoding/csv"
	"io"
	"testing"
	"time"

	"github.com/Tonyblaise/to-do/internal/models"
)

func TestExportTasksCSV_HeaderRow(t *testing.T) {
	r, err := ExportTasksCSV(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	records, err := csv.NewReader(r).ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 row (header only), got %d", len(records))
	}

	want := []string{"ID", "Title", "Description", "Status", "Priority", "Due Date", "Tags", "Created At", "Updated At"}
	for i, h := range want {
		if records[0][i] != h {
			t.Errorf("header[%d] = %q, want %q", i, records[0][i], h)
		}
	}
}

func TestExportTasksCSV_TaskData(t *testing.T) {
	due := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	tasks := []models.Task{
		{
			ID:          "id-1",
			Title:       "Buy milk",
			Description: "From the store",
			Status:      models.TaskStatusPending,
			Priority:    models.PriorityHigh,
			DueDate:     &due,
			Tags:        []models.Tag{{Name: "Shopping"}, {Name: "Urgent"}},
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:        "id-2",
			Title:     "Write tests",
			Status:    models.TaskStatusInProgress,
			Priority:  models.PriorityMedium,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	r, err := ExportTasksCSV(tasks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	records, err := csv.NewReader(r).ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV: %v", err)
	}

	if len(records) != 3 {
		t.Fatalf("expected 3 rows (header + 2 tasks), got %d", len(records))
	}

	row1 := records[1]
	if row1[0] != "id-1" {
		t.Errorf("row1 ID = %q, want %q", row1[0], "id-1")
	}
	if row1[1] != "Buy milk" {
		t.Errorf("row1 Title = %q, want %q", row1[1], "Buy milk")
	}
	if row1[3] != "pending" {
		t.Errorf("row1 Status = %q, want %q", row1[3], "pending")
	}
	if row1[4] != "high" {
		t.Errorf("row1 Priority = %q, want %q", row1[4], "high")
	}
	if row1[5] != "2026-03-10" {
		t.Errorf("row1 DueDate = %q, want %q", row1[5], "2026-03-10")
	}
	if row1[6] != "Shopping, Urgent" {
		t.Errorf("row1 Tags = %q, want %q", row1[6], "Shopping, Urgent")
	}
}

func TestExportTasksCSV_NilDueDate(t *testing.T) {
	tasks := []models.Task{
		{
			ID:        "id-1",
			Title:     "No due date",
			Status:    models.TaskStatusPending,
			Priority:  models.PriorityLow,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	r, err := ExportTasksCSV(tasks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	records, err := csv.NewReader(r).ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV: %v", err)
	}
	if records[1][5] != "" {
		t.Errorf("expected empty due date, got %q", records[1][5])
	}
}

func TestExportTasksCSV_NoTags(t *testing.T) {
	tasks := []models.Task{
		{
			ID:        "id-1",
			Title:     "No tags",
			Status:    models.TaskStatusPending,
			Priority:  models.PriorityLow,
			Tags:      nil,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	r, err := ExportTasksCSV(tasks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	records, _ := csv.NewReader(r).ReadAll()
	if records[1][6] != "" {
		t.Errorf("expected empty tags, got %q", records[1][6])
	}
}

func TestExportTasksCSV_ReturnsReadableContent(t *testing.T) {
	r, err := ExportTasksCSV([]models.Task{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read: %v", err)
	}
	if len(b) == 0 {
		t.Error("expected non-empty CSV output")
	}
}
