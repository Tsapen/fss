package fss

import (
	"time"
)

type (
	File struct {
		Name            string     `db:"name"`
		LastServerID    int64      `db:"last_server_id"`
		LastCommittedAt *time.Time `db:"last_committed_at"`
		Fragments       *int       `db:"fragments"`
	}

	Server struct {
		ID  int64  `db:"id"`
		URL string `db:"url"`
	}
)
