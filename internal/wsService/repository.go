package wsService

import (
	"gorm.io/gorm"
)

type WsRepository interface {
}

type wsRepository struct {
	db *gorm.DB
}

func NewWsRepository(db *gorm.DB) WsRepository {
	return &wsRepository{db: db}
}

// создадим пользователя, его пустые параметры и пустую авторизацию - применим транзакцию
// func (r *wsRepository) SaveData(date int64, val float64) (structs.USDEUR, error) {
// 	cq := structs.USDEUR{Date: date, Val: val}
// 	err := r.db.Create(&cq).Error
// 	return cq, err
// }
