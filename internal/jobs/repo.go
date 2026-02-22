package jobs

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{
		db: db,
	}
}

func (r *Repo) Create(ctx context.Context, total int) (Job, error) {
	j := Job{
		ID:         NewID(),
		Status:     Running,
		CreatedAt:  time.Now().UTC(),
		FinishedAt: nil,
		Total:      total,
		Done:       0,
		Error:      "",
	}

	_, err := r.db.ExecContext(ctx, "insert into jobs (id, status, created_at, finished_at, total, done, error) values ($1, $2, $3, $4, $5, $6, $7)",
		j.ID,
		j.Status,
		j.CreatedAt,
		j.FinishedAt,
		j.Total,
		j.Done,
		j.Error,
	)
	if err != nil {
		return Job{}, err
	}

	return j, nil
}

func (r *Repo) Get(ctx context.Context, id string) (Job, bool, error) {
	row := r.db.QueryRowContext(ctx, "select id, status, created_at, finished_at, total, done, error from jobs where id = $1", id)

	j := Job{}
	var finishedAt sql.NullTime
	var errMsg sql.NullString
	err := row.Scan(&j.ID, &j.Status, &j.CreatedAt, &finishedAt, &j.Total, &j.Done, &errMsg)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Job{}, false, nil
		}
		return Job{}, false, err
	}
	if finishedAt.Valid {
		t := finishedAt.Time
		j.FinishedAt = &t
	}
	if errMsg.Valid {
		j.Error = errMsg.String
	}

	return j, true, nil
}

func (r *Repo) MarkDone(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "update jobs set status = 'done', done = total, finished_at = now(), error= null where id = $1", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *Repo) MarkFailed(ctx context.Context, id string, msg string) error {
	res, err := r.db.ExecContext(ctx, "update jobs set status = 'failed', error = $2, finished_at = now() where id = $1", id, msg)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}

	return nil
}
