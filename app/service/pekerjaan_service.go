package service

import (
	"context"
	"errors"
	"gofiber-mongo/app/model"
	"gofiber-mongo/app/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

type PekerjaanService struct {
	Repo *repository.PekerjaanRepository
	DB   *mongo.Database
}

func NewPekerjaanService(repo *repository.PekerjaanRepository, db *mongo.Database) *PekerjaanService {
	return &PekerjaanService{
		Repo: repo,
		DB:   db,
	}
}

func (s *PekerjaanService) validateCreateRequest(req model.CreatePekerjaanRequest) error {
	if req.AlumniID == "" {
		return errors.New("alumni_id tidak boleh kosong")
	}
	if req.NamaPerusahaan == "" {
		return errors.New("nama_perusahaan tidak boleh kosong")
	}
	if req.PosisiJabatan == "" {
		return errors.New("posisi_jabatan tidak boleh kosong")
	}
	if req.BidangIndustri == "" {
		return errors.New("bidang_industri tidak boleh kosong")
	}
	if req.LokasiKerja == "" {
		return errors.New("lokasi_kerja tidak boleh kosong")
	}
	if req.TanggalMulaiKerja == "" {
		return errors.New("tanggal_mulai_kerja tidak boleh kosong")
	}
	if req.StatusPekerjaan == "" {
		return errors.New("status_pekerjaan tidak boleh kosong")
	}
	return nil
}

func (s *PekerjaanService) validateUpdateRequest(req model.UpdatePekerjaanRequest) error {
	if req.NamaPerusahaan == "" {
		return errors.New("nama_perusahaan tidak boleh kosong")
	}
	if req.PosisiJabatan == "" {
		return errors.New("posisi_jabatan tidak boleh kosong")
	}
	if req.BidangIndustri == "" {
		return errors.New("bidang_industri tidak boleh kosong")
	}
	if req.LokasiKerja == "" {
		return errors.New("lokasi_kerja tidak boleh kosong")
	}
	if req.TanggalMulaiKerja == "" {
		return errors.New("tanggal_mulai_kerja tidak boleh kosong")
	}
	if req.StatusPekerjaan == "" {
		return errors.New("status_pekerjaan tidak boleh kosong")
	}
	return nil
}

// HandleRestore godoc
// @Summary Restore pekerjaan dari trash
// @Description Mengembalikan pekerjaan yang sudah dihapus (soft delete) kembali ke data aktif
// @Tags Pekerjaan
// @Accept json
// @Produce json
// @Param id path string true "Pekerjaan ID"
// @Success 200 {object} map[string]interface{} "success response"
// @Failure 400 {object} map[string]interface{} "ID tidak valid"
// @Failure 404 {object} map[string]interface{} "Data tidak ditemukan"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /trash/pekerjaan/{id}/restore [put]
// @Security BearerAuth
func (s *PekerjaanService) Restore(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	userIDStr := c.Locals("user_id").(string)
	userID, _ := primitive.ObjectIDFromHex(userIDStr)

	idStr := c.Params("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID tidak valid"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get pekerjaan to check alumni_id
	pekerjaan, err := s.Repo.GetByID(ctx, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if pekerjaan == nil || !pekerjaan.IsDelete {
		return c.Status(404).JSON(fiber.Map{"error": "Data tidak ditemukan atau belum dihapus"})
	}

	if role == "admin" {
		err = s.Repo.RestoreByID(ctx, id)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Data pekerjaan berhasil direstore oleh admin",
		})
	}

	alumniRepo := repository.NewAlumniRepository(s.DB)
	alumni, err := alumniRepo.GetByID(ctx, userID)
	if err != nil || alumni == nil {
		return c.Status(403).JSON(fiber.Map{"error": "Data alumni tidak ditemukan"})
	}

	if alumni.ID != pekerjaan.AlumniID {
		return c.Status(403).JSON(fiber.Map{"error": "Anda tidak berhak me-restore data ini"})
	}

	err = s.Repo.RestoreByIDAndAlumni(ctx, id, alumni.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Data pekerjaan berhasil direstore oleh user",
	})
}

