package mongo

import (
	"context"

	"cinema-ticket/backend/internal/domain"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type AuditLogRepo struct{ col *mongo.Collection }

func NewAuditLogRepo(db *mongo.Database) *AuditLogRepo {
	return &AuditLogRepo{col: db.Collection("audit_logs")}
}

func (r *AuditLogRepo) Create(ctx context.Context, log *domain.AuditLog) error {
	log.ID = uuid.NewString()
	_, err := r.col.InsertOne(ctx, log)
	return err
}

func (r *AuditLogRepo) ListRecent(ctx context.Context, limit int) ([]*domain.AuditLog, error) {
	opts := options.Find().SetSort(bson.D{{Key: "occurred_at", Value: -1}}).SetLimit(int64(limit))
	cur, err := r.col.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	logs := []*domain.AuditLog{}
	if err := cur.All(ctx, &logs); err != nil {
		return nil, err
	}
	return logs, nil
}
