package article

import (
	"context"
	"errors"
	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
	"time"
)

type MongoDBArticleDAO struct {
	node    *snowflake.Node
	col     *mongo.Collection
	liveCol *mongo.Collection
}

func (m *MongoDBArticleDAO) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]Article, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoDBArticleDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	filter := bson.M{"id": id}
	var art Article
	err := m.liveCol.FindOne(ctx, filter).Decode(&art)
	return PublishedArticle{
		art,
	}, err
}

func (m *MongoDBArticleDAO) GetById(ctx context.Context, id int64) (Article, error) {
	filter := bson.M{"id": id}
	var art Article
	err := m.col.FindOne(ctx, filter).Decode(&art)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return Article{}, errors.New("article not found")
	}
	return art, err
}

func (m *MongoDBArticleDAO) GetByAuthor(ctx context.Context, uid int64, limit int, offset int) ([]Article, error) {
	filter := bson.M{"author_id": uid}
	findOptions := options.Find().SetLimit(int64(limit)).SetSkip(int64(offset))
	cursor, err := m.col.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err = cursor.Close(ctx)
		if err != nil {
			zap.L().Error("close cursor error", zap.Error(err))
		}
	}(cursor, ctx)
	var articles []Article
	for cursor.Next(ctx) {
		var article Article
		err = cursor.Decode(&article)
		if err != nil {
			return nil, err
		}
		articles = append(articles, article)
	}
	return articles, nil
}

func (m *MongoDBArticleDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	art.Id = m.node.Generate().Int64()
	_, err := m.col.InsertOne(ctx, &art)
	return art.Id, err
}

func (m *MongoDBArticleDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	art.Utime = now
	filter := bson.D{bson.E{"id", art.Id}, bson.E{"author_id", art.AuthorId}}
	set := bson.D{bson.E{"$set", bson.M{
		"title":   art.Title,
		"content": art.Content,
		"status":  art.Status,
		"utime":   art.Utime,
	}}}

	res, err := m.col.UpdateOne(ctx, filter, set)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("ID 不对或创作者不对")
	}
	return nil
}

func (m *MongoDBArticleDAO) Sync(ctx context.Context, art Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	if id > 0 {
		err = m.UpdateById(ctx, art)
	} else {
		id, err = m.Insert(ctx, art)
	}
	if err != nil {
		return 0, err
	}

	// 同步线上库
	art.Id = id
	now := time.Now().UnixMilli()
	art.Utime = now
	filter := bson.D{bson.E{"id", art.Id},
		bson.E{"author_id", art.AuthorId}}
	set := bson.D{bson.E{"$set", art},
		bson.E{"$setOnInsert", bson.D{bson.E{"ctime", now}}}}
	_, err = m.liveCol.UpdateOne(ctx, filter, set, options.Update().SetUpsert(true))
	return id, err
}

func (m *MongoDBArticleDAO) SyncStatus(ctx context.Context, id int64, authorId int64, status uint8) error {
	filter := bson.D{bson.E{"id", id}, bson.E{"author_id", authorId}}
	set := bson.D{bson.E{"$set", bson.D{bson.E{"status", status}}}}
	// upsert 语义
	res, err := m.col.UpdateOne(ctx, filter, set, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}
	if res.MatchedCount != 1 {
		return errors.New("ID 或者创作者不对")
	}
	_, err = m.liveCol.UpdateOne(ctx, filter, set)
	return err
}

func NewMongoDBArticleDAO(mdb *mongo.Database, node *snowflake.Node) ArticleDAO {
	return &MongoDBArticleDAO{
		col:     mdb.Collection("articles"),
		liveCol: mdb.Collection("published_articles"),
		node:    node,
	}
}
