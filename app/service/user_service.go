package service

import (
	"context"
	"gofiber-mongo/app/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"

	"github.com/gofiber/fiber/v2"
)

type UserService struct {
	Repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		Repo: repo,
	}
}

// HandleSoftDelete godoc
// @Summary Soft delete user
// @Description Menghapus user dengan soft delete (set is_delete flag)
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{} "success response"
// @Failure 400 {object} map[string]interface{} "ID tidak valid"
// @Failure 500 {object} map[string]interface{} "error"
// @Router /users/{id} [delete]
// @Security BearerAuth
func (s *UserService) SoftDelete(c *fiber.Ctx) error {
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
	return c.JSON(fiber.Map{"success": true, "message": "User berhasil dihapus (soft delete)"})
}
