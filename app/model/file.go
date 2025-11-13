package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Photo represents a photo file uploaded by alumni
type Photo struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	AlumniID  primitive.ObjectID `bson:"alumni_id" json:"alumni_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	FileName  string             `bson:"file_name" json:"file_name"`
	FilePath  string             `bson:"file_path" json:"file_path"`
	FileSize  int64              `bson:"file_size" json:"file_size"`
	FileType  string             `bson:"file_type" json:"file_type"`
	UploadedAt time.Time         `bson:"uploaded_at" json:"uploaded_at"`
	IsDelete  bool               `bson:"is_delete" json:"is_delete"`
}

// Certificate represents a certificate/diploma file uploaded by alumni
type Certificate struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	AlumniID  primitive.ObjectID `bson:"alumni_id" json:"alumni_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	FileName  string             `bson:"file_name" json:"file_name"`
	FilePath  string             `bson:"file_path" json:"file_path"`
	FileSize  int64              `bson:"file_size" json:"file_size"`
	FileType  string             `bson:"file_type" json:"file_type"`
	UploadedAt time.Time         `bson:"uploaded_at" json:"uploaded_at"`
	IsDelete  bool               `bson:"is_delete" json:"is_delete"`
}

// PhotoResponse is the response model for photo
type PhotoResponse struct {
	ID        string    `json:"id"`
	AlumniID  string    `json:"alumni_id"`
	UserID    string    `json:"user_id"`
	FileName  string    `json:"file_name"`
	FilePath  string    `json:"file_path"`
	FileSize  int64     `json:"file_size"`
	FileType  string    `json:"file_type"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// CertificateResponse is the response model for certificate
type CertificateResponse struct {
	ID        string    `json:"id"`
	AlumniID  string    `json:"alumni_id"`
	UserID    string    `json:"user_id"`
	FileName  string    `json:"file_name"`
	FilePath  string    `json:"file_path"`
	FileSize  int64     `json:"file_size"`
	FileType  string    `json:"file_type"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// UploadPhotoRequest is the request model for uploading photo
type UploadPhotoRequest struct {
	AlumniID string `form:"alumni_id"`
}

// UploadCertificateRequest is the request model for uploading certificate
type UploadCertificateRequest struct {
	AlumniID string `form:"alumni_id"`
}
