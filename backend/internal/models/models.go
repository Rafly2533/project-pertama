package models

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:80;not null" json:"username"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	FullName     string    `gorm:"size:160;not null" json:"full_name"`
	Role         string    `gorm:"size:30;not null;default:staff;check:role IN ('super_admin','staff')" json:"role"`
	IsActive     bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Category struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:120;not null" json:"name"`
	Slug        string    `gorm:"uniqueIndex;size:140;not null" json:"slug"`
	Description string    `gorm:"type:text" json:"description"`
	Icon        string    `gorm:"size:500" json:"icon"`
	Products    []Product `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"-"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Product struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	CategoryID  uint           `gorm:"not null;index" json:"category_id"`
	Category    Category       `json:"category"`
	Name        string         `gorm:"size:180;not null" json:"name"`
	Slug        string         `gorm:"uniqueIndex;size:200;not null" json:"slug"`
	Description string         `gorm:"type:text" json:"description"`
	PriceInfo   string         `gorm:"size:180;not null" json:"price_info"`
	IsActive    bool           `gorm:"not null;default:true;index" json:"is_active"`
	Images      []ProductImage `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"images"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type ProductImage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProductID uint      `gorm:"not null;index" json:"product_id"`
	ImageURL  string    `gorm:"size:1000;not null" json:"image_url"`
	IsPrimary bool      `gorm:"not null;default:false" json:"is_primary"`
	SortOrder int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Page struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Title           string    `gorm:"size:180;not null" json:"title"`
	Slug            string    `gorm:"uniqueIndex;size:200;not null" json:"slug"`
	Content         string    `gorm:"type:text;not null" json:"content"`
	MetaTitle       string    `gorm:"size:255" json:"meta_title"`
	MetaDescription string    `gorm:"type:text" json:"meta_description"`
	IsActive        bool      `gorm:"not null;default:true;index" json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type PaymentMethod struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	BankName      string    `gorm:"size:120;not null" json:"bank_name"`
	AccountNumber string    `gorm:"size:100;not null" json:"account_number"`
	AccountName   string    `gorm:"size:160;not null" json:"account_name"`
	QRCodeURL     string    `gorm:"size:1000" json:"qr_code_url"`
	IsActive      bool      `gorm:"not null;default:true;index" json:"is_active"`
	SortOrder     int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ShippingInfo struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Region        string    `gorm:"size:180;not null" json:"region"`
	EstimatedTime string    `gorm:"size:120;not null" json:"estimated_time"`
	ShippingCost  int64     `gorm:"not null;default:0;check:shipping_cost >= 0" json:"shipping_cost"`
	IsActive      bool      `gorm:"not null;default:true;index" json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Testimonial struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	CustomerName string    `gorm:"size:160;not null" json:"customer_name"`
	Content      string    `gorm:"type:text;not null" json:"content"`
	Rating       int       `gorm:"not null;check:rating >= 1 AND rating <= 5" json:"rating"`
	IsActive     bool      `gorm:"not null;default:true;index" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Banner struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Title     string    `gorm:"size:180;not null" json:"title"`
	Subtitle  string    `gorm:"type:text" json:"subtitle"`
	ImageURL  string    `gorm:"size:1000;not null" json:"image_url"`
	LinkURL   string    `gorm:"size:1000" json:"link_url"`
	IsActive  bool      `gorm:"not null;default:true;index" json:"is_active"`
	SortOrder int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Setting struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Key       string    `gorm:"uniqueIndex;size:120;not null" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Contact struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:160;not null" json:"name"`
	Email     string    `gorm:"size:255;not null;index" json:"email"`
	Subject   string    `gorm:"size:255;not null" json:"subject"`
	Message   string    `gorm:"type:text;not null" json:"message"`
	IsRead    bool      `gorm:"not null;default:false;index" json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
