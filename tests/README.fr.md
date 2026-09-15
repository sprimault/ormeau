> [🇬🇧 English](README.md) · [🇫🇷 Français](README.fr.md)

# Jeux de tests

`ddl/` contient les bases de référence, une par SGBD. Elles couvrent
délibérément les cas tordus : table sans clé primaire, clé primaire composite,
clé étrangère auto-référencée, table de jointure pure, colonne générée,
énumération par contrainte de contrôle, héritage par clé primaire étrangère,
identifiants réservés, accents.

`reference/` contient les cas d'inférence : un calque physique figé en entrée,
un calque logique attendu en sortie. L'inférence étant pure, ces cas se rejouent
sans base de données, et ce sont eux le vrai jeu de tests du projet.

Quand une heuristique est ajoutée, on ajoute d'abord le calque physique qui la
déclenche.

Ces calques physiques sont écrits à la main, sauf un : celui du cas `gescom`,
extraction réelle de `ddl/postgres.sql` par le pilote PostgreSQL. Un test
d'intégration le compare octet pour octet à une nouvelle extraction, et le même
fichier sert d'entrée à l'inférence. L'image du conteneur est épinglée sur une
version mineure, puisque la version du serveur figure dans le fichier : en
changer, c'est régénérer le calque et relire son diff.

La mise à jour groupée des attendus se fait derrière un drapeau, jamais
automatiquement : un attendu régénéré sans être relu ne teste plus rien. Trois
paquets le déclarent, et chacun se lance à part — un `./...` échouerait sur
tous les autres :

```bash
go test ./internal/inference/ -maj-attendus   # calques logiques attendus, aussi `make maj-attendus`
go test ./internal/calque/ -maj-attendus      # sérialisation du calque physique
make maj-calque-gescom                        # extraction de gescom, contre le conteneur
```

Après `make maj-calque-gescom`, `make maj-attendus` recalcule le calque logique
du cas, puis `ORMEAU_MAJ_ATTENDUS=1 composer test` régénère ses entités dans
`php/tests/Generation/attendus/`.

## Aller-retour

`allerretour/` porte le test qui vérifie la chaîne entière : la base de test
extraite, inférée, générée en entités Doctrine, recréée par Doctrine dans la base
vierge `allerretour` du conteneur, extraite à nouveau, puis comparée à
l'originale.

```bash
make aller-retour
```

Le diff n'est jamais vide : Doctrine ne recrée ni vues, ni contraintes CHECK, ni
types énumérés natifs, et nomme ses clés étrangères à sa façon. Ces écarts
forment une liste fermée, dans `allerretour/ecarts_test.go`, chacun avec sa
raison :

- **IMPOSSIBLE** : une limite de Doctrine ou de DBAL ;
- **VOULU** : une décision de l'outil, tolérée seulement quand le calque logique
  ou le générateur la prend pour cet objet ;
- **À COMBLER** : un manque connu, qui part avec le correctif qui le comble.

Le test échoue sur un écart que la liste ne couvre pas, et sur une entrée qui ne
couvre plus rien : la liste ne peut ni masquer une régression, ni garder une
tolérance devenue inutile.

Il exige PHP avec l'extension `pdo_pgsql` et les dépendances de `php/`
installées. `ORMEAU_PHP` désigne l'interpréteur quand ce n'est pas `php` — par
exemple une commande `docker run` qui monte le dépôt au même chemin —, et la
version d'ORM installée choisit la liste d'écarts. Le relevé du dernier passage
est écrit dans `.tmp/allerretour/ecarts.txt`.

Deux cibles ont leur liste, les deux bouts de la plage promise : ORM 3 avec
DBAL 4, depuis `composer.lock`, et ORM 2.14 avec DBAL 3, que la CI résout sans
le lock (job `aller-retour-orm2`). Une entrée qui ne vaut que pour l'une le dit ;
une entrée qui ne couvre rien sous une cible échoue, et doit être restreinte
explicitement.
