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
	if ctx.Err() != nil {
		return Job{}, ctx.Err()
	}

	j := Job{
		ID:         NewID(),
		Status:     Running,
		CreatedAt:  time.Now().UTC(),
		FinishedAt: nil,
		Total:      total,
		Done:       0,
		Error:      "",
	}

	_, err := r.db.Exec("insert into jobs (id, status, created_at, finished_at, total, done, error) values ($1, $2, $3, $4, $5, $6, $7)",
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

func (r *Repo) Get(ctx context.Context, id int) (Job, bool, error) {
	if ctx.Err() != nil {
		return Job{}, false, ctx.Err()
	}

	_, err := r.db.Exec("select id, status, created_at, finished_at, done, error from jobs where id = $1", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Job{}, false, nil
		}
		return Job{}, false, err
	}

	return Job{}, true, nil
}

func (r *Repo) MarkDone(ctx context.Context, id string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	_, err := r.db.Exec("update jobs set status = 'done', done = total, finished_at = now(), error= null where id = $1", id)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repo) MarkFailed(ctx context.Context, id string, msg string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	_, err := r.db.Exec("update jobs set status = 'failed', error = $2, finished_at = now() where id = $1", id, msg)
	if err != nil {
		return err
	}

	return nil
}
