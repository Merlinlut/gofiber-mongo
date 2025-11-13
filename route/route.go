package route

import (
	"gofiber-mongo/app/repository"
	"gofiber-mongo/app/service"
	"gofiber-mongo/middleware"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func RegisterRoutes(app *fiber.App, db *mongo.Database) {
	alumniRepo := repository.NewAlumniRepository(db)
	alumniService := service.NewAlumniService(alumniRepo)

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)

	pekerjaanRepo := repository.NewPekerjaanRepository(db)
	pekerjaanService := service.NewPekerjaanService(pekerjaanRepo, db)

	fileRepo := repository.NewFileRepository(db)
	fileService := service.NewFileService(fileRepo, "./uploads")

	api := app.Group("/api")

	// Restore pekerjaan dari trash
	api.Put("/trash/pekerjaan/:id/restore", middleware.AuthRequired(), pekerjaanService.Restore)

	// Hard delete pekerjaan
	api.Delete("/trash/pekerjaan/:id/permanent", middleware.AuthRequired(), pekerjaanService.HardDelete)

	// Trash pekerjaan
	api.Get("/trash/pekerjaan", middleware.AuthRequired(), pekerjaanService.GetTrashed)

	// user soft delete (admin only)
	api.Delete("/users/:id", middleware.AdminOnly(), userService.SoftDelete)

	// alumni soft delete (admin only)
	api.Delete("/alumni/:id", middleware.AdminOnly(), alumniService.SoftDelete)

	// pekerjaan soft delete
	api.Delete("/pekerjaan/:id", pekerjaanService.SoftDelete)

	// endpoint alumni tanpa pekerjaan
	api.Get("/alumni/tanpa-pekerjaan", alumniService.GetWithoutPekerjaan)

	// Alumni (protected)
	alumni := api.Group("/alumni", middleware.AuthRequired())
	alumni.Get("/", alumniService.GetAll)                 // admin + user
	alumni.Get("/:id", alumniService.GetByID)             // admin + user
	alumni.Post("/", middleware.AdminOnly(), alumniService.Create)
	alumni.Put("/:id", middleware.AdminOnly(), alumniService.Update)
	alumni.Delete("/:id", middleware.AdminOnly(), alumniService.Delete)

	// Pekerjaan (protected)
	pekerjaan := api.Group("/pekerjaan", middleware.AuthRequired())
	pekerjaan.Get("/", pekerjaanService.GetAll)                    // admin + user
	pekerjaan.Get("/:id", pekerjaanService.GetByID)                // admin + user
	pekerjaan.Get("/alumni/:alumni_id", middleware.AdminOnly(), pekerjaanService.GetByAlumniID)
	pekerjaan.Post("/", middleware.AdminOnly(), pekerjaanService.Create)
	pekerjaan.Put("/:id", middleware.AdminOnly(), pekerjaanService.Update)
	pekerjaan.Delete("/:id", middleware.AdminOnly(), pekerjaanService.Delete)

	RegisterFileRoutes(app, fileService)
}

func RegisterFileRoutes(app *fiber.App, fileService service.IFileService) {
	api := app.Group("/api")

	// File upload routes - require authentication
	files := api.Group("/files", middleware.AuthRequired())

	// Photo routes
	files.Post("/photo/upload", fileService.UploadPhoto)
	files.Get("/photo/:alumni_id", fileService.GetPhotoByAlumniID)
	files.Delete("/photo/:id", fileService.DeletePhoto)

	// Certificate routes
	files.Post("/certificate/upload", fileService.UploadCertificate)
	files.Get("/certificate/:alumni_id", fileService.GetCertificateByAlumniID)
	files.Delete("/certificate/:id", fileService.DeleteCertificate)
}
