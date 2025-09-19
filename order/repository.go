package order

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
	"github.com/master-wayne7/go-microservices/monitoring"
)

type Repository interface {
	Close()
	PutOrder(ctx context.Context, o Order) error
	GetOrdersForAccount(ctx context.Context, accountID string) ([]Order, error)
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
	return &PostgresRepository{
		db: db,
	}, nil
}

func (r *PostgresRepository) Close() {
	r.db.Close()
}

// Add DB accessor for metrics
func (r *PostgresRepository) DB() *sql.DB {
	return r.db
}

func (r *PostgresRepository) PutOrder(ctx context.Context, o Order) (err error) {
	startTx := time.Now()
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()
	startInsert := time.Now()
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO orders(id, created_at, account_id, total_price) VALUES($1, $2, $3, $4)",
		o.ID,
		o.CreatedAt,
		o.AccountID,
		o.TotalPrice,
	)
	if err != nil {
		return err
	}
	if r.metrics != nil {
		r.metrics.RecordDBQuery("insert", "orders", time.Since(startInsert))
	}

	stmt, _ := tx.PrepareContext(ctx, pq.CopyIn("order_products", "order_id", "product_id", "quantity"))

	for _, p := range o.Products {
		startCopy := time.Now()
		_, err = stmt.ExecContext(
			ctx,
			o.ID,
			p.ID,
			p.Quantity,
		)
		if err != nil {
			return err
		}
		if r.metrics != nil {
			r.metrics.RecordDBQuery("copy", "order_products", time.Since(startCopy))
		}
	}

	startCopyFlush := time.Now()
	_, err = stmt.ExecContext(ctx)
	if err != nil {
		return err
	}
	if r.metrics != nil {
		r.metrics.RecordDBQuery("copy", "order_products", time.Since(startCopyFlush))
	}

	stmt.Close()
	if r.metrics != nil {
		r.metrics.RecordDBQuery("tx", "orders", time.Since(startTx))
	}
	return
}

func (r *PostgresRepository) GetOrdersForAccount(ctx context.Context, accountId string) ([]Order, error) {
	start := time.Now()
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT
		o.id,
		o.created_at,
		o.account_id,
		o.total_price::money::numeric::float8,
		op.product_id,
		op.quantity
		FROM orders o JOIN order_products op ON(o.id = op.order_id)
		WHERE o.account_id = $1
		ORDER BY o.id
		`,
		accountId,
	)
	if err != nil {
		if r.metrics != nil {
			r.metrics.RecordDBQuery("select", "orders_join_order_products", time.Since(start))
		}
		return nil, err
	}
	defer rows.Close()
	orders := []Order{}
	lastOrder := &Order{}
	orderedProduct := &OrderedProduct{}
	products := []OrderedProduct{}
	order := &Order{}

	for rows.Next() {
		if err = rows.Scan(
			&order.ID,
			&order.CreatedAt,
			&order.AccountID,
			&order.TotalPrice,
			&orderedProduct.ID,
			&orderedProduct.Quantity,
		); err != nil {
			return nil, err
		}
		if lastOrder.ID != "" && lastOrder.ID != order.ID {
			lastOrder.Products = products
			newOrder := Order{
				ID:         lastOrder.ID,
				AccountID:  lastOrder.AccountID,
				CreatedAt:  lastOrder.CreatedAt,
				TotalPrice: lastOrder.TotalPrice,
				Products:   lastOrder.Products,
			}
			orders = append(orders, newOrder)
			products = []OrderedProduct{}
		}
		products = append(products, OrderedProduct{
			ID:       orderedProduct.ID,
			Quantity: orderedProduct.Quantity,
		})
		*lastOrder = *order
	}

	if lastOrder.ID != "" {
		lastOrder.Products = products
		newOrder := Order{
			ID:         lastOrder.ID,
			AccountID:  lastOrder.AccountID,
			CreatedAt:  lastOrder.CreatedAt,
			TotalPrice: lastOrder.TotalPrice,
			Products:   lastOrder.Products,
		}
		orders = append(orders, newOrder)
	}

	if err = rows.Err(); err != nil {
		if r.metrics != nil {
			r.metrics.RecordDBQuery("select", "orders_join_order_products", time.Since(start))
		}
		return nil, err
	}
	if r.metrics != nil {
		r.metrics.RecordDBQuery("select", "orders_join_order_products", time.Since(start))
	}
	return orders, nil
}

// Allow wiring metrics into repository
func (r *PostgresRepository) SetMetrics(mc *monitoring.MetricsCollector) {
	r.metrics = mc
}
