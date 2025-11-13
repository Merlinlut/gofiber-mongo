# Panduan Migrasi PostgreSQL ke MongoDB

## Daftar Isi
1. [Persiapan](#persiapan)
2. [Perbedaan Utama](#perbedaan-utama)
3. [Struktur Data](#struktur-data)
4. [Setup MongoDB](#setup-mongodb)
5. [Perubahan Kode](#perubahan-kode)
6. [Migrasi Data](#migrasi-data)
7. [Testing](#testing)
8. [Troubleshooting](#troubleshooting)

---

## Persiapan

### Install MongoDB Driver
\`\`\`bash
go get go.mongodb.org/mongo-driver/mongo
go get go.mongodb.org/mongo-driver/bson
go get go.mongodb.org/mongo-driver/bson/primitive
\`\`\`

### Environment Variables
Tambahkan ke `.env` atau Vercel environment variables:
\`\`\`
MONGODB_URI=mongodb://localhost:27017
DATABASE_NAME=alumni_db
APP_PORT=3000
\`\`\`

---

## Perbedaan Utama

### PostgreSQL vs MongoDB

| Aspek | PostgreSQL | MongoDB |
|-------|-----------|---------|
| **ID Type** | `int` (auto-increment) | `primitive.ObjectID` (BSON) |
| **Query Language** | SQL | BSON/Aggregation Pipeline |
| **Async** | Synchronous | Context-based (async) |
| **Filtering** | WHERE clauses | `bson.M{}` documents |
| **Pagination** | LIMIT/OFFSET | Skip/Limit options |
| **Search** | ILIKE operator | `$regex` with options |
| **Relationships** | Foreign keys | Document references |
| **Transactions** | ACID | Multi-document ACID (v4.0+) |

---

## Struktur Data

### Tabel PostgreSQL → Koleksi MongoDB

#### 1. Users Collection
**PostgreSQL:**
\`\`\`sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE,
    email VARCHAR(255) UNIQUE,
    password_hash VARCHAR(255),
    role VARCHAR(50),
    is_delete BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP
);
\`\`\`

**MongoDB:**
\`\`\`go
type User struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    Username  string             `bson:"username" json:"username"`
    Email     string             `bson:"email" json:"email"`
    Password  string             `bson:"password_hash" json:"-"`
    Role      string             `bson:"role" json:"role"`
    IsDelete  bool               `bson:"is_delete" json:"is_delete"`
    CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
\`\`\`

#### 2. Alumni Collection
**PostgreSQL:**
\`\`\`sql
CREATE TABLE alumni (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    nim VARCHAR(20) UNIQUE,
    nama VARCHAR(255),
    jurusan VARCHAR(100),
    angkatan INT,
    tahun_lulus INT,
    email VARCHAR(255),
    no_telepon VARCHAR(20),
    alamat TEXT,
    is_delete BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
\`\`\`

**MongoDB:**
\`\`\`go
type Alumni struct {
    ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    UserID     primitive.ObjectID `bson:"user_id" json:"user_id"`
    NIM        string             `bson:"nim" json:"nim"`
    Nama       string             `bson:"nama" json:"nama"`
    Jurusan    string             `bson:"jurusan" json:"jurusan"`
    Angkatan   int                `bson:"angkatan" json:"angkatan"`
    TahunLulus int                `bson:"tahun_lulus" json:"tahun_lulus"`
    Email      string             `bson:"email" json:"email"`
    NoTelepon  string             `bson:"no_telepon" json:"no_telepon"`
    Alamat     *string            `bson:"alamat" json:"alamat"`
    IsDelete   bool               `bson:"is_delete" json:"is_delete"`
    CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
    UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}
\`\`\`

#### 3. Pekerjaan Alumni Collection
**PostgreSQL:**
\`\`\`sql
CREATE TABLE pekerjaan_alumni (
    id SERIAL PRIMARY KEY,
    alumni_id INT REFERENCES alumni(id),
    nama_perusahaan VARCHAR(255),
    posisi_jabatan VARCHAR(255),
    bidang_industri VARCHAR(100),
    lokasi_kerja VARCHAR(255),
    gaji_range VARCHAR(50),
    tanggal_mulai_kerja DATE,
    tanggal_selesai_kerja DATE,
    status_pekerjaan VARCHAR(50),
    deskripsi_pekerjaan TEXT,
    is_delete BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
\`\`\`

**MongoDB:**
\`\`\`go
type PekerjaanAlumni struct {
    ID                  primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    AlumniID            primitive.ObjectID `bson:"alumni_id" json:"alumni_id"`
    NamaPerusahaan      string             `bson:"nama_perusahaan" json:"nama_perusahaan"`
    PosisiJabatan       string             `bson:"posisi_jabatan" json:"posisi_jabatan"`
    BidangIndustri      string             `bson:"bidang_industri" json:"bidang_industri"`
    LokasiKerja         string             `bson:"lokasi_kerja" json:"lokasi_kerja"`
    GajiRange           string             `bson:"gaji_range" json:"gaji_range"`
    TanggalMulaiKerja   time.Time          `bson:"tanggal_mulai_kerja" json:"tanggal_mulai_kerja"`
    TanggalSelesaiKerja *time.Time         `bson:"tanggal_selesai_kerja" json:"tanggal_selesai_kerja"`
    StatusPekerjaan     string             `bson:"status_pekerjaan" json:"status_pekerjaan"`
    DeskripsiPekerjaan  string             `bson:"deskripsi_pekerjaan" json:"deskripsi_pekerjaan"`
    IsDelete            bool               `bson:"is_delete" json:"is_delete"`
    CreatedAt           time.Time          `bson:"created_at" json:"created_at"`
    UpdatedAt           time.Time          `bson:"updated_at" json:"updated_at"`
}
\`\`\`

---

## Setup MongoDB

### Opsi 1: Local MongoDB
\`\`\`bash
# macOS dengan Homebrew
brew tap mongodb/brew
brew install mongodb-community
brew services start mongodb-community

# Linux
sudo apt-get install -y mongodb

# Windows
# Download dari https://www.mongodb.com/try/download/community
\`\`\`

### Opsi 2: MongoDB Atlas (Cloud)
1. Buat akun di https://www.mongodb.com/cloud/atlas
2. Buat cluster baru
3. Dapatkan connection string
4. Set `MONGODB_URI` dengan connection string

### Opsi 3: Docker
\`\`\`bash
docker run -d -p 27017:27017 --name mongodb mongo:latest
\`\`\`

---

## Perubahan Kode

### 1. Model Layer
**Sebelum (PostgreSQL):**
\`\`\`go
type User struct {
    ID        int       `json:"id"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}
\`\`\`

**Sesudah (MongoDB):**
\`\`\`go
type User struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    Username  string             `bson:"username" json:"username"`
    Email     string             `bson:"email" json:"email"`
    CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
\`\`\`

### 2. Repository Layer

**Sebelum (PostgreSQL):**
\`\`\`go
func (r *UserRepository) GetByID(id int) (*model.User, error) {
    var user model.User
    row := r.DB.QueryRow(`SELECT id, username, email FROM users WHERE id=$1`, id)
    err := row.Scan(&user.ID, &user.Username, &user.Email)
    return &user, err
}
\`\`\`

**Sesudah (MongoDB):**
\`\`\`go
func (r *UserRepositoryMongo) GetByID(ctx context.Context, id string) (*model.User, error) {
    objID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return nil, err
    }
    var user model.User
    filter := bson.M{"_id": objID}
    err = r.collection.FindOne(ctx, filter).Decode(&user)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, nil
        }
        return nil, err
    }
    return &user, nil
}
\`\`\`

### 3. Service Layer

**Sebelum (PostgreSQL):**
\`\`\`go
func (s *UserService) GetByID(c *fiber.Ctx) error {
    id, _ := strconv.Atoi(c.Params("id"))
    user, err := s.Repo.GetByID(id)
    // ...
}
\`\`\`

**Sesudah (MongoDB):**
\`\`\`go
func (s *UserServiceMongo) GetByID(c *fiber.Ctx) error {
    id := c.Params("id")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    user, err := s.Repo.GetByID(ctx, id)
    // ...
}
\`\`\`

### 4. Main.go

**Sebelum (PostgreSQL):**
\`\`\`go
db, err := sql.Open("postgres", os.Getenv("DB_DSN"))
\`\`\`

**Sesudah (MongoDB):**
\`\`\`go
func connectMongoDB() *mongo.Database {
    mongoURI := os.Getenv("MONGODB_URI")
    clientOptions := options.Client().ApplyURI(mongoURI)
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        log.Fatalf("Koneksi ke MongoDB gagal: %v", err)
    }
    
    err = client.Ping(ctx, nil)
    if err != nil {
        log.Fatalf("Ping ke MongoDB gagal: %v", err)
    }
    
    return client.Database("alumni_db")
}
\`\`\`

---

## Migrasi Data

### Script Migrasi dari PostgreSQL ke MongoDB

Buat file `scripts/migrate_to_mongodb.go`:

\`\`\`go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Connect to PostgreSQL
	pgDB, err := sql.Open("postgres", os.Getenv("DB_DSN"))
	if err != nil {
		log.Fatal("PostgreSQL connection failed:", err)
	}
	defer pgDB.Close()

	// Connect to MongoDB
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		log.Fatal("MongoDB connection failed:", err)
	}
	defer mongoClient.Disconnect(context.Background())

	mongoDB := mongoClient.Database(os.Getenv("DATABASE_NAME"))

	// Migrate Users
	migrateUsers(pgDB, mongoDB)
	
	// Migrate Alumni
	migrateAlumni(pgDB, mongoDB)
	
	// Migrate Pekerjaan
	migratePekerjaan(pgDB, mongoDB)

	fmt.Println("✓ Migrasi data selesai!")
}

func migrateUsers(pgDB *sql.DB, mongoDB *mongo.Database) {
	fmt.Println("Migrasi Users...")
	
	rows, err := pgDB.Query(`SELECT id, username, email, password_hash, role, created_at FROM users`)
	if err != nil {
		log.Fatal("Query users failed:", err)
	}
	defer rows.Close()

	collection := mongoDB.Collection("users")
	var documents []interface{}

	for rows.Next() {
		var id int
		var username, email, password, role string
		var createdAt time.Time

		if err := rows.Scan(&id, &username, &email, &password, &role, &createdAt); err != nil {
			log.Fatal("Scan failed:", err)
		}

		doc := map[string]interface{}{
			"_id":           primitive.NewObjectID(),
			"username":      username,
			"email":         email,
			"password_hash": password,
			"role":          role,
			"is_delete":     false,
			"created_at":    createdAt,
		}
		documents = append(documents, doc)
	}

	if len(documents) > 0 {
		_, err := collection.InsertMany(context.Background(), documents)
		if err != nil {
			log.Fatal("Insert users failed:", err)
		}
		fmt.Printf("✓ %d users berhasil dimigrasikan\n", len(documents))
	}
}

func migrateAlumni(pgDB *sql.DB, mongoDB *mongo.Database) {
	fmt.Println("Migrasi Alumni...")
	
	rows, err := pgDB.Query(`SELECT id, user_id, nim, nama, jurusan, angkatan, tahun_lulus, email, no_telepon, alamat, created_at, updated_at FROM alumni`)
	if err != nil {
		log.Fatal("Query alumni failed:", err)
	}
	defer rows.Close()

	collection := mongoDB.Collection("alumni")
	var documents []interface{}

	for rows.Next() {
		var id, userID, angkatan, tahunLulus int
		var nim, nama, jurusan, email, noTelepon string
		var alamat *string
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&id, &userID, &nim, &nama, &jurusan, &angkatan, &tahunLulus, &email, &noTelepon, &alamat, &createdAt, &updatedAt); err != nil {
			log.Fatal("Scan failed:", err)
		}

		doc := map[string]interface{}{
			"_id":         primitive.NewObjectID(),
			"user_id":     primitive.NewObjectID(), // Perlu mapping dari PostgreSQL user_id
			"nim":         nim,
			"nama":        nama,
			"jurusan":     jurusan,
			"angkatan":    angkatan,
			"tahun_lulus": tahunLulus,
			"email":       email,
			"no_telepon":  noTelepon,
			"alamat":      alamat,
			"is_delete":   false,
			"created_at":  createdAt,
			"updated_at":  updatedAt,
		}
		documents = append(documents, doc)
	}

	if len(documents) > 0 {
		_, err := collection.InsertMany(context.Background(), documents)
		if err != nil {
			log.Fatal("Insert alumni failed:", err)
		}
		fmt.Printf("✓ %d alumni berhasil dimigrasikan\n", len(documents))
	}
}

func migratePekerjaan(pgDB *sql.DB, mongoDB *mongo.Database) {
	fmt.Println("Migrasi Pekerjaan Alumni...")
	
	rows, err := pgDB.Query(`SELECT id, alumni_id, nama_perusahaan, posisi_jabatan, bidang_industri, lokasi_kerja, gaji_range, tanggal_mulai_kerja, tanggal_selesai_kerja, status_pekerjaan, deskripsi_pekerjaan, created_at, updated_at FROM pekerjaan_alumni`)
	if err != nil {
		log.Fatal("Query pekerjaan failed:", err)
	}
	defer rows.Close()

	collection := mongoDB.Collection("pekerjaan_alumni")
	var documents []interface{}

	for rows.Next() {
		var id, alumniID int
		var namaPerusahaan, posisiJabatan, bidangIndustri, lokasiKerja, gajiRange, statusPekerjaan, deskripsi string
		var tanggalMulai time.Time
		var tanggalSelesai *time.Time
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&id, &alumniID, &namaPerusahaan, &posisiJabatan, &bidangIndustri, &lokasiKerja, &gajiRange, &tanggalMulai, &tanggalSelesai, &statusPekerjaan, &deskripsi, &createdAt, &updatedAt); err != nil {
			log.Fatal("Scan failed:", err)
		}

		doc := map[string]interface{}{
			"_id":                   primitive.NewObjectID(),
			"alumni_id":             primitive.NewObjectID(), // Perlu mapping dari PostgreSQL alumni_id
			"nama_perusahaan":       namaPerusahaan,
			"posisi_jabatan":        posisiJabatan,
			"bidang_industri":       bidangIndustri,
			"lokasi_kerja":          lokasiKerja,
			"gaji_range":            gajiRange,
			"tanggal_mulai_kerja":   tanggalMulai,
			"tanggal_selesai_kerja": tanggalSelesai,
			"status_pekerjaan":      statusPekerjaan,
			"deskripsi_pekerjaan":   deskripsi,
			"is_delete":             false,
			"created_at":            createdAt,
			"updated_at":            updatedAt,
		}
		documents = append(documents, doc)
	}

	if len(documents) > 0 {
		_, err := collection.InsertMany(context.Background(), documents)
		if err != nil {
			log.Fatal("Insert pekerjaan failed:", err)
		}
		fmt.Printf("✓ %d pekerjaan berhasil dimigrasikan\n", len(documents))
	}
}
\`\`\`

---

## Query Patterns

### Filtering
**PostgreSQL:**
\`\`\`sql
SELECT * FROM alumni WHERE nama ILIKE '%john%' AND jurusan = 'Informatika'
\`\`\`

**MongoDB:**
\`\`\`go
filter := bson.M{
    "nama": bson.M{"$regex": "john", "$options": "i"},
    "jurusan": "Informatika",
}
cursor, err := collection.Find(ctx, filter)
\`\`\`

### Pagination
**PostgreSQL:**
\`\`\`sql
SELECT * FROM alumni LIMIT 10 OFFSET 20
\`\`\`

**MongoDB:**
\`\`\`go
opts := options.Find().SetSkip(20).SetLimit(10)
cursor, err := collection.Find(ctx, filter, opts)
\`\`\`

### Sorting
**PostgreSQL:**
\`\`\`sql
SELECT * FROM alumni ORDER BY created_at DESC
\`\`\`

**MongoDB:**
\`\`\`go
opts := options.Find().SetSort(bson.M{"created_at": -1})
cursor, err := collection.Find(ctx, filter, opts)
\`\`\`

### Aggregation
**PostgreSQL:**
\`\`\`sql
SELECT COUNT(*) FROM alumni WHERE is_delete = FALSE
\`\`\`

**MongoDB:**
\`\`\`go
count, err := collection.CountDocuments(ctx, bson.M{"is_delete": false})
\`\`\`

---

## Testing

### Test Endpoints

\`\`\`bash
# Login
curl -X POST http://localhost:3000/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'

# Get All Pekerjaan
curl -X GET "http://localhost:3000/api/pekerjaan?page=1&limit=10" \
  -H "Authorization: Bearer YOUR_TOKEN"

# Get Pekerjaan by ID
curl -X GET http://localhost:3000/api/pekerjaan/PEKERJAAN_ID \
  -H "Authorization: Bearer YOUR_TOKEN"

# Get Pekerjaan by Alumni ID
curl -X GET http://localhost:3000/api/pekerjaan/alumni/ALUMNI_ID \
  -H "Authorization: Bearer YOUR_TOKEN"

# Create Pekerjaan
curl -X POST http://localhost:3000/api/pekerjaan \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "alumni_id":"ALUMNI_ID",
    "nama_perusahaan":"PT Maju Jaya",
    "posisi_jabatan":"Software Engineer",
    "bidang_industri":"Technology",
    "lokasi_kerja":"Jakarta",
    "gaji_range":"10-15 juta",
    "tanggal_mulai_kerja":"2024-01-15",
    "status_pekerjaan":"Aktif",
    "deskripsi_pekerjaan":"Mengembangkan aplikasi web"
  }'

# Update Pekerjaan
curl -X PUT http://localhost:3000/api/pekerjaan/PEKERJAAN_ID \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "nama_perusahaan":"PT Maju Jaya Updated",
    "posisi_jabatan":"Senior Software Engineer",
    "bidang_industri":"Technology",
    "lokasi_kerja":"Jakarta",
    "gaji_range":"15-20 juta",
    "tanggal_mulai_kerja":"2024-01-15",
    "status_pekerjaan":"Aktif",
    "deskripsi_pekerjaan":"Mengembangkan aplikasi web dan mobile"
  }'

# Soft Delete Pekerjaan
curl -X DELETE http://localhost:3000/api/pekerjaan/PEKERJAAN_ID \
  -H "Authorization: Bearer YOUR_TOKEN"

# Get Trashed Pekerjaan
curl -X GET http://localhost:3000/api/trash/pekerjaan \
  -H "Authorization: Bearer YOUR_TOKEN"

# Restore Pekerjaan
curl -X PUT http://localhost:3000/api/trash/pekerjaan/PEKERJAAN_ID/restore \
  -H "Authorization: Bearer YOUR_TOKEN"

# Hard Delete Pekerjaan
curl -X DELETE http://localhost:3000/api/trash/pekerjaan/PEKERJAAN_ID/permanent \
  -H "Authorization: Bearer YOUR_TOKEN"
\`\`\`

---

## Troubleshooting

### Error: "connection refused"
**Solusi:** Pastikan MongoDB sudah running
\`\`\`bash
# Check status
mongosh

# Atau restart
brew services restart mongodb-community
\`\`\`

### Error: "invalid ObjectID"
**Solusi:** Pastikan ID format valid (24 hex characters)
\`\`\`go
objID, err := primitive.ObjectIDFromHex(id)
if err != nil {
    return nil, fmt.Errorf("ID tidak valid: %v", err)
}
\`\`\`

### Error: "context deadline exceeded"
**Solusi:** Tingkatkan timeout context
\`\`\`go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
\`\`\`

### Error: "no documents in result"
**Solusi:** Gunakan `mongo.ErrNoDocuments` untuk handle
\`\`\`go
err = collection.FindOne(ctx, filter).Decode(&data)
if err != nil {
    if err == mongo.ErrNoDocuments {
        return nil, nil // Data tidak ditemukan
    }
    return nil, err
}
\`\`\`

### Performance Issues
**Solusi:** Buat indexes
\`\`\`go
indexModel := mongo.IndexModel{
    Keys: bson.D{{Key: "email", Value: 1}},
}
collection.Indexes().CreateOne(ctx, indexModel)
\`\`\`

---

## Checklist Migrasi

- [ ] Install MongoDB driver
- [ ] Setup MongoDB (local/cloud)
- [ ] Update environment variables
- [ ] Update model structs dengan BSON tags
- [ ] Buat MongoDB repositories
- [ ] Buat MongoDB services
- [ ] Update main.go untuk MongoDB connection
- [ ] Update route.go untuk MongoDB repositories
- [ ] Migrasi data dari PostgreSQL
- [ ] Test semua endpoints
- [ ] Verifikasi data di MongoDB
- [ ] Deploy ke production

---

## Referensi

- [MongoDB Go Driver](https://pkg.go.dev/go.mongodb.org/mongo-driver)
- [BSON Documentation](https://pkg.go.dev/go.mongodb.org/mongo-driver/bson)
- [MongoDB Query Language](https://docs.mongodb.com/manual/reference/operator/query/)
- [Go Fiber Documentation](https://docs.gofiber.io/)
