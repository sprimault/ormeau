// Ce module n'existe que pour sortir web/ du module racine, et ne contient
// aucun code Go.
//
// node_modules héberge des paquets Go transitifs — flatted, tiré par ESLint, en
// embarque un — et `go build ./...` les compilerait avec le projet. La
// validation du dépôt dépendrait alors du contenu d'un paquet npm, et un jour
// échouerait sur du code que personne ici n'a écrit.
module github.com/sprimault/ormeau/web

go 1.27
