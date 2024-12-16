# Algeo02-23008

## Tugas Besar 2 Algeo

Projek ini dibuat dalam rangka memenuhi tugas besar 2 IF2123 - Aljabar Linear dan Geometri. Dalam tugas ini diimplementasikan suatu website yang mampu melakukan fitur pencarian dan pencocokan gambar dan lagu.

# Panduan

## Backend

### Buka VScode untuk go gin

```sh
code src/backend/go
# look on the readme there for further detail
```

Default .env

#### Contoh

DATABASE_NAME="algeotubes2"
DATABASE_USERNAME="postgres"
DATABASE_PASSWORD="postgres"
DATABASE_SERVER="localhost"

#### requirements

- Go

Jalankan:

```sh
go run .
```

### Buka VScode untuk fastapi python

```sh
code src/backend/server-midi-convert
# look on the readme there for further detail
```

Buat *virtual environment*:

```bash
python -m venv .venv
```

Jalankan *virtual environment*:

Untuk windows (git bash)
```bash
source .venv/Scripts/activate
```

Install dependencies:
```bash
pip install -r requirements.txt
```

Jalankan fastapi server:
```bash
fastapi dev main.py
```

### Buka VScode untuk frontend

```sh
code src/frontend
# look on the readme there for further detail
```

#### Development

Jalankan server dev:

```shellscript
npm run dev
```

#### Deployment

Build app:

```sh
npm run build
```

Jalankan app:

```sh
npm start
```

Pilih host untuk melakukan deploy.

## Developers

| Nama                 | NIM      
| :---:                | :---: 
| Varel Tiara          | 13523008
| Yonatan Edward Njoto | 13523036
| Benedict Presley     | 13523067
