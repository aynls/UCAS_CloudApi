package biz

import (
	"context"
	"github.com/looplj/axonhub/internal/authz"
	"github.com/looplj/axonhub/internal/ent"
	"github.com/looplj/axonhub/internal/ent/user"
	"strconv"
	"strings"
)

type CampusAccessUser struct {
	ID          int    `json:"id"`
	DisplayName string `json:"displayName"`
	IsSelf      bool   `json:"isSelf"`
}
type CampusAccessUsers struct {
	Users      []CampusAccessUser `json:"users"`
	NextCursor int                `json:"nextCursor"`
}

// This directory deliberately exposes neither email nor administrative details.
func (svc *CampusCatalogService) ListAccessUsers(ctx context.Context, search string, after int, ids []int) (*CampusAccessUsers, error) {
	current, projectID, err := campusCatalogIdentity(ctx)
	if err != nil {
		return nil, err
	}
	search = strings.TrimSpace(search)
	if len(search) > 512 || after < 0 || len(ids) > 10000 {
		return nil, ErrCampusCatalogInvalidInput
	}
	return authz.RunWithSystemBypass(ctx, "campus-access-user-directory", func(readCtx context.Context) (*CampusAccessUsers, error) {
		if _, err := svc.verifyCampusProjectAccess(readCtx, current, projectID); err != nil {
			return nil, err
		}
		query := svc.client.User.Query().Where(user.StatusEQ(user.StatusActivated), user.IDGT(after))
		if len(ids) > 0 {
			query.Where(user.IDIn(ids...))
		}
		if search != "" {
			id, _ := strconv.Atoi(strings.TrimPrefix(search, "#"))
			query.Where(user.Or(user.NicknameContainsFold(search), user.IDEQ(id)))
		}
		rows, err := query.Select(user.FieldID, user.FieldNickname).Order(ent.Asc(user.FieldID)).Limit(51).All(readCtx)
		if err != nil {
			return nil, err
		}
		result := &CampusAccessUsers{Users: []CampusAccessUser{}}
		if len(rows) > 50 {
			rows = rows[:50]
			result.NextCursor = rows[49].ID
		}
		for _, row := range rows {
			result.Users = append(result.Users, CampusAccessUser{ID: row.ID, DisplayName: CampusDisplayName(row.Nickname, CampusPublicAlias(projectID, row.ID)), IsSelf: row.ID == current.ID})
		}
		return result, nil
	})
}
