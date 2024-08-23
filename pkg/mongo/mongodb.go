package mongo

import (
	"context"
	"log"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type dbOperation string

const (
	Get              dbOperation = "Get"
	GetOne           dbOperation = "GetOne"
	GetSorted        dbOperation = "GetSorted"
	GetAggregate     dbOperation = "GetAggregate"
	Insert           dbOperation = "Insert"
	InsertMany       dbOperation = "InsertMany"
	Upsert           dbOperation = "Upsert"
	Replace          dbOperation = "Replace"
	Count            dbOperation = "Count"
	GetCursor        dbOperation = "GetCursor"
	BulkUpsert       dbOperation = "BulkUpsert"
	BulkWrite        dbOperation = "BulkWrite"
	FindOneAndUpdate dbOperation = "FindOneAndUpdate"
	DeleteMany       dbOperation = "DeleteMany"
	GetDistinct      dbOperation = "GetDistinct"
	Delete           dbOperation = "Delete"
)

func (c dbOperation) String() string {
	return string(c)
}

type MongoDB struct {
	dbName string
	client *mongo.Client
}

func NewMongoDB(ctx context.Context, dbName, clientURI string) *MongoDB {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(clientURI))
	if err != nil {
		log.Println("Err. mongo.Connect client:", err.Error())
	}
	return &MongoDB{dbName: dbName, client: client}
}

// InsertMany ...
func (db *MongoDB) InsertMany(ctx context.Context, collection string, data []interface{}) error {
	// defer newrelic.StartMongoDBDataSegment(ctx, collection, InsertMany.String()).End()
	c := db.client.Database(db.dbName).Collection(collection)
	// https://docs.mongodb.com/manual/reference/method/db.collection.insertMany/#behaviors
	ordered := false
	_, err := c.InsertMany(ctx, data, &options.InsertManyOptions{Ordered: &ordered})
	if err != nil {
		log.Println("Err. c.InsertMany,", err.Error())
	}
	return err
}

// Insert ...
func (db *MongoDB) Insert(ctx context.Context, collection string, data interface{}) error {
	// defer newrelic.StartMongoDBDataSegment(ctx, collection, Insert.String()).End()
	c := db.client.Database(db.dbName).Collection(collection)
	_, err := c.InsertOne(ctx, data)
	return err
}

func (db *MongoDB) BulkWrite(ctx context.Context, collection string, data []mongo.WriteModel) (*mongo.BulkWriteResult, error) {
	// defer newrelic.StartMongoDBDataSegment(ctx, collection, BulkWrite.String()).End()
	c := db.client.Database(db.dbName).Collection(collection)
	return c.BulkWrite(ctx, data)
}

func convertMapToBsonMapArray(m map[string]interface{}) []bson.M {
	mapArray := []bson.M{}
	for k, v := range m {
		mapArray = append(mapArray, bson.M{k: v})
	}
	return mapArray
}

func getQueryMapFromFilters(filters Filters) map[string]interface{} { // nolint: ignore-recursion
	m := make(map[string]interface{})
	for _, each := range filters {
		switch each.Operator {
		case EQUAL, IN_ARRAY:
			m[each.Key] = each.Value
		case IN:
			in := bson.M{"$in": each.Value}
			m[each.Key] = in
		case GREATER_THAN_EQUAL:
			gte := bson.M{"$gte": each.Value}
			m[each.Key] = gte
		case LESS_THAN_EQUAL:
			lte := bson.M{"$lte": each.Value}
			m[each.Key] = lte
		case GREATER_THAN:
			gt := bson.M{"$gt": each.Value}
			m[each.Key] = gt
		case BETWEEN:
			v := each.Value.(Range)
			b := bson.M{"$gte": v.Left, "$lte": v.Right}
			m[each.Key] = b
		case OR:
			subFilters := each.Value.(Filters)
			subQueryMap := getQueryMapFromFilters(subFilters)
			m[each.Key] = convertMapToBsonMapArray(subQueryMap)
		case LESS_THAN:
			lt := bson.M{"$lt": each.Value}
			m[each.Key] = lt
		case NOT_IN:
			nin := bson.M{"$nin": each.Value}
			m[each.Key] = nin
		case ALL:
			m[each.Key] = bson.M{"$all": each.Value}
		}
	}

	return m
}

