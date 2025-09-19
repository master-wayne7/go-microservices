package account

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
	"github.com/master-wayne7/go-microservices/monitoring"
)

type Repository interface {
	Close()
	// Expose DB for metrics
	DB() *sql.DB
	PutAccount(ctx context.Context, a Account) error
	GetAccountByID(ctx context.Context, id string) (*Account, error)
	ListAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error)
}

type PostgresRepository struct {
	db *sql.DB
	// Attach metrics collector for DB query metrics
	metrics *monitoring.MetricsCollector
}

func NewPostgresRepository(url string) (Repository, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &PostgresRepository{db: db}, nil
}

func (r *PostgresRepository) Close() {
	r.db.Close()
}

func (r *PostgresRepository) Ping() error {
	return r.db.Ping()
}

// Add DB accessor for metrics
func (r *PostgresRepository) DB() *sql.DB {
	return r.db
}

// Allow wiring metrics after repository creation
func (r *PostgresRepository) SetMetrics(mc *monitoring.MetricsCollector) {
	r.metrics = mc
}

func (r *PostgresRepository) PutAccount(ctx context.Context, a Account) error {
	start := time.Now()
	_, err := r.db.ExecContext(ctx, "INSERT INTO accounts(id,name) VALUES($1,$2)", a.ID, a.Name)
	if r.metrics != nil {
		r.metrics.RecordDBQuery("insert", "accounts", time.Since(start))
	}
	return err

}
func (r *PostgresRepository) GetAccountByID(ctx context.Context, id string) (*Account, error) {
	start := time.Now()
	row := r.db.QueryRowContext(ctx, "SELECT id, name FROM accounts WHERE id=$1", id)
	a := &Account{}
	if err := row.Scan(&a.ID, &a.Name); err != nil {
		if r.metrics != nil {
			r.metrics.RecordDBQuery("select", "accounts", time.Since(start))
		}
		return nil, err
	}
	if r.metrics != nil {
		r.metrics.RecordDBQuery("select", "accounts", time.Since(start))
	}
	return a, nil
}
func (r *PostgresRepository) ListAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error) {
	start := time.Now()
	rows, err := r.db.QueryContext(
		ctx,
		"SELECT id, name FROM accounts ORDER BY id DESC OFFSET $1 LIMIT $2",
		skip,
		take,
	)
	if err != nil {
		if r.metrics != nil {
			r.metrics.RecordDBQuery("select", "accounts", time.Since(start))
		}
		return nil, err
	}

	defer rows.Close()

	accounts := []Account{}

	for rows.Next() {
		a := Account{}
		if err := rows.Scan(&a.ID, &a.Name); err == nil {
			accounts = append(accounts, a)
		}
	}

	if err := rows.Err(); err != nil {
		if r.metrics != nil {
			r.metrics.RecordDBQuery("select", "accounts", time.Since(start))
		}
		return nil, err
	}
	if r.metrics != nil {
		r.metrics.RecordDBQuery("select", "accounts", time.Since(start))
	}
	return accounts, nil
}
