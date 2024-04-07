package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/alexedwards/flow"
	"github.com/hunterwilkins2/trolly/components"
	"github.com/hunterwilkins2/trolly/components/pages"
	"github.com/hunterwilkins2/trolly/internal/models"
	"github.com/hunterwilkins2/trolly/internal/service"
	"github.com/hunterwilkins2/trolly/internal/validator"
)

const PAGE_SIZE = 10
const SUGGEST_SIZE = 5

func (app *application) GroceryListPage(w http.ResponseWriter, r *http.Request) {
	basket, err := app.basket.GetItems(r.Context())
	if err != nil {
		app.logger.Error("could not get basket", "error", err.Error())
		r = r.WithContext(context.WithValue(r.Context(), components.FlashKey, "Could not get items"))
	} else {
		app.logger.Info("got basket", "basket", basket)
	}
	pages.GroceryList(basket).Render(r.Context(), w)
}

func (app *application) PantryPage(w http.ResponseWriter, r *http.Request) {
	var items []models.Item
	metadata, items, err := app.items.Search(r.Context(), "", 1, PAGE_SIZE, "times_bought")
	if err != nil {
		app.logger.Error("could not get items", "errors", err.Error())
		r = r.WithContext(context.WithValue(r.Context(), components.FlashKey, "Could not retrieve items"))
	}
	app.logger.Debug("got items", "items", items, "metadata", metadata)
	pages.Pantry("", "timesBought", metadata, items).Render(r.Context(), w)
}

func (app *application) LoginPage(w http.ResponseWriter, r *http.Request) {
	pages.Login(nil, nil).Render(r.Context(), w)
}

func (app *application) RegisterPage(w http.ResponseWriter, r *http.Request) {
	pages.Register(nil, nil).Render(r.Context(), w)
}

func (app *application) ValidateName(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	v := validator.New()
	models.ValidateName(v, name)
	if v.HasErrors() {
		w.Write([]byte(v.GetError("name").Error()))
		return
	}

	w.Write([]byte(""))
}

func (app *application) ValidateEmail(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	v := validator.New()
	models.ValidateEmail(v, email)
	if v.HasErrors() {
		w.Write([]byte(v.GetError("email").Error()))
		return
	}

	w.Write([]byte(""))
}

func (app *application) ValidatePassword(w http.ResponseWriter, r *http.Request) {
	password := r.FormValue("password")
	v := validator.New()
	models.ValidatePassword(v, password)
	if v.HasErrors() {
		w.Write([]byte(v.GetError("password").Error()))
		return
	}

	w.Write([]byte(""))
}

func (app *application) Register(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")
	user, err := app.users.Register(r.Context(), name, email, password)
	if err != nil {
		app.logger.Error("Failed to create user", "error", err.Error(), "name", name, "email", email)
		var ee map[string]error
		var v *validator.Validator
		if errors.As(err, &v) {
			ee = v.FieldErrors
		} else if err == models.ErrDuplicateEmail {
			r = r.WithContext(context.WithValue(r.Context(), components.FlashKey, "User with that email already exists"))
		} else {
			r = r.WithContext(context.WithValue(r.Context(), components.FlashKey, "Could not create account. Please try again."))
		}

		pages.Register(map[string]string{"name": name, "email": email}, ee).Render(r.Context(), w)
		return
	}
	app.logger.Info("created new user", "name", user.Name, "email", email)

	app.sessionManager.RenewToken(r.Context())
	app.sessionManager.Put(r.Context(), "userId", user.ID)
	app.sessionManager.Put(r.Context(), "userName", user.Name)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) Login(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	user, err := app.users.Login(r.Context(), email, password)
	if err != nil {
		app.logger.Error("Failed to log user in", "error", err.Error(), "email", email)
		var ee map[string]error
		var v *validator.Validator
		if errors.As(err, &v) {
			ee = v.FieldErrors
		} else if err == service.ErrInvalidCredentials {
			r = r.WithContext(context.WithValue(r.Context(), components.FlashKey, "Email or password is incorrect"))
		} else {
			r = r.WithContext(context.WithValue(r.Context(), components.FlashKey, "Could not log in. Please try again."))
		}
		pages.Login(map[string]string{"email": email}, ee).Render(r.Context(), w)
		return
	}

	app.logger.Info("login from", "name", user.Name, "email", email)

	app.sessionManager.RenewToken(r.Context())
	app.sessionManager.Put(r.Context(), "userId", user.ID)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) Logout(w http.ResponseWriter, r *http.Request) {
	app.sessionManager.RenewToken(r.Context())
	app.sessionManager.Remove(r.Context(), "userId")

	w.Header().Add("HX-Redirect", "/login")
}

