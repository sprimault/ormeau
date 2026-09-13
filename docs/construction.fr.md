> [🇬🇧 English](construction.md) · [🇫🇷 Français](construction.fr.md)

# Construction et distribution

## Une seule machine produit tout

Go croise nativement. Une machine Linux suffit à produire les binaires Windows,
macOS et Linux, pour amd64 comme pour arm64 :

```bash
make binaries
```

Cela ne tient qu'à une condition : **aucun cgo**. Tous les pilotes retenus sont
en Go pur, et c'est ce qui rend la matrice de compilation triviale. Le premier
pilote exigeant cgo imposerait un compilateur croisé par cible.

`CGO_ENABLED=0` est explicite dans chaque ligne de compilation, jamais implicite :
sur une machine Linux, cgo est actif par défaut, et un binaire lié dynamiquement
à la glibc refuserait de démarrer sur Alpine.

## Les binaires Linux ne dépendent d'aucune distribution

Un binaire statique n'a de dépendance ni à la glibc, ni à la musl, ni à quoi que
ce soit du système. Un seul `ormeau_linux_amd64` tourne sur Debian, Ubuntu, RHEL
et Alpine indifféremment.

Il n'existe donc pas de « version Debian » ou de « version Alpine ». Deux
architectures, pas davantage : amd64 et arm64.

## Ordre de construction

Le front se construit **avant** les binaires. Vite écrit dans
`internal/interface/embarque`, que `go:embed` lit depuis son propre paquet —
`embed` ne remonte pas au-dessus du répertoire où il est déclaré, et une copie
intermédiaire serait une étape de plus à oublier.

La cible `binaries` dépend de `web-build` pour cette raison, et un test vérifie
que le système de fichiers embarqué porte un `index.html` non vide. Sans lui, le
garde-fou reposerait sur la seule dépendance du Makefile, et un bundle absent se
découvrirait à l'ouverture d'une interface blanche.

## Compilation locale

L'antivirus d'un poste Windows peut mettre en quarantaine les exécutables
temporaires qu'écrit l'éditeur de liens : la compilation échoue alors sur un
accès refusé, sans rapport apparent avec le code. Le remède est de faire pointer
`GOTMPDIR` et `GOCACHE` dans `.tmp/`, à l'intérieur du dépôt — mais dans
`makefile.local`, que le Makefile inclut s'il existe et que git ignore, et non
dans le Makefile versionné : c'est une contrainte de ce poste, pas une décision
du projet. [`CONTRIBUTING.fr.md`](../CONTRIBUTING.fr.md) donne les lignes à
recopier.

`make build` écrit dans `.tmp/` ; `dist/` reste réservé aux artefacts de
publication.

## Images Docker

Les binaires étant déjà croisés, l'image multi-arch se construit **sans
émulation QEMU** : buildx renseigne `TARGETOS` et `TARGETARCH`, et chaque
plateforme reçoit le binaire correspondant.

```dockerfile
FROM alpine:3 AS certificats
RUN apk add --no-cache ca-certificates

FROM scratch
ARG TARGETOS TARGETARCH
COPY --from=certificats /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY dist/ormeau_${TARGETOS}_${TARGETARCH} /ormeau
ENTRYPOINT ["/ormeau"]
```

```bash
docker buildx build --platform linux/amd64,linux/arm64 \
  -t ghcr.io/sprimault/ormeau:$VERSION --push .
```

`FROM scratch` signifie qu'il n'y a aucune distribution dans l'image — et donc
**aucun certificat racine**. Sans la copie de `ca-certificates.crt`, toute
connexion TLS vers une base échoue avec une erreur de vérification difficile à
diagnostiquer. C'est le piège classique de `scratch`.

L'image tourne sous un utilisateur non privilégié, ce qui a une conséquence à
l'usage : elle n'écrit pas dans un volume appartenant à quelqu'un d'autre.
Extraire un calque demande donc de lui donner l'identité de l'appelant, faute de
quoi l'écriture est refusée sans que le message ne dise pourquoi :

```bash
docker run --rm --user "$(id -u):$(id -g)" \
  -e ORMEAU_DSN -v "$PWD:/sortie" \
  ghcr.io/sprimault/ormeau:VERSION extraire --sortie /sortie/gescom.calque.json
```

Abaisser l'utilisateur de l'image réglerait le symptôme et créerait un défaut :
un outil qui manipule des identifiants de production n'a aucune raison de
tourner en root.

Le conteneur reste un repli : il doit encore atteindre la base, ce que le binaire
natif fait sans configuration réseau.

## Signature : les deux frictions

Elles n'empêchent pas de publier, mais elles se documentent plutôt qu'elles ne se
découvrent.

**Windows.** Le binaire n'est pas signé, SmartScreen affiche un avertissement au
premier lancement. La signature est techniquement faisable depuis Linux avec
`osslsigncode`, mais elle exige un certificat de signature de code payant.

**macOS.** Le binaire se compile depuis Linux mais n'est ni signé ni notarié :
Gatekeeper le bloque au premier lancement. La notarisation exige un Mac et un
compte développeur Apple payant.

Dans les deux cas, le README prévient de l'avertissement et montre comment
vérifier l'archive avant de la lancer. Un avertissement de sécurité inexpliqué
sur un outil qui réclame un mot de passe de base de données arrête net un
utilisateur prudent — et il a raison de s'arrêter.
