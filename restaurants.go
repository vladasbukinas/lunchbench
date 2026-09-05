package main

import (
	"fmt"
	"strings"
)

type Special struct {
	Name        string
	Description string
	Price       float64
}

type Restaurant struct {
	Name        string
	Cuisine     string
	Description string
	Specials    []Special
}

func menu() []Restaurant {
	return []Restaurant{
		{
			Name:        "The Gilded Fork",
			Cuisine:     "New American",
			Description: "An upscale bistro serving seasonal plates in a converted brick warehouse downtown.",
			Specials: []Special{
				{Name: "Short Rib Melt", Description: "Braised short rib, gruyere, caramelized onion on sourdough with truffle fries.", Price: 16.50},
				{Name: "Harvest Grain Bowl", Description: "Farro, roasted squash, kale, pickled onion, maple-tahini dressing.", Price: 12.00},
				{Name: "Lobster Roll Sliders", Description: "Butter-poached lobster, brioche buns, old bay aioli, side of slaw.", Price: 18.75},
			},
		},
		{
			Name:        "Pho Saigon Corner",
			Cuisine:     "Vietnamese",
			Description: "A family-run shop ladling broth simmered for eighteen hours, plus banh mi to go.",
			Specials: []Special{
				{Name: "Lunch Pho Combo", Description: "Small beef pho with brisket and meatballs plus a spring roll.", Price: 11.25},
				{Name: "Lemongrass Chicken Banh Mi", Description: "Grilled lemongrass chicken, pickled carrot, cilantro on fresh baguette.", Price: 8.50},
				{Name: "Tofu Vermicelli Bowl", Description: "Crispy tofu, vermicelli, lettuce, herbs, nuoc cham on the side.", Price: 9.75},
			},
		},
		{
			Name:        "Nonna's Table",
			Cuisine:     "Italian",
			Description: "Hand-rolled pasta and slow ragus from a third-generation family kitchen.",
			Specials: []Special{
				{Name: "Pasta e Fagioli Lunch", Description: "Minestra with borlotti beans, ditalini, rosemary oil, crusty bread.", Price: 9.00},
				{Name: "Cacio e Pepe Taster", Description: "Tonnarelli tossed in pecorino romano and cracked black pepper.", Price: 12.50},
				{Name: "Chicken Parmigiana Wedge", Description: "Breaded cutlet, san marzano marinara, fior di latte on ciabatta.", Price: 13.25},
			},
		},
		{
			Name:        "Tandoor Junction",
			Cuisine:     "Indian",
			Description: "Clay-oven classics and thali plates with fresh naan made to order.",
			Specials: []Special{
				{Name: "Butter Chicken Thali", Description: "Butter chicken, dal, rice, naan, raita, and gulab jamun.", Price: 13.95},
				{Name: "Chana Masala Wrap", Description: "Spiced chickpeas, mint chutney, pickled onion in a roti wrap.", Price: 8.25},
				{Name: "TandooriPaneer Bowl", Description: "Charred paneer, basmati, grilled peppers, cilantro chutney.", Price: 11.50},
			},
		},
		{
			Name:        "La Barbacoa",
			Cuisine:     "Mexican",
			Description: "Slow-roasted meats, handmade tortillas, and a salsa bar with six salsas.",
			Specials: []Special{
				{Name: "Barbacoa Torta", Description: "Shredded beef, avocado, chipotle mayo, queso fresco on telera.", Price: 10.50},
				{Name: "Three Taco Lunch", Description: "Choice of asada, al pastor, or birria with rice and beans.", Price: 9.95},
				{Name: "Chile Relleno Plate", Description: "Fire-roasted poblano, queso oaxaca, tomato rice, frijoles.", Price: 12.25},
			},
		},
		{
			Name:        "Seaside Poke",
			Cuisine:     "Hawaiian-Japanese",
			Description: "Build-your-own poke bowls with fish delivered twice daily.",
			Specials: []Special{
				{Name: "Ahi Classic Bowl", Description: "Shoyu ahi, sushi rice, edamame, cucumber, sesame, furikake.", Price: 13.50},
				{Name: "Spicy Salmon Mini", Description: "Small bowl with salmon, sriracha aioli, avocado, seaweed salad.", Price: 10.75},
				{Name: "Tofu Garden Bowl", Description: "Marinated tofu, brown rice, mango, carrot, miso-ginger dressing.", Price: 9.50},
			},
		},
	}
}

func renderMenu(rs []Restaurant) string {
	var b strings.Builder

	for _, r := range rs {
		fmt.Fprintf(&b, "%s (%s): %s\n", r.Name, r.Cuisine, r.Description)
		for _, s := range r.Specials {
			fmt.Fprintf(&b, "  - %s ($%.2f): %s\n", s.Name, s.Price, s.Description)
		}
	}

	return b.String()
}
