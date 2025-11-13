package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gofiber-mongo/app/model"
	"time"
)

type AlumniRepository struct {
	collection *mongo.Collection
	userColl   *mongo.Collection
}

func (r *AlumniRepository) FindByID(ctx context.Context, alumniID string) (any, any) {
	panic("unimplemented")
}

func NewAlumniRepository(db *mongo.Database) *AlumniRepository {
	return &AlumniRepository{
		collection: db.Collection("alumni"),
		userColl:   db.Collection("users"),
	}
}

func (r *AlumniRepository) SoftDelete(ctx context.Context, alumniID primitive.ObjectID) error {
	// Soft delete alumni
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": alumniID}, bson.M{
		"$set": bson.M{"is_delete": true},
	})
	if err != nil {
		return err
	}

	// Also soft delete all related pekerjaan
	pekerjaanColl := r.collection.Database().Collection("pekerjaan_alumni")
	_, err = pekerjaanColl.UpdateMany(ctx, bson.M{"alumni_id": alumniID}, bson.M{
		"$set": bson.M{"is_delete": true},
	})
	return err
}

func (r *AlumniRepository) GetAll(ctx context.Context) ([]model.Alumni, error) {
	opts := options.Find().SetSort(bson.M{"created_at": -1})
	cursor, err := r.collection.Find(ctx, bson.M{"is_delete": false}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []model.Alumni
	if err = cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *AlumniRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*model.Alumni, error) {
	var alumni model.Alumni
	err := r.collection.FindOne(ctx, bson.M{"_id": id, "is_delete": false}).Decode(&alumni)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &alumni, nil
}

func (r *AlumniRepository) Create(ctx context.Context, req model.CreateAlumniRequest) (*model.Alumni, error) {
	alumni := model.Alumni{
		ID:         primitive.NewObjectID(),
		NIM:        req.NIM,
		Nama:       req.Nama,
		Jurusan:    req.Jurusan,
		Angkatan:   req.Angkatan,
		TahunLulus: req.TahunLulus,
		Email:      req.Email,
		NoTelepon:  req.NoTelepon,
		Alamat:     &req.Alamat,
		IsDelete:   false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	result, err := r.collection.InsertOne(ctx, alumni)
	if err != nil {
		return nil, err
	}
	alumni.ID = result.InsertedID.(primitive.ObjectID)
	return &alumni, nil
}

func (r *AlumniRepository) Update(ctx context.Context, id primitive.ObjectID, req model.UpdateAlumniRequest) (*model.Alumni, error) {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{
			"nama":        req.Nama,
			"jurusan":     req.Jurusan,
			"angkatan":    req.Angkatan,
			"tahun_lulus": req.TahunLulus,
			"email":       req.Email,
			"no_telepon":  req.NoTelepon,
			"alamat":      req.Alamat,
		},
	})
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *AlumniRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *AlumniRepository) GetAllWithFilter(ctx context.Context, search, sortBy, order string, limit, offset int) ([]model.Alumni, error) {
	// Build search filter
	filter := bson.M{"is_delete": false}
	if search != "" {
		filter = bson.M{
			"is_delete": false,
			"$or": []bson.M{
				{"nim": bson.M{"$regex": search, "$options": "i"}},
				{"nama": bson.M{"$regex": search, "$options": "i"}},
				{"jurusan": bson.M{"$regex": search, "$options": "i"}},
				{"email": bson.M{"$regex": search, "$options": "i"}},
			},
		}
	}

	// Build sort
	sortOrder := int32(-1)
	if order == "asc" {
		sortOrder = 1
	}
	opts := options.Find().
		SetSort(bson.M{sortBy: sortOrder}).
		SetSkip(int64(offset)).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []model.Alumni
	if err = cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *AlumniRepository) CountWithSearch(ctx context.Context, search string) (int64, error) {
	filter := bson.M{"is_delete": false}
	if search != "" {
		filter = bson.M{
			"is_delete": false,
			"$or": []bson.M{
				{"nim": bson.M{"$regex": search, "$options": "i"}},
				{"nama": bson.M{"$regex": search, "$options": "i"}},
				{"jurusan": bson.M{"$regex": search, "$options": "i"}},
				{"email": bson.M{"$regex": search, "$options": "i"}},
			},
		}
	}
	return r.collection.CountDocuments(ctx, filter)
}

func (r *AlumniRepository) GetWithoutPekerjaan(ctx context.Context) ([]model.Alumni, error) {
	pekerjaanColl := r.collection.Database().Collection("pekerjaan_alumni")

	// Get all alumni IDs that have pekerjaan
	cursor, err := pekerjaanColl.Find(ctx, bson.M{"is_delete": false})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var pekerjaanList []bson.M
	if err = cursor.All(ctx, &pekerjaanList); err != nil {
		return nil, err
	}

	// Extract alumni IDs
	alumniWithPekerjaan := make(map[primitive.ObjectID]bool)
	for _, p := range pekerjaanList {
		if alumniID, ok := p["alumni_id"].(primitive.ObjectID); ok {
			alumniWithPekerjaan[alumniID] = true
		}
	}

	// Get all alumni without pekerjaan
	opts := options.Find().SetSort(bson.M{"created_at": -1})
	cursor, err = r.collection.Find(ctx, bson.M{"is_delete": false}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []model.Alumni
	if err = cursor.All(ctx, &list); err != nil {
		return nil, err
	}

	// Filter out alumni that have pekerjaan
	var result []model.Alumni
	for _, alumni := range list {
		if !alumniWithPekerjaan[alumni.ID] {
			result = append(result, alumni)
		}
	}
	return result, nil
}
