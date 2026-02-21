package db

import "database/sql"

func EnsureSchema(db *sql.DB) error {
	_, err := db.Exec("create table if not exists jobs (id text primary key, status text not null, created_at timestamptz not null, finished_at timestamptz not null, total int not null, done int not null, error text null)")
	if err != nil {
		return err
	}

	return nil
}
