package postgres

import (
	"context"
	"database/sql/driver"
	"fmt"

	"github.com/evertonbiviatello/go-commons/store"
)

const Value = "$#"

// Table is the query builder table representation.
type Table[T any] struct {
	// Schema to use if you want to hard code it
	Schema string
	// Table name
	Table string
	// The fields you wish to reference for select, insert and update operations
	Fields []*Field[T]
	// Additional joins when fetching data from the table
	Joins string

	// Selector is a tool for fetching multiple rows from a table, using
	// queryp to filter results.
	//Selector[T]
	//
	//// Scanner is used
	//Scanner Scanner[T]

	// This is a callback that is used after fetching a row of data before
	// returning it.
	PostProcessRecord func(*T) error

	// The select portion of the query for just the fields in this table.
	// It should not include the SELECT keyword, just comma separated fields.
	// If this is not specified it will be built automatically based on the fields
	// provided above.
	SelectFields string
	// Additional fields you wish to select from the main query. Generally
	// associated with the Joins but could be anything. Just provide comma
	// separated field statements.
	SelectAdditionalFields string
	// The query used to get a record by ID. If not specified will be auto
	// generated.
	GetByIDQuery string
	// The query used to delete a record by ID. If not specified will be
	// auto generated.
	DeleteByIDQuery string
	// The query used to insert a record. If not specified will be auto generated.
	InsertQuery string
	// The query used to update a record. If not specified will be auto generated by ID.
	UpdateQuery string
	// The query used to upsert a record. If not specified will be auto generated by ID.
	UpsertQuery string
}

// Field is the field representation for each field in the table.
type Field[T any] struct {
	// The field name
	Name string
	// Is this field part of the record ID (a primary key). You can have
	// multiple ID fields on a record.
	ID bool
	// Should this field be used on a select statement. (used for auto
	// generating select statements.)
	Select bool
	// The value to use when inserting this field into the database. If you want
	// to use a positional argument, use the `Value` constant.
	Insert string
	// The value to use when updating this field in the database. If you want
	// to use a positional argument, use the `Value` constant.
	Update string
	// This function is used to fetch the value for insert or update from a record.
	Value func(*T) (driver.Value, error)
	// This is used to determine the value that should be returned if the
	// value is being returned in a COALESCED way. For example, if you left join this
	// table and there is no value, this would be the value returned if you use the
	// GenerateAdditionalFields(coalesce=true) to generate the AdditionalFields
	// string
	NullVal any
}

// GetByID fetches a single record by ID(s)
func (t *Table[T]) GetByID(ctx context.Context, db DB, ids ...interface{}) (*T, error) {
	var record = new(T)
	err := db.GetContext(ctx, record, t.GetByIDQuery, ids...)
	if err != nil {
		return nil, WrapError(err)
	}
	if t.PostProcessRecord != nil {
		if err := t.PostProcessRecord(record); err != nil {
			return nil, fmt.Errorf("post process record error: %w", err)
		}
	}
	return record, nil
}

// DeleteByID deletes a single record by ID(s)
func (t *Table[T]) DeleteByID(ctx context.Context, db DB, ids ...interface{}) error {
	result, err := db.ExecContext(ctx, t.DeleteByIDQuery, ids...)
	if err != nil {
		return WrapError(err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return WrapError(err)
	}
	if rowsAffected == 0 {
		return store.ErrNotFound
	}
	return nil
}

// Insert inserts a record
func (t *Table[T]) Insert(ctx context.Context, db DB, record *T, opts ...QueryOption) error {

	queryOptions := DefaultQueryOptions
	for _, opt := range opts {
		if err := opt(&queryOptions); err != nil {
			return fmt.Errorf("query option error: %w", err)
		}
	}

	var args []any
	for _, field := range t.Fields {
		if field.Value != nil {
			arg, err := field.Value(record)
			if err != nil {
				return fmt.Errorf("could not get arg for field %s: %w", field.Name, err)
			}
			args = append(args, arg)
		}
	}

	if queryOptions.IgnoreReturn {
		if _, err := db.ExecContext(ctx, t.InsertQuery, args...); err != nil {
			return WrapError(err)
		}
	} else {
		err := db.GetContext(ctx, record, t.InsertQuery, args...)
		if err != nil {
			return WrapError(err)
		}
		if t.PostProcessRecord != nil {
			if err := t.PostProcessRecord(record); err != nil {
				return fmt.Errorf("post process record error: %w", err)
			}
		}
	}
	return nil

}

// Updates a record using the Update query
func (t *Table[T]) Update(ctx context.Context, db DB, record *T, opts ...QueryOption) error {

	queryOptions := DefaultQueryOptions
	for _, opt := range opts {
		if err := opt(&queryOptions); err != nil {
			return fmt.Errorf("query option error: %w", err)
		}
	}

	var args []any
	for _, field := range t.Fields {
		if field.Value != nil {
			arg, err := field.Value(record)
			if err != nil {
				return fmt.Errorf("could not get arg for field %s: %w", field.Name, err)
			}
			args = append(args, arg)
		}
	}

	if queryOptions.IgnoreReturn {
		if _, err := db.ExecContext(ctx, t.UpdateQuery, args...); err != nil {
			return WrapError(err)
		}
	} else {
		err := db.GetContext(ctx, record, t.UpdateQuery, args...)
		if err != nil {
			return WrapError(err)
		}
		if t.PostProcessRecord != nil {
			if err := t.PostProcessRecord(record); err != nil {
				return fmt.Errorf("post process record error: %w", err)
			}
		}
	}
	return nil

}

// Upsert a record using the Upsert query.
func (t *Table[T]) Upsert(ctx context.Context, db DB, record *T, opts ...QueryOption) error {

	queryOptions := DefaultQueryOptions
	for _, opt := range opts {
		if err := opt(&queryOptions); err != nil {
			return fmt.Errorf("query option error: %w", err)
		}
	}

	var args []any
	for _, field := range t.Fields {
		if field.Value != nil {
			arg, err := field.Value(record)
			if err != nil {
				return fmt.Errorf("could not get arg for field %s: %w", field.Name, err)
			}
			args = append(args, arg)
		}
	}

	if queryOptions.IgnoreReturn {
		if _, err := db.ExecContext(ctx, t.UpsertQuery, args...); err != nil {
			return WrapError(err)
		}
	} else {
		err := db.GetContext(ctx, record, t.UpsertQuery, args...)
		if err != nil {
			return WrapError(err)
		}
		if t.PostProcessRecord != nil {
			if err := t.PostProcessRecord(record); err != nil {
				return fmt.Errorf("post process record error: %w", err)
			}
		}
	}
	return nil

}

// GetByQuery fetches a single record by the given query and values
func (t *Table[T]) GetByQuery(ctx context.Context, db DB, query string, values ...interface{}) (*T, error) {
	var record = new(T)
	err := db.GetContext(ctx, record, query, values...)
	if err != nil {
		return nil, WrapError(err)
	}
	if t.PostProcessRecord != nil {
		if err := t.PostProcessRecord(record); err != nil {
			return nil, fmt.Errorf("post process record error: %w", err)
		}
	}
	return record, nil
}
