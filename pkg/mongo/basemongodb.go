package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type BaseMongoDBClient interface {
	Insert(ctx context.Context, table string, data interface{}) (err error)
	InsertMany(ctx context.Context, table string, data []interface{}) (err error)
	Get(ctx context.Context, table string, filters Filters, offset string, limit int64, result interface{}) (next string, err error)
	GetSorted(ctx context.Context, table string, filters Filters, offset string, limit int64, sortKey []SortKey, results interface{}) (next string, err error)
	GetOne(ctx context.Context, table string, filters Filters, result interface{}) (err error)
	Update(ctx context.Context, table string, filters Filters, updates Updates) (err error)
	BulkWrite(ctx context.Context, table string, data []mongo.WriteModel) (*mongo.BulkWriteResult, error)
	Upsert(ctx context.Context, table string, filters Filters, updates Updates) (err error)
	Replace(ctx context.Context, table string, filters Filters, data interface{}) (err error)
	Count(ctx context.Context, table string, filters Filters) (count int, err error)
	Delete(ctx context.Context, table string, filters Filters) (deletedCount int64, err error)
	GetReaderDB() (db interface{}, err error)
	GetWriterDB() (db interface{}, err error)
	FindOneAndUpdate(ctx context.Context, table string, filters Filters, updates Updates, result interface{}) (err error)
	GetCursor(ctx context.Context, table string, filters Filters, sortKeys []SortKey, projections interface{}) (cursor interface{}, err error)
	GetAggregate(ctx context.Context, collection string, filters Filters, groupKeys GroupKeys, aggregateKeys AggregateKeys) (cursor interface{}, err error)
	DeleteMany(ctx context.Context, collection string, filters Filters) (deletedCount int64, err error)
	Distinct(ctx context.Context, collection string, fieldName string, filters Filters) (result []interface{}, err error)
}

// DataType datatypes used
type DataType uint16

// Operator operators used
type Operator uint16

// UpdateOperator operators used
type UpdateOperator string

func (u UpdateOperator) ToString() string {
	return string(u)
}

var NoItemFound = mongo.ErrNoDocuments

const (
	INT64        DataType = 1
	STRING       DataType = 2
	STRING_ARRAY DataType = 3
	INT          DataType = 4
	INT64_ARRAY  DataType = 5
	UINT64_ARRAY DataType = 6
	FLOAT64      DataType = 7
	TIME         DataType = 8
	BOOL         DataType = 9
)

const (
	EQUAL              Operator = 0
	IN                 Operator = 1
	GREATER_THAN_EQUAL Operator = 2
	LESS_THAN_EQUAL    Operator = 3
	GREATER_THAN       Operator = 4
	BETWEEN            Operator = 5
	IN_ARRAY           Operator = 6
	OR                 Operator = 7
	LESS_THAN          Operator = 8
	NOT_IN             Operator = 9
	ALL                Operator = 10
)

const (
	SET      UpdateOperator = "$set"
	UNSET    UpdateOperator = "$unset"
	PUSH     UpdateOperator = "$push"
	PULL     UpdateOperator = "$pull"
	INC      UpdateOperator = "$inc"
	ARRAY_IN UpdateOperator = "$in"
	NOT      UpdateOperator = "$not"
)

type Range struct {
	Left  interface{}
	Right interface{}
}

// Filter used to filter on fetch data
type Filter struct {
	Key      string
	Value    interface{}
	Type     DataType
	Operator Operator
}

// Filters : array of Filter
type Filters []Filter

// Append append function
func (f *Filters) Append(filters ...Filter) {
	*f = append(*f, filters...)
}

// Update used to Update data
type Update struct {
	Key            string
	Value          interface{}
	Type           DataType
	UpdateOperator UpdateOperator
}

// Updates : array of Update
type Updates []Update

// Append append function
func (u *Updates) Append(updates ...Update) {
	*u = append(*u, updates...)
}

type Order int8

var (
	ASC Order = 1
	DSC Order = -1
)

// SortKey is used to sort the result
type SortKey struct {
	Key   string
	Order Order
}

// GroupKeys : array of GroupKey
type GroupKeys []GroupKey

// GroupKey represents a field by which we group
type GroupKey struct {
	Key   string
	Value interface{}
}

// Append append function
func (g *GroupKeys) Append(groupKeys ...GroupKey) {
	*g = append(*g, groupKeys...)
}

// AggregateKeys : array of AggregateKey
type AggregateKeys []AggregateKey

// AggregateOperator operators used
type AggregateOperator uint16

const (
	SUM AggregateOperator = 1
	MIN AggregateOperator = 2
	MAX AggregateOperator = 3
)

// AggregateKey represents a fiedl which needs to be aggregated
type AggregateKey struct {
	Key      string
	Operator AggregateOperator
	Value    string
}

// Append append function
func (a *AggregateKeys) Append(aggregateKeys ...AggregateKey) {
	*a = append(*a, aggregateKeys...)
}

type BulkWriteResult struct {
	mongo.BulkWriteResult
	Err error
}

type Projection struct {
	Key   string
	Value interface{}
}

func mongoProjections(projections []Projection) interface{} {
	updates := bson.D{}
	for _, projection := range projections {
		updates = append(updates, bson.E{
			Key:   projection.Key,
			Value: projection.Value,
		})
	}
	return updates
}
