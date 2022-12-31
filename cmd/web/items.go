package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"trolly.hunterwilkins.dev/internal/models"
	"trolly.hunterwilkins.dev/internal/validator"
)

type itemAddForm struct {
	Item                string `form:"item"`
	InCart              bool   `form:"inCart"`
	validator.Validator `form:"-"`
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	id, _ := app.sessionManager.Get(r.Context(), "authenticatedUserID").(int)

	items, err := app.items.GetAll(id, "", true)
	if err != nil {
		app.serverError(w, err)
		return
	}

	var total float32
	for _, item := range items {
		total += item.Price
	}

	data := app.newTemplateData(r)
	data.Form = itemAddForm{}
	data.Items = items
	data.Total = total
	app.render(w, http.StatusOK, "home.html", data)
}

func (app *application) pantry(w http.ResponseWriter, r *http.Request) {
	uid, _ := app.sessionManager.Get(r.Context(), "authenticatedUserID").(int)

	items, err := app.items.GetAll(uid, "", false)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Form = itemAddForm{}
	data.Items = items
	app.render(w, http.StatusOK, "pantry.html", data)
}

func (app *application) addHomeItem(w http.ResponseWriter, r *http.Request) {
	app.addItem(w, r, "home.html", "/")
}

func (app *application) addPantryItem(w http.ResponseWriter, r *http.Request) {
	app.addItem(w, r, "pantry.html", "/pantry")
}

func (app *application) addItem(w http.ResponseWriter, r *http.Request, rerender, redirect string) {
	var form itemAddForm

	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Item), "item", "Item name cannot be blank")
	parts := strings.Split(form.Item, "$")

	var price float64
	if len(parts) == 2 {
		convPrice, err := strconv.ParseFloat(parts[1], 32)
		price = convPrice
		if err != nil {
			form.AddFieldError("item", fmt.Sprintf("%q is not a valid price", parts[1]))
		} else if price < 0 {
			form.AddFieldError("item", "Price cannot be less than 0")
		}
	}

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, rerender, data)
		return
	}
	name := strings.TrimSpace(parts[0])

	app.logger.Println(name)
	uid, _ := app.sessionManager.Get(r.Context(), "authenticatedUserID").(int)
	existingItem, err := app.items.GetByName(uid, name)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.logger.Println(existingItem)
	if existingItem != nil {
		existingItem.InCart = form.InCart
		err = app.items.Update(existingItem)
		if err != nil {
			app.serverError(w, err)
			return
		}
	} else {
		item := models.Item{
			Name:   name,
			Price:  float32(math.Round(price*100) / 100),
			InCart: form.InCart,
		}

		err = app.items.Insert(uid, &item)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}

	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

func (app *application) search(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	searchFor := params.ByName("query")
	id, _ := app.sessionManager.Get(r.Context(), "authenticatedUserID").(int)

	items, err := app.items.GetAll(id, searchFor, false)
	if err != nil {
		app.serverError(w, err)
		return
	}

	json.NewEncoder(w).Encode(items)
}

type updateItemForm struct {
	Name                string  `form:"name"`
	Price               string  `form:"price"`
	ConvPrice           float64 `form:"-"`
	InCart              bool    `form:"inCart"`
	Purchased           bool    `form:"purchased"`
	validator.Validator `form:"-"`
}

func (app *application) updateItem(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	item, err := app.items.Get(id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.ItemID = id
	data.Form = updateItemForm{
		Name:  item.Name,
		Price: strconv.FormatFloat(float64(item.Price), 'f', 2, 32),
	}
	app.render(w, http.StatusOK, "update.html", data)
}

func (app *application) updateHomeItems(w http.ResponseWriter, r *http.Request) {
	var form updateItemForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.logger.Println(err)
		app.clientError(w, http.StatusBadRequest)
		return
	}

	app.updateItemPost(w, r, &form)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) updatePantryItems(w http.ResponseWriter, r *http.Request) {
	var form updateItemForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	app.updateItemPost(w, r, &form)

	http.Redirect(w, r, "/pantry", http.StatusSeeOther)
}

func (app *application) updateItemForm(w http.ResponseWriter, r *http.Request) {
	var form updateItemForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Name), "name", "Name cannot be empty")
	if form.Price != "" {
		convPrice, err := strconv.ParseFloat(form.Price, 32)
		form.ConvPrice = convPrice
		if err != nil {
			form.AddFieldError("price", fmt.Sprintf("%q is not a number", form.Price))
		} else if convPrice < 0 {
			form.AddFieldError("price", "Price cannot be less than 0")
		}
	}

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "update.html", data)
		return
	}

	app.updateItemPost(w, r, &form)

	http.Redirect(w, r, "/pantry", http.StatusSeeOther)
}

func (app *application) updateItemPost(w http.ResponseWriter, r *http.Request, form *updateItemForm) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}
	item, err := app.items.Get(id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	if form.Name != "" {
		item.Name = form.Name
	}

	if form.Price != "" {
		item.Price = float32(math.Round(form.ConvPrice*100) / 100)
	}

	if r.Form.Get("purchased") != "" && form.Purchased != item.Purchased {
		item.Purchased = form.Purchased
		if item.Purchased {
			item.TimesPurchased++
		} else {
			item.TimesPurchased--
		}
	}

	if r.Form.Get("inCart") != "" && form.InCart != item.InCart {
		item.InCart = form.InCart
		if !item.InCart {
			item.Purchased = false
		} else {
			item.LastAdded = time.Now()
		}
	}

	err = app.items.Update(item)
	if err != nil {
		app.serverError(w, err)
		return
	}
}

func (app *application) deleteItem(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	err = app.items.Delete(id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, "/pantry", http.StatusSeeOther)
}

func (app *application) removeAll(w http.ResponseWriter, r *http.Request) {
	uid, _ := app.sessionManager.Get(r.Context(), "authenticatedUserID").(int)

	err := app.items.RemoveAllFromCart(uid)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
