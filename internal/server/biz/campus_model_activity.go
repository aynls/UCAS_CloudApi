package biz

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"

	"github.com/looplj/axonhub/internal/ent/request"
)

const campusModelActivityBuckets = 24

// CampusModelActivity contains aggregate successes only. It must never expose
// request identities, users, keys, prompts, or failures as model-health data.
type CampusModelActivity struct {
	WindowStart time.Time                   `json:"windowStart"`
	WindowEnd   time.Time                   `json:"windowEnd"`
	BucketHours int                         `json:"bucketHours"`
	Models      map[string]CampusModelUsage `json:"models"`
}

type CampusModelUsage struct {
	SuccessCount int   `json:"successCount"`
	Buckets      []int `json:"buckets"`
}

// Called only after project membership and the user's visible models have
// been resolved by GetResources. Request rows, unlike execution attempts,
// count retries only once. No new storage or background probes are required.
func (svc *CampusCatalogService) campusModelActivity(ctx context.Context, projectID int, models []string, now time.Time) (*CampusModelActivity, error) {
	result := &CampusModelActivity{
		WindowStart: now.UTC().Add(-24 * time.Hour), WindowEnd: now.UTC(),
		BucketHours: 1, Models: make(map[string]CampusModelUsage, len(models)),
	}
	for _, model := range models {
		result.Models[model] = CampusModelUsage{Buckets: make([]int, campusModelActivityBuckets)}
	}
	if len(models) == 0 {
		return result, nil
	}
	var rows []struct {
		ModelID string `json:"model_id"`
		Bucket  int    `json:"bucket"`
		Count   int    `json:"count"`
	}
	err := svc.client.Request.Query().Where(
		request.ProjectIDEQ(projectID), request.ModelIDIn(models...),
		request.SourceEQ(request.SourceAPI), request.StatusEQ(request.StatusCompleted),
		request.CreatedAtGTE(result.WindowStart), request.CreatedAtLT(result.WindowEnd),
	).Modify(func(s *sql.Selector) {
		// A parameterized CASE keeps bucketing portable across SQLite/MySQL/PG.
		expression := sql.ExprFunc(func(b *sql.Builder) {
			b.WriteString("CASE")
			for bucket := 0; bucket < campusModelActivityBuckets-1; bucket++ {
				b.WriteString(" WHEN ").Ident(s.C(request.FieldCreatedAt)).WriteString(" < ")
				b.Arg(result.WindowStart.Add(time.Duration(bucket+1) * time.Hour))
				b.WriteString(fmt.Sprintf(" THEN %d", bucket))
			}
			b.WriteString(" ELSE 23 END")
		})
		s.Select(s.C(request.FieldModelID)).
			AppendSelectExprAs(expression, "bucket").
			AppendSelectExprAs(sql.Expr("COUNT(*)"), "count").
			GroupBy(s.C(request.FieldModelID), "bucket")
	}).Scan(ctx, &rows)
	if err != nil {
		return nil, fmt.Errorf("aggregate campus model successes: %w", err)
	}
	for _, row := range rows {
		usage := result.Models[row.ModelID]
		usage.SuccessCount += row.Count
		usage.Buckets[row.Bucket] = row.Count
		result.Models[row.ModelID] = usage
	}
	return result, nil
}
