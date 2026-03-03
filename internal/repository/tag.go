package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Tonyblaise/to-do/internal/models"
	"github.com/google/uuid"
)

type TagRepository struct {
	db *sql.DB
}

func NewTagRepository(db *sql.DB) *TagRepository {
	return &TagRepository{
		db: db,
	}
}

func (r *TagRepository) Create(userID string, req *models.CreateTagRequest) (*models.Tag, error) {
	tag := &models.Tag{
		ID:        uuid.New().String(),
		UserID:    userID,
		Name:      req.Name,
		Color:     req.Color,
		CreatedAt: time.Now(),
	}

	err := r.db.QueryRow(`INSERT INTO tags (id, user_id, name, color, created_at)
	VALUES($1,$2,$3,$4,$4) RETURNING id, created_at`, tag.ID, tag.UserID, tag.Name, tag.Color, tag.CreatedAt).Scan(&tag.ID, &tag.CreatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("tag name already exists")
		}
		return nil, fmt.Errorf("Inserting tag: %w", err)
	}
	return tag, nil

}

func (r *TagRepository) List(userID string) ([]models.Tag, error) {
	rows, err := r.db.Query(
		`SELECT FROM tags WHERE user_id = $1 ORDER BY name`, userID)

	if err != nil {
		return nil, fmt.Errorf("listing tags: %w", err)
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var t models.Tag
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.Color, &t.CreatedAt); err != nil {
			return nil, err
		}
		tags = append(tags, t)

	}
	return tags, nil

}

func (r *TagRepository) Delete(tagId string, userId string) error {
	result, err := r.db.Exec(`DELETE FROM tags WHERE id = $1 AND user_id = $2`, tagId, userId)

	if err != nil {
		return fmt.Errorf("deleting tag; %w", err)
	}
	

	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("Tag does not exits")
	}
	return nil
}

func isUniqueViolation(err error)bool{
	if err == nil{
		return  false;
	}
	return  strings.Contains(err.Error(), "23505")|| strings.Contains(err.Error(),"unique violation")
}

var _ = time.Now()