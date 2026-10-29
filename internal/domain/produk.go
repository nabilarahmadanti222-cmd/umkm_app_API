package domain

type Produk struct {
    ID        int64   `json:"id"`
    UmkmID    int64   `json:"umkm_id"`
    Nama      string  `json:"nama"`
    Harga     float64 `json:"harga"`
    Deskripsi string  `json:"deskripsi"`
    Foto      string  `json:"foto"`
}