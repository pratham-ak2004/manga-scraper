-- Active: 1769191942250@@10.171.67.191@5000@manga_db
-- name: CreateManga :one
INSERT INTO Manga (title, url, status) VALUES ($1, $2, $3) RETURNING *;

-- name: CreateChapter :one
INSERT INTO Chapter (number, url, mangaId) VALUES ($1, $2, $3) RETURNING *;

-- name: CreatePage :one
INSERT INTO Page (index, url, filePath, chapterId, altText) VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: CreateTask :one
INSERT INTO Task (id, name, type, payload) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: CreateArchiveWithRange :one
INSERT INTO Archive (mangaId, filePath, startChapter, endChapter, complete) VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: UpdateMangaByID :one
UPDATE Manga SET title = $2, status = $3, poster = $4, description = $5, details = $6 WHERE id = $1 RETURNING *;

-- name: UpdateChapterByID :one
UPDATE Chapter SET url = $2 WHERE id = $1 RETURNING *;

-- name: UpdatePageDownloadedAt :one
UPDATE Page SET downloadedAt = $3 WHERE index = $1 AND chapterId = $2 RETURNING *;

-- name: UpdatePage :one
UPDATE Page SET url = $3, altText = $4, filePath = $5 WHERE index = $1 AND chapterId = $2 RETURNING *;

-- name: UpdateTaskByID :one
UPDATE Task SET data = $2, status = $3 WHERE id = $1 RETURNING *;

-- name: UpdateTaskStatusByID :one
UPDATE Task SET status = $2 WHERE id = $1 RETURNING *;

-- name: UpdateArchiveStatusAndSizeByID :one
UPDATE Archive SET status = $2, size = $3 WHERE id = $1 RETURNING *;

-- name: CreateChapterIfNotExists :one
INSERT INTO Chapter (number, url, mangaId)
VALUES ($1, $2, $3)
ON CONFLICT (number, mangaId) 
DO UPDATE SET 
    url = EXCLUDED.url
RETURNING *;

-- name: GetMangaByURL :one
SELECT * FROM Manga WHERE url = $1;

-- name: GetMangaByID :one
SELECT * FROM Manga WHERE id = $1;

-- name: GetMangaList :many
SELECT * FROM Manga;

-- name: GetChapterByURL :one
SELECT * FROM Chapter WHERE url = $1;

-- name: GetChapterByID :one
SELECT 
    c.*,
    m.title AS title
FROM Chapter c
INNER JOIN Manga m ON c.mangaId = m.id
WHERE c.id = $1;

-- name: GetChapterToReadByID :one
SELECT 
  c.*, 
  COALESCE((SELECT p.id FROM Chapter p WHERE p.number = c.number - 1 AND p.mangaId = c.mangaId), NULL) AS previous_chapter, 
  COALESCE((SELECT n.id FROM Chapter n WHERE n.number = c.number + 1 AND n.mangaId = c.mangaId), NULL) AS next_chapter,
  (SELECT m.title AS title FROM Manga m WHERE m.id = c.mangaId)
FROM Chapter c WHERE c.id = $1;

-- name: GetChapterByIndexAndManga :one
SELECT * FROM Chapter WHERE number = $1 AND mangaId = $2;

-- name: GetChaptersByMangaID :many
SELECT c.*, 
  ( SELECT COUNT(*) FROM Page WHERE chapterId = c.id ) AS page_count 
FROM Chapter c WHERE mangaId = $1 
ORDER BY number ASC;

-- name: GetChaptersByMangaIDWithPageCount :many
SELECT * FROM Chapter c LEFT JOIN (
    SELECT chapterId, COUNT(*) AS pageCount
    FROM Page
    GROUP BY chapterId
) p ON c.id = p.chapterId
WHERE c.mangaId = $1
ORDER BY c.number ASC;

-- name: GetPageByIndexAndChapterID :one
SELECT * FROM Page where index = $1 AND chapterId = $2;

-- name: GetPagesByChapterID :many
SELECT * FROM Page WHERE chapterId = $1 ORDER BY index ASC;

