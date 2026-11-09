package orchestrator

import (
	"context"
	"github.com/looplj/axonhub/internal/server/biz"
	"github.com/looplj/axonhub/llm/httpclient"
	"github.com/looplj/axonhub/llm/pipeline"
	"github.com/looplj/axonhub/llm/streams"
)

func (p *PersistentOutboundTransformer) checkUsageAccess(ctx context.Context, channel *biz.Channel) error {
	// Isolated transformer tests can supply a channel without persistence services.
	// Production orchestrators always install ChannelService.
	if p.state.ChannelService != nil {
		return p.state.ChannelService.CheckChannelUsageAccess(ctx, channel.ID)
	}
	if !biz.ChannelUsageAllowed(ctx, channel.Channel) {
		return biz.ChannelUsageDeniedError()
	}
	return nil
}

// The executor runs after admission/queue middleware, immediately before I/O.
type channelUsageExecutor struct {
	pipeline.Executor
	check func(context.Context) error
}

func (e *channelUsageExecutor) Do(ctx context.Context, request *httpclient.Request) (*httpclient.Response, error) {
	if err := e.check(ctx); err != nil {
		return nil, err
	}
	return e.Executor.Do(ctx, request)
}

func (e *channelUsageExecutor) DoStream(ctx context.Context, request *httpclient.Request) (streams.Stream[*httpclient.StreamEvent], error) {
	if err := e.check(ctx); err != nil {
		return nil, err
	}
	return e.Executor.DoStream(ctx, request)
}
