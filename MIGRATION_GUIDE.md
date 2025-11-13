# Panduan Migrasi dari PostgreSQL ke MongoDB - Modul 8

## Ringkasan Perubahan

Migrasi dari PostgreSQL ke MongoDB melibatkan perubahan fundamental dalam cara data disimpan dan diakses. Berikut adalah panduan lengkap untuk memahami dan mengimplementasikan migrasi ini.

---

## 1. Perbedaan Fundamental: PostgreSQL vs MongoDB

### PostgreSQL (Relational Database)
- **Struktur**: Tabel dengan baris dan kolom yang terstruktur
- **ID**: Integer auto-increment
- **Tipe Data**: Ketat (schema-defined)
- **Query**: SQL dengan JOIN untuk relasi
- **Transaksi**: ACID compliance penuh

### MongoDB (Document Database)
- **Struktur**: Koleksi dokumen JSON/BSON
- **ID**: ObjectID (12-byte unique identifier)
- **Tipe Data**: Fleksibel (schema-less)
- **Query**: BSON queries dengan aggregation pipeline
- **Transaksi**: Multi-document ACID (MongoDB 4.0+)

---

## 2. Instalasi Dependencies

\`\`\`bash
go get go.mongodb.org/mongo-driver/mongo
go get go.mongodb.org/mongo-driver/bson
go get go.mongodb.org/mongo-driver/bson/primitive
\`\`\`

---

## 3. Konfigurasi Environment

Tambahkan ke `.env`:

\`\`\`env
# MongoDB Configuration
MONGODB_URI=mongodb://localhost:27017
DATABASE_NAME=alumni_db
COLLECTION_PEKERJAAN=pekerjaan_alumni

# Tetap gunakan PostgreSQL untuk Alumni & Users (jika belum dimigrasikan)
DB_DSN=user=postgres password=yourpass dbname=alumni_db port=5432 sslmode=disable
\`\`\`

---

## 4. Perubahan Model Layer

### PostgreSQL Model (Lama)
\`\`\`go
type PekerjaanAlumni struct {
    ID    int       `json:"id"`                    // Integer ID
    AlumniID *int  `json:"alumni_id"`             // Foreign key
    // ... fields lainnya
    CreatedAt time.Time `json:"created_at"`
}
\`\`\`

### MongoDB Model (Baru)
\`\`\`go
type PekerjaanAlumni struct {
    ID    primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    AlumniID primitive.ObjectID `bson:"alumni_id" json:"alumni_id"`
    // ... fields lainnya dengan BSON tags
    CreatedAt time.Time `bson:"created_at" json:"created_at"`
}
\`\`\`

**Perubahan Kunci:**
- `int` → `primitive.ObjectID` untuk ID
- Tambah `bson` tags untuk serialisasi MongoDB
- `omitempty` untuk field yang opsional

---

## 5. Perubahan Repository Layer

### PostgreSQL (Lama)
\`\`\`go
func (r *PekerjaanRepository) GetByID(id int) (*model.PekerjaanAlumni, error) {
    var p model.PekerjaanAlumni
    row := r.DB.QueryRow(`SELECT ... FROM pekerjaan_alumni WHERE id=$1`, id)
    err := row.Scan(&p.ID, &p.AlumniID, ...)
    return &p, err
}
\`\`\`

### MongoDB (Baru)
\`\`\`go
func (r *PekerjaanRepository) GetByID(ctx context.Context, id string) (*model.PekerjaanAlumni, error) {
    objID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return nil, err
    }
    
    var pekerjaan model.PekerjaanAlumni
    err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&pekerjaan)
    return &pekerjaan, err
}
\`\`\`

**Perubahan Kunci:**
- Gunakan `context.Context` untuk operasi async
- Konversi string ID ke `ObjectID` dengan `ObjectIDFromHex()`
- Gunakan `bson.M{}` untuk query filters
- Gunakan `Decode()` untuk unmarshal dokumen

---

## 6. Operasi CRUD di MongoDB

### CREATE
\`\`\`go
// PostgreSQL: INSERT dengan RETURNING
result, err := r.DB.QueryRow(`INSERT INTO ... RETURNING id`, ...).Scan(&id)

// MongoDB: InsertOne
result, err := r.collection.InsertOne(ctx, pekerjaan)
insertedID := result.InsertedID.(primitive.ObjectID)
\`\`\`

### READ
\`\`\`go
// PostgreSQL: QueryRow + Scan
row := r.DB.QueryRow(`SELECT ... WHERE id=$1`, id)
err := row.Scan(&p.ID, &p.Name, ...)

// MongoDB: FindOne + Decode
err := r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&pekerjaan)
\`\`\`

### UPDATE
\`\`\`go
// PostgreSQL: Exec
r.DB.Exec(`UPDATE pekerjaan_alumni SET name=$1 WHERE id=$2`, name, id)

// MongoDB: UpdateOne
r.collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": bson.M{"name": name}})
\`\`\`

### DELETE
\`\`\`go
// PostgreSQL: Exec
r.DB.Exec(`DELETE FROM pekerjaan_alumni WHERE id=$1`, id)

// MongoDB: DeleteOne
r.collection.DeleteOne(ctx, bson.M{"_id": objID})
\`\`\`

---

## 7. Query Patterns

### Filter dengan Regex (Search)
\`\`\`go
// PostgreSQL: ILIKE
WHERE nama_perusahaan ILIKE '%search%'

// MongoDB: $regex
bson.M{"nama_perusahaan": bson.M{"$regex": search, "$options": "i"}}
\`\`\`

### Multiple Conditions (OR)
\`\`\`go
// PostgreSQL: OR
WHERE nama_perusahaan ILIKE $1 OR posisi_jabatan ILIKE $1

// MongoDB: $or
bson.M{"$or": []bson.M{
    {"nama_perusahaan": bson.M{"$regex": search, "$options": "i"}},
    {"posisi_jabatan": bson.M{"$regex": search, "$options": "i"}},
}}
\`\`\`

### Sorting & Pagination
\`\`\`go
// PostgreSQL: ORDER BY + LIMIT + OFFSET
ORDER BY created_at DESC LIMIT 10 OFFSET 0

// MongoDB: Sort + Skip + Limit
opts := options.Find().
    SetSort(bson.M{"created_at": -1}).
    SetLimit(10).
    SetSkip(0)
cursor, err := r.collection.Find(ctx, filter, opts)
\`\`\`

---

## 8. Soft Delete Implementation

### PostgreSQL
\`\`\`go
// Tambah kolom is_delete
UPDATE pekerjaan_alumni SET is_delete = TRUE WHERE id = $1

// Query dengan filter
SELECT ... FROM pekerjaan_alumni WHERE is_delete = FALSE
\`\`\`

### MongoDB
\`\`\`go
// Update dokumen
bson.M{"$set": bson.M{"is_delete": true, "deleted_at": time.Now()}}

// Query dengan filter
bson.M{"is_delete": false}
\`\`\`

---

## 9. Migrasi Data (Jika Diperlukan)

Jika Anda ingin memindahkan data dari PostgreSQL ke MongoDB:

\`\`\`go
// 1. Baca dari PostgreSQL
rows, err := db.Query(`SELECT * FROM pekerjaan_alumni`)

// 2. Konversi ke MongoDB format
var pekerjaan []model.PekerjaanAlumni
for rows.Next() {
    var p model.PekerjaanAlumni
    rows.Scan(&p.ID, &p.AlumniID, ...)
    pekerjaan = append(pekerjaan, p)
}

// 3. Insert ke MongoDB
collection.InsertMany(ctx, pekerjaan)
\`\`\`

---

## 10. Koneksi MongoDB di main.go

\`\`\`go
import "go.mongodb.org/mongo-driver/mongo"

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
        log.Fatal(err)
    }
    
    // Ping untuk verifikasi
    err = client.Ping(ctx, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    return client.Database(os.Getenv("DATABASE_NAME"))
}
\`\`\`

---

## 11. Keuntungan MongoDB untuk Modul Ini

✅ **Fleksibilitas**: Mudah menambah field baru tanpa migration
✅ **Scalability**: Horizontal scaling dengan sharding
✅ **Performance**: Query cepat untuk dokumen besar
✅ **Nested Data**: Bisa menyimpan relasi dalam satu dokumen
✅ **Aggregation**: Pipeline aggregation untuk analisis kompleks

---

## 12. Checklist Migrasi

- [ ] Install MongoDB driver
- [ ] Update `.env` dengan MONGODB_URI
- [ ] Buat model dengan BSON tags
- [ ] Implementasi repository dengan MongoDB operations
- [ ] Update service layer untuk context handling
- [ ] Update main.go untuk koneksi MongoDB
- [ ] Test semua endpoint CRUD
- [ ] Verifikasi soft delete & restore functionality
- [ ] Setup MongoDB indexes untuk performance
- [ ] Dokumentasi API endpoints

---

## 13. Troubleshooting

### Error: "ID tidak valid"
**Penyebab**: String ID tidak valid format ObjectID
**Solusi**: Pastikan ID adalah 24-character hex string

### Error: "no documents in result"
**Penyebab**: Dokumen tidak ditemukan
**Solusi**: Cek filter query dan pastikan dokumen ada

### Slow Query
**Penyebab**: Tidak ada index
**Solusi**: Buat index di MongoDB:
\`\`\`go
indexModel := mongo.IndexModel{
    Keys: bson.D{{Key: "alumni_id", Value: 1}},
}
collection.Indexes().CreateOne(ctx, indexModel)
\`\`\`

---

## Kesimpulan

Migrasi ke MongoDB memberikan fleksibilitas dan skalabilitas yang lebih baik. Pastikan untuk:
1. Memahami perbedaan paradigma (relational vs document)
2. Menggunakan context untuk operasi async
3. Menangani ObjectID dengan benar
4. Membuat index untuk query optimization
5. Testing menyeluruh sebelum production
