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

## Le miroir du paquet PHP

Composer ne lit un dépôt VCS que si `composer.json` est à sa racine. Le paquet
vit dans `php/` : il est donc publié dans
[sprimault/ormeau-doctrine](https://github.com/sprimault/ormeau-doctrine), où
`php/` devient la racine. Ce dépôt est un miroir en lecture seule : ni issues,
ni pull requests, et rien ne s'y écrit à la main.

`.github/workflows/miroir.yml` y pousse `git subtree split --prefix=php` à chaque
push sur `master`, et pour chaque tag `v*` appelé depuis `release.yml`. Le
split est déterministe, et celui d'un commit plus ancien est ancêtre de celui
d'un plus récent : le miroir avance toujours en avance rapide, et **rien ne s'y
pousse en force**. Un refus veut dire que le miroir a divergé ; il s'examine, il
ne s'écrase pas.

Le miroir ne porte que des versions où le paquet fonctionne : il commence à la
0.5.0, et aucun tag antérieur n'y sera poussé.

Packagist lit le miroir, pas le dépôt principal. Un webhook du miroir, sur
l'événement push, le prévient à chaque poussée de `miroir.yml` : un tag
apparaît sur Packagist sans rien ajouter à la publication. Le webhook porte
en secret le jeton d'API du compte Packagist qui maintient le paquet.

### Ce qui protège l'écriture

- La clé est une **clé de déploiement** du miroir : elle n'écrit sur aucun autre
  dépôt.
- Sa partie privée est le secret `MIROIR_CLE` de l'**environnement `miroir`** du
  dépôt principal, dont la règle de déploiement n'admet que `master` et les tags
  `v*`. Un workflow déclenché par une pull request tourne sur une autre
  référence : GitHub ne lui ouvre pas l'environnement, quoi que dise son
  fichier.
- Sur le miroir, deux règles : `master` ne se supprime pas, ne se force pas, et
  seule la clé de déploiement la met à jour, le propriétaire compris ; un tag
  `v*` ne se déplace pas, et seule la clé le crée. Le propriétaire peut
  supprimer un tag : c'est la reprise ci-dessous qui l'exige.

Mise en place, une fois, par le propriétaire des deux dépôts, dans un
répertoire temporaire : jamais dans le dépôt, où un `git add` l'emporterait, ni
dans `~/.ssh`, où elle resterait une clé capable d'écrire sur le miroir. Une
fois transmise, elle n'existe plus que dans le secret ; perdue ou à changer,
elle se remplace par une nouvelle, des deux côtés.

```bash
cd "$(mktemp -d)"
ssh-keygen -t ed25519 -N "" -C "miroir ormeau-doctrine" -f miroir
gh repo deploy-key add miroir.pub --repo sprimault/ormeau-doctrine --allow-write --title "miroir.yml"
gh secret set MIROIR_CLE --env miroir --repo sprimault/ormeau < miroir
rm miroir miroir.pub
```

Sous PowerShell, `-N ""` n'arrive pas à `ssh-keygen` : l'omettre, et valider
deux fois une phrase de passe vide. La redirection `<` n'existe pas non plus ;
passer la clé par `((Get-Content miroir) -join "`n") | gh secret set …`, qui
garde des fins de ligne LF, sans lesquelles ssh refuse la clé sur le runner.

### Publier une version

Le tag se pose **avant** la fusion de la pull request de clôture, pour
qu'aucune fenêtre ne sépare l'annonce de la version de sa disponibilité :

1. La pull request qui date la section du `CHANGELOG` est verte.
2. Le tag est posé sur sa tête et poussé :
   `git tag vX.Y.Z <tête de la PR> && git push origin vX.Y.Z`.
3. `release.yml` pousse d'abord le tag sur le miroir, et vérifie que Composer
   résout `sprimault/ormeau-doctrine:X.Y.Z` depuis le vrai miroir. Binaires,
   brouillon de release et image attendent cette preuve.
4. La pull request est fusionnée : le README qui annonce la version arrive sur
   `master` après que le miroir la porte.
5. Le brouillon est relu et publié.

### Si la publication échoue

Tant que le brouillon n'est pas publié, le numéro se réutilise. Ensuite, jamais :
un projet a pu verrouiller le commit du tag, et un correctif prend le numéro
suivant.

- **Le job du miroir échoue avant d'avoir poussé le tag** (secret absent, push
  refusé) : le tag n'existe que sur le dépôt principal, et rien d'autre n'est
  parti. Supprimer le tag, corriger sur la pull request, le reposer sur la
  nouvelle tête :

  ```bash
  git push origin :refs/tags/vX.Y.Z && git tag -d vX.Y.Z
  ```

- **Un job échoue après le tag du miroir, pour une cause passagère** (réseau,
  runner) : relancer les jobs échoués. Le split est le même, et repousser un tag
  identique ne change rien.

  ```bash
  gh run rerun <identifiant du run> --failed
  ```

- **La correction demande un commit** : le split change, et le tag du miroir ne
  se déplace pas. Supprimer le tag des deux côtés et le brouillon s'il existe,
  corriger, reposer le tag :

  ```bash
  gh api -X DELETE repos/sprimault/ormeau-doctrine/git/refs/tags/vX.Y.Z
  gh release delete vX.Y.Z --yes
  git push origin :refs/tags/vX.Y.Z && git tag -d vX.Y.Z
  ```

  La même suppression vaut pour une pull request abandonnée après la pose du
  tag.

- **`master` du miroir refuse l'avance rapide** : quelque chose y a été écrit
  sans venir du split. La règle du miroir ne l'autorise qu'à la clé : c'est donc
  que la clé a servi ailleurs. La révoquer (supprimer la clé de déploiement, en
  poser une nouvelle), comparer `git ls-remote` au split local de `master`, puis
  seulement remettre `master` du miroir sur le split — seul cas où un push forcé
  est permis, fait à la main par le propriétaire, règle désactivée le temps de
  l'opération.

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
