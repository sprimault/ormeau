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

Le front se construit **avant** les binaires. Si `web/dist` n'existe pas au
moment de la compilation, `embed` produit un système de fichiers vide et la
compilation réussit quand même : on publie alors des binaires avec une interface
blanche, sans qu'aucun avertissement ne le signale.

La cible `binaries` dépend de `web-build` pour cette raison, et un test vérifie que
le système de fichiers embarqué n'est pas vide.

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

Dans les deux cas, le README explique le contournement. Un avertissement de
sécurité inexpliqué sur un outil qui réclame un mot de passe de base de données
arrête net un utilisateur prudent — et il a raison de s'arrêter.
