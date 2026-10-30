// Démo du framework.
//
// Quitter : q ou Ctrl+C, hors d'une zone.
// Cet écran assemble les objets. Il ne décrit pas leur dessin.
// Chaque groupe dit quelle donnée est préparée, et qui la lit.
package main

import (
	"fmt"
	"strconv"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/palarchsys/GOBubbleCup/framework"
)

func main() {

	// -----------------------------------------------------------------
	// Titre
	// -----------------------------------------------------------------
	// Bandeau inactif. Il s'étire sur la largeur de sa ligne.

	titre := framework.Title("high-tui")

	// -----------------------------------------------------------------
	// Catalogue : mémoire, liste réactive, liste à coches
	// -----------------------------------------------------------------
	// Une entrée de mémoire a un id de rangement et un enregistrement data.
	// data.id devient l'id de la ligne. data.value devient le texte affiché.

	memoTeintes := framework.Memory()
	memoTeintes.Set(framework.Data{"id": "1", "data": framework.Data{"id": "cyan", "value": "Cyan"}})
	memoTeintes.Set(framework.Data{"id": "2", "data": framework.Data{"id": "magenta", "value": "Magenta"}})
	memoTeintes.Set(framework.Data{"id": "3", "data": framework.Data{"id": "jaune", "value": "Jaune"}})

	// La liste n'a pas ses lignes en dur. Elle les reconstruit depuis la mémoire.
	// Après Entrée sur Magenta, Value() vaut "magenta".
	teintes := framework.ListUniq("Teintes", framework.ReactiveObject(
		framework.ListItem,
		memoTeintes,
		map[string]string{"id": "id", "label": "value"},
	))

	// Lignes fixes. L'id sert à Values(), le second texte est affiché.
	pieces := framework.ListMult("Pièces",
		framework.ListItem("casque", "Casque"),
		framework.ListItem("veste", "Veste"),
		framework.ListItem("bottes", "Bottes"),
	)

	// La catégorie empile les deux listes sous un titre. Elle n'est pas un focus.
	catalogue := framework.Category("Catalogue", teintes, pieces)

	// -----------------------------------------------------------------
	// Saisie : les trois types de champ, et deux zones de texte
	// -----------------------------------------------------------------
	// Label à gauche, saisie à droite.
	// Le troisième argument est le nombre maximum de caractères. 0 ne borne pas.

	nom := framework.TextInput("Nom", framework.InputText, 24)
	port := framework.TextInput("Port", framework.InputNumber, 5)
	code := framework.TextInput("Code", framework.InputSecret, 0)

	// true : on peut écrire. Entrée insère une ligne, Échap sort.
	brouillon := framework.TextArea("Brouillon", "", true, 4)
	// false : lecture seule. Le bouton Echo écrira ici.
	sortie := framework.TextArea("Sortie", "En attente", false, 4)

	saisie := framework.Category("Saisie", nom, port, code, brouillon, sortie)

	// -----------------------------------------------------------------
	// Contrôles : un runtime lancé une fois, trois scripts
	// -----------------------------------------------------------------
	// Init pose id, label et path dans la mémoire, sans lancer les scripts.
	// Start, au clic, relit Functions puis lance controle1, controle2, controle3.
	// Chaque script reçoit le Nom et écrit son verdict : success, warning ou error.

	memoArgs := framework.Memory()
	controles := framework.Runtime(framework.RuntimeConfig{
		Loop:        false,
		StatusRegex: `(success|warning|error)`,
		Args:        memoArgs,
		Functions: func(args framework.Store) {
			args.Set(framework.Data{"id": "1", "data": framework.Data{"id": "nom", "value": nom.Value()}})
			valeur := args.Get("1").Get("data").Get("value").String()
			framework.Bash("controle1", "Contrôle 1", "scripts/controle1.sh", []string{valeur})
			framework.Bash("controle2", "Contrôle 2", "scripts/controle2.sh", []string{valeur})
			framework.Bash("controle3", "Contrôle 3", "scripts/controle3.sh", []string{valeur})
		},
	})
	controles.Init()

	// -----------------------------------------------------------------
	// Horloge : un runtime en boucle
	// -----------------------------------------------------------------
	// Timer est obligatoire quand Loop vaut true.
	// 0 voudrait dire 100 ms. Ici une seconde, pour lire l'heure.
	// Le sleep de fin de tour est celui du runtime. Le script ne fait qu'écrire l'heure.

	horloge := framework.Runtime(framework.RuntimeConfig{
		Loop:        true,
		Timer:       time.Second,
		StatusRegex: `(success|warning|error)`,
		Functions: func(args framework.Store) {
			framework.Bash("horloge", "Horloge", "scripts/horloge.sh", nil)
		},
	})
	horloge.Init()

	// Un seul watcher suit les deux mémoires.
	// args[0] est les contrôles, args[1] est l'horloge.
	// La fonction croise les deux et retourne une seule liste réactive.
	suiviCroise := framework.ReactiveWatcher(framework.WatcherConfig{
		Watcher: []framework.Store{controles.Memory, horloge.Memory},
		Functions: func(args []framework.Store) any {
			controles := args[0]
			horloge := args[1]
			forme := framework.Memory()
			for _, id := range []string{"controle1", "controle2", "controle3"} {
				texte := controles.Get(id).Get("status").String()
				if texte == "" {
					texte = controles.Get(id).Get("label").String()
				}
				forme.Set(id, texte)
			}
			heure := premiereLigne(horloge.Get("horloge").Get("return").String())
			if heure == "" {
				heure = "en attente"
			}
			forme.Set("heure", heure)
			return framework.ReactiveObject(
				framework.ListItem,
				forme,
				map[string]string{"id": "id", "label": "value"},
			)
		},
	})
	suivi := framework.ListUniq("Suivi", suiviCroise)

	// -----------------------------------------------------------------
	// Boutons
	// -----------------------------------------------------------------
	// Exécuter et Horloge ne font qu'une chose : lancer leur runtime.
	// Ajouter écrit une teinte. La liste réactive l'affiche.
	// Recharger oublie le choix des teintes. Les lignes ajoutées restent.
	// Echo lance une commande et écrit sa sortie dans la zone Sortie.

	executer := framework.Button("Exécuter", controles.Start())
	lancerHorloge := framework.Button("Horloge", horloge.Start())

	suivant := 4
	ajouter := framework.Button("Ajouter", framework.Call{Run: func() {
		id := strconv.Itoa(suivant)
		suivant++
		memoTeintes.Set(id, "Noir "+id)
	}})

	recharger := framework.Button("Recharger", framework.Reload{Target: teintes})

	echo := framework.Button("Echo", framework.Exec{
		Script: "echo",
		Args:   []string{"bonjour"},
		Into:   sortie,
	})

	// -----------------------------------------------------------------
	// Onglets et grille
	// -----------------------------------------------------------------
	// Une seule liste d'onglets. Ctrl+flèche change d'onglet depuis n'importe où.
	// Le focus de départ est la première zone de l'onglet Fiches.

	vues := framework.TabList(
		framework.TabItem("fiches", "Fiches"),
		framework.TabItem("saisie", "Saisie"),
		framework.TabItem("suivi", "Suivi"),
	)
	vues.SetTabs(
		framework.Tab("fiches", framework.Grid(framework.Row(catalogue))),
		framework.Tab("saisie", framework.Grid(framework.Row(saisie))),
		framework.Tab("suivi", framework.Grid(framework.Row(suivi))),
	)

	grille := framework.Grid(
		framework.Row(titre),
		framework.Row(vues),
		framework.Row(executer, lancerHorloge, ajouter, recharger, echo),
	)

	if _, err := tea.NewProgram(framework.New(framework.DefaultStyle(), grille)).Run(); err != nil {
		fmt.Println("Erreur:", err)
	}
}

// premiereLigne garde le texte avant le premier saut de ligne.
// La sortie du script est "12:00:00\nsuccess" : l'écran ne montre que l'heure.
func premiereLigne(texte string) string {
	for i := 0; i < len(texte); i++ {
		if texte[i] == '\n' {
			return texte[:i]
		}
	}
	return texte
}
