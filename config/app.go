/*
 * Contract ID: CTR-006
 * Service Name: SupermarketService
 * Description: Application composition root type. Kept outside wire.go (which
 *              carries the `wireinject` build tag and is excluded from normal
 *              builds) so it's visible to both wire_gen.go and main.go.
 */

package config

import (
	"github.com/dgraph-io/ristretto"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"supermarket-backend/controllers"
)

// App acts as the monolithic application composition root.
// It aggregates all components needed to run the HTTP (Gin) and gRPC servers.
type App struct {
	Config              *Config
	Logger              *zap.Logger
	DB                  *gorm.DB
	Cache               *ristretto.Cache
	UserController      *controllers.UserController
	CategoryController  *controllers.CategoryController
	ClientController    *controllers.ClientController
	ProductController   *controllers.ProductController
	PromotionController *controllers.PromotionController
	SaleController      *controllers.SaleController
}

// NewApp is the constructor for the App struct, binding all controllers, middlewares, and services.
func NewApp(
	cfg *Config,
	logger *zap.Logger,
	db *gorm.DB,
	cache *ristretto.Cache,
	userCtrl *controllers.UserController,
	catCtrl *controllers.CategoryController,
	clientCtrl *controllers.ClientController,
	prodCtrl *controllers.ProductController,
	promoCtrl *controllers.PromotionController,
	saleCtrl *controllers.SaleController,
) *App {
	return &App{
		Config:              cfg,
		Logger:              logger,
		DB:                  db,
		Cache:               cache,
		UserController:      userCtrl,
		CategoryController:  catCtrl,
		ClientController:    clientCtrl,
		ProductController:   prodCtrl,
		PromotionController: promoCtrl,
		SaleController:      saleCtrl,
	}
}