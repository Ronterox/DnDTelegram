// This file was AI-generated
package main

import "fmt"

func (p *Player) Inventory() string {
	if len(p.Equipment) == 0 {
		return "Tu inventario está vacío."
	}

	output := "=== Inventario ===\n"
	for i, item := range p.Equipment {
		switch v := item.(type) {
		case Weapon:
			output += fmt.Sprintf("%d. %s (%s) - Daño: %s %s\n", i+1, v.Name, v.Type, v.Damage, v.DamageType)
		case Armor:
			output += fmt.Sprintf("%d. %s (%s) - AC: %d\n", i+1, v.Name, v.Type, v.BaseAC)
		case Item:
			output += fmt.Sprintf("%d. %s - %s\n", i+1, v.Name, v.Desc)
		default:
			output += fmt.Sprintf("%d. %v\n", i+1, item)
		}
	}
	return output
}

func (p *Player) Skills() string {
	if len(p.Character.Skills) == 0 {
		return "No tienes habilidades definidas."
	}

	output := "=== Habilidades ===\n"
	for skill, description := range p.Character.Skills {
		output += fmt.Sprintf("- %s: %s\n", skill, description)
	}
	return output
}
