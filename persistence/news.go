package persistence

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type NewsPersistor struct {
	db *mongo.Collection
}

type News struct {
	NewsID    *primitive.ObjectID `json:"id" bson:"_id"`
	Title     string              `json:"title" bson:"title" keep:"update,create,omitempty" validate:"required,gt=0"`
	Project   primitive.ObjectID  `json:"project" bson:"project" keep:"update,create,omitempty" validate:"required"`
	Type      string              `json:"type" bson:"type" keep:"update,create,omitempty" validate:"required"`
	Timestamp int64               `json:"timestamp" bson:"timestamp"`
	Content   string              `json:"content" bson:"content" keep:"update,create,omitempty" validate:"required"`
	Thumbnail string              `json:"thumbnail" bson:"thumbnail" keep:"update,create,omitempty" validate:"required,url"`
	Featured  string              `json:"featured,omitempty" bson:"featured,omitempty" keep:"update,create,omitempty" validate:"omitempty,url"`
	TweetId   int64               `json:"tweetId" bson:"tweetId"`
}

func NewNewsPersistor(db *mongo.Collection) *NewsPersistor {
	return &NewsPersistor{db}
}

func (p *NewsPersistor) getAll(ctx context.Context, filter *bson.M, limit *int, skip *int) (*[]News, error) {

	opts := options.Find().SetSort(map[string]int{"timestamp": -1})

	if limit != nil {
		opts.SetLimit(int64(*limit))
	}
	if skip != nil {
		opts.SetLimit(int64(*skip))
	}
	if filter == nil {
		filter = &bson.M{}
	}

	cursor, err := p.db.Find(ctx, filter, opts)

	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	allNews := []News{}
	for cursor.Next(ctx) {
		var news News
		cursor.Decode(&news)
		allNews = append(allNews, news)
	}

	if err := cursor.Err(); err != nil {
		return nil, cursor.Err()
	}

	return &allNews, nil
}

func (p *NewsPersistor) GetById(ctx context.Context, id string) (*News, error) {

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	res := p.db.FindOne(ctx, bson.M{"_id": oid})
	if res.Err() != nil {
		return nil, res.Err()
	}

	var news News
	err = res.Decode(&news)

	if err != nil {
		return nil, err
	}

	return &news, nil
}

func (p *NewsPersistor) GetProjectNews(ctx context.Context, projectId string, query *string, limit *int, skip *int) (*[]News, error) {
	oid, err := primitive.ObjectIDFromHex(projectId)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"project": oid}

	if query != nil {
		regex := primitive.Regex{Pattern: *query, Options: "i"}
		filter = bson.M{"$and": []bson.M{
			filter,
			{"$or": []bson.M{
				{"_id": regex}, {"title": regex}, {"type": regex}, {"content": regex},
			}},
		}}
	}

	return p.getAll(ctx, &filter, limit, skip)
}

func (p *NewsPersistor) GetByQuery(ctx context.Context, query *string, limit *int, skip *int) (*[]News, error) {

	filter := bson.M{}

	if query != nil {
		regex := primitive.Regex{Pattern: *query, Options: "i"}
		filter = bson.M{"$or": []bson.M{
			{"_id": regex}, {"title": regex}, {"type": regex}, {"content": regex},
		}}
	}

	return p.getAll(ctx, &filter, limit, skip)
}

func (p *NewsPersistor) GetFeatured(ctx context.Context) (*[]News, error) {
	filter := bson.M{"$and": []bson.M{
		{"featured": bson.M{"$exists": "true"}},
		{"featured": bson.M{"$ne": nil}},
	}}

	return p.getAll(ctx, &filter, nil, nil)
}

func (p *NewsPersistor) GetProjects(ctx context.Context) (*[]interface{}, error) {
	res, err := p.db.Distinct(ctx, "project", bson.M{})
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (p *NewsPersistor) PostNews(ctx context.Context, document interface{}) (*News, error) {

	res, err := p.db.InsertOne(ctx, document)
	if err != nil {
		return nil, err
	}

	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		return p.GetById(ctx, oid.Hex())
	}
	return nil, errors.New("something went wrong while inserting")
}

func (p *NewsPersistor) PatchNews(ctx context.Context, id string, document interface{}) (*News, error) {

	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	res := p.db.FindOneAndUpdate(ctx, bson.M{"_id": oid}, bson.M{"$set": document}, options.FindOneAndUpdate().SetReturnDocument(options.After))
	if res.Err() != nil {
		return nil, res.Err()
	}

	var news News
	err = res.Decode(&news)

	if err != nil {
		return nil, err
	}

	return &news, nil
}

func (p *NewsPersistor) DeleteNews(ctx context.Context, id string) (bool, int64, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return false, 0, err
	}

	news, err := p.GetById(ctx, id)
	if err != nil {
		return false, 0, err
	}

	res, err := p.db.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return false, 0, err
	}

	if res.DeletedCount == 0 {
		return false, 0, errors.New("Couldn't delete news with id " + id)
	}

	return true, news.TweetId, nil
}

func (p *NewsPersistor) SetTweetId(ctx context.Context, newsId string, tweetId int64) (*News, error) {
	oid, err := primitive.ObjectIDFromHex(newsId)
	if err != nil {
		return nil, err
	}

	res := p.db.FindOneAndUpdate(ctx, bson.M{"_id": oid}, bson.M{"$set": bson.M{"tweetId": tweetId}}, options.FindOneAndUpdate().SetReturnDocument(options.After))
	if res.Err() != nil {
		return nil, res.Err()
	}

	var news News
	err = res.Decode(&news)

	if err != nil {
		return nil, err
	}

	return &news, nil
}
