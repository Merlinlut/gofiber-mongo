package repository

import (
	"context"
	"gofiber-mongo/app/model"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// IFileRepository defines the interface for file operations
type IFileRepository interface {
	// Photo operations
	CreatePhoto(ctx context.Context, photo *model.Photo) error
	FindPhotoByID(ctx context.Context, id string) (*model.Photo, error)
	FindPhotoByAlumniID(ctx context.Context, alumniID string) (*model.Photo, error)
	DeletePhoto(ctx context.Context, id string) error

	// Certificate operations
	CreateCertificate(ctx context.Context, cert *model.Certificate) error
	FindCertificateByID(ctx context.Context, id string) (*model.Certificate, error)
	FindCertificateByAlumniID(ctx context.Context, alumniID string) (*model.Certificate, error)
	DeleteCertificate(ctx context.Context, id string) error

	// Alumni ownership check
	CheckAlumniOwnership(ctx context.Context, alumniID string, userID primitive.ObjectID) (bool, error)
}

// FileRepository implements IFileRepository
type FileRepository struct {
	photoCollection       *mongo.Collection
	certificateCollection *mongo.Collection
	alumniCollection      *mongo.Collection
}

// NewFileRepository creates a new file repository
func NewFileRepository(db *mongo.Database) IFileRepository {
	return &FileRepository{
		photoCollection:       db.Collection("photos"),
		certificateCollection: db.Collection("certificates"),
		alumniCollection:      db.Collection("alumni"),
	}
}

// CheckAlumniOwnership checks if an alumni belongs to a specific user
func (r *FileRepository) CheckAlumniOwnership(ctx context.Context, alumniID string, userID primitive.ObjectID) (bool, error) {
	objID, err := primitive.ObjectIDFromHex(alumniID)
	if err != nil {
		return false, err
	}

	var alumni struct {
		UserID primitive.ObjectID `bson:"user_id"`
	}

	err = r.alumniCollection.FindOne(ctx, bson.M{
		"_id":       objID,
		"is_delete": false,
	}).Decode(&alumni)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}

	return alumni.UserID == userID, nil
}

// CreatePhoto saves a new photo to database
func (r *FileRepository) CreatePhoto(ctx context.Context, photo *model.Photo) error {
	photo.UploadedAt = time.Now()
	photo.IsDelete = false

	result, err := r.photoCollection.InsertOne(ctx, photo)
	if err != nil {
		return err
	}

	photo.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindPhotoByID retrieves a photo by ID
func (r *FileRepository) FindPhotoByID(ctx context.Context, id string) (*model.Photo, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var photo model.Photo
	err = r.photoCollection.FindOne(ctx, bson.M{"_id": objID, "is_delete": false}).Decode(&photo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &photo, nil
}

// FindPhotoByAlumniID retrieves a photo by alumni ID
func (r *FileRepository) FindPhotoByAlumniID(ctx context.Context, alumniID string) (*model.Photo, error) {
	objID, err := primitive.ObjectIDFromHex(alumniID)
	if err != nil {
		return nil, err
	}

	var photo model.Photo
	err = r.photoCollection.FindOne(ctx, bson.M{"alumni_id": objID, "is_delete": false}).Decode(&photo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &photo, nil
}

// DeletePhoto soft deletes a photo
func (r *FileRepository) DeletePhoto(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.photoCollection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"is_delete": true}})
	return err
}

// CreateCertificate saves a new certificate to database
func (r *FileRepository) CreateCertificate(ctx context.Context, cert *model.Certificate) error {
	cert.UploadedAt = time.Now()
	cert.IsDelete = false

	result, err := r.certificateCollection.InsertOne(ctx, cert)
	if err != nil {
		return err
	}

	cert.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// FindCertificateByID retrieves a certificate by ID
func (r *FileRepository) FindCertificateByID(ctx context.Context, id string) (*model.Certificate, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var cert model.Certificate
	err = r.certificateCollection.FindOne(ctx, bson.M{"_id": objID, "is_delete": false}).Decode(&cert)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &cert, nil
}

// FindCertificateByAlumniID retrieves a certificate by alumni ID
func (r *FileRepository) FindCertificateByAlumniID(ctx context.Context, alumniID string) (*model.Certificate, error) {
	objID, err := primitive.ObjectIDFromHex(alumniID)
	if err != nil {
		return nil, err
	}

	var cert model.Certificate
	err = r.certificateCollection.FindOne(ctx, bson.M{"alumni_id": objID, "is_delete": false}).Decode(&cert)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &cert, nil
}

// DeleteCertificate soft deletes a certificate
func (r *FileRepository) DeleteCertificate(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.certificateCollection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"is_delete": true}})
	return err
}   