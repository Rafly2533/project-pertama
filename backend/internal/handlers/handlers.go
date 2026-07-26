package handlers

import (
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"

	"intan-florist-api/config"
	"intan-florist-api/internal/models"
	"intan-florist-api/pkg/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Handler struct {
	DB     *gorm.DB
	Config config.Config
}

func success(c *gin.Context, status int, data any) { c.JSON(status, gin.H{"data": data}) }
func failure(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message, "message": message})
}

func (h Handler) Products(c *gin.Context) {
	page := positiveInt(c.Query("page"), 1)
	limit := positiveInt(c.Query("limit"), 12)
	if limit > 100 {
		limit = 100
	}
	query := h.DB.Model(&models.Product{}).Where("products.is_active = ?", true)
	if slug := strings.TrimSpace(c.Query("category")); slug != "" {
		query = query.Joins("JOIN categories ON categories.id = products.category_id").Where("categories.slug = ?", slug)
	}
	if search := strings.TrimSpace(c.Query("search")); search != "" {
		query = query.Where("products.name ILIKE ? OR products.description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		failure(c, 500, "failed to load products")
		return
	}
	var products []models.Product
	if err := query.Preload("Category").Preload("Images", func(db *gorm.DB) *gorm.DB { return db.Order("is_primary DESC, sort_order ASC, id ASC") }).Order("products.created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&products).Error; err != nil {
		failure(c, 500, "failed to load products")
		return
	}
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	c.JSON(http.StatusOK, gin.H{"data": products, "products": products, "meta": gin.H{"total": total, "totalPages": totalPages, "total_pages": totalPages, "page": page, "limit": limit}, "total": total, "totalPages": totalPages, "page": page, "limit": limit})
}

func (h Handler) ProductBySlug(c *gin.Context) {
	var product models.Product
	if err := h.DB.Preload("Category").Preload("Images", func(db *gorm.DB) *gorm.DB { return db.Order("is_primary DESC, sort_order ASC, id ASC") }).Where("slug = ? AND is_active = ?", c.Param("slug"), true).First(&product).Error; err != nil {
		notFoundOrError(c, err, "product")
		return
	}
	success(c, 200, product)
}

func (h Handler) Categories(c *gin.Context) { h.list(c, &[]models.Category{}, "name ASC") }
func (h Handler) PageBySlug(c *gin.Context) {
	var item models.Page
	if err := h.DB.Where("slug = ? AND is_active = ?", c.Param("slug"), true).First(&item).Error; err != nil {
		notFoundOrError(c, err, "page")
		return
	}
	success(c, 200, item)
}
func (h Handler) Banners(c *gin.Context) {
	h.activeList(c, &[]models.Banner{}, "sort_order ASC, id ASC")
}
func (h Handler) PaymentMethods(c *gin.Context) {
	h.activeList(c, &[]models.PaymentMethod{}, "sort_order ASC, id ASC")
}
func (h Handler) Shipping(c *gin.Context) { h.activeList(c, &[]models.ShippingInfo{}, "id ASC") }
func (h Handler) Testimonials(c *gin.Context) {
	h.activeList(c, &[]models.Testimonial{}, "created_at DESC")
}
func (h Handler) Settings(c *gin.Context) { h.list(c, &[]models.Setting{}, "key ASC") }

func (h Handler) list(c *gin.Context, destination any, order string) {
	if err := h.DB.Order(order).Find(destination).Error; err != nil {
		failure(c, 500, "failed to load data")
		return
	}
	success(c, 200, destination)
}
func (h Handler) activeList(c *gin.Context, destination any, order string) {
	if err := h.DB.Where("is_active = ?", true).Order(order).Find(destination).Error; err != nil {
		failure(c, 500, "failed to load data")
		return
	}
	success(c, 200, destination)
}

func (h Handler) CreateContact(c *gin.Context) {
	var input struct {
		Name    string `json:"name" binding:"required,min=2,max=160"`
		Email   string `json:"email" binding:"required,email,max=255"`
		Subject string `json:"subject" binding:"required,min=2,max=255"`
		Message string `json:"message" binding:"required,min=5,max=10000"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		failure(c, 400, "invalid contact data")
		return
	}
	item := models.Contact{Name: strings.TrimSpace(input.Name), Email: strings.ToLower(strings.TrimSpace(input.Email)), Subject: strings.TrimSpace(input.Subject), Message: strings.TrimSpace(input.Message)}
	if err := h.DB.Create(&item).Error; err != nil {
		failure(c, 500, "failed to submit contact")
		return
	}
	success(c, 201, item)
}

func (h Handler) Login(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		failure(c, 400, "username and password are required")
		return
	}
	var user models.User
	if err := h.DB.Where("username = ?", strings.TrimSpace(input.Username)).First(&user).Error; err != nil || !user.IsActive || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)) != nil {
		failure(c, 401, "invalid username or password")
		return
	}
	token, err := utils.GenerateToken(h.Config.JWTSecret, h.Config.JWTExpiresIn, user.ID, user.Username, user.Role)
	if err != nil {
		failure(c, 500, "failed to create token")
		return
	}
	c.JSON(200, gin.H{"data": gin.H{"token": token, "user": user}, "token": token, "user": user})
}
func (h Handler) Logout(c *gin.Context) { success(c, 200, gin.H{"message": "logged out"}) }
func (h Handler) Me(c *gin.Context) {
	var user models.User
	if err := h.DB.First(&user, c.GetUint("user_id")).Error; err != nil {
		notFoundOrError(c, err, "user")
		return
	}
	success(c, 200, user)
}

func (h Handler) ManagementList(resource string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var destination any
		switch resource {
		case "products":
			destination = &[]models.Product{}
		case "categories":
			destination = &[]models.Category{}
		case "testimonials":
			destination = &[]models.Testimonial{}
		case "banners":
			destination = &[]models.Banner{}
		case "pages":
			destination = &[]models.Page{}
		case "payments":
			destination = &[]models.PaymentMethod{}
		case "shipping":
			destination = &[]models.ShippingInfo{}
		case "users":
			destination = &[]models.User{}
		case "settings":
			destination = &[]models.Setting{}
		case "contacts":
			destination = &[]models.Contact{}
		default:
			failure(c, 404, "resource not found")
			return
		}
		query := h.DB.Order("created_at DESC")
		if resource == "products" {
			query = query.Preload("Category").Preload("Images")
		}
		if err := query.Find(destination).Error; err != nil {
			failure(c, 500, "failed to load data")
			return
		}
		success(c, 200, destination)
	}
}

func (h Handler) Create(resource string) gin.HandlerFunc {
	return func(c *gin.Context) {
		item, err := h.bindNew(c, resource)
		if err != nil {
			failure(c, 400, err.Error())
			return
		}
		if err = h.DB.Create(item).Error; err != nil {
			databaseError(c, err)
			return
		}
		if product, ok := item.(*models.Product); ok {
			h.DB.Preload("Category").Preload("Images").First(product, product.ID)
		}
		success(c, 201, item)
	}
}

func (h Handler) Update(resource string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := parseID(c)
		if !ok {
			return
		}
		if resource == "users" {
			h.updateUser(c, id)
			return
		}
		item, err := h.bindNew(c, resource)
		if err != nil {
			failure(c, 400, err.Error())
			return
		}
		result := h.DB.Model(item).Where("id = ?", id).Select("*").Omit("id", "created_at").Updates(item)
		if result.Error != nil {
			databaseError(c, result.Error)
			return
		}
		if result.RowsAffected == 0 {
			failure(c, 404, resource+" not found")
			return
		}
		if err := h.DB.First(item, id).Error; err != nil {
			notFoundOrError(c, err, resource)
			return
		}
		if product, ok := item.(*models.Product); ok {
			h.DB.Preload("Category").Preload("Images").First(product, product.ID)
		}
		success(c, 200, item)
	}
}

func (h Handler) Delete(resource string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := parseID(c)
		if !ok {
			return
		}
		if resource == "users" && id == c.GetUint("user_id") {
			failure(c, 400, "cannot delete your own account")
			return
		}
		item, err := modelFor(resource)
		if err != nil {
			failure(c, 404, err.Error())
			return
		}
		result := h.DB.Delete(item, id)
		if result.Error != nil {
			databaseError(c, result.Error)
			return
		}
		if result.RowsAffected == 0 {
			failure(c, 404, resource+" not found")
			return
		}
		success(c, 200, gin.H{"id": id})
	}
}

func (h Handler) AddProductImage(c *gin.Context) {
	productID, ok := parseID(c)
	if !ok {
		return
	}
	var product models.Product
	if err := h.DB.First(&product, productID).Error; err != nil {
		notFoundOrError(c, err, "product")
		return
	}
	var input struct {
		ImageURL  string `json:"image_url" binding:"required,url,max=1000"`
		IsPrimary bool   `json:"is_primary"`
		SortOrder int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		failure(c, 400, "valid image_url is required")
		return
	}
	item := models.ProductImage{ProductID: productID, ImageURL: input.ImageURL, IsPrimary: input.IsPrimary, SortOrder: input.SortOrder}
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if item.IsPrimary {
			if err := tx.Model(&models.ProductImage{}).Where("product_id = ?", productID).Update("is_primary", false).Error; err != nil {
				return err
			}
		}
		return tx.Create(&item).Error
	})
	if err != nil {
		databaseError(c, err)
		return
	}
	success(c, 201, item)
}
func (h Handler) DeleteProductImage(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	result := h.DB.Delete(&models.ProductImage{}, id)
	if result.Error != nil {
		databaseError(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		failure(c, 404, "image not found")
		return
	}
	success(c, 200, gin.H{"id": id})
}

func (h Handler) UpdateSetting(c *gin.Context) {
	var input struct {
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		failure(c, 400, "invalid value")
		return
	}
	var item models.Setting
	result := h.DB.Model(&item).Where("key = ?", c.Param("key")).Update("value", input.Value)
	if result.Error != nil {
		databaseError(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		failure(c, 404, "setting not found")
		return
	}
	h.DB.Where("key = ?", c.Param("key")).First(&item)
	success(c, 200, item)
}

func (h Handler) bindNew(c *gin.Context, resource string) (any, error) {
	item, err := modelFor(resource)
	if err != nil {
		return nil, err
	}
	if err = c.ShouldBindJSON(item); err != nil {
		return nil, errors.New("invalid request data")
	}
	switch value := item.(type) {
	case *models.Testimonial:
		if value.Rating < 1 || value.Rating > 5 {
			return nil, errors.New("rating must be between 1 and 5")
		}
	case *models.User:
		if value.Username == "" || value.FullName == "" {
			return nil, errors.New("username and full_name are required")
		}
	}
	return item, nil
}

func modelFor(resource string) (any, error) {
	switch resource {
	case "products":
		return &models.Product{}, nil
	case "categories":
		return &models.Category{}, nil
	case "testimonials":
		return &models.Testimonial{}, nil
	case "banners":
		return &models.Banner{}, nil
	case "pages":
		return &models.Page{}, nil
	case "payments":
		return &models.PaymentMethod{}, nil
	case "shipping":
		return &models.ShippingInfo{}, nil
	case "contacts":
		return &models.Contact{}, nil
	default:
		return nil, errors.New("resource not found")
	}
}

func (h Handler) CreateUser(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required,min=3,max=80"`
		Password string `json:"password" binding:"required,min=8,max=72"`
		FullName string `json:"full_name" binding:"required,max=160"`
		Role     string `json:"role" binding:"required,oneof=super_admin staff"`
		IsActive *bool  `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		failure(c, 400, "invalid user data")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		failure(c, 500, "failed to secure password")
		return
	}
	active := true
	if input.IsActive != nil {
		active = *input.IsActive
	}
	item := models.User{Username: strings.TrimSpace(input.Username), PasswordHash: string(hash), FullName: strings.TrimSpace(input.FullName), Role: input.Role, IsActive: active}
	if err = h.DB.Create(&item).Error; err != nil {
		databaseError(c, err)
		return
	}
	success(c, 201, item)
}
func (h Handler) updateUser(c *gin.Context, id uint) {
	var input struct {
		Username string `json:"username" binding:"required,min=3,max=80"`
		Password string `json:"password" binding:"omitempty,min=8,max=72"`
		FullName string `json:"full_name" binding:"required,max=160"`
		Role     string `json:"role" binding:"required,oneof=super_admin staff"`
		IsActive *bool  `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		failure(c, 400, "invalid user data")
		return
	}
	updates := map[string]any{"username": strings.TrimSpace(input.Username), "full_name": strings.TrimSpace(input.FullName), "role": input.Role}
	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
	}
	if input.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			failure(c, 500, "failed to secure password")
			return
		}
		updates["password_hash"] = string(hash)
	}
	result := h.DB.Model(&models.User{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		databaseError(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		failure(c, 404, "user not found")
		return
	}
	var user models.User
	h.DB.First(&user, id)
	success(c, 200, user)
}

func positiveInt(value string, fallback int) int {
	number, err := strconv.Atoi(value)
	if err != nil || number < 1 {
		return fallback
	}
	return number
}
func parseID(c *gin.Context) (uint, bool) {
	value, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || value == 0 {
		failure(c, 400, "invalid id")
		return 0, false
	}
	return uint(value), true
}
func notFoundOrError(c *gin.Context, err error, name string) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		failure(c, 404, name+" not found")
		return
	}
	failure(c, 500, "database error")
}
func databaseError(c *gin.Context, err error) {
	message := "database operation failed"
	lower := strings.ToLower(err.Error())
	if strings.Contains(lower, "duplicate") || strings.Contains(lower, "unique") {
		message = "resource already exists"
		failure(c, 409, message)
		return
	}
	if strings.Contains(lower, "foreign key") {
		message = "resource is in use"
		failure(c, 409, message)
		return
	}
	failure(c, 500, message)
}
