package biz

import (
	"context"
	"testing"
	"time"

	"github.com/looplj/axonhub/internal/authz"
	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/ent/enttest"
	"github.com/looplj/axonhub/internal/ent/project"
	"github.com/looplj/axonhub/internal/ent/request"
	"github.com/looplj/axonhub/internal/objects"
	"github.com/stretchr/testify/require"
)

func TestCampusModelActivityCountsOnlyVisibleProjectAPISuccesses(t *testing.T) {
	client := enttest.NewEntClient(t, "sqlite3", "file:campus_model_activity?mode=memory&_fk=0")
	defer client.Close()
	ctx := authz.WithTestBypass(ent.NewContext(context.Background(), client))
	p := client.Project.Create().SetName("Campus").SetStatus(project.StatusActive).SaveX(ctx)
	other := client.Project.Create().SetName("Other").SetStatus(project.StatusActive).SaveX(ctx)
	now := time.Date(2026, 10, 2, 12, 30, 0, 0, time.UTC)
	start := now.Add(-24 * time.Hour)
	add := func(projectID int, model string, source request.Source, status request.Status, at time.Time) {
		client.Request.Create().SetProjectID(projectID).SetModelID(model).SetSource(source).SetStatus(status).
			SetRequestBody(objects.JSONRawMessage(`{}`)).SetCreatedAt(at).SetUpdatedAt(at).SaveX(ctx)
	}
	add(p.ID, "visible", request.SourceAPI, request.StatusCompleted, start)
	add(p.ID, "visible", request.SourceAPI, request.StatusCompleted, start.Add(time.Hour-time.Nanosecond))
	add(p.ID, "visible", request.SourceAPI, request.StatusCompleted, start.Add(time.Hour))
	add(p.ID, "visible", request.SourceAPI, request.StatusCompleted, now.Add(-time.Second))
	add(p.ID, "visible", request.SourceAPI, request.StatusCompleted, start.Add(-time.Second))
	add(p.ID, "visible", request.SourceAPI, request.StatusCompleted, now)
	add(p.ID, "visible", request.SourceAPI, request.StatusFailed, now.Add(-time.Minute))
	add(p.ID, "visible", request.SourceAPI, request.StatusCanceled, now.Add(-time.Minute))
	add(p.ID, "visible", request.SourceTest, request.StatusCompleted, now.Add(-time.Minute))
	add(p.ID, "visible", request.SourcePlayground, request.StatusCompleted, now.Add(-time.Minute))
	add(p.ID, "private", request.SourceAPI, request.StatusCompleted, now.Add(-time.Minute))
	add(other.ID, "visible", request.SourceAPI, request.StatusCompleted, now.Add(-time.Minute))
	svc := &CampusCatalogService{client: client}
	result, err := svc.campusModelActivity(ctx, p.ID, []string{"visible", "unused"}, now)
	require.NoError(t, err)
	require.Len(t, result.Models, 2)
	require.Equal(t, 4, result.Models["visible"].SuccessCount)
	require.Len(t, result.Models["visible"].Buckets, 24)
	require.Equal(t, 2, result.Models["visible"].Buckets[0])
	require.Equal(t, 1, result.Models["visible"].Buckets[1])
	require.Equal(t, 1, result.Models["visible"].Buckets[23])
	require.Equal(t, 0, result.Models["unused"].SuccessCount)
	empty, err := svc.campusModelActivity(ctx, p.ID, nil, now)
	require.NoError(t, err)
	require.Empty(t, empty.Models)
}
