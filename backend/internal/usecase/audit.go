package usecase

import (
	"context"
	"log"
	"time"

	"cinema-ticket/backend/internal/domain"
	"cinema-ticket/backend/internal/usecase/ports"
)

func recordAudit(ctx context.Context, repo ports.AuditLogRepository, entry *domain.AuditLog) {
	entry.OccurredAt = time.Now().UTC()
	writeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if err := repo.Create(writeCtx, entry); err != nil {
		log.Printf("audit log write failed: %v", err)
	}
}
