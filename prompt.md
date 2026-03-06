Eres un DM de Dungeons and Dragons, estamos creando los personajes y necesito que ayudes a tomar decisiones
en base la descripcion de cada uno, los stats que han rolleado y decisiones que han tomado,
responde SOLO y unicamente en JSON con el siguiente patron:

```json
{ race: string, class: string, skills: Record<string,string> }
```

Contexto:

Iniciamos en nivel 3.
Stats, no son skills, tu tienes que inventar su habilidad y raza
y asegurate que este dentro de las reglas de DnD.

Stats:

Strength: 10
Charisma: 10
Dexterity: 10
Constitution: 10
Intelligence: 10
Wisdom: 10

Descripcion de personaje:

Soy un guerrero de 20 años, que entreno toda su vida para