// HandleHardDelete godoc
// @Summary Hard delete pekerjaan secara permanent
// @Description Menghapus pekerjaan secara permanent dari database (hanya data yang sudah soft delete)
// @Tags Pekerjaan
// @Accept json
// @Produce json
// @Param id path string true "Pekerjaan ID"
// @Success 200 {object} map[string]interface{} "success response"
// @Failure 400 {object} map[string]interface{} "ID tidak valid"
// @Failure 404 {object} map[string]interface{} "Data tidak ditemukan"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /trash/pekerjaan/{id}/permanent [delete]
// @Security BearerAuth
func (s *PekerjaanService) HardDelete(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	userIDStr := c.Locals("user_id").(string)
	userID, _ := primitive.ObjectIDFromHex(userIDStr)

	idStr := c.Params("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID tidak valid"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pekerjaan, err := s.Repo.GetByID(ctx, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if pekerjaan == nil || !pekerjaan.IsDelete {
		return c.Status(404).JSON(fiber.Map{"error": "Data tidak ditemukan atau belum dihapus (soft delete)"})
	}

	if role == "admin" {
		err = s.Repo.HardDeleteByID(ctx, id)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Data pekerjaan dihapus permanen oleh admin",
		})
	}

	alumniRepo := repository.NewAlumniRepository(s.DB)
	alumni, err := alumniRepo.GetByID(ctx, userID)
	if err != nil || alumni == nil {
		return c.Status(403).JSON(fiber.Map{"error": "Data alumni tidak ditemukan"})
	}

	if alumni.ID != pekerjaan.AlumniID {
		return c.Status(403).JSON(fiber.Map{"error": "Anda tidak berhak menghapus data pekerjaan ini"})
	}

	err = s.Repo.HardDeleteByIDAndAlumni(ctx, id, alumni.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Data pekerjaan dihapus permanen oleh user",
	})
}

// HandleGetTrashed godoc
// @Summary Get trashed pekerjaan
// @Description Mengambil daftar pekerjaan yang sudah dihapus (soft delete)
// @Tags Pekerjaan
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "trashed data list"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /trash/pekerjaan [get]
// @Security BearerAuth
func (s *PekerjaanService) GetTrashed(c *fiber.Ctx) error {
	role := c.Locals("role").(string)
	userIDStr := c.Locals("user_id").(string)
	userID, _ := primitive.ObjectIDFromHex(userIDStr)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if role == "admin" {
		data, err := s.Repo.GetTrashed(ctx)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Data pekerjaan yang sudah di-soft delete semua",
			"data":    data,
		})
	}

	alumniRepo := repository.NewAlumniRepository(s.DB)
	alumni, err := alumniRepo.GetByID(ctx, userID)
	if err != nil || alumni == nil {
		return c.Status(403).JSON(fiber.Map{"error": "Data alumni tidak ditemukan"})
	}

	ownData, err := s.Repo.GetTrashedByAlumni(ctx, alumni.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Data pekerjaan yang sudah di-soft delete oleh user",
		"data":    ownData,
	})
}

// HandleSoftDelete godoc
// @Summary Soft delete pekerjaan
// @Description Menghapus pekerjaan dengan soft delete (set is_delete flag)
// @Tags Pekerjaan
// @Accept json
// @Produce json
// @Param id path string true "Pekerjaan ID"
// @Success 200 {object} map[string]interface{} "success response"
// @Failure 400 {object} map[string]interface{} "ID tidak valid"
// @Failure 404 {object} map[string]interface{} "Data tidak ditemukan"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /pekerjaan/{id} [delete]
// @Security BearerAuth
func (s *PekerjaanService) SoftDelete(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID tidak valid"})
	}

	role := c.Locals("role").(string)
	userID := c.Locals("user_id").(primitive.ObjectID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if role == "admin" {
		err := s.Repo.SoftDelete(ctx, id)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"success": true, "message": "Pekerjaan berhasil dihapus oleh admin"})
	}

	pekerjaan, err := s.Repo.GetByID(ctx, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if pekerjaan == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Pekerjaan tidak ditemukan"})
	}

	alumniRepo := repository.NewAlumniRepository(s.DB)
	alumni, err := alumniRepo.GetByID(ctx, userID)
	if err != nil || alumni == nil {
		return c.Status(403).JSON(fiber.Map{"error": "Data alumni tidak ditemukan"})
	}

	if alumni.ID != pekerjaan.AlumniID {
		return c.Status(403).JSON(fiber.Map{"error": "Tidak boleh hapus pekerjaan orang lain"})
	}

	err = s.Repo.SoftDelete(ctx, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Pekerjaan berhasil dihapus"})
}

