package article

import (
	"context"
	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type MongoDBArticleDAO struct {
	nod     *snowflake.Node
	col     *mongo.Collection
	liveCol *mongo.Collection
}

func (m *MongoDBArticleDAO) Insert(ctx context.Context, art Article) (int64, error) {
	panic("implement me")
}

func (m *MongoDBArticleDAO) UpdateById(ctx context.Context, art Article) error {
	//TODO implement me
	panic("implement me")
}

func (m *MongoDBArticleDAO) Sync(ctx context.Context, art Article) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoDBArticleDAO) SyncStatus(ctx *gin.Context, id int64, authorId int64, status uint8) error {
	//TODO implement me
	panic("implement me")
}

func NewMongoDBArticleDAO(mdb *mongo.Database, node *snowflake.Node) ArticleDAO {
	return &MongoDBArticleDAO{
		col:     mdb.Collection("articles"),
		liveCol: mdb.Collection("published_articles"),
		nod:     node,
	}
}
