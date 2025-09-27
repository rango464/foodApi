package db

import (
	"fmt"
	"log"

	"github.com/RangoCoder/foodApi/internal/env"
	st "github.com/RangoCoder/foodApi/internal/structs"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func InitDB() (*gorm.DB, error) {
	// dsn := "hots=localhost user=postgres password=postgresspassword dbname=postgres port=5432 sslmode=disable"
	// docker run --name postgressGo -e POSTGRES_PASSWORD=postgresspassword -d -p 5432:5432 postgres
	// dsn1 := "postgres://postgres:postgresspassword@localhost:5432/"
	user := env.GetEnvVar("POSTGRES_USER")
	pass := env.GetEnvVar("POSTGRES_PASS")
	host := env.GetEnvVar("POSTGRES_HOST")
	port := env.GetEnvVar("POSTGRES_PORT")
	dsn := fmt.Sprintf("postgres://%v:%v@%v:%v/", user, pass, host, port)

	var err error

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	if err := db.AutoMigrate(&st.UserMailConfirm{}); err != nil {
		log.Fatal("Could not migrate: struct = AuthUser", err)
	}
	if err := db.AutoMigrate(&st.AuthUser{}); err != nil {
		log.Fatal("Could not migrate: struct = AuthUser", err)
	}

	if err := db.AutoMigrate(&st.ParamsUser{}); err != nil {
		log.Fatal("Could not migrate: struct = ParamsUser", err)
	}

	if err := db.AutoMigrate(&st.User{}); err != nil {
		log.Fatal("Could not migrate: struct = User ", err)
	}

	return db, nil
}
