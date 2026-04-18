//go:build wireinject
// +build wireinject

package injector

import (
	"log/slog"
	"net/http"

	"github.com/assidik12/go-restfull-api/config"
	handler "github.com/assidik12/go-restfull-api/internal/delivery/http/handler"
	"github.com/assidik12/go-restfull-api/internal/delivery/http/middleware"
	"github.com/assidik12/go-restfull-api/internal/delivery/http/route"

	"github.com/assidik12/go-restfull-api/internal/event"
	"github.com/assidik12/go-restfull-api/internal/infrastructure"
	"github.com/assidik12/go-restfull-api/internal/pkg/logger"
	mysql "github.com/assidik12/go-restfull-api/internal/repository/mysql"
	redis "github.com/assidik12/go-restfull-api/internal/repository/redis"
	service "github.com/assidik12/go-restfull-api/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/google/wire"
	"github.com/julienschmidt/httprouter"
)

var validatorSet = wire.NewSet(
	validator.New,
	wire.Value([]validator.Option{}),
)

// Setup Kafka
var eventSet = wire.NewSet(
	infrastructure.NewKafkaWriter,
	event.NewKafkaProducer,
	wire.Bind(new(event.Producer), new(*event.KafkaProducer)),
)

var userSet = wire.NewSet(
	mysql.NewUserRepository,
	service.NewUserService,
	handler.NewUserHandler,
)

var productSet = wire.NewSet(
	mysql.NewProductRepository,
	service.NewProductService,
	handler.NewProductHandler,
)

var transactionSet = wire.NewSet(
	mysql.NewTransactionRepository,
	service.NewTransactionService,
	handler.NewTransactionHandler,
)

// ProvideLogger initializes slog based on config environment.
func ProvideLogger(cfg config.Config) *slog.Logger {
	return logger.New(cfg.AppEnv)
}

// ExtractJwtSecret extracts the JWT secret from config for Wire to inject into NewUserService.
func ExtractJwtSecret(cfg config.Config) string {
	return cfg.JWTSecret
}

func InitializedServer(cfg config.Config) (*http.Server, func(), error) {
	wire.Build(
		// 1. Infrastructure (Singletons)
		infrastructure.DatabaseConnection,
		infrastructure.RedisConnection, 

		// 2. Utils & Wrappers
		redis.NewWrapper,
		validatorSet,
		ExtractJwtSecret,
		ProvideLogger,

		// 3. Feature Sets
		eventSet,
		userSet,
		productSet,
		transactionSet,

		// 4. HTTP & Routing
		route.NewRouter,
		wire.Bind(new(http.Handler), new(*httprouter.Router)),
		middleware.NewAuthMiddleware,
		config.NewServer,
	)
	return nil, nil, nil
}
