package services

import (
	"download-server/cmd/utils"
	"download-server/db/generated"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (celery *CeleryConn) SendNewPipelineMangaTask(id string, url string) error {
	celery.QueueLock.Lock()
	defer celery.QueueLock.Unlock()
	q, err := celery.Channel.QueueDeclare(
		"celery", // queue name
		true,     // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return err
	}

	data := struct {
		URL string `json:"url"`
		ID  string `json:"id"`
	}{
		URL: url,
		ID:  id,
	}

	body, id, err := GenerateCeleryTask("tasks.pipeline.manga.scrape", []any{data})
	if err != nil {
		return err
	}

	err = celery.Channel.Publish("", q.Name, false, false, amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp.Persistent,
	})

	if err == nil {
		go CreateTask(id, "tasks.pipeline.manga.scrape", generated.TaskTypePIPELINE)
		go celery.WaitForTaskResult(id, "tasks.pipeline.manga.scrape")
	}

	return err
}

func (celery *CeleryConn) SendNewPipelineChapterTask(id string, url string, name string, chapterNo string) error {
	celery.QueueLock.Lock()
	defer celery.QueueLock.Unlock()
	q, err := celery.Channel.QueueDeclare(
		"celery", // queue name
		true,     // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return err
	}

	data := struct {
		URL      string `json:"url"`
		ID       string `json:"id"`
		Name     string `json:"name"`
		Chapter  string `json:"chapter"`
		BasePath string `json:"base_path"`
	}{
		URL:      url,
		ID:       id,
		Name:     name,
		Chapter:  chapterNo,
		BasePath: utils.BaseDir,
	}

	body, id, err := GenerateCeleryTask("tasks.pipeline.chapter.scrape", []any{data})
	if err != nil {
		return err
	}

	err = celery.Channel.Publish("", q.Name, false, false, amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp.Persistent,
	})

	if err == nil {
		go CreateTask(id, "tasks.pipeline.chapter.scrape", generated.TaskTypePIPELINE)
		go celery.WaitForTaskResult(id, "tasks.pipeline.chapter.scrape")
	}

	return err
}

func (celery *CeleryConn) SendNewPipelinePageDownloadTask(url string, index int, filePath string, chapterId string) error {
	celery.QueueLock.Lock()
	defer celery.QueueLock.Unlock()
	q, err := celery.Channel.QueueDeclare(
		"celery", // queue name
		true,     // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return err
	}

	data := struct {
		URL       string `json:"url"`
		Index     int    `json:"index"`
		FilePath  string `json:"file_path"`
		ChapterId string `json:"chapter_id"`
	}{
		URL:       url,
		Index:     index,
		FilePath:  filePath,
		ChapterId: chapterId,
	}

	body, id, err := GenerateCeleryTask("tasks.pipeline.page.download", []any{data})
	if err != nil {
		return err
	}

	err = celery.Channel.Publish("", q.Name, false, false, amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp.Persistent,
	})

	if err == nil {
		go CreateTask(id, "tasks.pipeline.page.download", generated.TaskTypePIPELINE)
		go celery.WaitForTaskResult(id, "tasks.pipeline.page.download")
	}

	return err
}

func (celery *CeleryConn) SendMangaArchiveTask(manga generated.Manga, pages []generated.GetPagesByRangeAndMangaIDRow, archive generated.Archive) error {
	celery.QueueLock.Lock()
	defer celery.QueueLock.Unlock()
	q, err := celery.Channel.QueueDeclare(
		"celery", // queue name
		true,     // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return err
	}

	data := ArchivePayload{
		Manga:   manga,
		Archive: archive,
		Pages:   pages,
	}
	body, id, err := GenerateCeleryTask("tasks.archive.manga.cbz", []any{data})
	if err != nil {
		return err
	}

	err = celery.Channel.Publish("", q.Name, false, false, amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp.Persistent,
	})

	if err == nil {
		go CreateTask(id, "tasks.archive.manga.cbz", generated.TaskTypeREQUEST)
		go celery.WaitForTaskResult(id, "tasks.archive.manga.cbz")
	}

	return err
}
