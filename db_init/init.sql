DROP DATABASE IF EXISTS tasks;
CREATE DATABASE tasks;

\c tasks;

DROP TABLE IF EXISTS tasks_labels, tasks, labels, users;

-- пользователи системы
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

-- метки задач
CREATE TABLE labels (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

-- задачи
CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    opened BIGINT NOT NULL DEFAULT extract(epoch from now()), -- время создания задачи
    closed BIGINT DEFAULT 0, -- время выполнения задачи
    author_id INTEGER REFERENCES users(id) DEFAULT 0, -- автор задачи
    assigned_id INTEGER REFERENCES users(id) DEFAULT 0, -- ответственный
    title TEXT, -- название задачи
    content TEXT -- задачи
);

-- связь многие - ко- многим между задачами и метками
CREATE TABLE tasks_labels (
    task_id INTEGER REFERENCES tasks(id),
    label_id INTEGER REFERENCES labels(id)
);
-- наполнение БД начальными данными
INSERT INTO users (id, name) VALUES (0, 'default');

-- Наполнение тестовыми данными
-- Пользователи
INSERT INTO users (name) VALUES
    ('John Doe'),
    ('Jane Doe'),
    ('Bob Smith'),
    ('Alice Johnson'),
    ('Mike Brown');

-- Метки
INSERT INTO labels (name) VALUES
    ('Bug'),
    ('Feature'),
    ('Task'),
    ('Enhancement'),
    ('Documentation');

-- Задачи
INSERT INTO tasks (title, content, author_id, assigned_id) VALUES
    ('Fix login issue', 'The login feature is not working as expected.', 1, 2),
    ('Implement new feature', 'Implement a new feature to improve user experience.', 3, 1),
    ('Write documentation', 'Write documentation for the new feature.', 4, 5),
    ('Refactor code', 'Refactor the code to improve performance.', 2, 3),
    ('Test new feature', 'Test the new feature to ensure it works as expected.', 5, 1);

-- Повесить метки на таски
INSERT INTO tasks_labels (task_id, label_id) VALUES
    (1, 1),
    (2, 2),
    (3, 4),
    (4, 3),
    (5, 2),
    (1, 3),
    (2, 5);