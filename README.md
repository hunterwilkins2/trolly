# Trolly



Trolly is a mobile site designed for managing grocery lists and estimating grocery costs. Trolly tracks users purchases to allow users to quicky add recently and commonly bought items. Trolly is powered by Go and htmx and supports inline editing of items and prices, active search to quickly add and remove items, and inline parsing of user input to quickly add items without filling out forms, making editing and managing grocery lists from your phone a breeze.

![Trolly Screenshot](https://github.com/hunterwilkins2/trolly/blob/master/img/Screenshot%20from%202024-04-07%2014-11-15.png)

## Requirements

* Go v1.20+
* Tailwindcss
* docker

## Run

1. Create the database with `make db && make migrate`
2. Run the application with live reloading with `make run/live`
3. Open http://localhost:4000 to view the application

## Build

1. Create the datebase with `make db && make migrate`
2. Build with `make build`
3. Run the binary with `./bin/trolly`
4. Open http://localhost:4000 to view the application

## Helm

Deploy with kubernetes using helm

```
helm install trolly ../homelab-charts/ -n trolly --create-namespace -f devops/values.yaml 
```
