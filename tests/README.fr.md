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

La mise à jour groupée des attendus se fait avec `go test ./... -maj-attendus`,
jamais automatiquement : un attendu régénéré sans être relu ne teste plus rien.
