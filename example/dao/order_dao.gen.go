package dao

import (
	"context"

	"github.com/fengjx/daox/v2"
	"github.com/fengjx/daox/v2/example/dao/schema"
)

// RelFiller 关联查询
func (d *orderDao) RelFiller(ctx context.Context, list []*schema.AppOrder, p *daox.PreloadNode) error {
	if len(list) == 0 || p == nil {
		return nil
	}
	d.fillUser(ctx, list, p)
	return nil
}

func (d *orderDao) fillUser(ctx context.Context, list []*schema.AppOrder, p *daox.PreloadNode) error {
	n, ok := p.Node["User"]
	if !ok {
		return nil
	}
	refIDSet := make(map[int64]struct{})
	for _, o := range list {
		refIDSet[o.UserID] = struct{}{}
	}
	var refIDs []any
	for id := range refIDSet {
		refIDs = append(refIDs, id)
	}
	var refDao = daox.GetDao[*schema.AppUser]("app_user").WithPreloadNode(n)
	refList, err := refDao.ListByColumnsContext(ctx, daox.OfMultiKv("id", refIDs...))
	if err != nil {
		return err
	}
	refMap := make(map[int64]*schema.AppUser)
	for _, item := range refList {
		refMap[item.ID] = item
	}
	for _, o := range list {
		o.User = refMap[o.UserID]
	}
	return nil
}
