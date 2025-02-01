package storage

import (
	"fmt"
	"os"
	"testing"
)

var (
	ErrConnectDB       = fmt.Errorf("Unable to establish DB connection")
	ErrDBNotResponding = fmt.Errorf("DB not responding")
)

func postgresConf() Config {
	conf := Config{
		User:     "postgres",
		Password: os.Getenv("POSTGRES_PASSWORD"),
		Host:     "localhost",
		Port:     "5433",
		DBName:   "tasks",
	}

	return conf
}

func storageConnect() (*Storage, error) {
	conf := postgresConf()
	db, err := New(conf.ConString())
	if err != nil {
		return nil, ErrConnectDB
	}

	err = db.Ping()
	if err != nil {
		return nil, ErrDBNotResponding
	}

	return db, nil
}

func TestStorage_TasksAll(t *testing.T) {
	db, err := storageConnect()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	wantTasksCnt := 5
	tasks, err := db.Tasks(0, 0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(tasks) != wantTasksCnt {
		t.Errorf("tasks num: want %d, got %d", wantTasksCnt, len(tasks))
	}
}

func TestStorage_TasksByTaskIDExists(t *testing.T) {
	db, err := storageConnect()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	wantTasksCnt := 1
	for targetTaskID := 1; targetTaskID <= 5; targetTaskID++ {
		tasks, err := db.Tasks(targetTaskID, 0)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(tasks) != wantTasksCnt {
			t.Errorf("tasks num: want %d, got %d", wantTasksCnt, len(tasks))
			t.FailNow()
		}
		if tasks[0].ID != targetTaskID {
			t.Errorf("task ID: want %d, got %d", targetTaskID, tasks[0].ID)
		}
	}
}

func TestStorage_TasksByTaskIDDoesNotExist(t *testing.T) {
	db, err := storageConnect()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	wantTasksCnt := 0
	targetTaskID := 42
	tasks, err := db.Tasks(targetTaskID, 0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(tasks) != wantTasksCnt {
		t.Errorf("tasks num: want %d, got %d", wantTasksCnt, len(tasks))
	}
}

func TestStorage_TasksByAuthorIDExists(t *testing.T) {
	db, err := storageConnect()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	wantTasksCnt := 2
	targetAuthorID := 4
	tasks, err := db.Tasks(0, targetAuthorID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(tasks) != wantTasksCnt {
		t.Errorf("tasks num: want %d, got %d", wantTasksCnt, len(tasks))
		t.FailNow()
	}
	for _, task := range tasks {
		if task.AuthorID != targetAuthorID {
			t.Errorf("tasks AuthorID: want %d, got %d", targetAuthorID, tasks[0].AuthorID)
		}
	}
}

func TestStorage_TasksByAuthorIDDoesNotExist(t *testing.T) {
	db, err := storageConnect()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	wantTasksCnt := 0
	targetAuthorID := 42
	tasks, err := db.Tasks(0, targetAuthorID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(tasks) != wantTasksCnt {
		t.Errorf("tasks num: want %d, got %d", wantTasksCnt, len(tasks))
	}
}

func TestStorage_NewTasks(t *testing.T) {
	task1 := Task{
		Title:   "Task 1",
		Content: "This is the content of Task 1",
	}
	task2 := Task{
		Title:   "Task 2",
		Content: "This is the content of Task 2",
	}
	task3 := Task{
		Title:   "Task 3",
		Content: "This is the content of Task 3",
	}

	tests := []struct {
		name    string
		tasks   []Task
		wantErr error
	}{
		{"No tasks", []Task{}, ErrNoTasksToAdd},
		{"Three tasks", []Task{task1, task2, task3}, nil},
	}

	db, err := storageConnect()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks, err := db.Tasks(0, 0)
			if err != nil {
				t.Fatal(err)
			}
			taskNumBefore := len(tasks)

			err = db.NewTasks(tt.tasks)
			if err != tt.wantErr {
				t.Errorf("expected error %v, got error %v", tt.wantErr, err)
			}
			tasks, err = db.Tasks(0, 0)
			if err != nil {
				t.Fatal(err)
			}
			taskNumAfter := len(tasks)
			wantTaskNum := taskNumBefore + len(tt.tasks)
			if taskNumAfter != wantTaskNum {
				t.Errorf("new tasks wer'nt added: expected tasks %d, got tasks %d", wantTaskNum, taskNumAfter)
			}
		})
	}
}
