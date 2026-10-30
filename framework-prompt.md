# Prompt : Framework Bubble Tea V2

Figé. Ce fichier est le contrat. Ne pas inventer une autre syntaxe. Ne pas renommer un constructeur. Ne pas ajouter une option qui n'est pas écrite ici.

Le produit public s'appelle GOBubbleCup. `high-tui` n'est que le nom de code du dépôt de travail. Le module Go est `github.com/palarchsys/GOBubbleCup`. Le paquet est `framework`. `go run .` lance `main.go`.

Reproduire le framework, c'est respecter ce fichier à la lettre : mêmes constructeurs, mêmes arguments, mêmes valeurs par défaut, mêmes effets visibles. `tui/` est une ancienne démo : ne pas y toucher. Ne pas committer depuis le dépôt de travail `high-tui`. `AGENT.md` est le guide d'images : ne pas le réécrire.

Go ne permet pas qu'un type et une fonction portent le même nom. Les structs sont donc internes. Le code appelant n'utilise que les constructeurs courts. Chaque `New*` reste un alias du constructeur court.

## Principes

- Go, Bubble Tea V2, Lip Gloss V2.
- Un objet ne garde que ses données. Tout le dessin vient de l'unique `Style`.
- Réutiliser `gate`, `items`, `box`, `ListArg`.

## Style

```go
framework.DefaultStyle()
framework.New(framework.DefaultStyle(), grille)
tea.NewProgram(app).Run()
```

Champs de `Style`, figés dans `DefaultStyle` :

| Champ | Valeur | Rôle |
| `Background` | `#1e1e1e` | Fond. Égal à `Surface`. |
| `Surface` | `#1e1e1e` | Zones, gouttières, écran. |
| `Text` | `#d4d4d4` | Texte. |
| `Accent` | `#569cd6` | Titres et libellé d'onglet ouvert. |
| `Idle` | `#6b6b6b` | Contour au repos. |
| `Focus` | `#007acc` | Contour au focus, gras. |
| `Inside` | `#9cdcfe` | Contour à l'intérieur d'une zone. |
| `Width` | `22` | Largeur de cadre par défaut. |
| `Height` | `0` | `0` suit le contenu. |
| `PadX` | `1` | Une cellule à gauche et à droite du texte. |
| `PadY` | `0` | Aucune ligne vide dans un cadre. |
| `Margin` | `1` | Une cellule peinte entre deux colonnes. Horizontal seulement. |
| `Radius` | `1` | Coins `╭╮╰╯`. `0` donne des coins carrés. |
| `Thickness` | `1` | Un seul trait. |
| `Cursor` | `›` | Ligne survolée dans une liste ouverte. |
| `Mark` | `✓` | Ligne choisie. |
| `Track` | `│` | Piste de défilement. |
| `Thumb` | `█` | Curseur de défilement, seulement si le texte dépasse. |

`paper()` peint les zones, les écarts et `View.BackgroundColor`. Un espace sans fond laisse voir le terminal. Pas de ligne vide entre les rangées, entre les enfants d'une catégorie, ni au-dessus ou au-dessous de l'écran. L'inset extérieur n'ajoute que les colonnes gauche et droite. L'objet de droite absorbe le vide horizontal. La hauteur d'une ligne est celle du plus grand. Le dernier enfant d'une catégorie absorbe le vide vertical. `placed` garde la largeur d'origine. Une mesure imbriquée passe par `preferred()`.

## Grille

```go
framework.Row(a, b)
framework.Grid(
    framework.Row(titre),
    framework.Row(vues),
    framework.Row(nom, executer),
)
```

- `Row` est une ligne. `Grid` empile ces lignes.
- Un widget nul panique.
- Les flèches déplacent le focus entre zones et boutons actifs, sans boucle. Les inactifs sont sautés et gardent leur bande.
- La largeur de l'interface est la plus grande rangée, mesurée à la largeur d'origine.
- Un onglet ne dépasse pas cette largeur. Sa grille interne remplit le panneau. Elle mesure ses enfants avec `preferred()`. S'ils dépassent, ils rétrécissent depuis la droite.

## Title

```go
framework.Title("high-tui")
```

Inactif. Le bandeau s'étire sur la largeur de sa ligne. Une seule option : le texte.

## Category

```go
framework.Category("Catalogue", teintes, pieces)
```

- Argument 1 : sous-titre, écrit sur le trait haut.
- Arguments suivants : enfants, empilés bord à bord.
- Pas un arrêt de focus. Le focus entre dans les enfants actifs.
- `Reload()` propage aux enfants qui savent se recharger.
- Un enfant nul panique.

## ListItem

```go
framework.ListItem("casque", "Casque") // id, libellé
```

