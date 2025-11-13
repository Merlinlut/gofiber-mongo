package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gofiber-mongo/app/model"
	"gofiber-mongo/app/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// IFileService defines the interface for file operations
type IFileService interface {
	UploadPhoto(c *fiber.Ctx) error
	UploadCertificate(c *fiber.Ctx) error
	GetPhotoByAlumniID(c *fiber.Ctx) error
	GetCertificateByAlumniID(c *fiber.Ctx) error
	DeletePhoto(c *fiber.Ctx) error
	DeleteCertificate(c *fiber.Ctx) error
}

// FileService implements IFileService
type FileService struct {
	repo       repository.IFileRepository
	uploadPath string
}

// NewFileService creates a new file service
func NewFileService(repo repository.IFileRepository, uploadPath string) IFileService {
	return &FileService{
		repo:       repo,
		uploadPath: uploadPath,
	}
}

// HandleUploadPhoto godoc
// @Summary Upload photo
// @Description Upload foto alumni (JPG, PNG, max 1MB)
// @Tags Files
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Photo file"
// @Param alumni_id formData string true "Alumni ID"
// @Success 201 {object} map[string]interface{} "upload success"
// @Failure 400 {object} map[string]interface{} "File atau alumni_id tidak valid"
// @Failure 403 {object} map[string]interface{} "Permission denied"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /files/photo/upload [post]
// @Security BearerAuth
// UploadPhoto handles photo upload with validation
func (s *FileService) UploadPhoto(c *fiber.Ctx) error {
	// Get file from form
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "File tidak ditemukan",
			"error":   err.Error(),
		})
	}

	// Get alumni_id from form
	alumniID := c.FormValue("alumni_id")
	if alumniID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "alumni_id diperlukan",
		})
	}

	maxSize := int64(1 * 1024 * 1024) // 1MB
	if fileHeader.Size > maxSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Ukuran file foto tidak boleh lebih dari 1MB",
		})
	}

	allowedPhotoTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/jpg":  true,
	}

	contentType := fileHeader.Header.Get("Content-Type")
	if !allowedPhotoTypes[contentType] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format foto hanya boleh JPEG, JPG, atau PNG",
		})
	}

	// Check authorization
	userID := c.Locals("user_id").(primitive.ObjectID)
	role := c.Locals("role").(string)

	// Convert alumni_id to ObjectID
	alumniObjID, err := primitive.ObjectIDFromHex(alumniID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID alumni tidak valid",
		})
	}

	// Verify alumni exists and check ownership (only for non-admin)
	if role != "admin" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		isOwner, err := s.repo.CheckAlumniOwnership(ctx, alumniID, userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Gagal memverifikasi kepemilikan alumni",
				"error":   err.Error(),
			})
		}

		if !isOwner {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "Anda tidak memiliki izin untuk mengupload foto untuk alumni ini",
			})
		}
	}

	// Generate unique filename
	ext := filepath.Ext(fileHeader.Filename)
	newFileName := uuid.New().String() + ext
	uploadDir := filepath.Join(s.uploadPath, "photos")
	filePath := filepath.Join(uploadDir, newFileName)

	// Create directory if not exists
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat direktori upload",
			"error":   err.Error(),
		})
	}

	// Save file
	if err := c.SaveFile(fileHeader, filePath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menyimpan file",
			"error":   err.Error(),
		})
	}

	// Save metadata to database
	photo := &model.Photo{
		AlumniID:  alumniObjID,
		UserID:    userID,
		FileName:  newFileName,
		FilePath:  filePath,
		FileSize:  fileHeader.Size,
		FileType:  contentType,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.repo.CreatePhoto(ctx, photo); err != nil {
		os.Remove(filePath)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menyimpan metadata file",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Foto berhasil diupload",
		"data":    s.toPhotoResponse(photo),
	})
}

