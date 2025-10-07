package homeService

import (
	"github.com/RangoCoder/foodApi/internal/structs"
	"gorm.io/gorm"
)

type HomeRepository interface {
	Home() error
}

type homeRepository struct {
	db *gorm.DB
}

func NewHomeRepository(db *gorm.DB) HomeRepository {
	return &homeRepository{db: db}
}

func (r *homeRepository) Home() error {
	return r.db.Take(&structs.User{}).Error
}
