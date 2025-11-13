package main

import (
	"context"
	_ "gofiber-mongo/docs" // Import docs for Swagger
	"gofiber-mongo/app/model"
	"gofiber-mongo/app/repository"
	"gofiber-mongo/middleware"
	"gofiber-mongo/route"
	"gofiber-mongo/utils"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	fiberSwagger "github.com/swaggo/fiber-swagger" // Import fiber-swagger
)

// @title Alumni Management API
// @version 1.0
// @description API untuk mengelola data alumni dengan MongoDB menggunakan Clean Architecture
// @host localhost:3000
// @BasePath /api
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func connectMongoDB() *mongo.Database {
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
		log.Println("Peringatan: MONGODB_URI tidak disetel. Menggunakan default:", mongoURI)
	}

	databaseName := os.Getenv("DATABASE_NAME")
	if databaseName == "" {
		databaseName = "alumni_db"
		log.Println("Peringatan: DATABASE_NAME tidak disetel. Menggunakan default:", databaseName)
	}

	clientOptions := options.Client().ApplyURI(mongoURI)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("Koneksi ke MongoDB gagal: %v", err)
	}

	// Ping untuk verifikasi koneksi
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Ping ke MongoDB gagal: %v", err)
	}

	fmt.Println("✓ Berhasil terhubung ke MongoDB!")
	return client.Database(databaseName)
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("File .env tidak ditemukan, menggunakan environment variables sistem")
	}

	if err := os.Getenv("APP_PORT"); err == "" {
		os.Setenv("APP_PORT", "3000")
	}

	db := connectMongoDB()

	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024, // 10MB
	})

	app.Static("/uploads", "./uploads")

	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// =====================
	// REGISTER HANDLER
	// =====================
	app.Post("/api/register", func(c *fiber.Ctx) error {
		var req model.LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Request tidak valid"})
		}

		if req.Username == "" || req.Password == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Username dan password harus diisi"})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		userRepo := repository.NewUserRepository(db)

		existingUser, _ := userRepo.FindByUsername(ctx, req.Username)
		if existingUser != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Username sudah terdaftar"})
		}

		hashedPassword, err := utils.HashPassword(req.Password)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Gagal hash password"})
		}

		newUser := &model.User{
			Username:  req.Username,
			Email:     req.Username + "@example.com",
			Password:  hashedPassword,
			Role:      "user",
			IsDelete:  false,
			CreatedAt: time.Now(),
		}

		createdUser, err := userRepo.Create(ctx, newUser)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Gagal membuat user"})
		}

		return c.Status(201).JSON(fiber.Map{
			"message": "User berhasil dibuat",
			"user": fiber.Map{
				"id":       createdUser.ID,
				"username": createdUser.Username,
				"role":     createdUser.Role,
			},
		})
	})

	// =====================
	// LOGIN HANDLER
	// =====================
	app.Post("/api/login", func(c *fiber.Ctx) error {
		var req model.LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Request tidak valid"})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		userRepo := repository.NewUserRepository(db)
		user, err := userRepo.FindByUsername(ctx, req.Username)
		if err != nil || user == nil {
			user, err = userRepo.FindByEmail(ctx, req.Username)
			if err != nil || user == nil {
				return c.Status(401).JSON(fiber.Map{"error": "Username atau password salah"})
			}
		}

		if user.IsDelete {
			return c.Status(401).JSON(fiber.Map{"error": "User telah dihapus"})
		}

		if !utils.CheckPassword(req.Password, user.Password) {
			return c.Status(401).JSON(fiber.Map{"error": "Username atau password salah"})
		}

		token, err := utils.GenerateToken(user)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Gagal generate token"})
		}

		return c.JSON(model.LoginResponse{
			User:  *user,
			Token: token,
		})
	})

	// =====================
	// PROTECTED ROUTES
	// =====================
	app.Use("/api", middleware.AuthRequired())

	route.RegisterRoutes(app, db)

	// Run server
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}
	fmt.Println("Server jalan di http://localhost:" + port)
	fmt.Println("Swagger UI tersedia di http://localhost:" + port + "/swagger/index.html")
	log.Fatal(app.Listen(":" + port))
}
