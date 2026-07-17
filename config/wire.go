//go:build wireinject
// +build wireinject

/*
 * Contract ID: CTR-006
 * Service Name: SupermarketService
 * Description: Google Wire Dependency Injection configuration and injector setup.
 *              Declares all application provider sets, interface bindings,
 *              and compiles the complete composition root for SupermarketService.
 */

package config

import (
	"github.com/google/wire"

	"supermarket-backend/controllers"
	"supermarket-backend/repositories"
	"supermarket-backend/services"
)

// 1. Core Base Configuration, Logging, Caching, and DB providers
var BaseProviderSet = wire.NewSet(
	Load,
	NewLogger,
	NewDatabase,
	NewCache,
)

// 2. Repository Layer providers
var RepositoryProviderSet = wire.NewSet(
	repositories.NewUserRepository,
	repositories.NewCategoryRepository,
	repositories.NewClientRepository,
	repositories.NewProductRepository,
	repositories.NewPromotionRepository,
	repositories.NewSaleRepository,
)

// 3. Service Layer providers
var ServiceProviderSet = wire.NewSet(
	services.NewUserService,
	services.NewCategoryService,
	services.NewClientService,
	services.NewProductService,
	services.NewPromotionService,
	services.NewSaleService,
)

// 4. Presentation / Delivery Layer providers (REST Controllers)
var PresentationProviderSet = wire.NewSet(
	controllers.NewUserController,
	controllers.NewCategoryController,
	controllers.NewClientController,
	controllers.NewProductController,
	controllers.NewPromotionController,
	controllers.NewSaleController,
)

// InitializeApp is the main Google Wire injector function.
// It assembles the entire component graph starting from the environment variables.
func InitializeApp() (*App, error) {
	wire.Build(
		BaseProviderSet,
		RepositoryProviderSet,
		ServiceProviderSet,
		PresentationProviderSet,
		NewApp,
	)
	return nil, nil
}
