package models

import (
	"database/sql"
	"time"
)

type Item struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Price          float32   `json:"price"`
	TimesPurchased int       `json:"timesPurchased"`
	InCart         bool      `json:"inCart"`
	Purchased      bool      `json:"purchased"`
	LastAdded      time.Time `json:"lastAdded"`
}

type SortOptions struct {
	Price          bool
	TimesPurchased bool
	LastAdded      bool
}

type ItemModelInterface interface {
	Insert(uid int, item *Item) error
	Get(id int) (*Item, error)
	GetAll(uid int, name string, in_cart bool, sortby SortOptions) ([]*Item, error)
	GetByName(uid int, name string) (*Item, error)
	Update(item *Item) error
	Delete(id int) error
}

type ItemModel struct {
	DB *sql.DB
}

func (m ItemModel) Insert(uid int, item *Item) error {
	query := `
		INSERT INTO items(name, price, times_purchased, in_cart, purchased, uid, last_added)
		VALUES (?, ?, ?, ?, ?, ?, UTC_TIMESTAMP())
		RETURNING id, last_added
	`

	args := []any{item.Name, item.Price, 0, item.InCart, false, uid}

	err := m.DB.QueryRow(query, args...).Scan(&item.ID, &item.LastAdded)

	return err
}

func (m ItemModel) Get(id int) (*Item, error) {
	query := `
		SELECT id, name, price, times_purchased, in_cart, purchased, last_added
		FROM items
		WHERE id = ?
	`

	var item Item
	err := m.DB.QueryRow(query, id).Scan(
		&item.ID,
		&item.Name,
		&item.Price,
		&item.TimesPurchased,
		&item.InCart,
		&item.Purchased,
		&item.LastAdded,
	)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (m ItemModel) GetAll(uid int, name string, in_cart bool, sortby SortOptions) ([]*Item, error) {
	query := `
		SELECT id, name, price, times_purchased, in_cart, purchased, last_added
		FROM items
		WHERE uid = ? 
			AND (name LIKE CONCAT('%',?,'%') or ? = '')
			AND (in_cart = ? or ? = false)
	`

	args := []any{uid, name, name, in_cart, in_cart}
	rows, err := m.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	items := []*Item{}
	for rows.Next() {
		var item Item
		err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Price,
			&item.TimesPurchased,
			&item.InCart,
			&item.Purchased,
			&item.LastAdded,
		)

		if err != nil {
			return nil, err
		}

		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (m ItemModel) GetByName(uid int, name string) (*Item, error) {
	query := `
		SELECT id, name, price, times_purchased, in_cart, purchased, last_added
		FROM items
		WHERE uid = ? AND name = ?
	`

	var item Item
	err := m.DB.QueryRow(query, uid, name).Scan(
		&item.ID,
		&item.Name,
		&item.Price,
		&item.TimesPurchased,
		&item.InCart,
		&item.Purchased,
		&item.LastAdded,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return &item, nil
}

func (m ItemModel) Update(item *Item) error {
	query := `
		UPDATE items
		SET name = ?, price = ?, times_purchased = ?, in_cart = ?, purchased = ?, last_added = ?
		WHERE id = ?
	`

	args := []any{
		item.Name,
		item.Price,
		item.TimesPurchased,
		item.InCart,
		item.Purchased,
		item.LastAdded,
		item.ID,
	}

	row := m.DB.QueryRow(query, args...)
	if row.Err() != nil {
		return row.Err()
	}

	return nil
}

func (m ItemModel) Delete(id int) error {
	query := `
		DELETE FROM items
		WHERE id = ?
	`

	result, err := m.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNoRecord
	}

	return nil
}
