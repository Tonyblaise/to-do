package services

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"

	"github.com/Tonyblaise/to-do/internal/models"
)

func ExportTasksCSV(tasks []models.Task) (io.Reader, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.Write([]string{
		"ID", "Title", "Description", "Status", "Priority",
		"Due Date", "Tags", "Created At", "Updated At",
	}); err != nil {
		return nil, fmt.Errorf("writing CSV header: %w", err)
	}

	for _, t := range tasks {
		dueDate := ""

		if t.DueDate != nil {
			dueDate = t.DueDate.Format("2006-01-02")
		}

		tags := make([]string, len(t.Tags))
		for i, tag := range t.Tags {
			tags[i] = tag.Name
		}
		tagStr := ""
		for i, tg := range tags {
			if i > 0 {
				tagStr += ", "
			}
			tagStr += tg
		}
		if err := w.Write([]string{
			t.ID, t.Title, t.Description,
			string(t.Status), string(t.Priority),
			dueDate, tagStr,
			t.CreatedAt.Format("2006-01-02 15:04:05"),
			t.UpdatedAt.Format("2006-01-02 15:04:05"),
		}); err != nil {
			return nil, fmt.Errorf("writing CSV row: %w", err)
		}

	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("flushing CSV: %w", err)
	}
	return &buf, nil

}
