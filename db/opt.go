package db

import (
	"reflect"

	"github.com/RML7/go-sdk/xerrors"
)

// field — внутренний интерфейс, который реализуют Required и Nullable.
// Используется в Fields для полиморфного обхода полей структуры без
// хрупкого сравнения имён типов через reflection.
type field interface {
	isSet() bool
	appendTo(col string, m map[string]any)
}

// Required — поле NOT NULL колонки.
//   - не установлено (set=false) → поле не попадёт в UPDATE
//   - установлено (set=true)    → UPDATE SET col = Value
type Required[T any] struct {
	Value T
	set   bool
}

func NewRequired[T any](v T) Required[T] {
	return Required[T]{Value: v, set: true}
}

func (r *Required[T]) isSet() bool { return r.set }
func (r *Required[T]) appendTo(col string, m map[string]any) {
	m[col] = r.Value
}

// Nullable — поле NULL-able колонки.
//   - не установлено (set=false)      → поле не попадёт в UPDATE
//   - установлено, Value=nil (set=true) → UPDATE SET col = NULL
//   - установлено, Value=&v (set=true)  → UPDATE SET col = v
type Nullable[T any] struct {
	Value *T
	set   bool
}

func NewNullable[T any](v *T) Nullable[T] {
	return Nullable[T]{Value: v, set: true}
}

func (n *Nullable[T]) isSet() bool { return n.set }
func (n *Nullable[T]) appendTo(col string, m map[string]any) {
	if n.Value == nil {
		m[col] = nil
	} else {
		m[col] = *n.Value
	}
}

// fields конвертирует структуру с Optional/Required полями в map[string]any.
// Принимает как значение, так и указатель на структуру.
// Поля без db-тега или с set=false пропускаются.
func fields(updateStruct any) (map[string]any, error) {
	v := reflect.ValueOf(updateStruct)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, xerrors.New("updateStruct must be a struct or pointer to struct")
	}

	// Копируем в адресуемое значение, чтобы можно было брать Addr у полей.
	ptr := reflect.New(v.Type())
	ptr.Elem().Set(v)
	v = ptr.Elem()

	result := make(map[string]any)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fv := v.Field(i)
		ft := t.Field(i)

		if !ft.IsExported() {
			continue
		}

		col := ft.Tag.Get("db")
		if col == "" {
			continue
		}

		if f, ok := fv.Addr().Interface().(field); ok && f.isSet() {
			f.appendTo(col, result)
		}
	}

	return result, nil
}
