package biz

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/looplj/axonhub/internal/authz"
	"github.com/looplj/axonhub/internal/contexts"
	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/ent/channel"
	"github.com/looplj/axonhub/internal/ent/user"
	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/llm"
)

var ErrChannelUsageForbidden = errors.New("channel usage is not permitted for this user")

func ChannelUsageDeniedError() error {
	return fmt.Errorf("%w: %w", ErrChannelUsageForbidden, &llm.ResponseError{StatusCode: http.StatusForbidden, Detail: llm.ErrorDetail{Code: "channel_access_denied", Type: "permission_error", Message: "You do not have permission to use the channels providing this model."}})
}

// The authenticated key owner takes precedence even when a different web user is in context.
func ChannelUsageUserID(ctx context.Context) int {
	if key, ok := contexts.GetAPIKey(ctx); ok && key != nil {
		return key.UserID
	}
	if u, ok := contexts.GetUser(ctx); ok && u != nil {
		return u.ID
	}
	return 0
}

func ChannelUsageAllowed(ctx context.Context, ch *ent.Channel) bool {
	if ch.Settings == nil || ch.Settings.UsageAccess == nil {
		return true
	}
	access := ch.Settings.UsageAccess
	switch access.Mode {
	case objects.ChannelUsageAccessAll:
		return true
	case objects.ChannelUsageAccessSelected:
		id := ChannelUsageUserID(ctx)
		return id > 0 && slices.Contains(access.UserIDs, id)
	default:
		return false
	}
}

// Read authoritative settings for every dispatch, including a retry after queueing.
func (svc *ChannelService) CheckChannelUsageAccess(ctx context.Context, id int) error {
	return authz.RunWithSystemBypassVoid(ctx, "check-channel-usage-access", func(readCtx context.Context) error {
		ch, err := svc.entFromContext(readCtx).Channel.Query().Where(channel.IDEQ(id)).Select(channel.FieldID, channel.FieldSettings).Only(readCtx)
		if ent.IsNotFound(err) {
			return ChannelUsageDeniedError()
		}
		if err != nil {
			return fmt.Errorf("read channel usage access: %w", err)
		}
		if !ChannelUsageAllowed(ctx, ch) {
			return ChannelUsageDeniedError()
		}
		return nil
	})
}

func (svc *ChannelService) normalizeUsageAccess(ctx context.Context, access *objects.ChannelUsageAccess) error {
	if access == nil {
		return nil
	}
	if access.Mode == objects.ChannelUsageAccessAll {
		access.UserIDs = []int{}
		return nil
	}
	if access.Mode != objects.ChannelUsageAccessSelected || len(access.UserIDs) == 0 || len(access.UserIDs) > 10000 {
		return fmt.Errorf("invalid channel usage access: selected mode requires users")
	}
	ids := slices.Clone(access.UserIDs)
	slices.Sort(ids)
	ids = slices.Compact(ids)
	if ids[0] <= 0 {
		return fmt.Errorf("invalid channel usage access user ID")
	}
	count, err := authz.RunWithSystemBypass(ctx, "validate-channel-access-users", func(readCtx context.Context) (int, error) {
		return svc.entFromContext(readCtx).User.Query().Where(user.IDIn(ids...), user.StatusEQ(user.StatusActivated)).Count(readCtx)
	})
	if err != nil {
		return err
	}
	if count != len(ids) {
		return fmt.Errorf("channel usage access contains unknown or inactive users")
	}
	access.UserIDs = ids
	return nil
}

func cloneChannelUsageAccess(settings *objects.ChannelSettings) *objects.ChannelUsageAccess {
	if settings == nil || settings.UsageAccess == nil {
		return nil
	}
	return &objects.ChannelUsageAccess{Mode: settings.UsageAccess.Mode, UserIDs: slices.Clone(settings.UsageAccess.UserIDs)}
}

func (svc *ChannelService) preserveAndValidateUsageAccess(ctx context.Context, ch *ent.Channel, settings *objects.ChannelSettings) error {
	if settings.UsageAccess == nil {
		settings.UsageAccess = cloneChannelUsageAccess(ch.Settings)
		return nil
	}
	if err := svc.normalizeUsageAccess(ctx, settings.UsageAccess); err != nil {
		return err
	}
	previous := cloneChannelUsageAccess(ch.Settings)
	if previous == nil {
		previous = &objects.ChannelUsageAccess{Mode: objects.ChannelUsageAccessAll}
	}
	same := previous.Mode == settings.UsageAccess.Mode && slices.Equal(previous.UserIDs, settings.UsageAccess.UserIDs)
	currentUser, _ := contexts.GetUser(ctx)
	isOwner := currentUser != nil && currentUser.IsOwner
	if !same && ch.UserID != nil && ChannelUsageUserID(ctx) != *ch.UserID && !isOwner {
		return fmt.Errorf("only the donor or system owner can change donated channel usage access")
	}
	return nil
}

// AllowedUsageChannelIDs refreshes only ACL fields, keeping routing caches user-neutral.
func (svc *ChannelService) AllowedUsageChannelIDs(ctx context.Context, ids []int) (map[int]bool, error) {
	allowed := make(map[int]bool, len(ids))
	if len(ids) == 0 {
		return allowed, nil
	}
	err := authz.RunWithSystemBypassVoid(ctx, "filter-channel-usage-access", func(readCtx context.Context) error {
		rows, err := svc.entFromContext(readCtx).Channel.Query().Where(channel.IDIn(ids...)).Select(channel.FieldID, channel.FieldSettings).All(readCtx)
		if err != nil {
			return err
		}
		for _, row := range rows {
			allowed[row.ID] = ChannelUsageAllowed(ctx, row)
		}
		return nil
	})
	return allowed, err
}
