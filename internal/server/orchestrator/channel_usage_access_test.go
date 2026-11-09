package orchestrator

import (
	"context"
	"github.com/looplj/axonhub/internal/contexts"
	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/internal/server/biz"
	"github.com/looplj/axonhub/llm"
	"github.com/looplj/axonhub/llm/httpclient"
	"github.com/looplj/axonhub/llm/streams"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestChannelUsageSelectionFiltersStickyAndAliases(t *testing.T) {
	ctx, client := setupTest(t)
	channels := createTestChannels(t, ctx, client)
	svc := newTestChannelServiceForChannels(client)
	restricted := &objects.ChannelSettings{UsageAccess: &objects.ChannelUsageAccess{Mode: objects.ChannelUsageAccessSelected, UserIDs: []int{999}}}
	client.Channel.UpdateOneID(channels[0].ID).SetSettings(restricted).SaveX(ctx)
	// The selector intentionally holds stale public channel settings.
	candidates := []*ChannelModelsCandidate{}
	for _, ch := range channels[:2] {
		candidates = append(candidates, &ChannelModelsCandidate{Channel: &biz.Channel{Channel: ch}, Models: []biz.ChannelModelEntry{{RequestModel: "alias", ActualModel: "upstream"}}})
	}
	run := func(ctx context.Context) (*PersistenceState, error) {
		state := &PersistenceState{ChannelService: svc, CandidateSelector: &staticChannelSelector{candidates: append([]*ChannelModelsCandidate(nil), candidates...)}, UnifiedRoutes: biz.NewUnifiedRouteState()}
		_, err := selectCandidates(&PersistentInboundTransformer{state: state}).OnInboundLlmRequest(ctx, &llm.Request{Model: "alias"})
		return state, err
	}
	ctx = contexts.WithUser(ctx, &ent.User{ID: 123, IsOwner: true})
	state, err := run(contextWithSessionAffinityChannel(ctx, channels[0].ID))
	require.NoError(t, err)
	require.Len(t, state.ChannelModelsCandidates, 1)
	require.Equal(t, channels[1].ID, state.ChannelModelsCandidates[0].Channel.ID)
	client.Channel.UpdateOneID(channels[1].ID).SetSettings(restricted).SaveX(ctx)
	_, err = run(ctx)
	require.ErrorIs(t, err, biz.ErrChannelUsageForbidden)
	require.True(t, isNeutralAttemptError(ctx, err))
}

type accessCountingExecutor struct{ calls int }

func (e *accessCountingExecutor) Do(context.Context, *httpclient.Request) (*httpclient.Response, error) {
	e.calls++
	return &httpclient.Response{}, nil
}
func (e *accessCountingExecutor) DoStream(context.Context, *httpclient.Request) (streams.Stream[*httpclient.StreamEvent], error) {
	e.calls++
	return nil, nil
}

func TestChannelUsageExecutorRechecksAfterQueue(t *testing.T) {
	ctx, client := setupTest(t)
	channels := createTestChannels(t, ctx, client)
	svc := newTestChannelServiceForChannels(client)
	ctx = contexts.WithUser(ctx, &ent.User{ID: 123})
	downstream := &accessCountingExecutor{}
	executor := &channelUsageExecutor{Executor: downstream, check: func(ctx context.Context) error { return svc.CheckChannelUsageAccess(ctx, channels[0].ID) }}
	_, err := executor.Do(ctx, &httpclient.Request{})
	require.NoError(t, err)
	require.Equal(t, 1, downstream.calls)
	client.Channel.UpdateOneID(channels[0].ID).SetSettings(&objects.ChannelSettings{UsageAccess: &objects.ChannelUsageAccess{Mode: objects.ChannelUsageAccessSelected, UserIDs: []int{999}}}).SaveX(ctx)
	_, err = executor.Do(ctx, &httpclient.Request{})
	require.ErrorIs(t, err, biz.ErrChannelUsageForbidden)
	_, err = executor.DoStream(ctx, &httpclient.Request{})
	require.ErrorIs(t, err, biz.ErrChannelUsageForbidden)
	require.Equal(t, 1, downstream.calls, "neither normal nor streaming denial reaches upstream")
}
