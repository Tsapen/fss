DROP TABLE IF EXISTS files;

DROP TABLE IF EXISTS servers;

CREATE TABLE IF NOT EXISTS files (
    name VARCHAR(100) NOT NULL,
    last_server_id INT NOT NULL,
    last_committed_at TIMESTAMP,
    fragments INT,

    CONSTRAINT unique_files_name UNIQUE (name)
);

CREATE UNIQUE INDEX IF NOT EXISTS unique_index_files_name ON files (name);

CREATE TABLE IF NOT EXISTS servers (
    id SERIAL NOT NULL PRIMARY KEY,
    url VARCHAR(100) NOT NULL,

    CONSTRAINT unique_servers_url UNIQUE (url)
);