func (app *application) Search(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("item")
	orderBy := r.FormValue("orderBy")
	var page int
	pageStr := r.URL.Query().Get("page")
	page, _ = strconv.Atoi(pageStr)
	if page == 0 {
		page = 1
	}
	metadata, items, err := app.items.Search(r.Context(), query, page, PAGE_SIZE, orderBy)
	if err != nil {
		app.logger.Error("could not get items", "error", err)
	}

	app.logger.Debug("got items", "items", items, "metadata", metadata)
	pages.Pantry(query, orderBy, metadata, items).Render(r.Context(), w)
}

func (app *application) AddItem(w http.ResponseWriter, r *http.Request) {
	itemReq := r.FormValue("item")
	item, price := parseItem(itemReq)
	app.logger.Debug("parsed item", "item", item, "price", price)
	_, err := app.items.Add(r.Context(), item, price)
	if err != nil {
		app.logger.Error("unable to add item", "error", err.Error())
		r = r.WithContext(context.WithValue(r.Context(), components.FlashKey, "Could not add item. Please try again."))
	}
	metadata, items, err := app.items.Search(r.Context(), "", 1, PAGE_SIZE, "recentlyAdded")
	if err != nil {
		app.logger.Error("unable to get items", "error", err.Error())
		if _, ok := r.Context().Value(components.FlashKey).(string); !ok {
			r = r.WithContext(context.WithValue(r.Context(), components.FlashKey, "Could not retreive items"))
		}
	}
	app.logger.Debug("got items", "items", items, "metadata", metadata)
	pages.Pantry("", "timesBought", metadata, items).Render(r.Context(), w)
}

