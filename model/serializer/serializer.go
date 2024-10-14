package serializer

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"

	"gorm.io/gorm/schema"
)

type InterfaceSerializer[T any, V driver.Value] struct {
	Scanner func(V) T
	Valuer  func(T) V
}

// Scan implements serializer interface
func (i InterfaceSerializer[T, V]) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) error {
	if dbValue == nil {
		return nil
	}
	if v, ok := dbValue.(V); ok {
		return field.Set(ctx, dst, i.Scanner(v))
	}
	return fmt.Errorf("serializer: invalid value type %#v", dbValue)
}

// Value implements serializer interface
func (i InterfaceSerializer[T, V]) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	if fieldValue == nil {
		return nil, nil
	}
	if t, ok := fieldValue.(T); ok {
		return i.Valuer(t), nil
	}
	return nil, fmt.Errorf("serializer: invalid field type %#v", fieldValue)
}

var ErrorSerializer schema.SerializerInterface = InterfaceSerializer[error, string]{
	Scanner: errors.New,
	Valuer:  error.Error,
}