Deux arguments obligatoires. `id` est la valeur réactive. `label` est le texte affiché. Le type renvoyé s'appelle `Item`. `NewItem` est l'alias.

Dans une liste, la ligne est décalée de deux colonnes. La première, alignée sur le premier caractère du titre, porte `›` si le curseur est dessus, sinon `✓` si l'item est choisi, sinon un espace. La seconde est un espace.

## ListUniq

```go
framework.ListUniq("Teintes", framework.ListItem("cyan", "Cyan"))
framework.ListUniq("Teintes", objetReactif)
```

- Argument 1 : titre.
- Arguments suivants : des `ListItem`, ou un seul `ReactiveObject`.
- Un seul choix. `Value()` renvoie l'id confirmé par Entrée. Vide avant.
- `Reload()` oublie le choix. Sur une liste réactive, elle relit la mémoire.
- Haut et bas sont bornés. Entrée entre, puis Entrée choisit et sort. Le focus reste. Gauche et droite ne sortent pas.
- Au focus, contour bleu. À l'intérieur, contour bleu clair.

## ListMult

```go
framework.ListMult("Pièces",
    framework.ListItem("casque", "Casque"),
    framework.ListItem("veste", "Veste"),
)
```

- Mêmes arguments qu'une liste unique, sans objet réactif.
- Plusieurs choix. `Values()` renvoie les id cochés, dans l'ordre des lignes.
- Entrée entre. À l'intérieur, Entrée coche ou décoche et reste. Échap sort. Le focus reste.
- `Reload()` vide les coches et restaure les lignes d'origine.

## Memory

```go
memo := framework.Memory()
memo.Set(framework.Data{"id": "1", "data": framework.Data{"id": "cyan", "value": "Cyan"}})
memo.Set("etat", "en attente")
memo.Get("1").Get("data").Get("value").String() // "Cyan"
memo.Get("etat").Get("data").Get("id").String()   // "etat"
memo.Get("etat").Get("data").Get("value").String() // "en attente"
```

- `Memory()` crée une mémoire vide. `NewMemory` est l'alias.
- `Set(Data)` : la clé `"id"` est l'identifiant de l'entrée. La clé `"data"` est un `Data` dont les feuilles sont des chaînes. Un id vide panique. Une feuille qui n'est pas une chaîne panique.
- `Set(id, texte)` : forme courte. Elle écrit `data.id = id` et `data.value = texte`.
- `Get(clé)` se chaîne. Une clé absente donne un nœud vide. `String()` lit la feuille, ou `""`.
- `Set` réveille l'écran. Une liste déjà construite sur cette mémoire se reconstruit au dessin suivant.
- `SetList(clé, []*memory)` et `List(clé)` rangent des mémoires imbriquées. Le type des enfants est interne : on passe le résultat de `Memory()`.
- `ReadOnly()` partage les données et panique sur `Set` et `SetList`.
- L'ordre d'insertion des entrées est stable.

## ReactiveWatcher

```go
reactiveWatcher := framework.ReactiveWatcher(framework.WatcherConfig{
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
framework.ListUniq("Journal", reactiveWatcher)
```

- `Watcher` est une liste de mémoires. L'ordre est celui reçu par `Functions` : `args[0]`, `args[1]`, etc.
- Une liste vide, une mémoire nulle, ou `Functions` nulle paniquent.
- `Functions` est libre. Elle croise ces mémoires et retourne un résultat.
- Le framework rappelle `Functions` quand l'une des mémoires a changé. Le développeur n'écrit pas la boucle.
- Le résultat retourné est celui de `reactiveWatcher`. `Result()` le relit.
- S'il retourne un `ReactiveObject`, une `[]Item` ou un autre watcher, une `ListUniq` l'accepte.
- La fonction ne réécrit pas les mémoires qu'elle surveille. Elle écrit dans une autre mémoire.
- `NewReactiveWatcher` est l'alias.

## ReactiveObject

```go
framework.ReactiveObject(
    framework.ListItem,
    memo,
    map[string]string{"id": "id", "label": "value"},
)
```

- Argument 1 : `ListItem`, le constructeur appelé pour chaque entrée.
- Argument 2 : la mémoire. Un constructeur ou une mémoire nuls paniquent.
- Argument 3 : table figée. `"id"` et `"label"` sont les champs de `ListItem`. La valeur est le nom du champ lu dans l'entrée. Pour une mémoire ordinaire, `"id"` et `"value"` sont les champs de `data`. Pour la mémoire d'un runtime, les champs directs sont `id`, `label`, `path`, `status`, `return`.
- À chaque dessin, la liste relit la mémoire. L'id choisi et le curseur restent si cet id existe encore.
- `NewReactiveObject` est l'alias.

