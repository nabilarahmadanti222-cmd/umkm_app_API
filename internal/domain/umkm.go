package domain

type Umkm struct {
    ID        int64  `json:"id"`
    UserID    int64  `json:"user_id"`
    Nama      string `json:"nama"`
    Deskripsi string `json:"deskripsi"`
    Alamat    string `json:"alamat"`
    Foto      string `json:"foto"`
}