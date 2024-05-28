package postgres

import (
	"context"

	errorwrapper "github.com/ecumenos-social/error-wrapper"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Driver struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, url string) (*Driver, error) {
	pgClient := &Driver{}

	pool, err := pgxpool.Connect(ctx, url)
	if err != nil {
		return nil, errorwrapper.WrapMessage(err, "connect to database failure")
	}

	pgClient.pool = pool

	return pgClient, nil
}

func (c *Driver) Close() {
	c.pool.Close()
}

func (c *Driver) Ping(ctx context.Context) error {
	if err := c.pool.Ping(ctx); err != nil {
		return errorwrapper.WrapMessage(err, "ping database failure")
	}
	return nil
}

func (c *Driver) acquireConn(ctx context.Context) (*pgxpool.Conn, error) {
	conn, err := c.pool.Acquire(ctx)
	if err != nil {
		return nil, errorwrapper.WrapMessage(err, "acquire connection failure")
	}

	return conn, nil
}

func (c *Driver) QueryRow(ctx context.Context, query string, args ...interface{}) (pgx.Row, error) {
	conn, err := c.acquireConn(ctx)
	if err != nil {
		return nil, err
	}

	defer conn.Release()

	return conn.QueryRow(ctx, query, args...), nil
}

func (c *Driver) QueryRows(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	conn, err := c.acquireConn(ctx)
	if err != nil {
		return nil, err
	}

	defer conn.Release()

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, errorwrapper.WrapMessage(err, "query database failure")
	}

	return rows, nil
}

func (c *Driver) CountRows(ctx context.Context, query string, args ...interface{}) (int, error) {
	var count int

	conn, err := c.acquireConn(ctx)
	if err != nil {
		return count, err
	}

	defer conn.Release()

	if err = conn.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return 0, errorwrapper.WrapMessage(err, "query database row failure")
	}

	return count, nil
}

func (c *Driver) ExecuteQuery(ctx context.Context, query string, args ...interface{}) error {
	conn, err := c.acquireConn(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	if _, err = conn.Exec(ctx, query, args...); err != nil {
		return errorwrapper.WrapMessage(err, "execute query database failure")
	}
	return nil
}
