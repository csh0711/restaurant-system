package events

import "restaurant-system/shared/model"

type OrderEvent struct {
	OrderID model.OrderID    `json:"orderId"`
	TableID model.TableID    `json:"tableId"`
	Items   []model.MenuItem `json:"items"`
}

type DishPreparedEvent struct {
	OrderID model.OrderID    `json:"orderId"`
	TableID model.TableID    `json:"tableId"`
	Items   []model.MenuItem `json:"items"`
	Status  model.Status     `json:"status"`
}

type OrderServedEvent struct {
	OrderID model.OrderID    `json:"orderId"`
	TableID model.TableID    `json:"tableId"`
	Items   []model.MenuItem `json:"items"`
	Status  model.Status     `json:"status"`
}
