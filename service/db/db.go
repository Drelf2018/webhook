package db

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/Drelf2018/TypeGo/Reflect"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var Ref = Reflect.New[[]string](func(self *Reflect.Reflect[[]string], field reflect.StructField, elem reflect.Type) []string {
	if _, ok := elem.FieldByName(field.Name + "ID"); !ok {
		return []string{}
	}

	preloads := make([][]string, 0)
	if !self.GetType(field.Type, &preloads) {
		return []string{}
	}

	r := make([]string, 0)
	for _, preload := range preloads {
		for _, s := range preload {
			r = append(r, fmt.Sprintf("%v.%v", field.Name, s))
		}
	}

	if len(r) == 0 {
		return []string{field.Name}
	}
	return r
}, func(r *Reflect.Reflect[[]string]) {
	r.Alias = func(elem reflect.Type) []uintptr {
		return []uintptr{
			Reflect.Addr(elem),
			Reflect.Addr(reflect.SliceOf(elem)),
			Reflect.Addr(reflect.SliceOf(reflect.PtrTo(elem))),
		}
	}
})

type Model struct {
	ID uint64 `gorm:"primaryKey;autoIncrement" form:"-" json:"-"`
}

type NoCopy struct{}

// Lock is a no-op used by -copylocks checker from `go vet`.
func (*NoCopy) Lock()   {}
func (*NoCopy) Unlock() {}

type DB struct {
	*gorm.DB
	err    error
	noCopy NoCopy
}

func (db *DB) Close() error {
	if db.DB == nil {
		return nil
	}
	defer func() { db.DB = nil }()
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (db *DB) Error() error {
	return db.err
}

func (db *DB) SetDB(r *gorm.DB) *DB {
	*db = DB{DB: r}
	return db
}

func (db *DB) SetDialector(dialector gorm.Dialector) *DB {
	gdb, _ := gorm.Open(dialector, &gorm.Config{})
	return db.SetDB(gdb)
}

func (db *DB) SetSqlite(file string) *DB {
	return db.SetDialector(sqlite.Open(file))
}

func (db *DB) NoRecord() bool {
	return errors.Is(db.err, gorm.ErrRecordNotFound)
}

func (db *DB) Base(dest any, conds ...any) *DB {
	db.err = db.DB.First(dest, conds...).Error
	return db
}

func (db *DB) Select(dest any, fields []string, conds ...any) *DB {
	db.err = db.DB.Select(fields).First(dest, conds...).Error
	return db
}

func (db *DB) First(x any, conds ...any) bool {
	return !db.Base(x, conds...).NoRecord()
}

func (db *DB) FirstOrCreate(first, create func(), x any, conds ...any) {
	if db.First(x, conds...) {
		if first != nil {
			first()
		}
	} else {
		db.Create(x)
		if create != nil {
			create()
		}
	}
}

func Exists[T any](db *DB, conds ...any) bool {
	return db.First(new(T), conds...)
}

func (db *DB) preloadDB(in any) *gorm.DB {
	r := db.Model(in)
	for _, preload := range Ref.Get(in) {
		for _, s := range preload {
			r.Preload(s)
		}
	}
	return r.Preload(clause.Associations)
}

func (db *DB) Preload(t any, conds ...any) *DB {
	db.err = db.preloadDB(t).First(t, conds...).Error
	return db
}

func (db *DB) Preloads(t any, conds ...any) *DB {
	db.err = db.preloadDB(t).Find(t, conds...).Error
	return db
}

func (db *DB) AutoMigrate(dst ...any) *DB {
	Ref.Init(dst...)
	db.err = db.DB.AutoMigrate(dst...)
	return db
}
