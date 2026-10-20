package responses

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/looplj/axonhub/llm"
	"github.com/looplj/axonhub/llm/httpclient"
	"github.com/looplj/axonhub/llm/streams"
)

func TestProviderReasoningTextStreamIsVisible(t *testing.T) {
	events := []*httpclient.StreamEvent{
		{Type: "response.created", Data: []byte(`{"type":"response.created","response":{"id":"resp_reasoning","object":"response","created_at":1700000000,"model":"local","status":"in_progress","output":[]}}`)},
		{Type: "response.output_item.added", Data: []byte(`{"type":"response.output_item.added","output_index":0,"item":{"id":"rs_local","type":"reasoning","summary":[]}}`)},
		{Type: "response.reasoning_text.delta", Data: []byte(`{"type":"response.reasoning_text.delta","item_id":"rs_local","output_index":0,"content_index":0,"delta":"thinking"}`)},
		{Type: "response.reasoning_text.done", Data: []byte(`{"type":"response.reasoning_text.done","item_id":"rs_local","output_index":0,"content_index":0,"text":"thinking"}`)},
		{Type: "response.completed", Data: []byte(`{"type":"response.completed","response":{"id":"resp_reasoning","object":"response","created_at":1700000000,"model":"local","status":"completed","output":[]}}`)},
	}

	transformer, err := NewOutboundTransformer("https://example.com", "test")
	require.NoError(t, err)
	stream, err := transformer.TransformStream(t.Context(), nil, streams.SliceStream(events))
	require.NoError(t, err)
	chunks, err := streams.All(stream)
	require.NoError(t, err)
	var reasoning []string
	for _, chunk := range chunks {
		if chunk == nil || chunk == llm.DoneResponse {
			continue
		}
		for _, choice := range chunk.Choices {
			if choice.Delta != nil && choice.Delta.ReasoningContent != nil {
				reasoning = append(reasoning, *choice.Delta.ReasoningContent)
			}
		}
	}
	require.Equal(t, []string{"thinking"}, reasoning)

	inbound := NewInboundTransformer()
	replayed, err := inbound.TransformStream(t.Context(), streams.SliceStream(chunks))
	require.NoError(t, err)
	replayedEvents, err := streams.All(replayed)
	require.NoError(t, err)
	var summaries []string
	for _, event := range replayedEvents {
		if event == nil || event.Type != string(StreamEventTypeReasoningSummaryTextDelta) {
			continue
		}
		var decoded StreamEvent
		require.NoError(t, json.Unmarshal(event.Data, &decoded))
		summaries = append(summaries, decoded.Delta)
	}
	require.Equal(t, []string{"thinking"}, summaries)

	data, _, err := AggregateStreamChunks(t.Context(), events)
	require.NoError(t, err)
	var result Response
	require.NoError(t, json.Unmarshal(data, &result))
	require.Len(t, result.Output, 1)
	require.Equal(t, "thinking", result.Output[0].Summary[0].Text)
}

func TestProviderReasoningContentInNonStreamingResponse(t *testing.T) {
	var item Item
	require.NoError(t, json.Unmarshal([]byte(`{"type":"reasoning","content":[{"type":"reasoning_text","text":"full thought"}],"summary":[{"type":"summary_text","text":"short summary"}]}`), &item))
	message := convertOutputToMessage([]Item{item}, nil)
	require.Equal(t, "full thought", *message.ReasoningContent)

	item.Content = nil
	message = convertOutputToMessage([]Item{item}, nil)
	require.Equal(t, "short summary", *message.ReasoningContent)
}
