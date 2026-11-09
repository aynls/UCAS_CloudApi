package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/looplj/axonhub/internal/authz"
	"github.com/looplj/axonhub/internal/contexts"
	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/ent/channel"
	"github.com/looplj/axonhub/internal/objects"
	"github.com/looplj/axonhub/internal/pkg/xcache"
	"github.com/looplj/axonhub/llm"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestChannelUsageIdentityAndLegacy(t *testing.T) {
	ch := &ent.Channel{}
	require.True(t, ChannelUsageAllowed(context.Background(), ch))
	ch.Settings = &objects.ChannelSettings{UsageAccess: &objects.ChannelUsageAccess{Mode: objects.ChannelUsageAccessSelected, UserIDs: []int{7}}}
	owner := contexts.WithUser(context.Background(), &ent.User{ID: 8, IsOwner: true})
	require.False(t, ChannelUsageAllowed(owner, ch), "management does not grant usage")
	ctx := contexts.WithAPIKey(owner, &ent.APIKey{UserID: 7})
	require.True(t, ChannelUsageAllowed(ctx, ch), "key owner is authoritative")
	require.False(t, ChannelUsageAllowed(context.Background(), ch))
	ch.Settings.UsageAccess.Mode = "invalid"
	require.False(t, ChannelUsageAllowed(ctx, ch))
	var response *llm.ResponseError
	require.ErrorAs(t, ChannelUsageDeniedError(), &response)
	require.Equal(t, 403, response.StatusCode)
}

func TestChannelUsagePersistenceRevocationAndValidation(t *testing.T) {
	svc, client := setupTestChannelService(t)
	defer client.Close()
	ctx := authz.WithTestBypass(ent.NewContext(context.Background(), client))
	alice := client.User.Create().SetEmail("alice@example.test").SetPassword("hash").SetNickname("Alice").SaveX(ctx)
	bob := client.User.Create().SetEmail("bob@example.test").SetPassword("hash").SaveX(ctx)
	selected := func(ids ...int) *objects.ChannelSettings {
		return &objects.ChannelSettings{UsageAccess: &objects.ChannelUsageAccess{Mode: objects.ChannelUsageAccessSelected, UserIDs: ids}}
	}
	input := ent.CreateChannelInput{Type: channel.TypeOpenai, Name: "private", Credentials: objects.ChannelCredentials{APIKey: "test-key"}, SupportedModels: []string{"private-model"}, DefaultTestModel: "private-model", AutoSyncSupportedModels: new(false), Settings: selected(alice.ID, bob.ID, alice.ID)}
	ch, err := svc.CreateChannel(ctx, input)
	require.NoError(t, err)
	require.Equal(t, []int{alice.ID, bob.ID}, ch.Settings.UsageAccess.UserIDs)
	aliceCtx := contexts.WithUser(ctx, alice)
	require.NoError(t, svc.CheckChannelUsageAccess(aliceCtx, ch.ID))
	// Unrelated edits and clearSettings must never make the channel public.
	for _, update := range []*ent.UpdateChannelInput{{Settings: &objects.ChannelSettings{ExtraModelPrefix: "prefix"}}, {ClearSettings: true}} {
		updated, err := svc.UpdateChannel(ctx, ch.ID, update)
		require.NoError(t, err)
		require.Equal(t, ch.Settings.UsageAccess, updated.Settings.UsageAccess)
	}
	_, err = svc.UpdateChannel(ctx, ch.ID, &ent.UpdateChannelInput{Settings: selected(bob.ID)})
	require.NoError(t, err)
	require.ErrorIs(t, svc.CheckChannelUsageAccess(aliceCtx, ch.ID), ErrChannelUsageForbidden)
	require.NoError(t, svc.CheckChannelUsageAccess(contexts.WithUser(ctx, bob), ch.ID))
	require.True(t, ChannelUsageAllowed(aliceCtx, ch), "fixture intentionally retains stale cached settings")
	ids, err := svc.AllowedUsageChannelIDs(aliceCtx, []int{ch.ID})
	require.NoError(t, err)
	require.False(t, ids[ch.ID])
	for i, invalid := range []*objects.ChannelSettings{selected(), selected(-1), selected(999999), {UsageAccess: &objects.ChannelUsageAccess{Mode: "bogus"}}} {
		input.Name = fmt.Sprintf("invalid-%d", i)
		input.Settings = invalid
		_, err := svc.CreateChannel(ctx, input)
		require.Error(t, err)
	}
	require.Equal(t, 1, client.Channel.Query().CountX(ctx), "invalid creation is atomic")

	// The available catalog consults current ACLs even when its channel cache is stale.
	svc.SetEnabledChannelsForTest([]*Channel{{Channel: ch}})
	models := &ModelService{AbstractService: &AbstractService{db: client}, channelService: svc, systemService: &SystemService{AbstractService: &AbstractService{db: client}, Cache: xcache.NewFromConfig[ent.System](xcache.Config{Mode: xcache.ModeMemory})}}
	visible, err := models.ListEnabledModels(aliceCtx)
	require.NoError(t, err)
	require.Empty(t, visible)
	visible, err = models.ListEnabledModels(contexts.WithUser(ctx, bob))
	require.NoError(t, err)
	require.Len(t, visible, 1)
	input.Name = "public-alternative"
	input.Settings = nil
	alternative, err := svc.CreateChannel(ctx, input)
	require.NoError(t, err)
	svc.SetEnabledChannelsForTest([]*Channel{{Channel: ch}, {Channel: alternative}})
	visible, err = models.ListEnabledModels(aliceCtx)
	require.NoError(t, err)
	require.Len(t, visible, 1)
	require.Equal(t, "private-model", visible[0].ID)
	input.Name = "copied-private"
	input.Settings = nil
	duplicated, err := svc.DuplicateChannel(ctx, ch.ID, input)
	require.NoError(t, err)
	require.Equal(t, []int{bob.ID}, duplicated.Settings.UsageAccess.UserIDs, "omitted settings in a duplicate retain the source ACL")
	// A donor's ACL can be preserved by a manager but only changed by that donor.
	donation := client.Channel.Create().SetUser(alice).SetType(channel.TypeOpenai).SetName("donation").SetCredentials(objects.ChannelCredentials{APIKey: "key"}).SetSupportedModels([]string{"model"}).SetDefaultTestModel("model").SetSettings(selected(bob.ID)).SaveX(ctx)
	_, err = svc.UpdateChannel(contexts.WithUser(ctx, bob), donation.ID, &ent.UpdateChannelInput{Settings: selected(alice.ID)})
	require.ErrorContains(t, err, "only the donor")
}

