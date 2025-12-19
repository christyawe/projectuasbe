# Project UAS Backend – Achievement Management System

Sistem backend yang dibangun untuk memenuhi kebutuhan tugas UAS, yaitu **Sistem Manajemen Prestasi Mahasiswa** berbasis web.  
Project ini menggunakan **Golang**, **Fiber**, **PostgreSQL**, dan **MongoDB** serta mengikuti standar **Clean Architecture + Repository Pattern**, sesuai dengan SRS yang diberikan.

Sistem ini mendukung proses:
- Pelaporan prestasi oleh mahasiswa
- Verifikasi prestasi oleh dosen wali
- Monitoring oleh admin
- Penyimpanan data prestasi secara fleksibel menggunakan MongoDB
- Penyimpanan data referensi terstruktur menggunakan PostgreSQL

---

## ✨ **Daftar Isi**
1. [Tech Stack](#tech-stack)
2. [Fitur Utama](#fitur-utama)
3. [Arsitektur Project](#arsitektur-project)
4. [Struktur Folder](#struktur-folder)
5. [Instalasi](#instalasi)
6. [Konfigurasi ENV](#konfigurasi-env)
7. [Setup Database](#setup-database)
8. [Menjalankan Server](#menjalankan-server)
9. [Autentikasi](#autentikasi)
10. [Daftar API Endpoint](#daftar-api-endpoint)
11. [Workflow Status Prestasi](#workflow-status-prestasi)
12. [Testing](#testing)
13. [Lisensi](#lisensi)
14. [Author](#author)

---

## 🚀 **Tech Stack**

- **Go 1.22+**
- **Fiber v2**
- **PostgreSQL** (relational DB — users, lecturers, students, references)
- **MongoDB** (document DB — achievements)
- **JWT Authentication**
- **BCrypt Password Hashing**
- **Clean Architecture**
- **Repository Pattern**

---

## 🧩 **Fitur Utama**

### 👥 User Management
- Register & Login
- JWT Authentication
- 3 Role utama sesuai SRS:
  - **Admin** — akses penuh
  - **Mahasiswa** — unggah & kelola prestasi
  - **Dosen Wali** — verifikasi prestasi mahasiswa bimbingan

### 🧑‍🎓 Student Module
- Lihat detail mahasiswa
- Lihat mahasiswa berdasarkan dosen wali

### 🧑‍🏫 Lecturer Module
- Lihat data dosen
- Melihat daftar mahasiswa bimbingan

### 🏆 Achievement Module
Disimpan di **MongoDB**, mendukung:
- Dynamic field *details*
- Lampiran (attachments)
- Tagging
- Penilaian (points)
- Soft delete

### 🔗 Achievement Reference Module
Disimpan di **PostgreSQL**, berfungsi sebagai penghubung:
- mahasiswa → achievement MongoDB
- Menyimpan status verifikasi:
  - draft → submitted → verified/rejected → deleted

---

## 🏗️ **Arsitektur Project**

Sistem ini mengikuti prinsip **Clean Architecture**:

