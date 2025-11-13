package repository

import (
	"context"
	"gofiber-mongo/app/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type PekerjaanRepository struct {
	collection *mongo.Collection
}

func NewPekerjaanRepository(db *mongo.Database) *PekerjaanRepository {
	return &PekerjaanRepository{
		collection: db.Collection("pekerjaan_alumni"),
	}
}

func (r *PekerjaanRepository) RestoreByID(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{"is_delete": false},
	})
	return err
}

func (r *PekerjaanRepository) RestoreByIDAndAlumni(ctx context.Context, id primitive.ObjectID, alumniID primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id, "alumni_id": alumniID}, bson.M{
		"$set": bson.M{"is_delete": false},
	})
	return err
}

func (r *PekerjaanRepository) HardDeleteByID(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *PekerjaanRepository) HardDeleteByIDAndAlumni(ctx context.Context, id primitive.ObjectID, alumniID primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id, "alumni_id": alumniID})
	return err
}

func (r *PekerjaanRepository) GetTrashed(ctx context.Context) ([]model.PekerjaanTrashResponse, error) {
	opts := options.Find().SetSort(bson.M{"updated_at": -1})
	cursor, err := r.collection.Find(ctx, bson.M{"is_delete": true}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []model.PekerjaanTrashResponse
	if err = cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *PekerjaanRepository) GetTrashedByAlumni(ctx context.Context, alumniID primitive.ObjectID) ([]model.PekerjaanTrashResponse, error) {
	opts := options.Find().SetSort(bson.M{"updated_at": -1})
	cursor, err := r.collection.Find(ctx, bson.M{"is_delete": true, "alumni_id": alumniID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []model.PekerjaanTrashResponse
	if err = cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *PekerjaanRepository) SoftDelete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{"is_delete": true},
	})
	return err
}

func (r *PekerjaanRepository) GetAllWithFilter(ctx context.Context, search, sortBy, order string, limit, offset int) ([]model.PekerjaanAlumni, error) {
	filter := bson.M{"is_delete": false}
	if search != "" {
		filter = bson.M{
			"is_delete": false,
			"$or": []bson.M{
				{"nama_perusahaan": bson.M{"$regex": search, "$options": "i"}},
				{"posisi_jabatan": bson.M{"$regex": search, "$options": "i"}},
				{"bidang_industri": bson.M{"$regex": search, "$options": "i"}},
				{"lokasi_kerja": bson.M{"$regex": search, "$options": "i"}},
			},
		}
	}

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

	var list []model.PekerjaanAlumni
	if err = cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *PekerjaanRepository) CountWithSearch(ctx context.Context, search string) (int64, error) {
	filter := bson.M{"is_delete": false}
	if search != "" {
		filter = bson.M{
			"is_delete": false,
			"$or": []bson.M{
				{"nama_perusahaan": bson.M{"$regex": search, "$options": "i"}},
				{"posisi_jabatan": bson.M{"$regex": search, "$options": "i"}},
				{"bidang_industri": bson.M{"$regex": search, "$options": "i"}},
				{"lokasi_kerja": bson.M{"$regex": search, "$options": "i"}},
			},
		}
	}
	return r.collection.CountDocuments(ctx, filter)
}

func (r *PekerjaanRepository) GetAll(ctx context.Context) ([]model.PekerjaanAlumni, error) {
	opts := options.Find().SetSort(bson.M{"created_at": -1})
	cursor, err := r.collection.Find(ctx, bson.M{"is_delete": false}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []model.PekerjaanAlumni
	if err = cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *PekerjaanRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*model.PekerjaanAlumni, error) {
	var pekerjaan model.PekerjaanAlumni
	err := r.collection.FindOne(ctx, bson.M{"_id": id, "is_delete": false}).Decode(&pekerjaan)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &pekerjaan, nil
}

func (r *PekerjaanRepository) GetByAlumniID(ctx context.Context, alumniID primitive.ObjectID) ([]model.PekerjaanAlumni, error) {
	opts := options.Find().SetSort(bson.M{"created_at": -1})
	cursor, err := r.collection.Find(ctx, bson.M{"alumni_id": alumniID, "is_delete": false}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []model.PekerjaanAlumni
	if err = cursor.All(ctx, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *PekerjaanRepository) Create(ctx context.Context, req model.CreatePekerjaanRequest) (*model.PekerjaanAlumni, error) {
	alumniID, err := primitive.ObjectIDFromHex(req.AlumniID)
	if err != nil {
		return nil, err
	}

	tanggalMulai, err := time.Parse("2006-01-02", req.TanggalMulaiKerja)
	if err != nil {
		return nil, err
	}

	var tanggalSelesai *time.Time
	if req.TanggalSelesaiKerja != nil && *req.TanggalSelesaiKerja != "" {
		t, err := time.Parse("2006-01-02", *req.TanggalSelesaiKerja)
		if err != nil {
			return nil, err
		}
		tanggalSelesai = &t
	}

	now := time.Now()
	pekerjaan := model.PekerjaanAlumni{
		ID:                  primitive.NewObjectID(),
		AlumniID:            alumniID,
		NamaPerusahaan:      req.NamaPerusahaan,
		PosisiJabatan:       req.PosisiJabatan,
		BidangIndustri:      req.BidangIndustri,
		LokasiKerja:         req.LokasiKerja,
		GajiRange:           req.GajiRange,
		TanggalMulaiKerja:   tanggalMulai,
		TanggalSelesaiKerja: tanggalSelesai,
		StatusPekerjaan:     req.StatusPekerjaan,
		DeskripsiPekerjaan:  req.DeskripsiPekerjaan,
		IsDelete:            false,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	result, err := r.collection.InsertOne(ctx, pekerjaan)
	if err != nil {
		return nil, err
	}
	pekerjaan.ID = result.InsertedID.(primitive.ObjectID)
	return &pekerjaan, nil
}

func (r *PekerjaanRepository) Update(ctx context.Context, id primitive.ObjectID, req model.UpdatePekerjaanRequest) (*model.PekerjaanAlumni, error) {
	tanggalMulai, err := time.Parse("2006-01-02", req.TanggalMulaiKerja)
	if err != nil {
		return nil, err
	}

	var tanggalSelesai *time.Time
	if req.TanggalSelesaiKerja != nil && *req.TanggalSelesaiKerja != "" {
		t, err := time.Parse("2006-01-02", *req.TanggalSelesaiKerja)
		if err != nil {
			return nil, err
		}
		tanggalSelesai = &t
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{
			"nama_perusahaan":       req.NamaPerusahaan,
			"posisi_jabatan":        req.PosisiJabatan,
			"bidang_industri":       req.BidangIndustri,
			"lokasi_kerja":          req.LokasiKerja,
			"gaji_range":            req.GajiRange,
			"tanggal_mulai_kerja":   tanggalMulai,
			"tanggal_selesai_kerja": tanggalSelesai,
			"status_pekerjaan":      req.StatusPekerjaan,
			"deskripsi_pekerjaan":   req.DeskripsiPekerjaan,
			"updated_at":            time.Now(),
		},
	})
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *PekerjaanRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