// Get ...
func (db *MongoDB) Get(ctx context.Context, collection string, filters Filters, offset string, limit int64, results interface{}) (string, error) {
	// defer newrelic.StartMongoDBDataSegment(ctx, collection, Get.String()).End()
	var err error
	next := ""
	c := db.client.Database(db.dbName).Collection(collection)

	var skip int
	findOptions := options.Find()
	if limit != 0 {
		findOptions.SetLimit(int64(limit))
		if offset != "" {
			skip, err = strconv.Atoi(offset)
			if err != nil {
				log.Println("Err. converting offset -> skip:", err.Error())
				return "", err
			}
			findOptions.SetSkip(int64(skip))
		}
	}

	m := getQueryMapFromFilters(filters)
	log.Printf("Filters: %+v \n", m)
	cur, err := c.Find(ctx, bson.M(m), findOptions)
	if err != nil {
		log.Printf("Err. Mongo Find: %s \n", err)
		return next, err
	}

	err = cur.All(ctx, results)
	if err != nil {
		if err != mongo.ErrNilCursor {
			log.Println("Err. Mongo Find cur.All():", err)
		} else {
			log.Println("Err. Mongo Find cur.All():", err)
		}
		return next, err
	}

	// Close the cursor once finished
	cur.Close(ctx)

	if limit == 0 {
		log.Println("Empty limit value: skipping docs count")
		return next, nil
	}

	var total int64
	total, err = c.CountDocuments(ctx, bson.M(m))
	if err != nil {
		log.Println("Err. Mongo Find cur.Count():", err)
		return next, err
	}

	if int64(limit)+int64(skip) < total {
		next = strconv.Itoa(int(limit) + skip)
	}
	return next, nil
}

// GetOne ...
func (db *MongoDB) GetOne(ctx context.Context, collection string, filters Filters, projections []Projection, result interface{}) error {
	// defer newrelic.StartMongoDBDataSegment(ctx, collection, GetOne.String()).End()
	c := db.client.Database(db.dbName).Collection(collection)

	m := getQueryMapFromFilters(filters)
	findOptions := options.FindOne()
	if projections != nil {
		findOptions.SetProjection(mongoProjections(projections))
	}
	r := c.FindOne(ctx, bson.M(m), findOptions)
	err := r.Decode(result)
	if err != nil {
		return err
	}
	return nil
}

// Update ...
func (db *MongoDB) Update(ctx context.Context, collection string, filters Filters, updates Updates) error {
	// defer newrelic.StartMongoDBDataSegment(ctx, collection, "Update").End()
	c := db.client.Database(db.dbName).Collection(collection)

	filter := getQueryMapFromFilters(filters)

	docUpdates := mongoUpdates(updates)
	update := bson.M(docUpdates)
	log.Println("filters:", filter)
	log.Println("updates:", update)
	_, err := c.UpdateMany(ctx, filter, update)
	if err != nil {
		log.Println("Err. UpdateMany:", err.Error())
		return err
	}
	return nil
}

// Update ...
func (db *MongoDB) Upsert(ctx context.Context, collection string, filters Filters, updates Updates) error {
	// defer newrelic.StartMongoDBDataSegment(ctx, collection, Upsert.String()).End()
	c := db.client.Database(db.dbName).Collection(collection)
	op := options.UpdateOptions{}
	op = *op.SetUpsert(true)
	fm := make(map[string]interface{})
	for _, each := range filters {
		switch each.Operator {
		case EQUAL:
			fm[each.Key] = each.Value
		case IN:
			in := bson.M{"$in": bson.A{each.Value}}
			fm[each.Key] = in
		case GREATER_THAN_EQUAL:
			gte := bson.M{"$gte": bson.A{each.Value}}
			fm[each.Key] = gte
		case LESS_THAN:
			lt := bson.M{"$lt": bson.A{each.Value}}
			fm[each.Key] = lt
		default:
			log.Printf("Operator[%d] not implemented in mongo.Upsert", each.Operator)
		}
	}
	filter := bson.M(fm)

	um := make(map[string]interface{})
	for _, each := range updates {
		um[each.Key] = each.Value
	}
	update := bson.M{"$set": bson.M(um)}
	_, err := c.UpdateOne(ctx, filter, update, &op)
	if err != nil {
		log.Println("Err. UpdateOne:", err.Error())
		return err
	}
	return nil
}