func (app *application) DeleteItem(w http.ResponseWriter, r *http.Request) {
	itemIdStr := flow.Param(r.Context(), "id")
	itemId, _ := strconv.Atoi(itemIdStr)
	err := app.items.Remove(r.Context(), int64(itemId))
	if err != nil {
		app.logger.Error("could not delete item", "id", itemId, "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	app.logger.Info("deleted item", "id", itemId)
}

func parseItem(item string) (string, float32) {
	var itemName []byte
	var priceStart, priceEnd int
	var price float32
	for i := 0; i < len(item); i++ {
		if item[i] == '$' {
			i++
			priceStart = i
			for ; i < len(item) && item[i] != ' '; i++ {
			}
			priceEnd = i
			continue
		}
		itemName = append(itemName, item[i])
	}

	if priceStart != priceEnd {
		p, _ := strconv.ParseFloat(item[priceStart:priceEnd], 32)
		if p > 0 {
			price = float32(p)
		}
	}
	return strings.TrimSpace(string(itemName)), price
}

func (app *application) EditItemPage(w http.ResponseWriter, r *http.Request) {
	itemIdStr := r.URL.Query().Get("id")
	itemId, err := strconv.ParseInt(itemIdStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	item, err := app.items.Get(r.Context(), itemId)
	if err != nil {
		app.logger.Error("could not get item", "id", itemId, "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	pages.EditItem(item).Render(r.Context(), w)
}

func (app *application) EditItem(w http.ResponseWriter, r *http.Request) {
	itemIdStr := flow.Param(r.Context(), "id")
	itemId, err := strconv.ParseInt(itemIdStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	priceStr := r.FormValue("price")
	price, _ := strconv.ParseFloat(priceStr, 32)
	item, err := app.items.Update(r.Context(), itemId, name, float32(price))
	if err != nil {
		app.logger.Error("could not update item", "id", itemId, "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	pages.Item(item).Render(r.Context(), w)
}

func (app *application) AddItemToBasket(w http.ResponseWriter, r *http.Request) {
	itemIdStr := flow.Param(r.Context(), "itemId")
	itemId, err := strconv.ParseInt(itemIdStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	item, err := app.items.Get(r.Context(), itemId)
	if err != nil {
		app.logger.Error("could not get item", "id", itemId, "error", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	_, err = app.basket.AddItem(r.Context(), item)
	if err != nil {
		app.logger.Error("could not add item to basket", "it", itemId, "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	items, err := app.basket.GetItems(r.Context())
	if err != nil {
		app.logger.Error("could not get items", "error", err.Error)
		r = r.WithContext(context.WithValue(r.Context(), components.FlashKey, "Could not get items"))
	}
	pages.GroceryList(items).Render(r.Context(), w)
}

func (app *application) CreateNewItemAndAddToBasket(w http.ResponseWriter, r *http.Request) {
	itemReq := r.FormValue("item")
	itemName, price := parseItem(itemReq)
	app.logger.Debug("parsed item", "item", itemName, "price", price)

	var items models.Basket
	item, err := app.items.Add(r.Context(), itemName, price)
	app.logger.Debug("created new item", "item", item)
	if err != nil {
		app.logger.Error("unable to add item", "error", err.Error())
		r = r.WithContext(context.WithValue(r.Context(), components.FlashKey, "Could not add item. Please try again."))
	} else {
		_, err := app.basket.AddItem(r.Context(), item)
		if err != nil {
			app.logger.Error("unable to add item to basket", "error", err.Error())
			r = r.WithContext(context.WithValue(r.Context(), components.FlashKey, "Could not add item to basket. Please try again."))
		} else {
			basket, err := app.basket.GetItems(r.Context())
			if err != nil {
				app.logger.Error("could not get basket items", "error", err.Error())
				r = r.WithContext(context.WithValue(r.Context(), components.FlashKey, "Could not ge basket"))
			}
			items = basket
		}
	}

	pages.GroceryList(items).Render(r.Context(), w)
}

func (app *application) MarkPurchased(w http.ResponseWriter, r *http.Request) {
	idStr := flow.Param(r.Context(), "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	item, err := app.basket.TogglePurchased(r.Context(), id)
	if err != nil {
		app.logger.Error("could not update basket item status", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	app.logger.Info("updated item status", "item", item)
	pages.BasketItem(item).Render(r.Context(), w)
}

func (app *application) RemoveItemFromBasket(w http.ResponseWriter, r *http.Request) {
	idStr := flow.Param(r.Context(), "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = app.basket.RemoveItem(r.Context(), id)
	if err != nil {
		app.logger.Error("could not remove item", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	basket, err := app.basket.GetItems(r.Context())
	if err != nil {
		app.logger.Error("could not get basket", "error", err.Error())
		r = r.WithContext(context.WithValue(r.Context(), components.FlashKey, "Could not get basket"))
	}
	pages.GroceryList(basket).Render(r.Context(), w)
}

func (app *application) RemoveAllItems(w http.ResponseWriter, r *http.Request) {
	err := app.basket.RemoveAllItems(r.Context())
	if err != nil {
		app.logger.Error("could not remove all items", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	basket, err := app.basket.GetItems(r.Context())
	if err != nil {
		app.logger.Error("could not get basket", "error", err.Error())
		r = r.WithContext(context.WithValue(r.Context(), components.FlashKey, "Could not get basket"))
	}
	pages.GroceryList(basket).Render(r.Context(), w)
}

func (app *application) Suggest(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("item")
	if query == "" {
		return
	}
	_, items, err := app.items.Search(r.Context(), query, 1, SUGGEST_SIZE, "timesBought")
	if err != nil {
		app.logger.Error("could not get suggestion", "query", query, "error", err.Error())
		return
	}
	app.logger.Info("found suggestions", "query", query, "suggestions", items)
	pages.BasketSearch(items).Render(r.Context(), w)
}