func TestChannelUsagePublicProjectionAndDirectory(t *testing.T) {
	_, client := setupTestChannelService(t)
	defer client.Close()
	setup := authz.WithTestBypass(ent.NewContext(context.Background(), client))
	project := client.Project.Create().SetName("Campus").SaveX(setup)
	alice := client.User.Create().SetEmail("private-email@example.test").SetPassword("private-hash").SetNickname("Alice").SaveX(setup)
	bob := client.User.Create().SetEmail("bob@example.test").SetPassword("hash").SaveX(setup)
	client.UserProject.Create().SetUser(bob).SetProject(project).SaveX(setup)
	ch := client.Channel.Create().SetName("A contribution").SetStatus(channel.StatusEnabled).SetType(channel.TypeOpenai).SetUser(alice).SetRemark("A donor message").SetSupportedModels([]string{"secret-model"}).SetDefaultTestModel("secret-model").SetCredentials(objects.ChannelCredentials{APIKey: "secret-key"}).SetSettings(&objects.ChannelSettings{UsageAccess: &objects.ChannelUsageAccess{Mode: objects.ChannelUsageAccessSelected, UserIDs: []int{alice.ID}}}).SaveX(setup)
	svc := &CampusCatalogService{client: client}
	ctx := contexts.WithUser(setup, bob)
	ctx = contexts.WithProjectID(ctx, project.ID)
	resources, err := svc.listPublicChannels(ctx, project.ID, time.Now())
	require.NoError(t, err)
	require.Len(t, resources, 1)
	require.False(t, resources[0].CanUse)
	require.False(t, resources[0].CanProbe)
	require.Empty(t, resources[0].Models)
	require.Nil(t, resources[0].Health)
	require.Equal(t, "Alice", resources[0].Contributor)
	require.Equal(t, "A donor message", resources[0].Description)
	encoded, err := json.Marshal(resources)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "secret-model")
	require.NotContains(t, string(encoded), "userIDs")
	_, _, err = svc.PrepareChannelProbe(ctx, ch.ID, nil)
	require.ErrorIs(t, err, ErrChannelUsageForbidden)
	_, err = svc.GetChannelHealth(ctx, ch.ID)
	require.ErrorIs(t, err, ErrChannelUsageForbidden)
	users, err := svc.ListAccessUsers(ctx, "Alice", 0, nil)
	require.NoError(t, err)
	require.Len(t, users.Users, 1)
	require.Equal(t, alice.ID, users.Users[0].ID)
	encoded, err = json.Marshal(users)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "email")
	require.NotContains(t, string(encoded), "hash")
	users, err = svc.ListAccessUsers(ctx, fmt.Sprint(bob.ID), 0, nil)
	require.NoError(t, err)
	require.True(t, users.Users[0].IsSelf)
}