// Replace ...
func (db *MongoDB) Replace(ctx context.Context, collection string, filters Filters, data interface{}) error {
	// defer newrelic.StartMongoDBDataSegment(ctx, collection, Replace.String()).End()
	fm := make(map[string]interface{})
	for _, each := range filters {
		switch each.Operator {
		case EQUAL:
			fm[each.Key] = each.Value
		case IN:
			in := bson.M{"$in": bson.A{each.Value}}
			fm[each.Key] = in
		}
	}
	filter := bson.M(fm)
	c := db.client.Database(db.dbName).Collection(collection)
	_, err := c.ReplaceOne(ctx, filter, data)
	if err != nil {
		log.Println("Err. c.ReplaceOne,", err.Error())
	}
	return nil
}

// Count ...
func (db *MongoDB) Count(ctx context.Context, collection string, filters Filters) (int, error) {
	// defer newrelic.StartMongoDBDataSegment(ctx, collection, Count.String()).End()
	var err error
	c := db.client.Database(db.dbName).Collection(collection)

	m := getQueryMapFromFilters(filters)
	r, err := c.CountDocuments(ctx, bson.M(m), options.Count())
	if err != nil {
		log.Println("Count Documents Err:", err.Error())
		return 0, err
	}

	return int(r), err
}

// GetSorted ...
// eg. 	sortKeys := make(db.SortKeys)
//
//	sortKeys["created_at"] = db.DSC
//	sortKeys["name"] = db.ASC
func (db *MongoDB) GetSorted(ctx context.Context, collection string, filters Filters, offset string, limit int64, sortKeys []SortKey, results interface{}) (string, error) {
	// defer newrelic.StartMongoDBDataSegment(ctx, collection, GetSorted.String()).End()
	var err error
	next := ""
	c := db.client.Database(db.dbName).Collection(collection)
	var skip int
	findOptions := options.Find()
	if sortKeys != nil {
		sort := make(map[string]Order)
		for _, each := range sortKeys {
			sort[each.Key] = each.Order
		}
		findOptions.SetSort(sort)
	}
	if limit != 0 {
		findOptions.SetLimit(int64(limit))
		if offset != "" {
			skip, err = strconv.Atoi(offset)
			if err != nil {
				log.Println("Err. converting offset -> skip:", err.Error())
				return "", err
			}
			findOptions.SetSkip(int64(skip))
		}
	}

	m := getQueryMapFromFilters(filters)
	log.Printf("Filters: %+v \n", m)
	cur, err := c.Find(ctx, bson.M(m), findOptions)
	if err != nil {
		log.Printf("Err. Mongo Find: %s \n", err)
		return next, err
	}

	err = cur.All(ctx, results)
	if err != nil {
		if err != mongo.ErrNilCursor {
			log.Println("Err. Mongo Find cur.All():", err)
		} else {
			log.Println("Err. Mongo Find cur.All():", err)
		}
		return next, err
	}

	// Close the cursor once finished
	err = cur.Close(ctx)
	if err != nil {
		log.Println("Err. Mongo Find cur.Close():", err)
	}

	if limit == 0 {
		log.Println("Empty limit value: skipping docs count")
		return next, nil
	}

	var total int64
	total, err = c.CountDocuments(ctx, bson.M(m))
	if err != nil {
		log.Println("Err. Mongo CountDocuments():", err)
		return next, err
	}

	if int64(limit)+int64(skip) < total {
		next = strconv.Itoa(int(limit) + skip)
	}
	return next, nil
}

// GetCursor ...
func (db *MongoDB) GetCursor(ctx context.Context, collection string, filters Filters, sortKeys []SortKey, projections interface{}) (interface{}, error) {
	// defer newrelic.StartMongoDBDataSegment(ctx, collection, GetCursor.String()).End()
	var err error
	c := db.client.Database(db.dbName).Collection(collection)
	findOptions := options.Find().SetProjection(projections)
	if sortKeys != nil {
		sort := make(map[string]Order)
		for _, each := range sortKeys {
			sort[each.Key] = each.Order
		}
		findOptions.SetSort(sort)
	}
	m := getQueryMapFromFilters(filters)
	cur, err := c.Find(ctx, bson.M(m), findOptions)
	if err != nil {
		log.Printf("Err. Mongo Find: %s \n", err)
		return nil, err
	}

	return cur, nil
}

