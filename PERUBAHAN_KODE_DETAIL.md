# Detail Perubahan Kode: PostgreSQL ke MongoDB

## 1. Model Layer

### User Model
**Perubahan:**
- ID: `int` → `primitive.ObjectID`
- Tambah field `IsDelete` untuk soft delete
- Tambah BSON tags untuk serialisasi

\`\`\`go
// Sebelum
type User struct {
    ID        int       `json:"id"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    Role      string    `json:"role"`
    CreatedAt time.Time `json:"created_at"`
}

// Sesudah
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

### Alumni Model
**Perubahan:**
- ID: `int` → `primitive.ObjectID`
- Tambah `UserID` field untuk relasi dengan User
- Tambah `IsDelete` field
- Tambah BSON tags

\`\`\`go
// Sebelum
type Alumni struct {
    ID         int       `json:"id"`
    NIM        string    `json:"nim"`
    Nama       string    `json:"nama"`
    // ...
}

// Sesudah
type Alumni struct {
    ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    UserID     primitive.ObjectID `bson:"user_id" json:"user_id"`
    NIM        string             `bson:"nim" json:"nim"`
    Nama       string             `bson:"nama" json:"nama"`
    IsDelete   bool               `bson:"is_delete" json:"is_delete"`
    // ...
}
\`\`\`

### PekerjaanAlumni Model
**Perubahan:**
- ID: `int` → `primitive.ObjectID`
- AlumniID: `*int` → `primitive.ObjectID`
- Tambah `IsDelete` field
- Tambah BSON tags

\`\`\`go
// Sebelum
type PekerjaanAlumni struct {
    ID       int    `json:"id"`
    AlumniID *int   `json:"alumni_id"`
    // ...
}

// Sesudah
type PekerjaanAlumni struct {
    ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    AlumniID primitive.ObjectID `bson:"alumni_id" json:"alumni_id"`
    IsDelete bool               `bson:"is_delete" json:"is_delete"`
    // ...
}
\`\`\`

---

## 2. Repository Layer

### User Repository

**Sebelum (PostgreSQL):**
\`\`\`go
type UserRepository struct {
    DB *sql.DB
}

func (r *UserRepository) SoftDelete(userID int) error {
    _, err := r.DB.Exec(`UPDATE users SET is_delete = TRUE WHERE id = $1`, userID)
    return err
}
\`\`\`

**Sesudah (MongoDB):**
\`\`\`go
type UserRepositoryMongo struct {
    collection *mongo.Collection
}

func NewUserRepositoryMongo(db *mongo.Database) *UserRepositoryMongo {
    return &UserRepositoryMongo{
        collection: db.Collection("users"),
    }
}

func (r *UserRepositoryMongo) FindByUsername(ctx context.Context, username string) (*model.User, error) {
    var user model.User
    filter := bson.M{"username": username}
    err := r.collection.FindOne(ctx, filter).Decode(&user)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, nil
        }
        return nil, err
    }
    return &user, nil
}

func (r *UserRepositoryMongo) SoftDelete(ctx context.Context, userID string) error {
    objID, err := primitive.ObjectIDFromHex(userID)
    if err != nil {
        return err
    }
    filter := bson.M{"_id": objID}
    update := bson.M{"$set": bson.M{"is_delete": true}}
    _, err = r.collection.UpdateOne(ctx, filter, update)
    return err
}
\`\`\`

### Alumni Repository

**Sebelum (PostgreSQL):**
\`\`\`go
func (r *AlumniRepository) GetAll() ([]model.Alumni, error) {
    rows, err := r.DB.Query(`
        SELECT id, nim, nama, jurusan, angkatan, tahun_lulus,
               email, no_telepon, alamat, created_at, updated_at 
        FROM alumni ORDER BY created_at DESC
    `)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var list []model.Alumni
    for rows.Next() {
        var a model.Alumni
        if err := rows.Scan(&a.ID, &a.NIM, &a.Nama, ...); err != nil {
            return nil, err
        }
        list = append(list, a)
    }
    return list, nil
}
\`\`\`

**Sesudah (MongoDB):**
\`\`\`go
func (r *AlumniRepositoryMongo) GetAll(ctx context.Context) ([]model.Alumni, error) {
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
\`\`\`

### Pekerjaan Repository

**Sebelum (PostgreSQL):**
\`\`\`go
func (r *PekerjaanRepository) GetByAlumniID(alumniID int) ([]model.PekerjaanAlumni, error) {
    rows, err := r.DB.Query(`
        SELECT id, alumni_id, nama_perusahaan, posisi_jabatan, ...
        FROM pekerjaan_alumni 
        WHERE alumni_id=$1 ORDER BY created_at DESC
    `, alumniID)
    // ... scan rows
}
\`\`\`

**Sesudah (MongoDB):**
\`\`\`go
func (r *PekerjaanRepository) GetByAlumniID(ctx context.Context, alumniID string) ([]model.PekerjaanAlumni, error) {
    objID, err := primitive.ObjectIDFromHex(alumniID)
    if err != nil {
        return nil, fmt.Errorf("Alumni ID tidak valid: %v", err)
    }

    opts := options.Find().SetSort(bson.M{"created_at": -1})
    cursor, err := r.collection.Find(ctx, bson.M{"alumni_id": objID, "is_delete": false}, opts)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    var pekerjaan []model.PekerjaanAlumni
    if err = cursor.All(ctx, &pekerjaan); err != nil {
        return nil, err
    }
    return pekerjaan, nil
}
\`\`\`

---

## 3. Service Layer

### User Service

**Sebelum (PostgreSQL):**
\`\`\`go
type UserService struct {
    Repo *repository.UserRepository
}

func (s *UserService) SoftDelete(c *fiber.Ctx) error {
    id, _ := strconv.Atoi(c.Params("id"))
    err := s.Repo.SoftDelete(id)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }
    return c.JSON(fiber.Map{"success": true, "message": "User berhasil dihapus"})
}
\`\`\`

**Sesudah (MongoDB):**
\`\`\`go
type UserServiceMongo struct {
    Repo *repository.UserRepositoryMongo
}

func (s *UserServiceMongo) SoftDelete(c *fiber.Ctx) error {
    userID := c.Locals("user_id").(string)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    err := s.Repo.SoftDelete(ctx, userID)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }
    return c.JSON(fiber.Map{"success": true, "message": "User berhasil dihapus (soft delete)"})
}
\`\`\`

### Alumni Service

**Sebelum (PostgreSQL):**
\`\`\`go
func (s *AlumniService) GetByID(c *fiber.Ctx) error {
    id, err := strconv.Atoi(c.Params("id"))
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "ID tidak valid"})
    }
    alumni, err := s.Repo.GetByID(id)
    if err != nil {
        if err == sql.ErrNoRows {
            return c.Status(404).JSON(fiber.Map{"error": "Alumni tidak ditemukan"})
        }
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }
    return c.JSON(fiber.Map{"success": true, "data": alumni})
}
\`\`\`

**Sesudah (MongoDB):**
\`\`\`go
func (s *AlumniServiceMongo) GetByID(c *fiber.Ctx) error {
    id := c.Params("id")
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
\`\`\`

---

## 4. Main Application

**Sebelum (PostgreSQL):**
\`\`\`go
import (
    "database/sql"
    _ "github.com/lib/pq"
)

func main() {
    db, err := sql.Open("postgres", os.Getenv("DB_DSN"))
    if err != nil {
        log.Fatal("Gagal koneksi DB:", err)
    }
    defer db.Close()

    // ... setup routes dengan db *sql.DB
    route.RegisterRoutes(app, db)
}
\`\`\`

**Sesudah (MongoDB):**
\`\`\`go
import (
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func connectMongoDB() *mongo.Database {
    mongoURI := os.Getenv("MONGODB_URI")
    if mongoURI == "" {
        mongoURI = "mongodb://localhost:27017"
    }

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

    return client.Database(os.Getenv("DATABASE_NAME"))
}

func main() {
    db := connectMongoDB()
    
    // ... setup routes dengan db *mongo.Database
    route.RegisterRoutes(app, db)
}
\`\`\`

---

## 5. Route Configuration

**Sebelum (PostgreSQL):**
\`\`\`go
func RegisterRoutes(app *fiber.App, db *sql.DB) {
    alumniRepo := &repository.AlumniRepository{DB: db}
    alumniService := &service.AlumniService{Repo: alumniRepo}
    
    pekerjaanRepo := &repository.PekerjaanRepository{DB: db}
    pekerjaanService := &service.PekerjaanService{Repo: pekerjaanRepo}
    
    // ... setup routes
}
\`\`\`

**Sesudah (MongoDB):**
\`\`\`go
func RegisterRoutes(app *fiber.App, db *mongo.Database) {
    alumniRepo := repository.NewAlumniRepositoryMongo(db)
    alumniService := service.NewAlumniServiceMongo(alumniRepo)
    
    pekerjaanRepo := repository.NewPekerjaanRepository(db)
    pekerjaanService := service.NewPekerjaanService(pekerjaanRepo, db)
    
    // ... setup routes (sama seperti sebelumnya)
}
\`\`\`

---

## 6. Query Pattern Comparison

### SELECT Query

**PostgreSQL:**
\`\`\`sql
SELECT id, nama, email FROM alumni WHERE is_delete = FALSE ORDER BY created_at DESC LIMIT 10 OFFSET 0
\`\`\`

**MongoDB:**
\`\`\`go
filter := bson.M{"is_delete": false}
opts := options.Find().
    SetSort(bson.M{"created_at": -1}).
    SetLimit(10).
    SetSkip(0)
cursor, err := collection.Find(ctx, filter, opts)
\`\`\`

### INSERT Query

**PostgreSQL:**
\`\`\`sql
INSERT INTO alumni (nim, nama, email, created_at, updated_at) 
VALUES ('123456', 'John', 'john@example.com', NOW(), NOW()) 
RETURNING id
\`\`\`

**MongoDB:**
\`\`\`go
alumni := model.Alumni{
    ID:        primitive.NewObjectID(),
    NIM:       "123456",
    Nama:      "John",
    Email:     "john@example.com",
    CreatedAt: time.Now(),
    UpdatedAt: time.Now(),
}
result, err := collection.InsertOne(ctx, alumni)
\`\`\`

### UPDATE Query

**PostgreSQL:**
\`\`\`sql
UPDATE alumni SET nama = 'Jane', updated_at = NOW() WHERE id = 1
\`\`\`

**MongoDB:**
\`\`\`go
filter := bson.M{"_id": objID}
update := bson.M{
    "$set": bson.M{
        "nama":       "Jane",
        "updated_at": time.Now(),
    },
}
_, err := collection.UpdateOne(ctx, filter, update)
\`\`\`

### DELETE Query

**PostgreSQL:**
\`\`\`sql
UPDATE alumni SET is_delete = TRUE WHERE id = 1  -- Soft delete
DELETE FROM alumni WHERE id = 1                   -- Hard delete
\`\`\`

**MongoDB:**
\`\`\`go
// Soft delete
filter := bson.M{"_id": objID}
update := bson.M{"$set": bson.M{"is_delete": true}}
_, err := collection.UpdateOne(ctx, filter, update)

// Hard delete
_, err := collection.DeleteOne(ctx, filter)
\`\`\`

### SEARCH Query

**PostgreSQL:**
\`\`\`sql
SELECT * FROM alumni WHERE nama ILIKE '%john%' OR email ILIKE '%john%'
\`\`\`

**MongoDB:**
\`\`\`go
filter := bson.M{
    "$or": []bson.M{
        {"nama": bson.M{"$regex": "john", "$options": "i"}},
        {"email": bson.M{"$regex": "john", "$options": "i"}},
    },
}
cursor, err := collection.Find(ctx, filter)
\`\`\`

---

## 7. Error Handling

### PostgreSQL Error Handling

\`\`\`go
if err == sql.ErrNoRows {
    return c.Status(404).JSON(fiber.Map{"error": "Data tidak ditemukan"})
}
\`\`\`

### MongoDB Error Handling

\`\`\`go
if err == mongo.ErrNoDocuments {
    return nil, nil  // atau return error sesuai kebutuhan
}
if err != nil {
    return nil, err
}
\`\`\`

---

## 8. Context Usage

### PostgreSQL (Synchronous)
\`\`\`go
func (r *UserRepository) GetByID(id int) (*model.User, error) {
    // Langsung query tanpa context
    row := r.DB.QueryRow(`SELECT ... WHERE id = $1`, id)
}
\`\`\`

### MongoDB (Async with Context)
\`\`\`go
func (r *UserRepositoryMongo) GetByID(ctx context.Context, id string) (*model.User, error) {
    // Menggunakan context untuk timeout
    err := r.collection.FindOne(ctx, filter).Decode(&user)
}
\`\`\`

---

## Summary of Changes

| Aspek | PostgreSQL | MongoDB |
|-------|-----------|---------|
| **Connection** | `sql.Open()` | `mongo.Connect()` |
| **ID Type** | `int` | `primitive.ObjectID` |
| **Query** | SQL strings | BSON documents |
| **Async** | Synchronous | Context-based |
| **Error** | `sql.ErrNoRows` | `mongo.ErrNoDocuments` |
| **Pagination** | LIMIT/OFFSET | Skip/Limit |
| **Search** | ILIKE | $regex |
| **Soft Delete** | UPDATE is_delete | $set is_delete |
| **Serialization** | - | BSON tags |

---

**Semua perubahan telah diimplementasikan dan siap digunakan!**