// HandleGetAll godoc
// @Summary Get all pekerjaan
// @Description Mengambil daftar semua pekerjaan dengan pagination dan filter
// @Tags Pekerjaan
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param sortBy query string false "Sort field" default(created_at)
// @Param order query string false "Sort order (asc/desc)" default(desc)
// @Param search query string false "Search by nama perusahaan or posisi"
// @Success 200 {object} map[string]interface{} "pekerjaan list with metadata"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /pekerjaan [get]
// @Security BearerAuth
func (s *PekerjaanService) GetAll(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if limit < 1 {
		limit = 10
	}
	sortBy := c.Query("sortBy", "created_at")
	order := strings.ToLower(c.Query("order", "desc"))
	if order != "asc" {
		order = "desc"
	}
	search := c.Query("search", "")

	offset := (page - 1) * limit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	list, err := s.Repo.GetAllWithFilter(ctx, search, sortBy, order, limit, offset)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	total, err := s.Repo.CountWithSearch(ctx, search)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	pages := (int(total) + limit - 1) / limit

	meta := model.MetaInfo{
		Page:   page,
		Limit:  limit,
		Total:  int(total),
		Pages:  pages,
		SortBy: sortBy,
		Order:  order,
		Search: search,
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    list,
		"meta":    meta,
	})
}

// HandleGetByID godoc
// @Summary Get pekerjaan by ID
// @Description Mengambil data pekerjaan berdasarkan ID
// @Tags Pekerjaan
// @Accept json
// @Produce json
// @Param id path string true "Pekerjaan ID"
// @Success 200 {object} map[string]interface{} "pekerjaan data"
// @Failure 400 {object} map[string]interface{} "ID tidak valid"
// @Failure 404 {object} map[string]interface{} "Pekerjaan tidak ditemukan"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /pekerjaan/{id} [get]
// @Security BearerAuth
func (s *PekerjaanService) GetByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID tidak valid"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := s.Repo.GetByID(ctx, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if data == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Pekerjaan tidak ditemukan"})
	}
	return c.JSON(fiber.Map{"success": true, "data": data})
}

// HandleGetByAlumniID godoc
// @Summary Get pekerjaan by Alumni ID
// @Description Mengambil daftar pekerjaan berdasarkan Alumni ID
// @Tags Pekerjaan
// @Accept json
// @Produce json
// @Param alumni_id path string true "Alumni ID"
// @Success 200 {object} map[string]interface{} "pekerjaan list"
// @Failure 400 {object} map[string]interface{} "Alumni ID tidak valid"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /pekerjaan/alumni/{alumni_id} [get]
// @Security BearerAuth
func (s *PekerjaanService) GetByAlumniID(c *fiber.Ctx) error {
	alumniIDStr := c.Params("alumni_id")
	alumniID, err := primitive.ObjectIDFromHex(alumniIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Alumni ID tidak valid"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := s.Repo.GetByAlumniID(ctx, alumniID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": data})
}

// HandleCreate godoc
// @Summary Create new pekerjaan
// @Description Membuat data pekerjaan baru
// @Tags Pekerjaan
// @Accept json
// @Produce json
// @Param body body model.CreatePekerjaanRequest true "Pekerjaan data"
// @Success 201 {object} map[string]interface{} "created pekerjaan"
// @Failure 400 {object} map[string]interface{} "Request tidak valid"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /pekerjaan [post]
// @Security BearerAuth
func (s *PekerjaanService) Create(c *fiber.Ctx) error {
	var req model.CreatePekerjaanRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Request tidak valid"})
	}

	if err := s.validateCreateRequest(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	newData, err := s.Repo.Create(ctx, req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"success": true, "data": newData})
}

// HandleUpdate godoc
// @Summary Update pekerjaan
// @Description Memperbarui data pekerjaan berdasarkan ID
// @Tags Pekerjaan
// @Accept json
// @Produce json
// @Param id path string true "Pekerjaan ID"
// @Param body body model.UpdatePekerjaanRequest true "Pekerjaan data"
// @Success 200 {object} map[string]interface{} "updated pekerjaan"
// @Failure 400 {object} map[string]interface{} "ID tidak valid"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /pekerjaan/{id} [put]
// @Security BearerAuth
func (s *PekerjaanService) Update(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID tidak valid"})
	}

	var req model.UpdatePekerjaanRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Request tidak valid"})
	}

	if err := s.validateUpdateRequest(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updated, err := s.Repo.Update(ctx, id, req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": updated})
}

// HandleDelete godoc
// @Summary Hard delete pekerjaan (actual deletion)
// @Description Menghapus pekerjaan secara permanent dari database
// @Tags Pekerjaan
// @Accept json
// @Produce json
// @Param id path string true "Pekerjaan ID"
// @Success 200 {object} map[string]interface{} "success response"
// @Failure 400 {object} map[string]interface{} "ID tidak valid"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /pekerjaan/{id} [delete]
// @Security BearerAuth
func (s *PekerjaanService) Delete(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID tidak valid"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Repo.Delete(ctx, id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Pekerjaan berhasil dihapus"})
}
