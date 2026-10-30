# GOBubbleCup

GOBubbleCup est un framework Go pour écrire une interface TUI avec peu de code. Vous décrivez des objets, une mémoire et, si besoin, des scripts. Le framework dessine l'écran, déplace le focus et met les listes à jour quand les données changent.

Le nom de code du projet de travail est `high-tui`. Le produit publié est GOBubbleCup.

L'interface de démonstration se lance ainsi :

```bash
go run .
```

Quitter : `q` ou Ctrl+C, en dehors d'une zone de saisie.

## Installation

```bash
go get github.com/palarchsys/GOBubbleCup@beta-0.1
```

Go 1.27 ou plus récent. Le module dépend de Bubble Tea v2 et de Lip Gloss v2.

```go
import "github.com/palarchsys/GOBubbleCup/framework"
```

## Premier programme

```go
package main

import (
    "fmt"

    tea "charm.land/bubbletea/v2"
    "github.com/palarchsys/GOBubbleCup/framework"
)

func main() {
    titre := framework.Title("Bonjour")
    action := framework.Button("Quitter", framework.Call{Run: func() {}})
    grille := framework.Grid(
        framework.Row(titre),
        framework.Row(action),
    )
    app := framework.New(framework.DefaultStyle(), grille)
    if _, err := tea.NewProgram(app).Run(); err != nil {
        fmt.Println(err)
    }
}
```

`Row` aligne des objets. `Grid` empile ces lignes. `New` construit l'application. `tea.NewProgram` l'affiche.

Une légende des touches est ajoutée sous la grille. Vous n'avez pas à la créer.

## Ce que vous construisez

Vous ne choisissez pas les couleurs, les marges ni les cadres. Un seul style les fixe. Vos objets ne portent que leurs données : un titre, des lignes, un texte, une action.

Le focus se déplace avec les flèches entre les zones actives. Entrée ouvre une liste ou un champ, ou déclenche un bouton. Il n'y a pas de boucle en bout de liste ni en bout d'écran.

Une mémoire peut alimenter une liste. Quand vous appelez `Set`, l'écran se redessine et la liste relit ses lignes.

## Style

`framework.DefaultStyle()` est le thème livré. Il reprend le thème sombre de Visual Studio Code.

| Rôle | Couleur |
| Texte | `#d4d4d4` |
| Titres | `#569cd6` |
| Contour au repos | `#6b6b6b` |
| Contour au focus | `#007acc` |
| Contour pendant l'édition | `#9cdcfe` |
| Fond de l'écran et des zones | `#1e1e1e` |

Les mesures sont fixes : une cellule de padding horizontal, pas de ligne vide dans un cadre, une cellule entre deux colonnes. L'objet le plus à droite d'une ligne occupe l'espace restant. La hauteur d'une ligne est celle de son plus grand objet.

## Clavier

| Touche | Effet |
| Flèches | Déplacent le focus. Pas de boucle. |
| Entrée | Ouvre une liste ou un champ, ou déclenche un bouton. |
| Échap | Sort d'une liste multiple ou d'un champ. Le focus reste sur la zone. |
| Ctrl+gauche, Ctrl+droite | Changent d'onglet, depuis n'importe où. |
| `q` | Quitte, sauf dans un champ où c'est une lettre. |
| Ctrl+C | Quitte toujours. |

Dans une liste unique, Entrée choisit la ligne et sort de la liste. Dans une liste multiple, Entrée coche ou décoche et vous restez dans la liste.

## Objets

Les noms ci-dessous sont les constructeurs. Des alias `NewTitle`, `NewListUniq`, etc. existent et font la même chose.

### Titre

```go
framework.Title("high-tui")
```

Bandeau inactif. Il s'étire sur la largeur de sa ligne.

### Catégorie

```go
framework.Category("Catalogue", teintes, pieces)
```

Le premier argument est le sous-titre, écrit sur le bord haut. Les suivants sont les enfants, empilés. La catégorie n'est pas un arrêt de focus : les flèches entrent dans ses enfants.

### Ligne de liste

```go
framework.ListItem("casque", "Casque") // id, texte affiché
```

L'id est la valeur que vous relisez plus tard. Le texte est ce que l'écran montre.

Dans la liste, un `›` marque la ligne survolée quand la liste est ouverte. Un `✓` marque une ligne choisie.

### Liste à choix unique

```go
teintes := framework.ListUniq("Teintes",
    framework.ListItem("cyan", "Cyan"),
    framework.ListItem("magenta", "Magenta"),
)
// Après Entrée sur Magenta : teintes.Value() == "magenta"
// Avant toute confirmation : teintes.Value() == ""
```

Vous pouvez aussi lui donner un objet réactif à la place des lignes. Voir plus bas.

`Reload` oublie le choix. Si la liste est réactive, elle relit la mémoire.

### Liste à choix multiple

```go
pieces := framework.ListMult("Pièces",
    framework.ListItem("casque", "Casque"),
    framework.ListItem("veste", "Veste"),
)
// pieces.Values() renvoie les id cochés, dans l'ordre des lignes.
```

