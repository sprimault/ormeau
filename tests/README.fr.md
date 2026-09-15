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
