# Tucil 3 IF2211 Strategi Algoritma — Voxelisasi 3D dengan Divide and Conquer

## Identitas Pembuat

| NIM       | Nama                    |
|-----------|---------------------|
| 13524028 | Muhammad Nur Majiid |
| 13524106 | Michael James Liman |

## Penjelasan Singkat Program

Program ini mengimplementasikan **Voxelisasi 3D** menggunakan pendekatan algoritma **Divide and Conquer** dengan struktur data **Octree**. Voxelisasi adalah proses mengubah representasi model 3D berbasis permukaan (mesh segitiga dari file OBJ) menjadi representasi berbasis volume (kumpulan kubus-kubus kecil / voxel). Program ini juga menyajikan fitur OBJ Viewer untuk melakukan visualisasi dari sebuah file OBJ.

Alur kerja program utama (Voxelization):
1. Membaca file model 3D berformat `.obj` untuk mengekstrak vertex dan segitiga
2. Menghitung bounding box yang melingkupi seluruh model
3. Membangun pohon Octree secara rekursif dengan **concurrency** (goroutine pada level dangkal) untuk membagi ruang 3D dan mendeteksi area yang terisi model
4. Mengumpulkan titik-titik pusat voxel dari daun-daun Octree
5. Mengekspor hasil voxelisasi ke file `.obj` baru

## Dependency

Program ini ditulis dalam bahasa **Go (Golang)** dan hanya menggunakan **standard library**, tanpa dependency eksternal.

| Dependency         | Versi Minimum | Keterangan                    |
|--------------------|---------------|-------------------------------|
| Go (Golang)        | 1.24          | Bahasa pemrograman utama      |

Package standard library yang digunakan:
- `bufio` — Buffered I/O untuk membaca dan menulis file
- `fmt` — Formatted I/O (parsing dan output)
- `math` — Fungsi matematika (Min, Max, Pow)
- `os` — Operasi file (Open, Create)
- `strconv` — Konversi string ke integer
- `strings` — Manipulasi string (TrimSpace, Fields, Index)
- `sync` — Sinkronisasi concurrency (WaitGroup)
- `ebiten` — 2D Game Engine untuk Go

## Cara Instalasi dan Menjalankan Program

### Prasyarat
Pastikan **Go** sudah terinstal di sistem Anda:
```bash
go version
```
Jika belum terinstal, unduh dari [https://go.dev/dl/](https://go.dev/dl/)

### Langkah Instalasi
1. Clone repository ini:
```bash
git clone https://github.com/MichaelJamesL/Tucil2_13524106_13524028.git
cd Tucil2_13524106_13524028
```

2. (Opsional) Verifikasi modul:
```bash
cat go.mod
```
sesuaikan dengan versi go yang anda gunakan

### Menjalankan Program
```bash
go run .
```
Lalu, ikuti petunjuk program
Untuk menu voxelization (2), program akan membaca file input dan menghasilkan file output berupa `voxelized.obj` di direktori utama. Untuk menu OBJ Viewer (1), program akan menampilkan window GUI baru sebagai visualisasi dari file OBJ yang dimasukkan.

### Build (Opsional)
```bash
go build .
/bin/tucil2-stima.exe (windows)
```

## Directory Tree

```
C:.
│   .gitignore
│   go.mod                                      # Go module definition
│   go.sum                                      # Go module's checksum
│   main.go                                     # Program utama (entry point)
│   README.md
│
|───bin
│       tucil2-stima.exe                        # Executable program
|───docs
│       IF2211 Laporan Tugas Kecil 2.pdf        # Laporan
├───src                                         # Source code utama program (package src)
│       objviewer.go
│       stats.go
│       voxelization.go
│
└───test                                        # Input dan output testing
    ├───in
    │       cow.obj
    │       line.obj
    │       pumpkin.obj
    │       teapot.obj
    │
    └───out
            cow-v.obj
            line-v.obj
            pumpkin.obj
            teapot-v.obj
```