## TextInput

```go
framework.TextInput("Nom", framework.InputText, 24)
```

- Argument 1 : label, affiché à gauche.
- Argument 2 : `InputText`, `InputNumber` ou `InputSecret`.
- Argument 3 : nombre maximum de caractères. `0` ne borne pas.
- La saisie est à droite du label, sur la même ligne. Elle occupe la place restante.
- `Value()` renvoie le texte. Un secret s'affiche en `•`.
- Entrée ouvre. Entrée ou Échap referme. Gauche et droite déplacent le curseur. Retour arrière efface. `InputNumber` ne garde que les chiffres.
- Dans le champ, `q` est une lettre. Ctrl+C quitte quand même.
- `Reload()` restaure le texte d'origine.

## TextArea

```go
framework.TextArea("Notes", "En attente", false, 8)
```

- Argument 1 : titre.
- Argument 2 : texte initial.
- Argument 3 : `true` si l'on peut écrire, `false` en lecture.
- Argument 4 : nombre minimum de lignes visibles. Une valeur inférieure à 1 devient 1.
- La barre `█` n'apparaît que si le texte dépasse la fenêtre.
- Éditable : Entrée insère une ligne, Échap sort, les flèches suivent le curseur.
- Lecture : Entrée ouvre, haut et bas défilent, Entrée ou Échap sort.
- `Reload()` restaure le texte d'origine. `SetText` ne change pas cet original.

## Button et actions

```go
framework.Button("Exécuter", runtimeA.Start())
framework.Button("Ajouter", framework.Call{Run: func() { memo.Set("4", "Noir") }})
framework.Button("Recharger", framework.Reload{Target: teintes})
framework.Button("Echo", framework.Exec{Script: "echo", Args: []string{"ok"}, Into: zone})
```

- `Button(label, action)`. Entrée déclenche l'action et n'ouvre pas la zone. Une action nulle ne fait rien.
- `Call` a un seul champ : `Run func()`.
- `Reload` a un seul champ : `Target`. Cibles : `ListUniq`, `ListMult`, `TextInput`, `TextArea`, `Category`, `TabList`.
- `Exec` a trois champs : `Script`, `Args`, `Into`. `Into` est une zone de texte. S'il est posé, la sortie y est écrite sans quitter l'écran. Sinon le script prend le terminal. Un script vide ne fait rien.

## TabList, TabItem, Tab

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

- `TabItem(id, label)`. Un id vide panique.
- `Tab(id, grille)`. Un id vide ou une grille nulle paniquent.
- `TabList` reçoit les items. `SetTabs` reçoit les grilles, liées par le même `id`.
- Une interface n'a qu'une seule `TabList`. Une deuxième panique.
- Pas de sous-titre. Pas un arrêt de focus.
- Au chargement, le focus est la première zone de l'onglet affiché.
- Ctrl+gauche et Ctrl+droite agissent depuis n'importe où, même dans une zone. S'il existe un onglet dans cette direction, le runtime d'interface sort de la zone actuelle et pose le focus sur la première zone du nouvel onglet. Sinon la touche ne fait rien. Pas de boucle.
- Entrée ne change pas d'onglet.
- Le premier item est affiché. `SelectedID()` renvoie son id.

Dessin figé, référence `retour-attendu.png`. `tab.png` est un schéma : ses couleurs ne sont pas celles du thème.

- Onglets collés, sans écart. Traits partagés : `┬` en haut, `│` devant le libellé, `┴` en bas.
- Coins hauts `╭` et `╮` si `Radius` vaut 1. Le libellé a un espace avant et un espace après.
- Contour des onglets et du panneau : gris de repos `#6b6b6b`, jamais le bleu de focus, jamais gras.
- Libellé de l'onglet affiché : accent, gras. Les autres : texte.
- L'onglet affiché n'a pas de trait bas. Bord gauche : `│` s'il est le premier, sinon `┘`. Bord droit : `│` s'il touche le bord du panneau, sinon `└`.
- Un onglet fermé a un trait bas. Entre deux onglets fermés : `┴`.
- Premier onglet fermé : coin gauche `├`. `╰` est interdit ici. Dernier onglet fermé qui touche le bord droit : coin `┤`.
- À droite des onglets, le trait haut du panneau continue jusqu'au coin.

## Runtime

```go
runtimeA := framework.Runtime(framework.RuntimeConfig{
    Loop: false,
    Timer: 0,
    Args: memoArgs,
    StatusRegex: `(success|warning|error)`,
    Functions: func(args framework.Store) {
        framework.Bash("controle1", "Contrôle 1", "scripts/controle1.sh", []string{
            args.Get("1").Get("data").Get("value").String(),
        })
    },
})
runtimeA.Init()
framework.Button("Exécuter", runtimeA.Start())
```

