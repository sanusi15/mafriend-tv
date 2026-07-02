package config

import (
	"fmt"
	"log"
	"mafriend-tv/internal/model" // 💡 Sesuaikan dengan path modul lu
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB() *gorm.DB {
	username := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASS")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, host, port, dbName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Menampilkan query SQL di terminal biar gampang di-debug
	})
	if err != nil {
		log.Fatalf("Gagal terhubung ke MySQL via GORM: %v", err)
	}

	log.Println("Koneksi MySQL via GORM Berhasil!")

	// 🚀 1. JALANKAN AUTOMIGRATE
	log.Println("Menjalankan AutoMigrate...")
	err = db.AutoMigrate(&model.User{})
	if err != nil {
		log.Fatalf("Gagal menjalankan AutoMigrate: %v", err)
	}
	log.Println("AutoMigrate Berhasil!")

	// 🚀 2. JALANKAN SEEDER
	seedUsers(db)

	// Set konfigurasi connection pool bawaan
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Minute * 5)

	return db
}

// Fungsi Seeder untuk data awal
func seedUsers(db *gorm.DB) {
	var count int64
	db.Model(&model.User{}).Count(&count)

	// Jika tabel masih kosong, isi dengan data tiruan (sample data)
	if count == 0 {
		log.Println("Tabel tbl_users kosong, menyuntikkan data seed...")

		dummyUsers := []model.User{
			{
				GoogleID: "google-dummy-11111",
				Name:     "Admin Mafriend",
				Email:    "admin@mafriendtv.com",
				Picture:  "https://lh3.googleusercontent.com/a/default-user=s96-c",
			},
			{
				GoogleID: "google-dummy-22222",
				Name:     "Bot Ganteng",
				Email:    "bot.ganteng@gmail.com",
				Picture:  "https://lh3.googleusercontent.com/a/default-user=s96-c",
			},
		}

		// Insert batch sekaligus
		if err := db.Create(&dummyUsers).Error; err != nil {
			log.Printf("Gagal menyuntikkan data seed: %v", err)
		} else {
			log.Println("Data seed sukses disuntikkan!")
		}
	} else {
		log.Println("Data seed dilewati karena tabel sudah berisi data.")
	}
}