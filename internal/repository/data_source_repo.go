package repository

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/muyue/comic-harmony-backend/internal/model"
)

type DataSourceRepository interface {
	Create(ctx context.Context, ds *model.DataSource) error
	GetByID(ctx context.Context, id int64) (*model.DataSource, error)
	ListByUser(ctx context.Context, userID int64) ([]model.DataSource, error)
	Update(ctx context.Context, ds *model.DataSource) error
	Delete(ctx context.Context, id int64) error
	GetAll(ctx context.Context) ([]model.DataSource, error)
}

type dataSourceRepo struct {
	pool *pgxpool.Pool
}

func NewDataSourceRepository(pool *pgxpool.Pool) DataSourceRepository {
	return &dataSourceRepo{pool: pool}
}

func encodePassword(pwd string) string {
	return base64.StdEncoding.EncodeToString([]byte(pwd))
}

func decodePassword(encoded string) string {
	b, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return ""
	}
	return string(b)
}

func scanDataSource(row pgx.Row) (*model.DataSource, error) {
	var ds model.DataSource
	var lastHealth *time.Time
	var pwd string
	err := row.Scan(
		&ds.ID, &ds.UserID, &ds.Name, &ds.SourceType,
		&ds.URL, &ds.Username, &pwd, &ds.RootPath,
		&ds.Enabled, &lastHealth, &ds.CreatedAt, &ds.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	ds.Password = decodePassword(pwd)
	ds.LastHealth = lastHealth
	return &ds, nil
}

func (r *dataSourceRepo) Create(ctx context.Context, ds *model.DataSource) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO data_sources (user_id, name, source_type, url, username, password, root_path, enabled)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, created_at, updated_at`,
		ds.UserID, ds.Name, ds.SourceType, ds.URL,
		ds.Username, encodePassword(ds.Password),
		ds.RootPath, ds.Enabled,
	).Scan(&ds.ID, &ds.CreatedAt, &ds.UpdatedAt)
	return err
}

func (r *dataSourceRepo) GetByID(ctx context.Context, id int64) (*model.DataSource, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, user_id, name, source_type, url, username, password, root_path,
		        enabled, last_health, created_at, updated_at
		 FROM data_sources WHERE id = $1`, id)
	return scanDataSource(row)
}

func (r *dataSourceRepo) ListByUser(ctx context.Context, userID int64) ([]model.DataSource, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, name, source_type, url, username, password, root_path,
		        enabled, last_health, created_at, updated_at
		 FROM data_sources WHERE user_id = $1 ORDER BY created_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sources []model.DataSource
	for rows.Next() {
		ds, err := scanDataSource(rows)
		if err != nil {
			return nil, err
		}
		sources = append(sources, *ds)
	}
	return sources, nil
}

func (r *dataSourceRepo) Update(ctx context.Context, ds *model.DataSource) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE data_sources SET name=$1, source_type=$2, url=$3, username=$4,
		 password=$5, root_path=$6, enabled=$7, updated_at=NOW()
		 WHERE id=$8`,
		ds.Name, ds.SourceType, ds.URL, ds.Username,
		encodePassword(ds.Password), ds.RootPath, ds.Enabled, ds.ID,
	)
	return err
}

func (r *dataSourceRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM data_sources WHERE id = $1`, id)
	return err
}

func (r *dataSourceRepo) GetAll(ctx context.Context) ([]model.DataSource, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, name, source_type, url, username, password, root_path,
		        enabled, last_health, created_at, updated_at
		 FROM data_sources ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sources []model.DataSource
	for rows.Next() {
		ds, err := scanDataSource(rows)
		if err != nil {
			return nil, err
		}
		sources = append(sources, *ds)
	}
	return sources, nil
}

var _ DataSourceRepository = (*dataSourceRepo)(nil)
