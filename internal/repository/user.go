package repository

import (
	"database/sql"
	"errors"
	"fmt"
	
	"time"

	"github.com/Tonyblaise/to-do/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct{
	db *sql.DB
}


func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

var ErrUserNotFound = errors.New("user not found")
var ErrEmailTaken = errors.New("email already registered")

func (r *UserRepository) Create(email, password, firstName, lastName string) (*models.User, error) {
  
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return nil, fmt.Errorf("hashing password: %w", err)
    }

   
    user := &models.User{
        ID:           uuid.New().String(),
        Email:        email,
        FirstName:    firstName,
        LastName:     lastName,
        PasswordHash: string(hash),
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
    }

   
    query := `
        INSERT INTO users (id, email, first_name, last_name, password_hash, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, email, first_name, last_name, created_at, updated_at`

  
    err = r.db.QueryRow(
        query, 
        user.ID, user.Email, user.FirstName, user.LastName, user.PasswordHash, user.CreatedAt, user.UpdatedAt,
    ).Scan(
        &user.ID, &user.Email, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt,
    )

    if err != nil {
        if isUniqueViolation(err) {
            return nil, ErrEmailTaken
        }
        return nil, fmt.Errorf("inserting user: %w", err)
    }

    return user, nil
}
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(`
		SELECT id, email, password_hash, created_at, updated_at
		FROM users WHERE email = $1`, email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying user: %w", err)
	}

	return user, nil
}


func(r *UserRepository) GetById(id string)(*models.User, error){
	user := &models.User{}

	err:= r.db.QueryRow(
		`SELECT id, email,first_name, last_name, created_at, updated_at
		FROM users WHERE id = $1`, id,

	).Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows{
		return  nil, ErrUserNotFound
	}

	if err != nil{
		return  nil, fmt.Errorf("querying user: %w", err)
	}
	return  user, nil
}


func(r *UserRepository) VerifyPassword(user *models.User, password string) bool{
	return  bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) ==nil
}
