-- Active: 1769751283287@@127.0.0.1@5432@manga_db
-- CREATE DATABASE manga_scraper;

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE "Status" AS ENUM ('ONGOING', 'COMPLETED', 'UPCOMING');
CREATE TYPE "TaskType" AS ENUM ('PIPELINE', 'REQUEST');
CREATE TYPE "TaskStatus" AS ENUM ('PENDING', 'SUCCESS', 'RETRY', 'FAILURE', 'STARTED', 'COMMITTED');
CREATE TYPE "ArchiveStatus" AS ENUM ('PENDING', 'COMPLETED', 'FAILED');

CREATE TABLE IF NOT EXISTS Manga (
    id VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid(),
    url TEXT UNIQUE NOT NULL,
    
    title VARCHAR(255),
    status "Status" DEFAULT 'UPCOMING',
    poster TEXT,
    details JSON,
    description TEXT,

    createdAt TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updatedAt TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX manga_index ON Manga (title, status);

CREATE TABLE IF NOT EXISTS Chapter (
    id VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid(),
    number FLOAT NOT NULL,
    url TEXT NOT NULL,

    mangaId VARCHAR(36) NOT NULL REFERENCES Manga(id) ON DELETE CASCADE,
    createdAt TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updatedAt TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chapter_number_idx UNIQUE (number, mangaId)
);

CREATE INDEX chapter_index ON Chapter (mangaId, number);

CREATE TABLE IF NOT EXISTS Page (
    index INT NOT NULL,
    url TEXT NOT NULL,
    
    altText VARCHAR(255),
    filePath TEXT NOT NULL,
    downloadedAt TIMESTAMP(3),

    chapterId VARCHAR(36) NOT NULL REFERENCES Chapter(id) ON DELETE CASCADE,
    createdAt TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updatedAt TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (index, chapterId)
);

CREATE INDEX page_index ON Page (chapterId, index);

CREATE TABLE Archive (
    id VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid(),
    
    filePath TEXT NOT NULL,
    startChapter FLOAT DEFAULT 0,
    endChapter FLOAT DEFAULT 0,
    status "ArchiveStatus" DEFAULT 'PENDING',
    complete BOOLEAN DEFAULT FALSE,
    size BIGINT DEFAULT 0,
    
    mangaId VARCHAR(36) NOT NULL REFERENCES Manga(id) ON DELETE CASCADE,

    createdAt TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updatedAt TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX archive_index ON Archive(mangaId, startChapter, endChapter);

CREATE TABLE IF NOT EXISTS Task (
    id TEXT UNIQUE PRIMARY KEY NOT NULL,

    name VARCHAR(255) NOT NULL,
    type "TaskType" DEFAULT 'REQUEST',
    status "TaskStatus" DEFAULT 'PENDING',
    data JSON,
    payload JSON,

    createdAt TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updatedAt TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX task_index ON Task (status, updatedAt, name);

CREATE OR REPLACE FUNCTION set_current_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updatedAt = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER updated_at_Manga_trigger
BEFORE UPDATE ON Manga
FOR EACH ROW
EXECUTE FUNCTION set_current_timestamp();

CREATE TRIGGER updated_at_Chapter_trigger
BEFORE UPDATE ON Chapter
FOR EACH ROW
EXECUTE FUNCTION set_current_timestamp();

CREATE TRIGGER updated_at_Archive_trigger
BEFORE UPDATE ON Archive
FOR EACH ROW
EXECUTE FUNCTION set_current_timestamp();

CREATE TRIGGER updated_at_Page_trigger
BEFORE UPDATE ON Page
FOR EACH ROW
EXECUTE FUNCTION set_current_timestamp();

CREATE TRIGGER updated_at_Task_trigger
BEFORE UPDATE ON Task
FOR EACH ROW
EXECUTE FUNCTION set_current_timestamp();