`Reload` enlève les coches.

### Champ d'une ligne

```go
nom := framework.TextInput("Nom", framework.InputText, 24)
port := framework.TextInput("Port", framework.InputNumber, 5)
code := framework.TextInput("Code", framework.InputSecret, 0)

nom.Value()
```

Le label est à gauche, la saisie à droite, sur la même ligne.

| Type | Effet |
| `InputText` | Texte libre. |
| `InputNumber` | Chiffres seulement. |
| `InputSecret` | Affiché avec `•`. |

Le troisième argument est le nombre maximum de caractères. `0` ne limite pas.

Entrée ou Échap referme le champ. `Reload` restaure le texte d'origine.

### Zone de texte

```go
brouillon := framework.TextArea("Brouillon", "", true, 4)
lecture := framework.TextArea("Sortie", "En attente", false, 4)
```

Arguments : titre, texte initial, `true` pour écrire, nombre minimum de lignes visibles. Une barre de défilement n'apparaît que si le texte dépasse.

En écriture, Entrée insère une ligne. En lecture, Entrée ouvre la zone et les flèches haut et bas défilent. `Reload` restaure le texte initial.

### Bouton

```go
framework.Button("Exécuter", uneAction)
```

Entrée déclenche l'action. Le bouton ne s'ouvre pas.

Trois actions existent.

```go
framework.Call{Run: func() {
    memo.Set("4", "Noir")
}}

framework.Reload{Target: teintes}

framework.Exec{
    Script: "echo",
    Args:   []string{"bonjour"},
    Into:   sortie, // zone de texte, ou nil pour prendre le terminal
}
```

`Reload` accepte une liste, un champ, une zone de texte, une catégorie ou une liste d'onglets. Sur une catégorie ou une liste d'onglets, le rechargement se propage aux enfants.

### Onglets

```go
vues := framework.TabList(
    framework.TabItem("fiches", "Fiches"),
    framework.TabItem("notes", "Notes"),
)
vues.SetTabs(
    framework.Tab("fiches", framework.Grid(framework.Row(catalogue))),
    framework.Tab("notes", framework.Grid(framework.Row(journal))),
)
```

Une application ne contient qu'une liste d'onglets. Une seconde arrête le programme.

La barre n'est pas un arrêt de focus. Au lancement, le focus est sur la première zone de l'onglet affiché. Ctrl+gauche et Ctrl+droite changent d'onglet depuis n'importe où, y compris pendant une saisie. S'il n'y a pas d'onglet dans cette direction, la touche ne fait rien. Entrée ne change pas d'onglet.

L'onglet ouvert n'a pas de trait en bas : il rejoint le panneau. Les autres sont des boîtes fermées. Tout le contour reste gris. Seul le libellé de l'onglet ouvert est en couleur de titre.

Le contenu d'un onglet remplit le panneau. Il ne peut pas élargir l'écran au-delà de la largeur déjà décidée par les autres lignes.

## Mémoire

```go
memo := framework.Memory()

memo.Set(framework.Data{
    "id": "1",
    "data": framework.Data{"id": "cyan", "value": "Cyan"},
})
memo.Set("etat", "en attente")

memo.Get("1").Get("data").Get("value").String()   // "Cyan"
memo.Get("etat").Get("data").Get("value").String() // "en attente"
```

`Set(id, texte)` est la forme courte : elle enregistre `data.id` et `data.value`.

`Get` se chaîne. Une clé absente donne une chaîne vide, pas une erreur.

`Set` réveille l'écran. L'ordre d'insertion est conservé.

`ReadOnly()` renvoie la même mémoire en lecture. Un `Set` sur cette vue arrête le programme.

## Liste réactive

```go
teintes := framework.ListUniq("Teintes", framework.ReactiveObject(
    framework.ListItem,
    memo,
    map[string]string{"id": "id", "label": "value"},
))
```

La table se lit ainsi : l'id de la ligne vient de `data.id`, le texte affiché vient de `data.value`.

Chaque dessin relit la mémoire. Un `memo.Set(...)` ajoute ou remplace une ligne. Le choix déjà confirmé reste si son id existe encore.

## Runtime et scripts

Un runtime lance des scripts bash, en arrière-plan, et écrit leur résultat dans sa mémoire.

```go
runtimeA := framework.Runtime(framework.RuntimeConfig{
    Loop:        false,
    Timer:       0,
    Args:        memoArgs,
    StatusRegex: `(success|warning|error)`,
    Functions: func(args framework.Store) {
        framework.Bash(
            "controle1",
            "Contrôle 1",
            "scripts/controle1.sh",
            []string{args.Get("1").Get("data").Get("value").String()},
        )
    },
})
runtimeA.Init()
executer := framework.Button("Exécuter", runtimeA.Start())
```

`Init` enregistre les scripts. Vous pouvez déjà lire leur id, leur libellé et leur chemin. `status` et `return` sont vides. Aucun script ne tourne.