-- name: GetAllPendingTask :many
SELECT * FROM Task WHERE status != 'SUCCESS' OR status != 'COMMITTED' ORDER BY createdAt ASC;

-- name: GetAllCompletedTask :many
SELECT * FROM Task WHERE status = 'SUCCESS' ORDER BY updatedAt DESC;

-- name: GetAllTasks :many
SELECT * FROM Task 
ORDER BY 
  CASE status
    WHEN 'STARTED' THEN 1
    WHEN 'FAILURE' THEN 2
    WHEN 'RETRY' THEN 3
    WHEN 'SUCCESS' THEN 4
    WHEN 'PENDING' THEN 5
    WHEN 'COMMITTED' THEN 4
  END ASC,
  CASE type
    WHEN 'REQUEST' THEN 1
    WHEN 'PIPELINE' THEN 2
  END ASC,
  name DESC, 
  updatedAt DESC;

-- name: GetTaskByID :one
SELECT * FROM Task WHERE id = $1;

-- name: GetArchiveByMangaID :one
SELECT * FROM Archive WHERE mangaId = $1 AND startChapter = $2 AND endChapter = $3;

-- name: GetArchiveByMangaIDIfComplete :one
SELECT * FROM Archive WHERE mangaId = $1 AND complete = TRUE;

-- name: GetArchivesByMangaID :many
SELECT * FROM Archive WHERE mangaId = $1 ORDER BY createdAt DESC;

-- name: GetPagesByRangeAndMangaID :many
SELECT 
    p.index,
    p.url,
    p.altText,
    p.filePath,
    p.downloadedAt,
    p.chapterId,
    c.number AS chapterNumber,
    c.url AS chapterUrl
FROM Page p
INNER JOIN Chapter c ON p.chapterId = c.id
WHERE c.mangaId = $1 AND c.number >= $2 AND c.number <= $3
ORDER BY c.number ASC, p.index ASC;

-- name: ListMangaDashboard :many
SELECT
    m.id                          AS manga_id,
    m.title                       AS title,
    m.url                         AS manga_url,
    m.status                      AS status,
    m.poster                      AS poster,

    m.createdAt                   AS manga_created_at,
    m.updatedAt                   AS manga_updated_at,

    COUNT(DISTINCT c.id)::INT     AS total_chapters,
    COUNT(p.url)::INT             AS total_pages,

    COALESCE(MAX(c.number), 0)    AS latest_chapter_number,
    MAX(c.updatedAt)              AS latest_chapter_updated_at

FROM Manga m
LEFT JOIN Chapter c
    ON c.mangaId = m.id
LEFT JOIN Page p
    ON p.chapterId = c.id

GROUP BY
    m.id,
    m.title,
    m.url,
    m.status,
    m.poster,
    m.createdAt,
    m.updatedAt

ORDER BY
    m.updatedAt DESC;

-- name: ListMangaDetails :one
SELECT 
COUNT(DISTINCT c.id)::INT AS chapter_count,
COUNT(p.url)::INT AS page_count,
COALESCE(COUNT(p.url)::INT, 0)::INT / NULLIF(COUNT(DISTINCT c.id)::INT, 0) AS avg_pages_per_chapter
FROM Manga m
LEFT JOIN Chapter c
  ON c.mangaId = m.id
LEFT JOIN Page p
  ON p.chapterId = c.id
WHERE m.id = $1;

-- name: DashboardTaskDetails :one
SELECT 
    COUNT(*) FILTER (WHERE status = 'PENDING')::INT AS pending_tasks,
    COUNT(*) FILTER (WHERE status = 'RETRY')::INT AS retry_tasks,
    COUNT(*) FILTER (WHERE status IN ('SUCCESS', 'COMMITTED'))::INT AS successful_tasks,
    COUNT(*) FILTER (WHERE status = 'FAILURE')::INT AS failed_tasks,
    COUNT(*) FILTER (WHERE status = 'STARTED')::INT AS started_tasks
FROM Task;

-- name: DeleteTaskByID :one 
DELETE FROM Task WHERE id = $1 RETURNING *;

