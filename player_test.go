// This file was AI-generated
package main

import (
	"strings"
	"testing"
)

func TestPlayerInventoryEmpty(t *testing.T) {
	player := &Player{
		Character: Character{
			Name:      "TestPlayer",
			Equipment: []any{},
		},
	}

	result := player.Inventory()
	expected := "Tu inventario está vacío."

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestPlayerInventoryWithWeapon(t *testing.T) {
	player := &Player{
		Character: Character{
			Name: "TestPlayer",
			Equipment: []any{
				Weapon{
					Item: Item{
						Type: "Weapon",
						Name: "Espada",
						Desc: "Una espada afilada",
					},
					Damage:     "1d8",
					DamageType: "Slashing",
				},
			},
		},
	}

	result := player.Inventory()
	expectedSubstring := "Espada"

	if !strings.Contains(result, expectedSubstring) {
		t.Errorf("Expected inventory to contain '%s', got '%s'", expectedSubstring, result)
	}
}

func TestPlayerInventoryWithArmor(t *testing.T) {
	player := &Player{
		Character: Character{
			Name: "TestPlayer",
			Equipment: []any{
				Armor{
					Item: Item{
						Type: "Armor",
						Name: "Armadura de cuero",
						Desc: "Armadura ligera",
					},
					BaseAC:      12,
					MinStrength: 0,
				},
			},
		},
	}

	result := player.Inventory()
	expectedSubstring := "Armadura de cuero"

	if !strings.Contains(result, expectedSubstring) {
		t.Errorf("Expected inventory to contain '%s', got '%s'", expectedSubstring, result)
	}
}

func TestPlayerInventoryWithItem(t *testing.T) {
	player := &Player{
		Character: Character{
			Name: "TestPlayer",
			Equipment: []any{
				Item{
					Type: "Item",
					Name: "Poción de curación",
					Desc: "Restaura 2d4+2 puntos de vida",
				},
			},
		},
	}

	result := player.Inventory()
	expectedSubstring := "Poción de curación"

	if !strings.Contains(result, expectedSubstring) {
		t.Errorf("Expected inventory to contain '%s', got '%s'", expectedSubstring, result)
	}
}

func TestPlayerSkillsEmpty(t *testing.T) {
	player := &Player{
		Character: Character{
			Name:   "TestPlayer",
			Skills: map[string]string{},
		},
	}

	result := player.Skills()
	expected := "No tienes habilidades definidas."

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestPlayerSkillsWithEntries(t *testing.T) {
	player := &Player{
		Character: Character{
			Name: "TestPlayer",
			Skills: map[string]string{
				"Atletismo":  "Saltar, trepar y nadar",
				"Percepción": "Notar detalles en el entorno",
			},
		},
	}

	result := player.Skills()

	if len(result) == 0 {
		t.Error("Expected skills output to not be empty")
	}

	expectedSubstring := "Atletismo"

	if !strings.Contains(result, expectedSubstring) {
		t.Errorf("Expected skills to contain '%s', got '%s'", expectedSubstring, result)
	}
}