`Start()` ne lance rien au moment où vous écrivez le bouton. Il renvoie l'action du bouton. Le lancement a lieu quand l'utilisateur valide Exécuter. Les scripts partent alors en arrière-plan. Si le runtime tournait déjà, il s'arrête puis repart du début.

`Bash` s'écrit seulement dans `Functions`. Le chemin est relatif au dossier depuis lequel vous avez lancé le programme. Le script reçoit ses arguments en lecture seule. Un délai ou un `sleep` se met dans le script, pas dans le bouton.

Après le départ, le slot se lit ainsi :

```go
runtimeA.Memory.Get("controle1").Get("id").String()
runtimeA.Memory.Get("controle1").Get("label").String()
runtimeA.Memory.Get("controle1").Get("path").String()
runtimeA.Memory.Get("controle1").Get("status").String()
runtimeA.Memory.Get("controle1").Get("return").String()
```

`status` vaut `pending` pendant le script, puis `success`, `warning` ou `error`. `return` est la sortie brute. `Memory` est en lecture seule : c'est le runtime qui l'écrit.

`Status()` du runtime, ce n'est pas le verdict du script. C'est le cycle : `not_initialized`, `pending`, `running`, `terminated`.

`Stop` annule l'exécution et remet ce cycle à `not_initialized`. `Pause` retient le runtime entre deux scripts. `Resume` le relâche. Un `Resume` sans pause ne fait rien.

### Boucle

```go
framework.Runtime(framework.RuntimeConfig{
    Loop:  true,
    Timer: time.Second, // 0 signifie 100 ms
})
```

`Timer` est obligatoire quand `Loop` vaut `true`. À la fin de chaque tour, le runtime attend cette durée, puis recommence. `Stop` coupe l'attente. Sans boucle, `Timer` est ignoré.

### Plusieurs runtimes

Chaque `Runtime(...)` est indépendant. Vous pouvez en lancer plusieurs. Ils écrivent chacun dans leur mémoire.

La regex par défaut est `(success|warning|error)`. Elle s'applique à la sortie du script. Le premier groupe, ou le texte entier, doit être l'un de ces trois mots. Sinon le status du slot est `error`. Une regex invalide arrête le programme au moment de la construction.

Un plantage dans le conteneur termine le cycle et ajoute le texte au journal `Log()`.

## Watcher

Le watcher croise une ou plusieurs mémoires et produit un résultat. Vous écrivez la mise en forme. Le framework rappelle votre fonction quand l'une des mémoires change.

```go
reactiveWatcher := framework.ReactiveWatcher(framework.WatcherConfig{
    Watcher: []framework.Store{controles.Memory, horloge.Memory},
    Functions: func(args []framework.Store) any {
        forme := framework.Memory()
        forme.Set("controle1", args[0].Get("controle1").Get("status").String())
        forme.Set("heure", args[1].Get("horloge").Get("return").String())
        return framework.ReactiveObject(
            framework.ListItem,
            forme,
            map[string]string{"id": "id", "label": "value"},
        )
    },
})
journal := framework.ListUniq("Suivi", reactiveWatcher)
```

`args[0]` est la première mémoire de `Watcher`, `args[1]` la deuxième. Le résultat de la fonction devient la valeur du watcher. `Result()` la relit.

Une `ListUniq` accepte ce watcher si la fonction retourne un `ReactiveObject`, une liste de `ListItem`, ou un autre watcher.

N'écrivez pas dans les mémoires surveillées depuis cette fonction. Écrivez dans une mémoire neuve, comme `forme` ci-dessus. Une liste vide, une mémoire nulle ou une fonction nulle arrêtent le programme.

Chaîne typique : un runtime asynchrone écrit `status` et `return`, le watcher remet ces champs en forme, la liste affiche le résultat, sans boucle d'attente dans votre bouton.

## Démonstration

```bash
go run .
```

Trois onglets :

| Onglet | Contenu |
| Fiches | Teintes réactives et pièces à cocher. |
| Saisie | Nom, port, code secret, brouillon et sortie. |
| Suivi | Verdicts des trois scripts et heure, croisés par un watcher. |

Les boutons :

| Bouton | Effet |
| Exécuter | Lance une fois `scripts/controle1.sh`, `controle2.sh` et `controle3.sh` avec le Nom. |
| Horloge | Relance `scripts/horloge.sh` chaque seconde. |
| Ajouter | Ajoute une teinte `Noir`. |
| Recharger | Oublie le choix des teintes. |
| Echo | Écrit `bonjour` dans la zone Sortie. |

Ctrl+gauche et Ctrl+droite passent d'un onglet à l'autre.

## Limites utiles

- Une seule liste d'onglets par application.
- Les scripts bash sont des fichiers. Leur chemin part du dossier de lancement.
- Le framework peint un thème unique. Il n'y a pas d'option de couleur par objet.
- `q` dans un champ saisit la lettre q.

## Développement

```bash
go test -count=1 ./...
```
