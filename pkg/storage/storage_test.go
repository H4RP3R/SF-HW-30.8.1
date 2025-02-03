package storage

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"
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

func TestStorage_Tasks(t *testing.T) {
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

func TestStorage_TasksAll(t *testing.T) {
	db, err := storageConnect()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	wantTasksCnt := 5
	tasks, err := db.TasksAll()
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

func TestStorage_TasksByID(t *testing.T) {
	db, err := storageConnect()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	tasks, err := db.TasksAll()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	for _, targetTask := range tasks {
		task, err := db.TaskByID(targetTask.ID)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if !reflect.DeepEqual(task, targetTask) {
			t.Errorf("Tasks do not match: %+v, %+v", task, targetTask)
		}
	}
}

func TestStorage_TasksByIDDoesNotExist(t *testing.T) {
	db, err := storageConnect()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	targetID := 99999
	_, err = db.TaskByID(targetID)
	if !errors.Is(err, ErrTaskNotFound) {
		t.Errorf("error: want %v, got %v", ErrTaskNotFound, err)
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
	tasks, err := db.TasksByAuthorID(targetAuthorID)
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
	tasks, err := db.TasksByAuthorID(targetAuthorID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(tasks) != wantTasksCnt {
		t.Errorf("tasks num: want %d, got %d", wantTasksCnt, len(tasks))
	}
}

func TestStorage_TasksByLabel(t *testing.T) {
	db, err := storageConnect()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	tests := []struct {
		name        string
		targetLabel string
		wantTaskCnt int
		wantError   error
	}{
		{"label exists, 2 tasks", "Feature", 2, nil},
		{"label exists, 1 task", "Bug", 1, nil},
		{"label doesn't exist, 0 tasks, 2 tasks", "chore", 0, nil},
		{"empty label, 0 task", "", 0, ErrEmptyLabel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks, err := db.TasksByLabel(tt.targetLabel)
			if !errors.Is(err, tt.wantError) {
				t.Errorf("error: want %v, got %v", tt.wantError, err)
			}
			if len(tasks) != tt.wantTaskCnt {
				t.Errorf("Tasks num: want %d, got %d", tt.wantTaskCnt, len(tasks))
			}
		})
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

func TestStorage_UpdateTaskAllFields(t *testing.T) {
	db, err := storageConnect()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	targetTaskID := 3
	newClosed := time.Now().Unix() + 1000
	newAssignedID := 4
	newTitle := "New Title"
	newContent := "New Text"
	err = db.UpdateTask(targetTaskID, newAssignedID, newClosed, newTitle, newContent)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	tasks, err := db.Tasks(targetTaskID, 0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(tasks) == 0 {
		t.Fatalf("Unable to find target task id:%d", targetTaskID)
	}
	updatedTask := tasks[0]

	if updatedTask.AssignedID != newAssignedID {
		t.Errorf("task.assigned_id: want %d, got %d", newAssignedID, updatedTask.AssignedID)
	}
	if updatedTask.Closed != newClosed {
		t.Errorf("task.assigned_id: want %d, got %d", newClosed, updatedTask.Closed)
	}
	if updatedTask.Title != newTitle {
		t.Errorf("task.assigned_id: want %s, got %s", newTitle, updatedTask.Title)
	}
	if updatedTask.Content != newContent {
		t.Errorf("task.assigned_id: want %s, got %s", newContent, updatedTask.Content)
	}
}

func TestStorage_UpdateTaskPartialFields(t *testing.T) {
	db, err := storageConnect()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	targetTaskID := 5
	type param struct {
		Closed     int64
		AssignedID int
		Title      string
		Content    string
	}

	tests := []struct {
		name string
		param
	}{
		{"field: closed", param{Closed: time.Now().Unix() + 1000}},
		{"field: assigned_id", param{AssignedID: 1}},
		{"field: title", param{Title: "New title"}},
		{"field: content", param{Content: "New content"}},
		{"multiple fields: closed and assigned_id", param{Closed: time.Now().Unix() + 1000, AssignedID: 2}},
		{"multiple fields: title and content", param{Title: "New title", Content: "New content"}},
		{"all fields", param{Closed: time.Now().Unix() + 1000, AssignedID: 3, Title: "New title", Content: "New content"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks, err := db.Tasks(targetTaskID, 0)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if len(tasks) == 0 {
				t.Fatalf("Unable to find target task id:%d", targetTaskID)
			}
			taskBeforeUpdate := tasks[0]

			err = db.UpdateTask(targetTaskID, tt.AssignedID, tt.Closed, tt.Title, tt.Content)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			tasks, err = db.Tasks(targetTaskID, 0)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if len(tasks) == 0 {
				t.Fatalf("Unable to find target task id:%d", targetTaskID)
			}
			updatedTask := tasks[0]

			if tt.AssignedID == 0 && updatedTask.AssignedID != taskBeforeUpdate.AssignedID {
				t.Errorf("task.assigned_id: want %d, got %d", taskBeforeUpdate.AssignedID, updatedTask.AssignedID)
			}
			if tt.Closed == 0 && updatedTask.Closed != taskBeforeUpdate.Closed {
				t.Errorf("task.assigned_id: want %d, got %d", taskBeforeUpdate.Closed, updatedTask.Closed)
			}
			if tt.Title == "" && updatedTask.Title != taskBeforeUpdate.Title {
				t.Errorf("task.assigned_id: want %s, got %s", taskBeforeUpdate.Title, updatedTask.Title)
			}
			if tt.Content == "" && updatedTask.Content != taskBeforeUpdate.Content {
				t.Errorf("task.assigned_id: want %s, got %s", taskBeforeUpdate.Content, updatedTask.Content)
			}
		})
	}
}

func TestStorage_DeleteTaskIDExists(t *testing.T) {
	db, err := storageConnect()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	tasks, err := db.Tasks(0, 0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(tasks) == 0 {
		t.Fatal("Test DB is empty")
	}

	const tasksToDel = 3
	expectedTasksNum := len(tasks) - tasksToDel

	taskIdToDelete := make([]int, tasksToDel, tasksToDel)
	for i := expectedTasksNum; i < len(tasks); i++ {
		taskIdToDelete = append(taskIdToDelete, tasks[i].ID)
	}

	for _, id := range taskIdToDelete {
		err := db.DeleteTask(id)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		tasks, err := db.Tasks(0, 0)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		for _, task := range tasks {
			if task.ID == id {
				t.Errorf("Task with id:%d wasn't deleted", id)
			}
		}
	}

	tasks, err = db.Tasks(0, 0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(tasks) != expectedTasksNum {
		t.Errorf("Tasks num: want %d, got %d", expectedTasksNum, len(tasks))
	}
}

func TestStorage_DeleteTaskIDDoesNotExist(t *testing.T) {
	db, err := storageConnect()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	targetID := 9999
	err = db.DeleteTask(targetID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	tasks, err := db.Tasks(0, 0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	for _, task := range tasks {
		if task.ID == targetID {
			t.Errorf("Task with id:%d wasn't deleted", targetID)
		}
	}
}

func TestStorage_NewTask(t *testing.T) {
	db, err := storageConnect()
	if err != nil {
		t.Fatal(err)
	}

	deleteNewTask := func(id int) {
		err := db.DeleteTask(id)
		if err != nil {
			t.Errorf("Can't remove new task: %v", err)
		}
		db.Close()
	}

	newTask := Task{
		Title:   "New Test Task",
		Content: "Some content",
	}

	newTaskID, err := db.NewTask(newTask)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	t.Cleanup(func() {
		deleteNewTask(newTaskID)
	})

	_, err = db.TaskByID(newTaskID)
	if errors.Is(err, ErrTaskNotFound) {
		t.Errorf("Can't find new task in DB: %v", err)
	} else if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
