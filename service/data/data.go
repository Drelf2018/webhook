package data

import (
	"errors"

	"github.com/Drelf2018/resource"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Model struct {
	ID uint64 `gorm:"primaryKey;autoIncrement" form:"-" json:"-"`
}

var (
	db     *gorm.DB
	folder string
	public resource.Explorer
)

func Public() resource.Explorer {
	return public
}

func SetPublic(name string, r resource.Explorer) {
	folder = name
	public = r
	public.MkdirAll()
}

func SetDB(r *gorm.DB) *gorm.DB {
	db = r
	db.AutoMigrate(new(Post))
	return db
}

func SetDialector(dialector gorm.Dialector) *gorm.DB {
	db, _ = gorm.Open(dialector, &gorm.Config{})
	return SetDB(db)
}

func SetSqlite(file string) *gorm.DB {
	return SetDialector(sqlite.Open(file))
}

func Exists[T any](conds ...any) bool {
	return !errors.Is(db.First(new(T), conds...).Error, gorm.ErrRecordNotFound)
}

func Update(x any, conds ...any) bool {
	return !errors.Is(db.First(x, conds...).Error, gorm.ErrRecordNotFound)
}

func PreloadDB(in any) (r *gorm.DB) {
	r = db.Model(in)
	s := ParseStruct(in)
	for i, l := 0, len(s); i < l; i++ {
		r.Preload(s[i])
	}
	return r.Preload(clause.Associations)
}

func Preload[T any](t *T, conds ...any) error {
	return PreloadDB(new(T)).First(t, conds...).Error
}

func Preloads[T any](t *[]T, conds ...any) error {
	return PreloadDB(new(T)).Find(t, conds...).Error
}