package service

import (
	"context"
	"fmt"

	"github.com/hunterwilkins2/trolly/internal/models"
)

type ItemService struct {
	repository *models.ItemRepository
}

func NewItemService(r *models.ItemRepository) *ItemService {
	return &ItemService{
		repository: r,
	}
}

func (s *ItemService) Search(ctx context.Context, query string, page int, pageSize int, orderBy string) (models.Metadata, []models.Item, error) {
	return s.repository.GetAll(ctx, query, page, pageSize, orderBy)
}

func (s *ItemService) Add(ctx context.Context, name string, price float32) (models.Item, error) {
	item := &models.Item{
		Name:  name,
		Price: price,
	}
	err := s.repository.Create(ctx, item)
	if err != nil {
		return models.Item{}, err
	}
	return *item, nil
}

func (s *ItemService) Update(ctx context.Context, id int64, name string, price float32) (models.Item, error) {
	item, err := s.repository.GetById(ctx, id)
	if err != nil {
		return models.Item{}, err
	}
	fmt.Println(item)

	if name != "" {
		item.Name = name
	}
	if price != 0 {
		item.Price = price
	}
	err = s.repository.Update(ctx, &item)
	if err != nil {
		return models.Item{}, err
	}

	return item, nil
}

func (s *ItemService) Get(ctx context.Context, id int64) (models.Item, error) {
	item, err := s.repository.GetById(ctx, id)
	if err != nil {
		return models.Item{}, err
	}
	return item, nil
}

func (s *ItemService) Remove(ctx context.Context, id int64) error {
	return s.repository.Delete(ctx, id)
}
