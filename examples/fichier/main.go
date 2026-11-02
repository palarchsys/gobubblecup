// Lecture et écriture du fichier notes.txt, affichées par un watcher.
//
// Lancer depuis la racine du module : go run ./examples/fichier
// Le fichier est créé dans le dossier courant. Il est ignoré par git.
package main

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/palarchsys/gobubblecup/framework"
)

func main() {
	memo := framework.Memory()
	note := framework.TextArea("Note", "une note", true, 6, true)

	// La liste ne lit pas la zone. Elle lit l'entrée « file » de la mémoire.
	apercu := framework.ReactiveWatcher(framework.WatcherConfig{
		Watcher: []framework.Store{memo},
		Functions: func(args []framework.Store) any {
			forme := framework.Memory()
			texte := args[0].Get("file").Get("data").Get("value").String()
			forme.Set("file", uneLigne(texte))
			return framework.ReactiveObject(
				framework.ListItem,
				forme,
				map[string]string{"id": "id", "label": "value"},
			)
		},
	})

	charger := framework.Button("Charger", framework.Call{Run: func() {
		framework.ReadFile{Path: "notes.txt", Into: memo}.Act()
		note.SetText(memo.Get("file").Get("data").Get("value").String())
	}})
	enregistrer := framework.Button("Enregistrer", framework.Call{Run: func() {
		memo.Set("file", note.Text())
		framework.WriteFile{Path: "notes.txt", From: memo}.Act()
	}})

	grille := framework.Grid(
		framework.Row(framework.Title("gobubblecup")),
		framework.Row(note),
		framework.Row(framework.ListUniq("Fichier", apercu)),
		framework.Row(charger, enregistrer),
	)
	style := framework.DefaultStyle()
	style.Height = style.Width * 3 / 4
	app := framework.New(style, grille)
	if _, err := tea.NewProgram(app).Run(); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("fichier:", memo.Get("file").Get("data").Get("value").String())
}

func uneLigne(texte string) string {
	texte = strings.TrimSpace(strings.ReplaceAll(texte, "\n", " "))
	if texte == "" {
		return "vide"
	}
	return texte
}
