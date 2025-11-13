# Setup MongoDB untuk Proyek Go Fiber Alumni

## Quick Start

### 1. Install Dependencies
\`\`\`bash
go get go.mongodb.org/mongo-driver/mongo
go get go.mongodb.org/mongo-driver/bson
go get go.mongodb.org/mongo-driver/bson/primitive
\`\`\`

### 2. Setup Environment Variables
Buat atau update file `.env`:
\`\`\`env
MONGODB_URI=mongodb://localhost:27017
DATABASE_NAME=alumni_db
APP_PORT=3000
JWT_SECRET=your_secret_key_here
\`\`\`

### 3. Start MongoDB

**Option A: Local Installation**
\`\`\`bash
# macOS
brew services start mongodb-community

# Linux
sudo systemctl start mongod

# Windows
# Buka Services dan start MongoDB
\`\`\`

**Option B: Docker**
\`\`\`bash
docker run -d -p 27017:27017 --name mongodb mongo:latest
\`\`\`

**Option C: MongoDB Atlas (Cloud)**
1. Daftar di https://www.mongodb.com/cloud/atlas
2. Buat cluster
3. Dapatkan connection string
4. Set `MONGODB_URI` dengan connection string

### 4. Run Application
\`\`\`bash
go run main.go
\`\`\`

Server akan berjalan di `http://localhost:3000`

---

## Struktur Koleksi MongoDB

### Users Collection
\`\`\`json
{
  "_id": ObjectId("..."),
  "username": "admin",
  "email": "admin@example.com",
  "password_hash": "hashed_password",
  "role": "admin",
  "is_delete": false,
  "created_at": ISODate("2024-01-01T00:00:00Z")
}
\`\`\`

### Alumni Collection
\`\`\`json
{
  "_id": ObjectId("..."),
  "user_id": ObjectId("..."),
  "nim": "123456",
  "nama": "John Doe",
  "jurusan": "Informatika",
  "angkatan": 2020,
  "tahun_lulus": 2024,
  "email": "john@example.com",
  "no_telepon": "08123456789",
  "alamat": "Jl. Merdeka No. 1",
  "is_delete": false,
  "created_at": ISODate("2024-01-01T00:00:00Z"),
  "updated_at": ISODate("2024-01-01T00:00:00Z")
}
\`\`\`

### Pekerjaan Alumni Collection
\`\`\`json
{
  "_id": ObjectId("..."),
  "alumni_id": ObjectId("..."),
  "nama_perusahaan": "PT Maju Jaya",
  "posisi_jabatan": "Software Engineer",
  "bidang_industri": "Technology",
  "lokasi_kerja": "Jakarta",
  "gaji_range": "10-15 juta",
  "tanggal_mulai_kerja": ISODate("2024-01-15T00:00:00Z"),
  "tanggal_selesai_kerja": null,
  "status_pekerjaan": "Aktif",
  "deskripsi_pekerjaan": "Mengembangkan aplikasi web",
  "is_delete": false,
  "created_at": ISODate("2024-01-01T00:00:00Z"),
  "updated_at": ISODate("2024-01-01T00:00:00Z")
}
\`\`\`

---

## API Endpoints

### Authentication
- `POST /api/login` - Login user

### Alumni Management
- `GET /api/alumni` - Get all alumni (with pagination)
- `GET /api/alumni/:id` - Get alumni by ID
- `POST /api/alumni` - Create alumni (admin only)
- `PUT /api/alumni/:id` - Update alumni (admin only)
- `DELETE /api/alumni/:id` - Delete alumni (admin only)
- `GET /api/alumni/tanpa-pekerjaan` - Get alumni without jobs

### Pekerjaan Alumni Management
- `GET /api/pekerjaan` - Get all jobs (with pagination)
- `GET /api/pekerjaan/:id` - Get job by ID
- `GET /api/pekerjaan/alumni/:alumni_id` - Get jobs by alumni (admin only)
- `POST /api/pekerjaan` - Create job (admin only)
- `PUT /api/pekerjaan/:id` - Update job (admin only)
- `DELETE /api/pekerjaan/:id` - Soft delete job

### Trash Management
- `GET /api/trash/pekerjaan` - Get trashed jobs
- `PUT /api/trash/pekerjaan/:id/restore` - Restore job
- `DELETE /api/trash/pekerjaan/:id/permanent` - Hard delete job

---

## Query Parameters

### Pagination & Filtering
\`\`\`
GET /api/pekerjaan?page=1&limit=10&search=PT&sortBy=created_at&order=desc
\`\`\`

Parameters:
- `page` - Page number (default: 1)
- `limit` - Items per page (default: 10)
- `search` - Search keyword
- `sortBy` - Sort field (default: created_at)
- `order` - Sort order: asc or desc (default: desc)

---

## File Structure

\`\`\`
go-fiber/
├── app/
│   ├── model/
│   │   ├── auth.go              # User model
│   │   ├── alumni.go            # Alumni model
│   │   └── pekerjaan.go         # Pekerjaan model
│   ├── repository/
│   │   ├── user_repository_mongo.go
│   │   ├── alumni_repository_mongo.go
│   │   └── pekerjaan_repository_mongo.go
│   └── service/
│       ├── user_service_mongo.go
│       ├── alumni_service_mongo.go
│       └── pekerjaan_service_mongo.go
├── middleware/
│   ├── auth.go                  # Authentication middleware
│   └── admin.go                 # Admin authorization
├── route/
│   └── route.go                 # Route definitions
├── utils/
│   ├── jwt.go                   # JWT utilities
│   └── password.go              # Password hashing
├── main.go                      # Application entry point
├── .env                         # Environment variables
├── MONGODB_MIGRATION_GUIDE.md   # Migration guide
└── SETUP_MONGODB.md             # This file
\`\`\`

---

## Common Issues & Solutions

### Issue: "MONGODB_URI not set"
**Solution:** Add to `.env` or environment variables
\`\`\`env
MONGODB_URI=mongodb://localhost:27017
\`\`\`

### Issue: "Connection refused"
**Solution:** Ensure MongoDB is running
\`\`\`bash
# Check if MongoDB is running
mongosh

# Or restart
brew services restart mongodb-community
\`\`\`

### Issue: "Invalid ObjectID"
**Solution:** Ensure ID is valid 24-character hex string
\`\`\`bash
# Valid: 507f1f77bcf86cd799439011
# Invalid: 123
\`\`\`

### Issue: "Context deadline exceeded"
**Solution:** Increase timeout in context
\`\`\`go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
\`\`\`

---

## Verification

### Check MongoDB Connection
\`\`\`bash
mongosh
> show databases
> use alumni_db
> show collections
\`\`\`

### Test API
\`\`\`bash
# Login
curl -X POST http://localhost:3000/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'

# Get all pekerjaan
curl -X GET "http://localhost:3000/api/pekerjaan" \
  -H "Authorization: Bearer YOUR_TOKEN"
\`\`\`

---

## Next Steps

1. Update database credentials in `.env`
2. Run `go run main.go`
3. Test endpoints with Postman or curl
4. Migrate data from PostgreSQL (see MONGODB_MIGRATION_GUIDE.md)
5. Deploy to production

---

## Support

For issues or questions:
- Check MONGODB_MIGRATION_GUIDE.md for detailed migration steps
- Review MongoDB documentation: https://docs.mongodb.com/
- Check Go MongoDB driver: https://pkg.go.dev/go.mongodb.org/mongo-driver
