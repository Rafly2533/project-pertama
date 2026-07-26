package repository

import (
	"fmt"

	"intan-florist-api/config"
	"intan-florist-api/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(cfg config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)})
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Category{}, &models.Product{}, &models.ProductImage{}, &models.Page{}, &models.PaymentMethod{}, &models.ShippingInfo{}, &models.Testimonial{}, &models.Banner{}, &models.Setting{}, &models.Contact{}); err != nil {
		return nil, fmt.Errorf("migrate database: %w", err)
	}
	if err := Seed(db, cfg); err != nil {
		return nil, fmt.Errorf("seed database: %w", err)
	}
	return db, nil
}

func Seed(db *gorm.DB, cfg config.Config) error {
	return db.Transaction(func(tx *gorm.DB) error {
		hash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		admin := models.User{Username: cfg.AdminUsername, PasswordHash: string(hash), FullName: cfg.AdminFullName, Role: "super_admin", IsActive: true}
		if err := tx.Where("username = ?", cfg.AdminUsername).FirstOrCreate(&admin).Error; err != nil {
			return err
		}
		settings := []models.Setting{{Key: "logo", Value: ""}, {Key: "wa_number", Value: "6281234567890"}, {Key: "email", Value: "info@intanflorist.com"}, {Key: "address", Value: "Jakarta, Indonesia"}, {Key: "meta_title", Value: "Intan Florist"}, {Key: "meta_description", Value: "Bunga segar untuk setiap momen istimewa"}}
		for _, item := range settings {
			if err := tx.Where("key = ?", item.Key).FirstOrCreate(&item).Error; err != nil {
				return err
			}
		}
		categories := []models.Category{{Name: "Buket Bunga", Slug: "buket-bunga", Description: "Buket bunga segar untuk berbagai momen"}, {Name: "Bunga Papan", Slug: "bunga-papan", Description: "Bunga papan untuk ucapan dan acara"}}
		for _, item := range categories {
			if err := tx.Where("slug = ?", item.Slug).FirstOrCreate(&item).Error; err != nil {
				return err
			}
		}
		pages := []models.Page{{Title: "Tentang Kami", Slug: "tentang-kami", Content: "Intan Florist menyediakan rangkaian bunga berkualitas untuk setiap momen.", MetaTitle: "Tentang Intan Florist", IsActive: true}, {Title: "Cara Pemesanan", Slug: "cara-pemesanan", Content: "Hubungi kami melalui WhatsApp dan sampaikan detail pesanan Anda.", IsActive: true}}
		for _, item := range pages {
			if err := tx.Where("slug = ?", item.Slug).FirstOrCreate(&item).Error; err != nil {
				return err
			}
		}
		seeds := []any{&models.Banner{Title: "Intan Florist", Subtitle: "Bunga segar untuk momen istimewa", ImageURL: "https://placehold.co/1600x600", IsActive: true}, &models.Testimonial{CustomerName: "Pelanggan Intan", Content: "Bunganya segar dan pelayanannya ramah.", Rating: 5, IsActive: true}, &models.PaymentMethod{BankName: "Bank Central Asia", AccountNumber: "0000000000", AccountName: "Intan Florist", IsActive: true}, &models.ShippingInfo{Region: "Area Kota", EstimatedTime: "1-2 jam", ShippingCost: 25000, IsActive: true}}
		for _, item := range seeds {
			if err := tx.FirstOrCreate(item).Error; err != nil {
				return err
			}
		}
		return nil
	})
}