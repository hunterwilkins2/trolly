package service

import (
	"context"

	"github.com/hunterwilkins2/trolly/internal/models"
)

type BasketService struct {
	repository *models.BasketRepository
}

func NewBasketService(r *models.BasketRepository) *BasketService {
	return &BasketService{
		repository: r,
	}
}

func (s *BasketService) GetItems(ctx context.Context) (models.Basket, error) {
	return s.repository.Get(ctx)
}

func (s *BasketService) GetItem(ctx context.Context, basketId int64) (models.BasketItem, error) {
	return s.repository.GetItem(ctx, basketId)
}

func (s *BasketService) AddItem(ctx context.Context, item models.Item) (models.BasketItem, error) {
	return s.repository.Add(ctx, item)
}

func (s *BasketService) TogglePurchased(ctx context.Context, basketId int64) (models.BasketItem, error) {
	item, err := s.GetItem(ctx, basketId)
	if err != nil {
		return models.BasketItem{}, err
	}
	err = s.repository.TogglePurchased(ctx, &item)
	if err != nil {
		return models.BasketItem{}, err
	}
	return item, nil
}

func (s *BasketService) RemoveItem(ctx context.Context, basketId int64) error {
	return s.repository.Remove(ctx, basketId)
}

func (s *BasketService) RemoveAllItems(ctx context.Context) error {
	return s.repository.RemoveAll(ctx)
}
