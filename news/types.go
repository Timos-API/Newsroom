package news

import "go.mongodb.org/mongo-driver/bson/primitive"

type News struct {
	NewsID    *primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Title     string              `json:"title" bson:"title" validate:"required"`
	Project   primitive.ObjectID  `json:"project" bson:"project" validate:"required"`
	Type      string              `json:"type" bson:"type"`
	Timestamp *int64              `json:"timestamp,omitempty" bson:"timestamp,omitempty"`
	Content   string              `json:"content" bson:"content" validate:"required"`
	Thumbnail string              `json:"thumbnail" bson:"thumbnail" validate:"required,url"`
	Featured  *string             `json:"featured" bson:"featured"`
}

type Exception struct {
	Message string `json:"message,omitempty"`
}
