# gobubblecup

gobubblecup est un framework Go pour écrire une interface en terminal. Vous décrivez des objets, une mémoire et, si besoin, des scripts. Le framework dessine l'écran, déplace le focus et met les listes à jour quand les données changent.

Le module publié est `github.com/palarchsys/gobubblecup`. Le paquet à importer est `framework`. La version de ce guide est `v1.1.0-beta` (étiquette `beta-1.1`).

Vous ne choisissez pas les couleurs, les marges ni les cadres. `DefaultStyle` les fixe. Vos objets ne portent que leurs données.

## Prérequis

- Go **1.27.3** ou plus récent. Vérifiez avec `go version`.
- Un terminal qui accepte les couleurs.
- Git, pour `go get`.

L'archive officielle de Go 1.27.3 pour Linux amd64 est [go1.27.3.linux-amd64.tar.gz](https://go.dev/dl/go1.27.3.linux-amd64.tar.gz). Après l'installation, `go version` doit afficher `go1.27.3` ou une version plus récente. Les autres systèmes sont sur [go.dev/dl](https://go.dev/dl/).

Le framework dépend de Bubble Tea v2 et de Lip Gloss v2. `go get` les ajoute au projet. Vous n'avez pas à les installer à part.

## Créer un projet

Dans un dossier vide, hors de ce dépôt :

```bash
mkdir mon-tui
cd mon-tui
go mod init example.com/mon-tui
go get github.com/palarchsys/gobubblecup@v1.1.0-beta
```

`go mod init` crée `go.mod`. Le chemin `example.com/mon-tui` est le nom de votre module : remplacez-le par le vôtre.

`go get` télécharge le framework et écrit la dépendance. `go.mod` contient alors une ligne de cette forme :

```
require github.com/palarchsys/gobubblecup v1.1.0-beta
```

`go.sum` enregistre les sommes de contrôle. Conservez les deux fichiers.

## Le programme

Enregistrez ce fichier sous `main.go`, à côté de `go.mod`. Le dépôt reprend le même programme dans `examples/hello/main.go`.

```go
package main

import (
    "fmt"

    tea "charm.land/bubbletea/v2"
    "github.com/palarchsys/gobubblecup/framework"
)

func main() {
    memo := framework.Memory()
    memo.Set("1", "Cyan")
    memo.Set("2", "Magenta")

    colors := framework.ListUniq("Colors", framework.ReactiveObject(
        framework.ListItem,
        memo,
        map[string]string{"id": "id", "label": "value"},
    ))
    name := framework.TextInput("Name", framework.InputText, 24, true)
    add := framework.Button("Add", framework.Call{Run: func() {
        memo.Set("3", "Yellow")
    }})

    grid := framework.Grid(
        framework.Row(framework.Title("gobubblecup")),
        framework.Row(framework.Category("Catalog", colors)),
        framework.Row(name, add),
    )
    style := framework.DefaultStyle()
    style.Height = style.Width * 3 / 4
    app := framework.New(style, grid)
    if _, err := tea.NewProgram(app).Run(); err != nil {
        fmt.Println(err)
        return
    }
    fmt.Println("choice:", colors.Value())
    fmt.Println("name:", name.Value())
}
```

`Row` aligne des objets. `Grid` empile ces lignes. `New` construit l'application. `tea.NewProgram` l'affiche et bloque jusqu'à la sortie.

Une légende des touches est ajoutée sous la grille. Vous n'avez pas à la créer.

## Compiler

Depuis le dossier qui contient `main.go` et `go.mod` :

```bash
go mod tidy
go build -o mon-tui .
./mon-tui
```

`go mod tidy` complète `go.mod` et `go.sum` avec Bubble Tea, Lip Gloss et leurs dépendances.

`go build -o mon-tui .` compile le paquet `main` du dossier courant et écrit le binaire `mon-tui` à cet endroit. Sans `-o`, le binaire prend le nom du dossier.

`./mon-tui` lance l'interface et efface le terminal. À la sortie, l'écran précédent revient. Quitter : Ctrl+C. Les scripts encore en cours sont arrêtés. Après la fermeture, le programme affiche l'id choisi dans la liste et le texte du champ.

Pour lancer sans écrire de binaire :

```bash
go run .
```

Pour recompiler après un changement de `main.go`, relancez `go build -o mon-tui .`. Un changement du framework demande un nouveau `go get` vers la version voulue, puis `go build`.

### Depuis ce dépôt

Si vous avez cloné `github.com/palarchsys/gobubblecup` :

```bash
git clone https://github.com/palarchsys/gobubblecup.git
cd gobubblecup
go test -count=1 ./framework
go build -o hello ./examples/hello
./hello
```

`go test` vérifie le framework. `go build -o hello ./examples/hello` produit le même exemple que ci-dessus. Ce dépôt n'a pas de `package main` à la racine : la compilation d'une application se fait dans votre module, ou dans `examples/hello`.

Les autres programmes sont dans `examples/`. Le sommaire et la commande de chacun sont dans [examples/README.md](examples/README.md).

### Autre système

Le binaire produit sur cette machine vise l'OS et l'architecture courants. Pour un autre couple :

```bash
GOOS=linux GOARCH=amd64 go build -o mon-tui .
GOOS=windows GOARCH=amd64 go build -o mon-tui.exe .
```

`go env GOOS GOARCH` affiche la cible par défaut.

## Clavier

Le focus parcourt les zones actives de la grille affichée, de haut-gauche vers droite-bas. Les objets inactifs sont sautés. La barre d'onglets n'est pas un arrêt.

| Touche | Effet |
| --- | --- |
| ← → | Zone précédente ou suivante. Après la dernière, retour à la première. |
| ↑ ↓ | Item précédent ou suivant dans la liste qui a le focus. Bornés au premier et au dernier. |
| Entrée | Choisit ou décoche l'item focalisé, ou déclenche le bouton. Dans un champ d'une ligne, elle ne change pas le focus. |
| Échap | Ferme une zone de texte ouverte. |
| Tab | Onglet suivant. Après le dernier, retour au premier. Un seul onglet : la touche ne fait rien. |
| Retour arrière | Efface le caractère à gauche du curseur. |
| Suppr | Efface le caractère sous le curseur. |
| Inser | Bascule entre insertion et écrasement dans un champ. |
| Ctrl+V | Colle le presse-papier dans le champ ou la zone éditable. |
| Ctrl+C | Quitte, y compris dans un champ, et arrête les scripts en cours. |

Une liste ne s'ouvre pas. Dès qu'elle a le focus, haut et bas déplacent son item. Entrée confirme cet item. La liste se souvient de la ligne et la retrouve quand le focus revient. Au départ, c'est la première ligne. Cette ligne garde un fond `#37373d`, même quand la liste n'a plus le focus. Il n'y a pas de marqueur `›`. Un `✓` marque une ligne choisie.

Dans un champ d'une ligne, gauche et droite déplacent le curseur. Au bord, elles passent à la zone voisine. Le curseur recouvre le caractère : il n'ajoute pas de colonne. En fin de texte, il occupe une colonne après le dernier caractère. Haut et bas ne quittent pas le champ.

Dans une zone de texte éditable, Entrée insère une ligne et Échap referme. En lecture, Entrée ouvre, haut et bas défilent, Entrée ou Échap referment.

## Style

`framework.DefaultStyle()` reprend le thème sombre de Visual Studio Code.

| Rôle | Couleur |
| --- | --- |
| Texte | `#d4d4d4` |
| Titres | `#569cd6` |
| Contour au repos | `#6b6b6b` |
| Contour au focus | `#007acc` |
| Contour pendant l'édition d'une zone de texte | `#9cdcfe` |
| Fond de l'écran et des zones | `#1e1e1e` |

Les mesures sont fixes : une cellule de padding horizontal, aucune ligne vide dans un cadre, une cellule entre deux colonnes. L'objet le plus à droite d'une ligne occupe l'espace restant. Sans `Height`, une ligne a la hauteur de son plus grand objet. Une fenêtre connue ne fait alors que réduire ce qui dépasse : le contenu court ne descend pas jusqu'en bas. Une `Height` positive est le format de l'écran, légende comprise. Les listes et les zones se partagent ce qui reste sous la légende, sans aller plus bas que ce format. Deux listes d'une même ligne se partagent cette hauteur. La liste n'en montre qu'une fenêtre : le titre reste, et le curseur aussi. La zone garde au moins son nombre de lignes, et prend la hauteur encore libre dans le format. Les démos posent `Height` aux trois quarts de la largeur par défaut.

`framework.Color("#ffee00")` produit une couleur pour une pastille. Un status de slot s'écrit tel quel : `pending`, `success`, `warning` ou `error`. Le thème choisit la teinte. Le thème des cadres ne se change pas objet par objet.

## Objets

Chaque constructeur court a un alias `New…` qui fait la même chose. `NewRuntime` est l'exception : il ne reprend pas `Runtime`.

### Titre

```go
framework.Title("gobubblecup")
```

Bandeau inactif. Il s'étire sur la largeur de sa ligne. Il ne prend pas le focus.

### Catégorie

```go
framework.Category("Catalogue", teintes, pieces)
```

Le premier argument est le sous-titre, écrit sur le bord haut. Les suivants sont les enfants, empilés bord à bord. La catégorie n'est pas un arrêt de focus : les flèches entrent dans ses enfants actifs. Un enfant nul arrête le programme. `Reload` se propage aux enfants qui savent se recharger.

### Ligne de liste

```go
framework.ListItem("casque", "Casque")
framework.ListItem("casque", "Casque").WithDot("#ffee00")
```

Deux arguments : l'id, puis le texte affiché. L'id est la valeur que vous relisez. `WithDot` pose une pastille ronde à droite. La couleur est un `#rrggbb`, ou l'un des mots `pending`, `success`, `warning`, `error`. Une couleur vide, ou un autre mot, ne dessine rien. Une liste désactivée peint ces quatre mots en gris.

Dans la ligne, la première colonne porte `✓` si l'item est choisi, sinon un espace. Elle est alignée sur le premier caractère du titre.

### Liste à choix unique

```go
teintes := framework.ListUniq("Teintes",
    framework.ListItem("cyan", "Cyan"),
    framework.ListItem("magenta", "Magenta"),
)
```

Le premier argument est le titre. Les suivants sont des `ListItem`, ou un seul `ReactiveObject`, ou un seul `ReactiveWatcher`.

`Value()` renvoie l'id confirmé par Entrée. Il est vide avant le premier choix. Un second Entrée sur le même item le laisse choisi. Un nouveau choix remplace le précédent.

`Focus()` renvoie l'id de la ligne focalisée, ou `""` si la liste est vide. `SetFocus(id)` place le curseur sur cet id. Un id absent ne le déplace pas. `SetDot(id, couleur)` change la pastille. Sur une liste réactive qui mappe `dot`, cette écriture passe par la mémoire.

`Bind(mémoire, clé)` écrit cet id focalisé, tout de suite, sous `data.value`. Haut, bas, `SetFocus`, `Reload`, et un id qui disparaît de la liste, réécrivent la même clé. Le même id n'est pas réécrit. Entrée ne publie pas : `Value()` reste le choix confirmé.

`Reload` oublie le choix et revient à la première ligne. Sur une liste réactive, les entrées reviennent à celles présentes à la création de l'objet.

### Liste à choix multiple

```go
pieces := framework.ListMult("Pièces",
    framework.ListItem("casque", "Casque"),
    framework.ListItem("veste", "Veste"),
)
```

Mêmes lignes qu'une liste unique, sans objet réactif. `Values()` renvoie les id cochés, dans l'ordre des lignes. Entrée coche ou décoche l'item focalisé. `Focus`, `SetFocus`, `SetDot` et `Bind` suivent la liste unique. Entrée ne change pas la clé de `Bind`. `Reload` enlève les coches et restaure les lignes d'origine.

### Champ d'une ligne

```go
nom := framework.TextInput("Nom", framework.InputText, 24, true)
port := framework.TextInput("Port", framework.InputNumber, 5, true)
code := framework.TextInput("Code", framework.InputSecret, 0, false)

nom.Value()
nom.SetText("Ada")
```

Le label est à gauche, la saisie à droite, sur la même ligne.

| Type | Effet |
| --- | --- |
| `InputText` | Texte libre. |
| `InputNumber` | Chiffres seulement. |
| `InputSecret` | Affiché avec `•`. `Value()` renvoie le texte clair. |

Le troisième argument est le nombre maximum de caractères. `0` ne limite pas. Le quatrième argument dessine le cadre. Sans cadre, des champs empilés se touchent.

Le focus suffit pour écrire. Entrée ne referme pas le champ. `Reload` restaure le texte d'origine. Ctrl+V, comme le collage du terminal, insère le presse-papier. Un retour à la ligne ne passe pas dans ce champ.

### Zone de texte

```go
brouillon := framework.TextArea("Brouillon", "", true, 4, true)
lecture := framework.TextArea("Sortie", "En attente", false, 4, true)
```

Arguments : titre, texte initial, `true` pour écrire, nombre minimum de lignes visibles, puis le cadre. Une valeur de lignes inférieure à 1 devient 1. La barre `█` n'apparaît que si le texte dépasse. `Reload` restaure le texte initial. Coller insère le texte quand la zone est éditable, y compris les retours à la ligne.

```go
zone := framework.TextView(runtimeA.Memory, lien, "courant")
```

`TextView` est une zone en lecture, sans titre, avec un cadre et quatre lignes au minimum. À chaque dessin elle relit la clé dans `lien`, puis le `return` du slot de cet id. `return` est à la racine du slot, pas sous `data`. Un id vide, ou un slot absent, donne un texte vide. Changer d'id remet le défilement en haut. Si le même texte s'allonge et que le bas était déjà visible, le bas le reste. Une fenêtre remontée ne redescend pas toute seule. La zone n'écrit ni dans la mémoire du runtime, ni dans `lien`.

### Bouton

```go
framework.Button("Ajouter", framework.Call{Run: func() {
    memo.Set("3", "Jaune")
}})
```

Entrée déclenche l'action. Le bouton ne s'ouvre pas.

| Action | Rôle |
| --- | --- |
| `Call{Run: fn}` | Appelle votre fonction. |
| `Reload{Target: zone}` | Recharge une liste, un champ, une zone de texte, une catégorie ou une liste d'onglets. |
| `Exec{Script, Args, Into}` | Lance un programme. `Into` reçoit la sortie. Sans `Into`, le programme prend le terminal le temps de son exécution. |

Si `Into` est un champ ou une zone placés dans un onglet ou une autre fenêtre, cet écran s'affiche et l'objet prend le focus. La sortie est alors visible. Le saut de ligne final est retiré. En cas d'erreur, le message est ajouté au texte.

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

Une application ne contient qu'une liste d'onglets. Une seconde arrête le programme. La barre n'a pas de titre. Au lancement, le focus est la première zone de l'onglet affiché. Tab change d'onglet depuis n'importe où, y compris pendant une saisie, et reprend au premier. Entrée ne change pas d'onglet. Le contour des onglets et du panneau reste gris. Seul le libellé de l'onglet ouvert prend la couleur de titre.

L'onglet ouvert n'a pas de trait en bas : il rejoint le panneau. Un premier onglet fermé joint le panneau avec `├`. Un dernier onglet fermé au bord droit utilise `┤`.

Le contenu d'un onglet remplit le panneau. Il ne peut pas élargir l'écran au-delà de la largeur déjà décidée par les autres lignes.

La barre peut venir d'une mémoire. Le constructeur du `ReactiveObject` est alors `TabItem`. Les grilles restent liées par l'id : une grille est cachée tant que son id n'est pas dans la mémoire.

```go
memoVues.Set("fiches", "Fiches")
vues := framework.TabList(framework.ReactiveObject(
    framework.TabItem,
    memoVues,
    map[string]string{"id": "id", "label": "value"},
))
```

`memoVues.Set("plus", "Plus")` ajoute l'onglet Plus. Changer la valeur d'un id renomme son libellé. L'onglet affiché reste le même id.

### Activer et désactiver

`Disable()` grise un objet et le retire du focus. Il reste dessiné. `Enable()` le rétablit. Sur une catégorie ou une liste d'onglets, l'appel s'applique aussi au contenu qui n'était pas déjà désactivé.

## Mémoire

```go
memo := framework.Memory()

memo.Set(framework.Data{
    "id": "1",
    "data": framework.Data{"id": "cyan", "value": "Cyan", "color": "#ffee00"},
})
memo.Set("etat", "en attente")

memo.Get("1").Get("data").Get("value").String()    // "Cyan"
memo.Get("etat").Get("data").Get("value").String()  // "en attente"
```

`Set(id, texte)` est la forme courte : elle enregistre `data.id` et `data.value`.

`Get` se chaîne. Une clé absente donne une chaîne vide, pas une erreur. `Set` réveille l'écran. L'ordre d'insertion est conservé. Un id vide, ou une feuille qui n'est pas une chaîne, arrête le programme.

`ReadOnly()` renvoie la même mémoire en lecture. Un `Set` sur cette vue arrête le programme.

`SetList(clé, []*memory)` et `List(clé)` rangent des mémoires imbriquées. On passe le résultat de `Memory()`.

## Liste réactive

```go
teintes := framework.ListUniq("Teintes", framework.ReactiveObject(
    framework.ListItem,
    memo,
    map[string]string{"id": "id", "label": "value", "dot": "color"},
))
```

La table nomme les champs lus dans chaque entrée. `id` et `label` sont obligatoires pour une ligne. `dot` est facultatif : il nomme la couleur de la pastille, un `#rrggbb` ou un status. Pour une mémoire ordinaire, ces noms sont des champs de `data`. Pour la mémoire d'un runtime, les champs directs sont `id`, `label`, `path`, `status` et `return`.

```go
map[string]string{"id": "id", "label": "label", "dot": "status"}
```

Chaque dessin relit la mémoire. `memo.Set` ajoute ou remplace une ligne. Le choix confirmé et la ligne focalisée restent si leur id existe encore. Sinon le focus revient à la première ligne.

## Runtime et scripts

Un runtime lance des scripts bash en arrière-plan et écrit leur résultat dans sa mémoire.

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

`Start()` ne lance rien au moment où vous écrivez le bouton. Il renvoie l'action du bouton. Le lancement a lieu quand l'utilisateur valide Exécuter. Si le runtime tournait déjà, il s'arrête puis repart du début.

`Bash` s'écrit seulement dans `Functions`. Un `Bash` ailleurs, ou un id ou un chemin vides, arrête le programme. Sans `Dir`, le chemin est relatif au dossier depuis lequel vous avez lancé le binaire. Avec `Dir`, il part de ce répertoire. Le script reçoit ses arguments en lecture seule. `BashEnv` ajoute le même script, plus une map de clés. Une map nulle ou vide équivaut à `Bash`. Une clé absente ne retire pas l'environnement hérité. Le même id remplace le job.

Après le départ, le slot se lit ainsi :

```go
runtimeA.Memory.Get("controle1").Get("status").String()
runtimeA.Memory.Get("controle1").Get("return").String()
```

`status` reste vide tant que le script n'a pas démarré. Il vaut `pending` le temps de son exécution, puis `success`, `warning` ou `error`. Un seul slot est `pending` : celui qui tourne. `return` grandit à chaque ligne complète, avant la fin du processus. La regex ne classe le slot qu'une fois, sur le texte entier.

Après `Init`, `Memory.Set` ajoute sur un slot déjà créé un champ dont le nom n'est pas `id`, `label`, `path`, `status` ou `return`. Ces cinq noms arrêtent le programme. Un id qui n'est pas encore un slot l'arrête aussi. Le champ se lit dans `data` et reste pendant le script.

Un slot `error` arrête la file, même si `Loop` vaut vrai. Les scripts pas encore lancés restent vides. `warning` et `success` laissent continuer.

`Status()` du runtime n'est pas le verdict du script. C'est le cycle : `not_initialized`, `pending`, `running`, `terminated`.

`Stop` annule l'exécution, remet ce cycle à `not_initialized` et vide `status` et `return` de chaque slot. `Cancel` annule le bash en cours et laisse le tableau lisible : le slot encore `pending` passe à `error` avec le texte déjà reçu, le cycle vaut `terminated`. Si rien ne tourne, `Cancel` revient tout de suite. Ctrl+C quitte l'écran et fait ce `Cancel` sur chaque runtime. `Pause` retient le runtime entre deux scripts. `Resume` le relâche. Un `Resume` sans pause ne fait rien.

Un second `Start` pendant l'exécution appelle `Stop`, puis relance. Après la fin, `Start` relance sans vider les slots : chaque script remet son propre `return` à vide en démarrant.

Avec `Loop: true`, `Timer` est obligatoire. `0` vaut 100 ms. À la fin de chaque tour, le runtime attend cette durée, puis recommence. `Stop` coupe l'attente. Sans boucle, `Timer` est ignoré.

Chaque `Runtime(...)` est indépendant. La regex par défaut est `(success|warning|error)`. Elle s'applique à la sortie du script. Le premier groupe, ou le texte entier, doit être l'un de ces trois mots. Sinon le status du slot est `error`. Une regex invalide arrête le programme à la construction.

Un plantage dans le conteneur termine le cycle et ajoute le texte au journal `Log()`.

Les scripts sont des fichiers à côté du binaire lancé, pas des fichiers compilés dans le programme. Pour les embarquer, copiez-les au déploiement, ou placez-les dans le dossier de lancement.

```go
lien := framework.Memory()
rt := framework.Runtime(framework.RuntimeConfig{
    Dir:         root,
    Loop:        false,
    StatusRegex: `(success|warning|error)`,
    Functions: func(framework.Store) {
        framework.Bash("un", "Un", "scripts/un.sh", nil)
        framework.BashEnv("deux", "Deux", "scripts/deux.sh", nil,
            map[string]string{"MODE": "strict"})
    },
})
parent := framework.ListUniq("Groupes", groupes...)
parent.Bind(lien, "parent")
enfant := framework.ListUniq("Scripts", framework.ReactiveWatcher(...)) // filtre sur lien / parent
enfant.Bind(lien, "courant")
zone := framework.TextView(rt.Memory, lien, "courant")
```

Le watcher de l'enfant suit `lien` et `rt.Memory`. Il ne garde que les scripts du parent publié. `dot` reçoit le status. La zone lit le `return` du courant. Rien de tout cela n'est écrit à la main dans `Update`.

Ctrl+C a déjà appelé `Cancel`. Si le programme quitte autrement, `rt.Cancel()` coupe encore les scripts en cours. Un second appel ne fait rien.

## Watcher

Le watcher croise une ou plusieurs mémoires. Le framework rappelle votre fonction quand l'une d'elles change.

```go
suivi := framework.ReactiveWatcher(framework.WatcherConfig{
    Watcher: []framework.Store{runtimeA.Memory, runtimeB.Memory},
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
journal := framework.ListUniq("Suivi", suivi)
```

`args[0]` est la première mémoire de `Watcher`. Le résultat de la fonction devient la valeur du watcher. `Result()` la relit. Une `ListUniq` accepte ce watcher si la fonction retourne un `ReactiveObject`, une liste de `ListItem`, ou un autre watcher.

N'écrivez pas dans les mémoires surveillées depuis cette fonction. Écrivez dans une mémoire neuve. Une liste vide, une mémoire nulle ou une fonction nulle arrêtent le programme.

## Fenêtres

```go
fenetre := framework.Window()
page := framework.Grid(framework.Row(framework.Button("Suivante", fenetre.Next())))
autre := framework.Grid(framework.Row(framework.Button("Précédente", fenetre.Prev())))
fenetre.Use(page, autre)
app := framework.NewWindow(framework.DefaultStyle(), fenetre)
```

Une seule grille est visible. Les boutons font le tour. À l'arrivée sur une grille, le focus se place sur son premier élément. La compilation reste `go build -o mon-tui .` : `NewWindow` remplace `New` dans le même `main`.

## Fichier

```go
memo.Set("file", "bonjour")
framework.Button("Charger", framework.ReadFile{Path: "notes.txt", Into: memo})
framework.Button("Enregistrer", framework.WriteFile{Path: "notes.txt", From: memo})
```

Le texte lu et écrit est la valeur `file` de la mémoire. Le chemin est relatif au dossier de lancement du binaire. Un chemin vide ou une mémoire absente arrête le programme.

## Barre de progression

```go
un := framework.Variable()
deux := framework.Variable()
etapes := framework.Variables(un, deux)
barre := framework.ProgressBar(etapes)
un.Set(1)
etapes.Add(framework.Variable())
```

Une variable vaut `0` au départ. `Set(1)` la passe à 1. La barre suit la liste : `Set` la remplace, `Add` ajoute une variable. Seules les variables à `0` ou `1` comptent. Le pourcentage affiché au centre est la part de ces variables qui valent `1`. La barre n'est pas un arrêt de focus.

## Image

```go
import _ "embed"

//go:embed logo.png
var logoPNG []byte

func init() {
    framework.IncludeImage("logo.png", logoPNG)
}

// dans la grille
framework.Image("logo.png", 80)
```

Le fichier PNG, JPEG ou GIF est compilé dans le binaire par `//go:embed`. `IncludeImage` le décode sous le nom que `Image` réutilise. Après `go build`, ce fichier n'est plus relu sur le disque. Un nom vide, des octets illisibles, un nom absent ou une largeur inférieure à 1 arrêtent le programme.

Le second argument de `Image` est la largeur en colonnes. La hauteur suit le ratio : une cellule est comptée deux fois plus haute que large, `round(largeur × ratio × 0.5)`, au minimum 1. L'image est réduite en Lanczos3. Chaque pixel devient un signe de la rampe ` .,:;i1tfLCG08@`. Le signe prend la couleur du pixel. L'espace de la rampe reste un espace. L'image est un arrêt de focus. Entrée ne l'ouvre pas.

`//go:embed` exige que `logo.png` soit dans le paquet du fichier source, à côté de `main.go`, au moment de `go build`.

## Limites

- Une seule liste d'onglets par application.
- Le thème est unique. `Disable` grise un objet sans le masquer.
- `q` dans un champ saisit la lettre q. `q` ne quitte pas. Ctrl+C quitte, y compris dans un champ, et arrête les runtimes en cours.
- Les scripts bash ne sont pas compilés dans le binaire. Leur chemin part du dossier de lancement.
- Un widget nul dans une grille ou une catégorie arrête le programme.
