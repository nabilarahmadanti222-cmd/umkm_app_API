package handler

import (
    "errors"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"
)

// errInvalidFileType dikembalikan saat ekstensi berkas yang diunggah bukan
// salah satu dari allowedImageExt.
var errInvalidFileType = errors.New("tipe berkas tidak didukung, hanya jpg/jpeg/png")

var allowedImageExt = map[string]bool{
    ".jpg":  true,
    ".jpeg": true,
    ".png":  true,
}

// saveUploadedFile menyimpan berkas dari form-field fieldName ke <uploadDir>/<subDir>/.
// Mengembalikan path RELATIF (contoh: "umkm/169...jpg"), bukan path disk absolut,
// agar nilai yang sama bisa langsung dipakai klien sebagai "<base_url>/uploads/<foto>".
// Mengembalikan string kosong (bukan error) apabila pengguna tidak mengunggah foto,
// karena foto bersifat opsional. Hanya menerima berkas berekstensi .jpg/.jpeg/.png.
func saveUploadedFile(r *http.Request, fieldName, uploadDir, subDir string) (string, error) {
    file, header, err := r.FormFile(fieldName)
    if err == http.ErrMissingFile {
        return "", nil
    }
    if err != nil {
        return "", err
    }
    defer file.Close()

    ext := strings.ToLower(filepath.Ext(header.Filename))
    if !allowedImageExt[ext] {
        return "", errInvalidFileType
    }

    targetDir := filepath.Join(uploadDir, subDir)
    if err := os.MkdirAll(targetDir, 0755); err != nil {
        return "", err
    }

    filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

    dst, err := os.Create(filepath.Join(targetDir, filename))
    if err != nil {
        return "", err
    }
    defer dst.Close()

    if _, err := io.Copy(dst, file); err != nil {
        return "", err
    }

    return subDir + "/" + filename, nil
}

// deleteFile menghapus berkas lama di <uploadDir>/<relPath>. Dipanggil setelah
// update/delete berhasil di database, agar tidak ada berkas foto yang menumpuk.
func deleteFile(uploadDir, relPath string) {
    if relPath == "" {
        return
    }
    _ = os.Remove(filepath.Join(uploadDir, relPath))
}