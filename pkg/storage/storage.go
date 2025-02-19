package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	ErrNoTasksToAdd = fmt.Errorf("empty tasks slice")
	ErrTaskNotFound = fmt.Errorf("task not found")
	ErrEmptyLabel   = fmt.Errorf("label cannot be empty")
)

// Хранилище данных.
type Storage struct {
	db *pgxpool.Pool
}

func (s *Storage) Ping() error {
	return s.db.Ping(context.Background())
}

func (s *Storage) Close() {
	s.db.Close()
}

// Конструктор, принимает строку подключения к БД.
func New(constr string) (*Storage, error) {
	db, err := pgxpool.Connect(context.Background(), constr)
	if err != nil {
		return nil, err
	}
	s := Storage{
		db: db,
	}
	return &s, nil
}

// Задача.
type Task struct {
	ID         int
	Opened     int64
	Closed     int64
	AuthorID   int
	AssignedID int
	Title      string
	Content    string
}

// Deprecated: Tasks возвращает список задач из БД.
func (s *Storage) Tasks(taskID, authorID int) ([]Task, error) {
	rows, err := s.db.Query(context.Background(), `
		SELECT 
			id,
			opened,
			closed,
			author_id,
			assigned_id,
			title,
			content
		FROM tasks
		WHERE
			($1 = 0 OR id = $1) AND
			($2 = 0 OR author_id = $2)
		ORDER BY id;
	`,
		taskID,
		authorID,
	)
	if err != nil {
		return nil, err
	}
	var tasks []Task
	// итерирование по результату выполнения запроса
	// и сканирование каждой строки в переменную
	for rows.Next() {
		var t Task
		err = rows.Scan(
			&t.ID,
			&t.Opened,
			&t.Closed,
			&t.AuthorID,
			&t.AssignedID,
			&t.Title,
			&t.Content,
		)
		if err != nil {
			return nil, err
		}
		// добавление переменной в массив результатов
		tasks = append(tasks, t)

	}
	// ВАЖНО не забыть проверить rows.Err()
	return tasks, rows.Err()
}

// TasksAll возвращает список задач из БД.
func (s *Storage) TasksAll() ([]Task, error) {
	ctx := context.Background()
	rows, err := s.db.Query(ctx, `
		SELECT
			id,
			opened,
			closed,
			author_id,
			assigned_id,
			title,
			content
		FROM tasks
	`)
	if err != nil {
		return nil, err
	}

	var tasks []Task
	for rows.Next() {
		var t Task
		err = rows.Scan(
			&t.ID,
			&t.Opened,
			&t.Closed,
			&t.AuthorID,
			&t.AssignedID,
			&t.Title,
			&t.Content,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, rows.Err()
}

// TasksByID возвращает задачу по ID.
func (s *Storage) TaskByID(taskID int) (Task, error) {
	ctx := context.Background()
	var task Task
	err := s.db.QueryRow(ctx, `
		SELECT
			id,
			opened,
			closed,
			author_id,
			assigned_id,
			title,
			content
		FROM tasks
		WHERE id = $1
	`,
		taskID,
	).Scan(
		&task.ID,
		&task.Opened,
		&task.Closed,
		&task.AuthorID,
		&task.AssignedID,
		&task.Title,
		&task.Content,
	)

	if err == pgx.ErrNoRows {
		return task, ErrTaskNotFound
	}

	return task, err
}

// TaskByAuthorID возвращает список задач из БД по ID автора.
func (s *Storage) TasksByAuthorID(authorID int) ([]Task, error) {
	ctx := context.Background()
	rows, err := s.db.Query(ctx, `
		SELECT
			id,
			opened,
			closed,
			author_id,
			assigned_id,
			title,
			content
		FROM tasks
		WHERE author_id = $1
	`,
		authorID,
	)
	if err != nil {
		return nil, err
	}

	var tasks []Task
	for rows.Next() {
		var t Task
		err = rows.Scan(
			&t.ID,
			&t.Opened,
			&t.Closed,
			&t.AuthorID,
			&t.AssignedID,
			&t.Title,
			&t.Content,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, rows.Err()
}

// TasksByLabel возвращает список задач из БД по метке.
func (s *Storage) TasksByLabel(label string) ([]Task, error) {
	if label == "" {
		return nil, ErrEmptyLabel
	}

	ctx := context.Background()
	rows, err := s.db.Query(ctx, `
		SELECT
			t.id,
			t.opened,
			t.closed,
			t.author_id,
			t.assigned_id,
			t.title,
			t.content
		FROM tasks AS t
		JOIN tasks_labels AS tl
		ON tl.task_id = t.id
		JOIN labels AS l
		ON tl.label_id = l.id
		WHERE l.name = $1
	`,
		label)
	if err != nil {
		return nil, err
	}

	var tasks []Task
	for rows.Next() {
		var t Task
		err = rows.Scan(
			&t.ID,
			&t.Opened,
			&t.Closed,
			&t.AuthorID,
			&t.AssignedID,
			&t.Title,
			&t.Content,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, rows.Err()
}

// NewTask создаёт новую задачу и возвращает её id.
func (s *Storage) NewTask(t Task) (int, error) {
	var id int
	err := s.db.QueryRow(context.Background(), `
		INSERT INTO tasks (title, content)
		VALUES ($1, $2) RETURNING id;
		`,
		t.Title,
		t.Content,
	).Scan(&id)
	return id, err
}

// NewTasks создает несколько новых задач
func (s *Storage) NewTasks(tasks []Task) error {
	if len(tasks) == 0 {
		return ErrNoTasksToAdd
	}

	ctx := context.Background()
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	batch := new(pgx.Batch)
	for _, t := range tasks {
		batch.Queue(`
        INSERT INTO tasks (title, content)
		VALUES ($1, $2);
        `,
			t.Title,
			t.Content,
		)
	}

	res := tx.SendBatch(ctx, batch)
	err = res.Close()
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// UpdateTask обновляет задачу по id.
// Обновляет соответствующие атрибуты в случае если передан не нулевой параметр.
// Обновление происходит в один SQL запрос.
func (s *Storage) UpdateTask(taskID, assignedID int, closed int64, title, content string) error {
	ctx := context.Background()
	_, err := s.db.Exec(ctx, `
		UPDATE tasks
		SET
			closed = CASE WHEN $2 > 0 THEN $2 ELSE closed END,
			assigned_id = CASE WHEN $3 > 0 THEN $3 ELSE assigned_id END,
			title = CASE WHEN $4 <> '' THEN $4 ELSE title END,
			content = CASE WHEN $5 <> '' THEN $5 ELSE content END
		WHERE id = $1
	`,
		taskID,
		closed,
		assignedID,
		title,
		content,
	)
	if err != nil {
		return err
	}

	return nil
}

// DeleteTask удаляет задачу по ID.
func (s *Storage) DeleteTask(taskID int) error {
	ctx := context.Background()
	_, err := s.db.Exec(ctx, `
		DELETE FROM tasks
		WHERE id = $1
	`,
		taskID,
	)
	if err != nil {
		return err
	}

	return nil
}