// HandleUploadCertificate godoc
// @Summary Upload certificate
// @Description Upload sertifikat alumni (PDF, max 2MB)
// @Tags Files
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Certificate file"
// @Param alumni_id formData string true "Alumni ID"
// @Success 201 {object} map[string]interface{} "upload success"
// @Failure 400 {object} map[string]interface{} "File atau alumni_id tidak valid"
// @Failure 403 {object} map[string]interface{} "Permission denied"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /files/certificate/upload [post]
// @Security BearerAuth
// UploadCertificate handles certificate upload with validation
func (s *FileService) UploadCertificate(c *fiber.Ctx) error {
	// Get file from form
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "File tidak ditemukan",
			"error":   err.Error(),
		})
	}

	// Get alumni_id from form
	alumniID := c.FormValue("alumni_id")
	if alumniID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "alumni_id diperlukan",
		})
	}

	maxSize := int64(2 * 1024 * 1024) // 2MB
	if fileHeader.Size > maxSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Ukuran file sertifikat tidak boleh lebih dari 2MB",
		})
	}

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType != "application/pdf" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format sertifikat hanya boleh PDF",
		})
	}

	// Check authorization
	userID := c.Locals("user_id").(primitive.ObjectID)
	role := c.Locals("role").(string)

	// Convert alumni_id to ObjectID
	alumniObjID, err := primitive.ObjectIDFromHex(alumniID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID alumni tidak valid",
		})
	}

	// Verify alumni exists and check ownership (only for non-admin)
	if role != "admin" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		isOwner, err := s.repo.CheckAlumniOwnership(ctx, alumniID, userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Gagal memverifikasi kepemilikan alumni",
				"error":   err.Error(),
			})
		}

		if !isOwner {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "Anda tidak memiliki izin untuk mengupload sertifikat untuk alumni ini",
			})
		}
	}

	// Generate unique filename
	ext := filepath.Ext(fileHeader.Filename)
	newFileName := uuid.New().String() + ext
	uploadDir := filepath.Join(s.uploadPath, "certificates")
	filePath := filepath.Join(uploadDir, newFileName)

	// Create directory if not exists
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat direktori upload",
			"error":   err.Error(),
		})
	}

	// Save file
	if err := c.SaveFile(fileHeader, filePath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menyimpan file",
			"error":   err.Error(),
		})
	}

	// Save metadata to database
	cert := &model.Certificate{
		AlumniID:  alumniObjID,
		UserID:    userID,
		FileName:  newFileName,
		FilePath:  filePath,
		FileSize:  fileHeader.Size,
		FileType:  contentType,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.repo.CreateCertificate(ctx, cert); err != nil {
		os.Remove(filePath)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menyimpan metadata file",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Sertifikat berhasil diupload",
		"data":    s.toCertificateResponse(cert),
	})
}

// HandleGetPhotoByAlumniID godoc
// @Summary Get photo by Alumni ID
// @Description Mengambil foto berdasarkan Alumni ID
// @Tags Files
// @Accept json
// @Produce json
// @Param alumni_id path string true "Alumni ID"
// @Success 200 {object} map[string]interface{} "photo data"
// @Failure 404 {object} map[string]interface{} "Foto tidak ditemukan"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /files/photo/{alumni_id} [get]
// @Security BearerAuth
// GetPhotoByAlumniID retrieves photo by alumni ID
func (s *FileService) GetPhotoByAlumniID(c *fiber.Ctx) error {
	alumniID := c.Params("alumni_id")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	photo, err := s.repo.FindPhotoByAlumniID(ctx, alumniID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data foto",
			"error":   err.Error(),
		})
	}

	if photo == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Foto tidak ditemukan",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Foto berhasil diambil",
		"data":    s.toPhotoResponse(photo),
	})
}

// HandleGetCertificateByAlumniID godoc
// @Summary Get certificate by Alumni ID
// @Description Mengambil sertifikat berdasarkan Alumni ID
// @Tags Files
// @Accept json
// @Produce json
// @Param alumni_id path string true "Alumni ID"
// @Success 200 {object} map[string]interface{} "certificate data"
// @Failure 404 {object} map[string]interface{} "Sertifikat tidak ditemukan"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /files/certificate/{alumni_id} [get]
// @Security BearerAuth
// GetCertificateByAlumniID retrieves certificate by alumni ID
func (s *FileService) GetCertificateByAlumniID(c *fiber.Ctx) error {
	alumniID := c.Params("alumni_id")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cert, err := s.repo.FindCertificateByAlumniID(ctx, alumniID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data sertifikat",
			"error":   err.Error(),
		})
	}

	if cert == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Sertifikat tidak ditemukan",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Sertifikat berhasil diambil",
		"data":    s.toCertificateResponse(cert),
	})
}