Options de `RuntimeConfig`, toutes nommées, aucune autre :

| Option | Type | Défaut | Règle |
| `Loop` | `bool` | `false` | `true` recommence les scripts après chaque tour. |
| `Timer` | `time.Duration` | `100ms` si `Loop` est vrai | Obligatoire dès que `Loop` vaut true. `0` ou une valeur négative valent 100 ms. Attente à la fin du tour, avant le tour suivant. Ignoré si `Loop` est faux. `Stop` interrompt l'attente. |
| `Args` | `Store` | mémoire vide | Mémoire lue par `Functions`. |
| `Functions` | `func(Store)` | aucune | Appelée par `Init` et par chaque `Start`. C'est ici, et seulement ici, qu'on appelle `Bash`. |
| `StatusRegex` | `string` | `(success|warning|error)` | Regex appliquée à la sortie brute. Le premier groupe, ou le texte entier, doit valoir `success`, `warning` ou `error`. Sans correspondance : `error`. Une regex invalide panique. |

Méthodes :

- `Init()` remplit les slots. Aucun script ne part. `Status()` reste `not_initialized`.
- `Start()` renvoie une action. Le lancement a lieu au clic, pas à l'appel qui construit le bouton. Les scripts partent en arrière-plan. `Functions` est relue avant le départ. Si le runtime tourne déjà, il s'arrête puis repart du début.
- `Stop()` annule, attend, remet `Status()` à `not_initialized`, vide `status` et `return` de chaque slot. `id`, `label` et `path` restent.
- `Pause()` retient le runtime entre deux scripts. `Resume()` le relâche. `Resume` sans pause ne fait rien.
- `Status()` vaut `not_initialized`, `pending`, `running` ou `terminated`.
- `Log()` accumule les verdicts et le texte d'un plantage.
- `Memory` est un champ en lecture seule. `Set` dessus panique.

Un plantage dans le conteneur passe `Status()` à `terminated` et ajoute le texte à `Log()`. Plusieurs runtime tournent en même temps. Chaque écriture mémoire réveille l'écran.

## Bash

Appelé seulement dans `Functions` :

```go
framework.Bash(id, label, path, args)
```

- `id` vide ou `path` vide : panique.
- `path` est relatif au répertoire de lancement de `main`. La commande est `bash`, ce chemin, puis `args`. Le script lit ses arguments, il ne les modifie pas.
- Le slot `Memory.Get(id)` expose exactement ces feuilles : `id`, `label`, `path`, `status`, `return`.
- Après `Init` : `id`, `label`, `path` sont remplis. `status` et `return` sont vides.
- Après `Start` : `status` vaut `pending`, puis `success`, `warning` ou `error`. `return` est la sortie brute.
- Un `Stop` pendant l'exécution ne remplace pas une sortie déjà écrite par une sortie vide.
- Les délais et les `sleep` appartiennent au script, pas au code qui crée le bouton.

## Legend

Pas de constructeur dans l'écran. `App.View` pose la légende sous la grille. Elle liste les flèches, Entrée, Ctrl+flèches, Échap et `q`. Elle est inactive.

## Clavier

- Flèches : focus entre zones et boutons actifs, sans boucle.
- Entrée ouvre une liste ou un champ, ou déclenche un bouton.
- `q` quitte hors d'une zone. Ctrl+C quitte toujours.
- Ctrl+gauche et Ctrl+droite changent d'onglet selon la règle de `TabList`.
- Liste unique : Entrée choisit et sort. Liste multiple : Entrée coche et reste, Échap sort.
- Champ : Entrée ou Échap referme.

## Démo figée

`main.go` regroupe les objets par rôle. Il n'utilise que les constructeurs de ce fichier.

- Fiches : mémoire des teintes, `ReactiveObject`, `ListUniq`, `ListMult`, `Category`.
- Saisie : `TextInput` en `InputText`, `InputNumber` et `InputSecret`, `TextArea` éditable et en lecture.
- Suivi : un runtime sans boucle (`controle1.sh`, `controle2.sh`, `controle3.sh`) et un runtime `Loop: true` avec `Timer` d'une seconde (`scripts/horloge.sh`).
- Chaque runtime a `Init()`, puis un bouton dont l'action est seulement `Start()`.
- Un `ReactiveWatcher` suit les deux mémoires de runtime, croise leurs slots et alimente la liste Suivi.
- Ajouter fait `Set` sur la mémoire des teintes. Recharger vise cette liste. Echo est un `Exec` vers la zone Sortie.
