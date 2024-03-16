package main

import (
	"context"
	"fmt"
	"log"

	bankaccount "github.com/ahmadnaufal/openidea-shopifyx/internal/bank_account"
	"github.com/ahmadnaufal/openidea-shopifyx/internal/config"
	"github.com/ahmadnaufal/openidea-shopifyx/internal/image"
	"github.com/ahmadnaufal/openidea-shopifyx/internal/product"
	"github.com/ahmadnaufal/openidea-shopifyx/internal/user"
	"github.com/ahmadnaufal/openidea-shopifyx/pkg/jwt"
	"github.com/ahmadnaufal/openidea-shopifyx/pkg/middleware"
	"github.com/ahmadnaufal/openidea-shopifyx/pkg/s3"

	"github.com/ansrivas/fiberprometheus/v2"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
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
	// custom middleware to set all method not allowed response to not found
	app.Use(middleware.CustomMiddleware404())

	jwtProvider := jwt.NewJWTProvider(cfg.JWTSecret)

	db := connectToDB(cfg.Database, cfg.Env)

	trxProvider := config.NewTransactionProvider(db)

	awsCfg, err := awsConfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}

	s3Provider := s3.NewS3Provider(awsCfg, cfg.S3.Bucket, cfg.S3.Region, cfg.S3.ID, cfg.S3.SecretKey)

	userRepo := user.NewUserRepo(db)
	user.UserRepoImpl = &userRepo
	user.JwtProvider = &jwtProvider
	user.SaltCost = cfg.BcryptSalt

	productRepo := product.NewProductRepo(db)
	product.ProductRepoImpl = &productRepo
	product.TrxProvider = &trxProvider
	product.UserRepoImpl = &userRepo

	bankAccountRepo := bankaccount.NewBankAccountRepo(db)
	bankaccount.BankAccountRepoImpl = &bankAccountRepo
	product.BankAccountRepoImpl = &bankAccountRepo

	image.S3ProviderImpl = &s3Provider

	// setup instrumentation
	prometheus := fiberprometheus.New("shopifyx")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	// register routes
	user.RegisterRoute(app)
	product.RegisterRoute(app, jwtProvider)
	bankaccount.RegisterRoute(app, jwtProvider)
	image.RegisterRoute(app, jwtProvider)

	addr := fmt.Sprintf(":%s", cfg.AppPort)

	log.Fatal(app.Listen(addr))
}

func connectToDB(dbCfg config.DatabaseConfig, env string) *sqlx.DB {
	var dsn string
	if env == "production" {
		dsn = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=verify-full&sslrootcert=ap-southeast-1-bundle.pem",
			dbCfg.Username, dbCfg.Password, dbCfg.Host,
			dbCfg.Port, dbCfg.Name,
		)
	} else {
		dsn = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			dbCfg.Username, dbCfg.Password, dbCfg.Host,
			dbCfg.Port, dbCfg.Name,
		)
	}

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
