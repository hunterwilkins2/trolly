package models

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/hunterwilkins2/trolly/components"
)

var (
	ErrBasketItemNotFound = errors.New("basket item could not be found")
)

type BasketItem struct {
	BasketID  int64
	Purchased bool
	Item
}

type Basket struct {
	Items []BasketItem
	Total float32
}

type BasketRepository struct {
	db *sql.DB
}

func NewBasketRepository(db *sql.DB) *BasketRepository {
	return &BasketRepository{
		db: db,
	}
}

func (r *BasketRepository) Get(ctx context.Context) (Basket, error) {
	userId := ctx.Value(components.UserKey).(uuid.UUID)
	stmt := `SELECT b.id, b.purchased, i.id, i.name, i.price, i.times_bought 
	FROM basket b 
	INNER JOIN items i 
	ON i.id = b.item_id 
	WHERE b.user_id = ?`

	rows, err := r.db.QueryContext(ctx, stmt, userId)
	if err != nil {
		return Basket{}, err
	}
	items := []BasketItem{}
	total := 0.0
	for rows.Next() {
		var item BasketItem
		err := rows.Scan(&item.BasketID, &item.Purchased, &item.ID, &item.Name, &item.Price, &item.TimesBought)
		if err != nil {
			return Basket{}, err
		}
		total += float64(item.Price)
		items = append(items, item)
	}

	return Basket{
		Items: items,
		Total: float32(total),
	}, nil
}

func (r *BasketRepository) GetItem(ctx context.Context, basketId int64) (BasketItem, error) {
	userId := ctx.Value(components.UserKey).(uuid.UUID)
	stmt := `SELECT b.id, b.purchased, i.id, i.name, i.price, i.times_bought
	FROM basket b
	INNER JOIN items i
	ON i.id = b.item_id
	WHERE b.user_id = ? AND b.id = ?`

	var item BasketItem
	err := r.db.QueryRowContext(ctx, stmt, userId, basketId).Scan(
		&item.BasketID, &item.Purchased, &item.ID, &item.Name, &item.Price, &item.TimesBought,
	)
	if err != nil {
		return BasketItem{}, ErrBasketItemNotFound
	}

	return item, nil

}

func (r *BasketRepository) Add(ctx context.Context, item Item) (BasketItem, error) {
	userId := ctx.Value(components.UserKey).(uuid.UUID)
	stmt := `INSERT INTO basket (user_id, item_id)
	VALUES (?, ?)`

	row, err := r.db.ExecContext(ctx, stmt, userId, item.ID)
	if err != nil {
		return BasketItem{}, err
	}

	id, err := row.LastInsertId()
	if err != nil {
		return BasketItem{}, err
	}

	return BasketItem{
		id,
		false,
		item,
	}, nil
}

func (r *BasketRepository) TogglePurchased(ctx context.Context, item *BasketItem) error {
	userId := ctx.Value(components.UserKey).(uuid.UUID)
	stmt := `UPDATE basket SET purchased = ? WHERE user_id = ? AND id = ?`
	row, err := r.db.ExecContext(ctx, stmt, !item.Purchased, userId, item.BasketID)
	if err != nil {
		return err
	}
	n, err := row.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrBasketItemNotFound
	}

	item.Purchased = !item.Purchased
	return nil
}

func (r *BasketRepository) Remove(ctx context.Context, basketId int64) error {
	userId := ctx.Value(components.UserKey).(uuid.UUID)
	stmt := `DELETE FROM basket WHERE user_id = ? AND id = ?`

	row, err := r.db.ExecContext(ctx, stmt, userId, basketId)
	if err != nil {
		return err
	}
	n, err := row.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrBasketItemNotFound
	}

	return nil
}

func (r *BasketRepository) RemoveAll(ctx context.Context) error {
	userId := ctx.Value(components.UserKey).(uuid.UUID)
	stmt := `DELETE FROM basket WHERE user_id = ?`

	row, err := r.db.ExecContext(ctx, stmt, userId)
	if err != nil {
		return err
	}
	n, err := row.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrBasketItemNotFound
	}

	return nil
}
