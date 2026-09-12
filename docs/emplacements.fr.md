> [🇬🇧 English](emplacements.md) · [🇫🇷 Français](emplacements.fr.md)

# Où Ormeau range quoi

Deux emplacements, séparés par ce qu'ils contiennent et surtout par qui en est
propriétaire.

|  | données du travail | données de la machine |
|---|---|---|
| où | votre projet | le répertoire de configuration du système |
| quoi | les trois fichiers d'une base | préférences, profils, brouillons |
| versionné | oui, avec votre code | jamais |
| partagé avec l'équipe | oui | jamais |
| perdre le fichier coûte | le travail d'arbitrage | une ressaisie |

La règle tient en une phrase : **ce qui décrit la base va dans le projet, ce qui
décrit votre poste reste sur votre poste.**

## Les données du travail

Le répertoire de travail est celui d'où vous avez lancé la commande, ou celui
que `--repertoire` désigne. L'interface l'affiche en permanence et permet d'en
changer sans relancer ; un profil de connexion peut le mémoriser.

```
gescom.calque.json       le constat, extrait de la base
gescom.decisions.yaml    vos arbitrages
gescom.logique.json      le modèle objet, consommé par le générateur
```

En pratique, le projet Symfony d'où partira `bin/console ormeau:generer`. Ces
fichiers se commitent : c'est ce qui rend la deuxième extraction supportable,
six mois plus tard.

Un calque porte les noms de tables, les noms de colonnes et les commentaires
métier d'une base — et, avec `--echantillonner`, des valeurs réelles. Commitez-le
dans le dépôt du client, jamais ailleurs.

## Les données de la machine

| système | emplacement |
|---|---|
| Windows | `%AppData%\ormeau\` |
| Linux | `$XDG_CONFIG_HOME/ormeau/`, sinon `~/.config/ormeau/` |
| macOS | `~/Library/Application Support/ormeau/` |

Créé au premier lancement, en `0700` — le fichier de profils porte des hôtes et
des comptes de bases de production, que les autres comptes de la machine n'ont
pas à lire.

```
ormeau/
  preferences.yaml     thème, langue, tailles des panneaux
  profils.yaml         connexions enregistrées
  cle.bin              clé de chiffrement des mots de passe enregistrés
  etat/                jetable — s'efface sans rien perdre
    LISEZMOI.txt
    sessions/          brouillons d'arbitrage en cours
```

**`ORMEAU_CONFIG_DIR` remplace cet emplacement.** Pratique pour un usage
portable — le binaire et sa configuration sur la même clé — et c'est aussi ce
qui permet aux tests de ne pas écrire dans votre configuration réelle.

### preferences.yaml

Thème, langue, et la taille des panneaux que vous avez réglés. Elles vivent là
et non dans le navigateur pour une raison précise : l'interface écoute sur un
port tiré à chaque lancement, l'origine de la page change donc à chaque fois, et
tout stockage local repartirait vide.

Un fichier illisible ou mal formé n'empêche pas de démarrer : l'interface
s'ouvre sur ses valeurs par défaut et le dit une fois.

### profils.yaml

Les connexions que vous avez nommées : SGBD, hôte, port, utilisateur, base, et
le répertoire de travail associé. Choisir un profil vous emmène donc aussi dans
le bon projet.

Un profil se crée en donnant un nom avant de se connecter — il est enregistré
une fois la connexion réussie, et il n'y a pas d'autre bouton. Un profil qui
nomme une base y garde la session : les autres bases du serveur ne sont pas
proposées. C'est un cadrage de travail et non une restriction de droits, que
seul le serveur peut poser.

### Le mot de passe enregistré

Seulement si vous cochez la case prévue, décochée par défaut. Il est alors
chiffré en AES-256-GCM avec `cle.bin`, tirée au hasard à la première écriture.

**Ce que cela protège** : une lecture accidentelle — un fichier ouvert par
erreur, une sauvegarde parcourue, un `cat` pendant un partage d'écran.

**Ce que cela ne protège pas** : quelqu'un qui a accès à votre session. La clé
vit sur le même disque que le fichier qu'elle chiffre. C'est le même modèle que
DBeaver ou SSMS, et [`SECURITY.fr.md`](../SECURITY.fr.md) le détaille. Qui a
besoin davantage laisse la case décochée : le mot de passe est alors ressaisi à
chaque connexion et n'est écrit nulle part.

`cle.bin` effacée, ou `profils.yaml` recopié depuis une autre machine : les
profils restent utilisables, sans leur mot de passe, et l'écran le dit une fois.

### etat/

Jetable, et son `LISEZMOI.txt` le rappelle. On y trouve les brouillons
d'arbitrage : ce que vous avez décidé sans l'avoir encore enregistré dans
`<base>.decisions.yaml`.

Un brouillon porte l'empreinte du calque et celle du fichier de décisions dont
il découle. Si l'un des deux a changé — une réextraction, un fichier retouché
dans un éditeur —, il est écarté et l'écran l'annonce : **le fichier de
décisions fait foi**, et rejouer un brouillon par-dessus une version antérieure
réintroduirait en silence ce qu'on venait d'en retirer.

Son nom porte la base et une empreinte du répertoire de travail : deux projets
ayant chacun une base `gescom` ne se confondent pas.

## Ce qui ne survit pas à un relancement

L'interface tire un port libre à chaque lancement, ce qui a deux conséquences
visibles.

Le **cookie de session** meurt avec l'onglet, et le jeton de l'URL ne sert
qu'une fois : rouvrir une adresse notée plus tôt ne donne pas accès à l'API.

Le **fragment d'URL** — `#arbitrage/gescom` — survit à un rechargement de la
page, pas à un relancement du binaire : l'adresse a changé de port. C'est voulu,
et c'est pourquoi le brouillon d'arbitrage, lui, est rangé dans `etat/`.