func (db *MongoDB) GetAggregate(ctx context.Context, collection string, filters Filters, groupKeys GroupKeys, aggregateKeys AggregateKeys) (interface{}, error) {
	// defer newrelic.StartMongoDBDataSegment(ctx, collection, GetAggregate.String()).End()
	c := db.client.Database(db.dbName).Collection(collection)
	// Filter results in match stage
	m := getQueryMapFromFilters(filters)
	matchStage := bson.D{{
		Key:   "$match",
		Value: bson.M(m)}}
	// Group by certain fields and then aggregate in group stage
	groupByMap := bson.D{}
	for _, each := range groupKeys {
		groupByMap = append(groupByMap, bson.E{
			Key:   each.Key,
			Value: each.Value})
	}
	aggregateMap := bson.D{}
	aggregateMap = append(aggregateMap, bson.E{
		Key:   "_id",
		Value: groupByMap})
	for _, each := range aggregateKeys {
		switch each.Operator {
		case SUM:
			aggregateMap = append(aggregateMap, bson.E{
				Key:   each.Key,
				Value: bson.M{"$sum": each.Value}})
		case MIN:
			aggregateMap = append(aggregateMap, bson.E{
				Key:   each.Key,
				Value: bson.M{"$min": each.Value}})
		case MAX:
			aggregateMap = append(aggregateMap, bson.E{
				Key:   each.Key,
				Value: bson.M{"$max": each.Value}})
		default:
			log.Printf("Operator[%d] not implemented in mongo.GetAggregate", each.Operator)
		}
	}
	groupStage := bson.D{{
		Key:   "$group",
		Value: aggregateMap}}
	cur, err := c.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage})
	if err != nil {
		log.Printf("Err. Mongo Aggregate: %s \n", err)
		return nil, err
	}
	return cur, nil
}

func (db *MongoDB) Delete(ctx context.Context, collection string, filters Filters) (deletedCount int64, err error) {
	// defer newrelic.StartMongoDBDataSegment(ctx, collection, Delete.String()).End()
	c := db.client.Database(db.dbName).Collection(collection)
	m := getQueryMapFromFilters(filters)
	res, err := c.DeleteOne(ctx, bson.M(m))
	return res.DeletedCount, err
}

// GetReaderDB
func (db *MongoDB) GetReaderDB() (interface{}, error) {
	return db.client.Database(db.dbName), nil
}

// GetWriterDB
func (db *MongoDB) GetWriterDB() (interface{}, error) {
	return db.client.Database(db.dbName), nil
}

// FindOneAndUpdate
func (db *MongoDB) FindOneAndUpdate(ctx context.Context, collection string, filters Filters, updates Updates, result interface{}) error {
	// defer newrelic.StartMongoDBDataSegment(ctx, collection, FindOneAndUpdate.String()).End()
	c := db.client.Database(db.dbName).Collection(collection)
	f := getQueryMapFromFilters(filters)

	docUpdates := mongoUpdates(updates)
	update := bson.M(docUpdates)

	returnNewDocType := options.After
	op := &options.FindOneAndUpdateOptions{
		ReturnDocument: &returnNewDocType,
	}
	r := c.FindOneAndUpdate(ctx, bson.M(f), update, op)
	err := r.Decode(result)
	if err != nil {
		return err
	}
	return nil
}

func (db *MongoDB) DeleteMany(ctx context.Context, collection string, filters Filters) (deletedCount int64, err error) {
	// defer newrelic.StartMongoDBDataSegment(ctx, collection, DeleteMany.String()).End()
	c := db.client.Database(db.dbName).Collection(collection)
	m := getQueryMapFromFilters(filters)
	res, err := c.DeleteMany(ctx, bson.M(m))
	return res.DeletedCount, err
}

func (db *MongoDB) Distinct(ctx context.Context, collection string, fieldName string, filters Filters) ([]interface{}, error) {
	// defer newrelic.StartMongoDBDataSegment(ctx, collection, GetDistinct.String()).End()
	c := db.client.Database(db.dbName).Collection(collection)
	m := getQueryMapFromFilters(filters)
	return c.Distinct(ctx, fieldName, bson.M(m))
}

func mongoUpdates(updates Updates) map[string]interface{} {
	docUpdates := make(map[string]interface{})
	for _, each := range updates {
		if each.UpdateOperator == PUSH {
			exist, ok := docUpdates[each.UpdateOperator.ToString()]
			if !ok {
				docUpdates[each.UpdateOperator.ToString()] = bson.M{
					each.Key: bson.M{"$each": each.Value},
				}
			} else {
				exist.(bson.M)[each.Key] = bson.M{"$each": each.Value}
			}
		} else {
			exist, ok := docUpdates[each.UpdateOperator.ToString()]
			if !ok {
				docUpdates[each.UpdateOperator.ToString()] = bson.M{
					each.Key: each.Value,
				}
			} else {
				exist.(bson.M)[each.Key] = each.Value
			}
		}
	}
	return docUpdates
}
