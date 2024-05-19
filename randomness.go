package main

import "math/rand"

var (
	colors = []string{
		"Red",
		"Green",
		"Blue",
		"Yellow",
		"Purple",
		"Orange",
		"Pink",
		"Brown",
		"Black",
		"White",
		"Gray",
		"Turquoise",
		"Magenta",
		"Gold",
		"Silver",
		"Lime",
		"Olive",
		"Maroon",
		"Teal",
		"Navy",
		"Coral",
		"Mint",
		"Lavender",
		"Indigo",
		"Aquamarine",
		"Azure",
		"Crimson",
		"Fuchsia",
		"Peach",
		"Periwinkle",
		"Plum",
		"Sand",
		"Sienna",
		"Tan",
		"Violet",
		"Amber",
		"Charcoal",
		"Chartreuse",
		"Cyan",
		"Emerald",
		"Jade",
		"Rose",
		"Sapphire",
		"Scarlet",
		"Sepia",
		"Tangerine",
		"Topaz",
		"Chocolate",
		"Eggplant",
		"Glacier",
		"Ivory",
		"Jasper",
		"Latte",
		"Lemon",
		"Lilac",
		"Opal",
		"Pear",
		"Ruby",
		"Slate",
		"SteelBlue",
		"Tuscan",
		"Ultramarine",
		"Wine",
	}

	space = []string{
		"Andromeda",
		"Whirlpool",
		"Triangulum",
		"Canis",
		"Ursa",
		"Carina",
		"Centaurus",
		"Antlia",
		"Pisces",
		"Seahorse",
		"Medusa",
		"Dragonfly",
		"Phoenix",
		"Orion",
		"Pegasus",
		"Hercules",
		// Solar System
		"Mercury",
		"Venus",
		"Earth",
		"Mars",
		"Jupiter",
		"Saturn",
		"Uranus",
		"Neptune",
		"Pluto",

		// Star Wars
		"Tatooine",
		"Alderaan",
		"Dagobah",
		"Hoth",
		"Endor",
		"Mustafar",
		"Naboo",
		"Kamino",
		"Geonosis",

		// Star Trek
		"Vulcan",
		"Kronos",
		"Risa",
		"Ferenginar",
		"Romulus",
		"Bajor",
		"Betazed",
		"Cardassia",

		// Other Shows
		"Arrakis",   // From Dune
		"Caprica",   // From Battlestar Galactica
		"Krypton",   // From Superman
		"Gallifrey", // From Doctor Who
		"Pandora",   // From Avatar
		"Terminus",  // From Foundation
		"Trantor",   // From Foundation
		"Namek",     // From Dragon Ball Z
		"Cybertron", // From Transformers
		"Aegis",     // From Killjoys
	}
)

func randomColor() string {
	return colors[rand.Intn(len(colors))]
}

func randomSpace() string {
	return space[rand.Intn(len(space))]
}
