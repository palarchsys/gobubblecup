# Démonstrations

Chaque dossier est un programme. Lancez-le depuis la racine du module, pour que les scripts et `notes.txt` soient trouvés :

```bash
go run ./examples/bonjour
```

Ctrl+C quitte et arrête les scripts en cours. Dans un champ, `q` est une lettre. Le lancement efface le terminal. À la sortie, l'écran précédent revient.

Chaque démo fixe `Height` aux trois quarts de la largeur par défaut (`Width * 3 / 4`). L'écran garde ce format. Il ne descend pas jusqu'en bas du terminal.

| Dossier | Ce qu'il montre |
| --- | --- |
| `bonjour` | Mémoire, liste réactive, champ, bouton |
| `listes` | Choix unique, choix multiple, pastilles, rechargement |
| `saisie` | Texte, nombre, secret, zone éditable, zone en lecture |
| `memoire` | `Set(Data)`, objet réactif, pastille écrite dans la mémoire |
| `onglets` | Deux onglets fixes. Tab passe au suivant |
| `vues` | Onglets relus dans une mémoire. L'onglet absent reste masqué |
| `runtime` | Trois scripts enchaînés une fois. La liste suit leurs statuts |
| `horloge` | Boucle, attente d'une seconde, pause et reprise |
| `suivi` | Un watcher croise deux runtimes dans une seule liste |
| `fenetres` | Deux grilles. Suivante et Précédente font le tour |
| `fichier` | Lecture et écriture de `notes.txt`, résumées par un watcher |
| `progression` | Barre sur des variables 0 ou 1. Une autre valeur est ignorée |
| `echo` | La sortie d'un script arrive dans une zone |
| `griser` | Une catégorie grise ses enfants encore actifs |
| `image` | PNG compilé dans le binaire |

Les scripts bash sont dans `examples/scripts/`. Le runtime les lance avec `bash`. Ils doivent afficher `success`, `warning` ou `error`.
