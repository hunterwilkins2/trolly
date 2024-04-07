package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/hunterwilkins2/trolly/components"
)

var (
	ErrItemNotFound = errors.New("item does not exist")
)

type Item struct {
	ID          int64
	Name        string
	Price       float32
	TimesBought int
}

type Metadata struct {
	CurrentPage  int
	PageSize     int
	FirstPage    int
	LastPage     int
	TotalRecords int
}

func calculateMetadata(totalRecords, page, pageSize int) Metadata {
	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    0,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}

type ItemRepository struct {
	db *sql.DB
}

func NewItemRepository(db *sql.DB) *ItemRepository {
	return &ItemRepository{
		db: db,
	}
}

func (r *ItemRepository) GetAll(ctx context.Context, search string, page int, pageSize int, orderBy string) (Metadata, []Item, error) {
	userId := ctx.Value(components.UserKey).(uuid.UUID)
	stmt := fmt.Sprintf(`
	SELECT count(*) OVER(), id, name, price, times_bought
	FROM items	
	WHERE user_id = ? AND (? = '' OR name LIKE CONCAT('%%', ?, '%%'))
	ORDER BY %s DESC, id DESC 
	LIMIT ? OFFSET ?
	`, orderedBy(orderBy))

	rows, err := r.db.QueryContext(ctx, stmt, userId.String(), search, search, pageSize, (page-1)*pageSize)
	if err != nil {
		return Metadata{}, nil, err
	}

	totalRecords := 0
	items := []Item{}
	for rows.Next() {
		item := Item{}
		err := rows.Scan(&totalRecords, &item.ID, &item.Name, &item.Price, &item.TimesBought)
		if err != nil {
			return Metadata{}, nil, err
		}
		items = append(items, item)
	}
	return calculateMetadata(totalRecords, page, pageSize), items, nil
}

func orderedBy(col string) string {
	switch col {
	case "recentlyPurchased":
		return "last_purchase_date"
	case "recentlyAdded":
		return "created_at"
	case "timeBought":
		return "times_bought"
	default:
		return "times_bought"
	}
}

func (r *ItemRepository) GetById(ctx context.Context, id int64) (Item, error) {
	userId := ctx.Value(components.UserKey).(uuid.UUID)
	stmt := `SELECT id, name, price, times_bought FROM items
	WHERE user_id = ? AND id = ?`

	item := &Item{}
	err := r.db.QueryRowContext(ctx, stmt, userId, id).Scan(&item.ID, &item.Name, &item.Price, &item.TimesBought)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Item{}, ErrItemNotFound
		}
		return Item{}, err
	}
	return *item, nil
}

func (r *ItemRepository) Create(ctx context.Context, item *Item) error {
	userId := ctx.Value(components.UserKey).(uuid.UUID)
	stmt := `INSERT INTO items (name, price, user_id)
	VALUES (?, ?, ?)
	RETURNING id`

	err := r.db.QueryRowContext(ctx, stmt, item.Name, item.Price, userId.String()).Scan(&item.ID)
	if err != nil {
		return err
	}

	return nil
}

func (r *ItemRepository) Update(ctx context.Context, item *Item) error {
	userId := ctx.Value(components.UserKey).(uuid.UUID)
	stmt := `UPDATE items
	SET name=?, price=?
	WHERE user_id=? AND id=?`

	_, err := r.db.ExecContext(ctx, stmt, item.Name, item.Price, userId, item.ID)
	if err != nil {
		return err
	}
	return nil
}

func (r *ItemRepository) Delete(ctx context.Context, id int64) error {
	userId := ctx.Value(components.UserKey).(uuid.UUID)
	stmt := `DELETE FROM items WHERE user_id = ? AND id = ?`

	row, err := r.db.ExecContext(ctx, stmt, userId, id)
	if err != nil {
		return err
	}
	n, err := row.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrItemNotFound
	}

	return nil
}
