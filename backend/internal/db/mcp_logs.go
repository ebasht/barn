package db

import (
	"context"
	"github.com/google/uuid"
)

// QueryDeploymentLogPage bounds database work as well as the MCP response.
func (q *Queries) QueryDeploymentLogPage(ctx context.Context, id uuid.UUID, after int64, limit int) ([]DeploymentLog, error) {
	rows, err := q.db.Query(ctx, `SELECT id,deployment_id,level,message,created_at FROM deployment_logs WHERE deployment_id=$1 AND id>$2 ORDER BY id LIMIT $3`, id, after, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []DeploymentLog{}
	for rows.Next() {
		var r DeploymentLog
		if err := rows.Scan(&r.ID, &r.DeploymentID, &r.Level, &r.Message, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
