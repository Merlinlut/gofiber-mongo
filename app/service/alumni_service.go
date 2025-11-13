package service

import (
	"context"
	"gofiber-mongo/app/model"
	"gofiber-mongo/app/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type AlumniService struct {
	Repo *repository.AlumniRepository
}

func NewAlumniService(repo *repository.AlumniRepository) *AlumniService {
	return &AlumniService{
		Repo: repo,
	}
}

// HandleSoftDelete godoc
// @Summary Soft delete alumni
// @Description Menghapus alumni dengan soft delete
// @Tags Alumni
// @Accept json
// @Produce json
// @Param id path string true "Alumni ID"
// @Success 200 {object} map[string]interface{} "success response"
// @Failure 400 {object} map[string]interface{} "ID tidak valid"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /alumni/{id} [delete]
// @Security BearerAuth
func (s *AlumniService) SoftDelete(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID tidak valid"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = s.Repo.SoftDelete(ctx, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Alumni + riwayat pekerjaan berhasil dihapus (soft delete)"})
}

// HandleGetAll godoc
// @Summary Get all alumni
// @Description Mengambil daftar semua alumni dengan pagination dan filter
// @Tags Alumni
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param sortBy query string false "Sort field" default(created_at)
// @Param order query string false "Sort order (asc/desc)" default(desc)
// @Param search query string false "Search by nama or nim"
// @Success 200 {object} map[string]interface{} "alumni list with metadata"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /alumni [get]
// @Security BearerAuth
func (s *AlumniService) GetAll(c *fiber.Ctx) error {
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

	alumniList, err := s.Repo.GetAllWithFilter(ctx, search, sortBy, order, limit, offset)
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
		"data":    alumniList,
		"meta":    meta,
	})
}

// HandleGetByID godoc
// @Summary Get alumni by ID
// @Description Mengambil data alumni berdasarkan ID
// @Tags Alumni
// @Accept json
// @Produce json
// @Param id path string true "Alumni ID"
// @Success 200 {object} map[string]interface{} "alumni data"
// @Failure 400 {object} map[string]interface{} "ID tidak valid"
// @Failure 404 {object} map[string]interface{} "Alumni tidak ditemukan"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /alumni/{id} [get]
// @Security BearerAuth
func (s *AlumniService) GetByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID tidak valid"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	alumni, err := s.Repo.GetByID(ctx, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if alumni == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Alumni tidak ditemukan"})
	}
	return c.JSON(fiber.Map{"success": true, "data": alumni})
}

// HandleCreate godoc
// @Summary Create new alumni
// @Description Membuat data alumni baru
// @Tags Alumni
// @Accept json
// @Produce json
// @Param body body model.CreateAlumniRequest true "Alumni data"
// @Success 201 {object} map[string]interface{} "created alumni"
// @Failure 400 {object} map[string]interface{} "Request tidak valid"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /alumni [post]
// @Security BearerAuth
func (s *AlumniService) Create(c *fiber.Ctx) error {
	var req model.CreateAlumniRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Request tidak valid"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	newAlumni, err := s.Repo.Create(ctx, req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"success": true, "data": newAlumni})
}

// HandleUpdate godoc
// @Summary Update alumni
// @Description Memperbarui data alumni berdasarkan ID
// @Tags Alumni
// @Accept json
// @Produce json
// @Param id path string true "Alumni ID"
// @Param body body model.UpdateAlumniRequest true "Alumni data"
// @Success 200 {object} map[string]interface{} "updated alumni"
// @Failure 400 {object} map[string]interface{} "ID tidak valid"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /alumni/{id} [put]
// @Security BearerAuth
func (s *AlumniService) Update(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "ID tidak valid"})
	}

	var req model.UpdateAlumniRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Request tidak valid"})
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
// @Summary Hard delete alumni (actual deletion)
// @Description Menghapus alumni secara permanent dari database
// @Tags Alumni
// @Accept json
// @Produce json
// @Param id path string true "Alumni ID"
// @Success 200 {object} map[string]interface{} "success response"
// @Failure 400 {object} map[string]interface{} "ID tidak valid"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /alumni/{id} [delete]
// @Security BearerAuth
func (s *AlumniService) Delete(c *fiber.Ctx) error {
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
	return c.JSON(fiber.Map{"success": true, "message": "Alumni berhasil dihapus"})
}

// HandleGetWithoutPekerjaan godoc
// @Summary Get alumni without pekerjaan
// @Description Mengambil daftar alumni yang belum memiliki data pekerjaan
// @Tags Alumni
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "alumni list"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /alumni/tanpa-pekerjaan [get]
// @Security BearerAuth
func (s *AlumniService) GetWithoutPekerjaan(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := s.Repo.GetWithoutPekerjaan(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"jumlah":  len(data),
		"data":    data,
		"success": true,
	})
}
