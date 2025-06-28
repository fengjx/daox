package dao

import (
	"context"

	"github.com/fengjx/daox/v2"
	"github.com/fengjx/daox/v2/example/dao/schema"
)

// RelFiller 关联查询
func (d *userDao) RelFiller(ctx context.Context, users []*schema.AppUser, p *daox.PreloadNode) error {
	if len(users) == 0 || p == nil {
		return nil
	}
	userIDs := make([]any, 0, len(users))
	for _, u := range users {
		userIDs = append(userIDs, u.ID)
	}

	// Orders
	if n, ok := p.Node["Orders"]; ok {
		var _orderDao = daox.GetDao[*schema.AppOrder]("app_order").WithPreloadNode(n)
		orders, err := _orderDao.ListByColumnsContext(ctx, daox.OfMultiKv("user_id", userIDs...))
		if err != nil {
			return err
		}
		orderMap := make(map[int64][]*schema.AppOrder)
		for _, o := range orders {
			orderMap[o.UserID] = append(orderMap[o.UserID], o)
		}
		for _, u := range users {
			u.Orders = orderMap[u.ID]
		}
	}

	// Cards
	if n, ok := p.Node["Card"]; ok {
		var _cardDao = daox.GetDao[*schema.AppCard]("app_card").WithPreloadNode(n)
		cards, err := _cardDao.ListByColumnsContext(ctx, daox.OfMultiKv("user_id", userIDs...))
		if err != nil {
			return err
		}
		cardMap := make(map[int64]*schema.AppCard)
		for _, c := range cards {
			cardMap[c.UserID] = c
		}
		for _, u := range users {
			u.Card = cardMap[u.ID]
		}
	}
	return nil
}
