-- name: CreateManga :one
INSERT INTO Manga (title, url, status) VALUES ($1, $2, $3) RETURNING *;

-- name: CreateChapter :one
INSERT INTO Chapter (number, url, mangaId) VALUES ($1, $2, $3) RETURNING *;

-- name: CreatePage :one
INSERT INTO Page (index, url, chapterId) VALUES ($1, $2, $3) RETURNING *;

-- name: CreateTask :one
INSERT INTO Task (id, name) VALUES ($1, $2) RETURNING *;

-- name: UpdateMangaByID :one
UPDATE Manga SET title = $2, status = $3, poster = $4, description = $5, details = $6 WHERE id = $1 RETURNING *;

-- name: CreateChapterIfNotExists :one
INSERT INTO Chapter (number, url, mangaId)
VALUES ($1, $2, $3)
ON CONFLICT (number, mangaId) 
DO UPDATE SET 
    url = EXCLUDED.url
RETURNING *;

-- name: GetMangaByURL :one
SELECT * FROM Manga WHERE url = $1;

-- name: GetChapterByURL :one
SELECT * FROM Chapter WHERE url = $1;

-- name: GetAllTask :many
SELECT * FROM Task ORDER BY createdAt ASC;

-- name: DeleteTaskByID :one 
DELETE FROM Task WHERE id = $1 RETURNING *;
