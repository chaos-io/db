package db

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/plugin/dbresolver"
)

var _ Provider = &provider{}

//go:generate mockgen -destination=mocks/db.go -package=mocks . Provider
type Provider interface {
	NewSession(ctx context.Context, opts ...Option) *gorm.DB
	Transaction(ctx context.Context, fn func(tx *gorm.DB) error, opts ...Option) error
}

type provider struct {
	db *gorm.DB
}

func (p *provider) NewSession(ctx context.Context, opts ...Option) *gorm.DB {
	session := p.db

	opt := &option{}
	for _, fn := range opts {
		fn(opt)
	}

	if opt.tx != nil {
		session = opt.tx
	}
	if opt.debug {
		session = session.Debug()
	}
	if opt.master {
		session = session.Clauses(dbresolver.Write)
	}
	if opt.deleted {
		session = session.Unscoped()
	}
	if opt.SelectForUpdate {
		session = session.Clauses(clause.Locking{Strength: "UPDATE"})
	}

	return session.WithContext(ctx)
}

func (p *provider) Transaction(ctx context.Context, fn func(tx *gorm.DB) error, opts ...Option) error {
	session := p.NewSession(ctx, opts...)
	return session.Transaction(fn)
}

type Option func(*option)

type option struct {
	tx *gorm.DB

	debug           bool
	master          bool
	deleted         bool
	SelectForUpdate bool
}

func WithDebug() Option {
	return func(o *option) {
		o.debug = true
	}
}

func WithMaster() Option {
	return func(o *option) {
		o.master = true
	}
}

func WithDeleted() Option {
	return func(o *option) {
		o.deleted = true
	}
}

func WithSelectForUpdate() Option {
	return func(o *option) {
		o.SelectForUpdate = true
	}
}
