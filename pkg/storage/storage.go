package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var ErrNoTasksToAdd = fmt.Errorf("empty tasks slice")

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

// Tasks возвращает список задач из БД.
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