// HandleDeletePhoto godoc
// @Summary Delete photo
// @Description Menghapus foto berdasarkan ID
// @Tags Files
// @Accept json
// @Produce json
// @Param id path string true "Photo ID"
// @Success 200 {object} map[string]interface{} "success response"
// @Failure 403 {object} map[string]interface{} "Permission denied"
// @Failure 404 {object} map[string]interface{} "Foto tidak ditemukan"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /files/photo/{id} [delete]
// @Security BearerAuth
// DeletePhoto deletes a photo
func (s *FileService) DeletePhoto(c *fiber.Ctx) error {
	photoID := c.Params("id")
	userID := c.Locals("user_id").(primitive.ObjectID)
	role := c.Locals("role").(string)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	photo, err := s.repo.FindPhotoByID(ctx, photoID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data foto",
			"error":   err.Error(),
		})
	}

	if photo == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Foto tidak ditemukan",
		})
	}

	// Check authorization: only owner or admin can delete
	if role != "admin" && photo.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Anda tidak memiliki izin untuk menghapus foto ini",
		})
	}

	// Delete file from storage
	if err := os.Remove(photo.FilePath); err != nil {
		fmt.Println("Warning: Gagal menghapus file dari storage:", err)
	}

	// Soft delete from database
	if err := s.repo.DeletePhoto(ctx, photoID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus foto",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Foto berhasil dihapus",
	})
}

// HandleDeleteCertificate godoc
// @Summary Delete certificate
// @Description Menghapus sertifikat berdasarkan ID
// @Tags Files
// @Accept json
// @Produce json
// @Param id path string true "Certificate ID"
// @Success 200 {object} map[string]interface{} "success response"
// @Failure 403 {object} map[string]interface{} "Permission denied"
// @Failure 404 {object} map[string]interface{} "Sertifikat tidak ditemukan"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /files/certificate/{id} [delete]
// @Security BearerAuth
// DeleteCertificate deletes a certificate
func (s *FileService) DeleteCertificate(c *fiber.Ctx) error {
	certID := c.Params("id")
	userID := c.Locals("user_id").(primitive.ObjectID)
	role := c.Locals("role").(string)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cert, err := s.repo.FindCertificateByID(ctx, certID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data sertifikat",
			"error":   err.Error(),
		})
	}

	if cert == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Sertifikat tidak ditemukan",
		})
	}

	// Check authorization: only owner or admin can delete
	if role != "admin" && cert.UserID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Anda tidak memiliki izin untuk menghapus sertifikat ini",
		})
	}

	// Delete file from storage
	if err := os.Remove(cert.FilePath); err != nil {
		fmt.Println("Warning: Gagal menghapus file dari storage:", err)
	}

	// Soft delete from database
	if err := s.repo.DeleteCertificate(ctx, certID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus sertifikat",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Sertifikat berhasil dihapus",
	})
}

// Helper functions
func (s *FileService) toPhotoResponse(photo *model.Photo) *model.PhotoResponse {
	return &model.PhotoResponse{
		ID:         photo.ID.Hex(),
		AlumniID:   photo.AlumniID.Hex(),
		UserID:     photo.UserID.Hex(),
		FileName:   photo.FileName,
		FilePath:   photo.FilePath,
		FileSize:   photo.FileSize,
		FileType:   photo.FileType,
		UploadedAt: photo.UploadedAt,
	}
}

func (s *FileService) toCertificateResponse(cert *model.Certificate) *model.CertificateResponse {
	return &model.CertificateResponse{
		ID:         cert.ID.Hex(),
		AlumniID:   cert.AlumniID.Hex(),
		UserID:     cert.UserID.Hex(),
		FileName:   cert.FileName,
		FilePath:   cert.FilePath,
		FileSize:   cert.FileSize,
		FileType:   cert.FileType,
		UploadedAt: cert.UploadedAt,
	}
}
