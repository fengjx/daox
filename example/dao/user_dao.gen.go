package dao

import (
	"context"

	"github.com/fengjx/daox/v2"
	"github.com/fengjx/daox/v2/example/dao/schema"
)

// RelFiller 关联查询
func (d *userDao) RelFiller(ctx context.Context, list []*schema.AppUser, p *daox.PreloadNode) error {
	if len(list) == 0 || p == nil {
		return nil
	}
	// Orders
	d.fillOrders(ctx, list, p)

	// Cards
	d.fillCard(ctx, list, p)
	return nil
}

func (d *userDao) fillOrders(ctx context.Context, list []*schema.AppUser, p *daox.PreloadNode) error {
	n, ok := p.Node["Orders"]
	if !ok {
		return nil
	}
	ids := make([]any, 0, len(list))
	for _, u := range list {
		ids = append(ids, u.ID)
	}
	var refDao = daox.GetDao[*schema.AppOrder]("app_order").WithPreloadNode(n)
	redList, err := refDao.ListByColumnsContext(ctx, daox.OfMultiKv("user_id", ids...))
	if err != nil {
		return err
	}
	refMap := make(map[int64][]*schema.AppOrder)
	for _, o := range redList {
		refMap[o.UserID] = append(refMap[o.UserID], o)
	}
	for _, u := range list {
		u.Orders = refMap[u.ID]
	}
	return nil
}

func (d *userDao) fillCard(ctx context.Context, list []*schema.AppUser, p *daox.PreloadNode) error {
	n, ok := p.Node["Card"]
	if !ok {
		return nil
	}
	ids := make([]any, 0, len(list))
	for _, u := range list {
		ids = append(ids, u.ID)
	}
	var refDao = daox.GetDao[*schema.AppCard]("app_card").WithPreloadNode(n)
	refList, err := refDao.ListByColumnsContext(ctx, daox.OfMultiKv("user_id", ids...))
	if err != nil {
		return err
	}
	refMap := make(map[int64]*schema.AppCard)
	for _, c := range refList {
		refMap[c.UserID] = c
	}
	for _, u := range list {
		u.Card = refMap[u.ID]
	}
	return nil
}
