package services

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func (celery *CeleryConn) SendNewMangaTask(id string, url string) error {
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

	body, id, err := GenerateCeleryTask("tasks.manga.scrape", []any{data})
	if err != nil {
		return err
	}

	err = celery.Channel.Publish("", q.Name, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})

	if err == nil {
		go CreateTask(id, "tasks.manga.scrape")
		go celery.WaitForTaskResult(id, "tasks.manga.scrape")
	}

	return err
}

func (celery *CeleryConn) SendNewChapterTask(id string, url string) error {
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

	body, id, err := GenerateCeleryTask("tasks.chapter.scrape", []any{data})
	if err != nil {
		return err
	}

	err = celery.Channel.Publish("", q.Name, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})

	if err == nil {
		go CreateTask(id, "tasks.chapter.scrape")
		go celery.WaitForTaskResult(id, "tasks.chapter.scrape")
	}

	return err
}
