package main

import (
	"fmt"
	"log"

	"github.com/ahmadnaufal/openidea-shopifyx/internal/config"
	"github.com/ahmadnaufal/openidea-shopifyx/internal/product"
	"github.com/ahmadnaufal/openidea-shopifyx/internal/user"
	"github.com/ahmadnaufal/openidea-shopifyx/pkg/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.InitializeConfig()

	app := fiber.New()
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(compress.New())

	jwtProvider := jwt.NewJWTProvider(cfg.JWTSecret)

	db := connectToDB(cfg.Database)

	trxProvider := config.NewTransactionProvider(db)

	userRepo := user.NewUserRepo(db)
	user.UserRepoImpl = &userRepo
	user.JwtProvider = &jwtProvider
	user.SaltCost = cfg.BcryptSalt

	productRepo := product.NewProductRepo(db)
	product.ProductRepoImpl = &productRepo
	product.TrxProvider = &trxProvider
	product.UserRepoImpl = &userRepo

	user.RegisterRoute(app)
	product.RegisterRoute(app, jwtProvider)

	addr := fmt.Sprintf(":%s", cfg.AppPort)

	log.Fatal(app.Listen(addr))
}

func connectToDB(dbCfg config.DatabaseConfig) *sqlx.DB {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbCfg.Username, dbCfg.Password, dbCfg.Host,
		dbCfg.Port, dbCfg.Name,
	)

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	db.SetMaxOpenConns(dbCfg.MaxOpenConnection)
	db.SetMaxIdleConns(dbCfg.MaxIdleConnection)

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	return db
}
