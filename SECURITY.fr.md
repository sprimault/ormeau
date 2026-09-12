> [🇬🇧 English](SECURITY.md) · [🇫🇷 Français](SECURITY.fr.md)

# Sécurité

## Signaler une vulnérabilité

**Ne pas ouvrir d'issue publique pour une faille de sécurité.**

Utiliser le bouton **Report a vulnerability** de l'onglet Security du dépôt
(signalement privé GitHub). Le rapport reste confidentiel jusqu'à la
publication d'un correctif.

Aucune adresse de contact n'est publiée, et il n'existe pas d'autre canal.

## Versions prises en charge

Ormeau en est à ses débuts et rien n'est encore publié. Quand des versions
existeront, seule la dernière sera corrigée ; il n'y aura pas de rétroportage.

## Périmètre

Ormeau lit une base et écrit des fichiers. Il n'écrit jamais dans la base qu'il
introspecte, et c'est un invariant, pas une intention. Deux choses appellent de
l'attention et définissent l'essentiel de ce qui compte ici comme une faille :
**le DSN est le seul secret manipulé**, et **un calque est le schéma de la base
d'un client**.

Sont **dans** le périmètre :

- le DSN, ou son mot de passe, apparaissant là où il ne devrait pas — une ligne
  de journal, un message d'erreur, le calque, une réponse d'API de l'interface
  locale, un vidage mémoire ;
- toute écriture atteignant la base introspectée, y compris pendant
  l'échantillonnage ;
- une injection SQL par un identifiant du catalogue — un nom de table, de
  colonne ou de schéma qui s'échapperait dans une requête au lieu d'être
  échappé ;
- une chaîne SQL arbitraire atteignant la base depuis l'interface locale :
  l'API expose des points d'entrée fixes, et aucune requête n'a à transiter
  depuis le navigateur ;
- l'interface locale écoutant ailleurs que sur l'adresse de bouclage, ou
  joignable depuis le réseau ;
- des données métier se retrouvant dans un calque au-delà du plafond de
  cardinalité configuré, ou sans que `--echantillonner` ait été demandé ;
- une traversée de répertoire à l'écriture d'un calque, d'un fichier de
  décisions ou d'entités générées ;
- une exécution de code à la lecture d'un calque, d'un fichier de décisions ou
  d'une entité existante ;
- une dépendance vulnérable réellement atteignable depuis le code d'Ormeau.

### Mot de passe enregistré

L'écran de connexion propose d'enregistrer le mot de passe d'un profil, sur une
case décochée par défaut. Il est alors chiffré en AES-256-GCM dans
`profils.yaml`, avec une clé propre à l'installation, tirée au hasard et rangée
à côté en `0600`.

**Ce que cela protège** : une lecture accidentelle. Un fichier ouvert par
erreur, une sauvegarde parcourue, un `cat` pendant un partage d'écran, un
regard par-dessus l'épaule.

**Ce que cela ne protège pas** : quelqu'un qui a accès à votre session. La clé
vit sur le même disque que le fichier qu'elle chiffre, et qui peut lire l'un
peut lire l'autre. C'est le même modèle que DBeaver ou SSMS, et il est dit ici
plutôt que laissé à deviner. Qui a besoin davantage laisse la case décochée : le
mot de passe est alors ressaisi à chaque connexion et n'est écrit nulle part.

Ne sont **pas** des vulnérabilités :

- l'avertissement SmartScreen sous Windows et le blocage Gatekeeper sous macOS.
  Les binaires ne sont ni signés ni notariés ; c'est documenté dans
  [`docs/construction.fr.md`](docs/construction.fr.md), avec la raison ;
- un calque contenant des noms de tables, de colonnes et des commentaires
  métier. C'est sa fonction. Protéger le fichier une fois produit relève de
  l'exploitant — le README le dit, et le `.gitignore` tient les calques hors de
  ce dépôt ;
- le fait de se connecter au DSN fourni par l'utilisateur, y compris vers un
  hôte qui ne lui appartient pas. Le choix de la base relève de l'exploitant,
  pas de l'outil ;
- la valeur `POSTGRES_PASSWORD` de `tests/docker-compose.yml`. Elle appartient
  à un conteneur de test éphémère monté depuis `tests/ddl/`, et ne donne accès
  à rien ;
- l'épuisement de ressources provoqué par l'introspection d'un très gros
  schéma. L'outil tourne sur la machine de l'exploitant, contre une base dont
  il a déjà les accès ;
- le fait de pouvoir déchiffrer un mot de passe enregistré depuis la session de
  l'utilisateur qui l'a enregistré. C'est ce que la section ci-dessus annonce,
  et la case à cocher le dit à l'écran ;
- une sortie de scanner automatique sans reproduction.

## Traitement

Un rapport doit porter une reproduction : quelle version, quelle commande, quel
SGBD, et ce qu'un attaquant obtient que les sections ci-dessus ne lui donnent
pas déjà. Un rapport sans reproduction est clos.

Le projet est maintenu bénévolement, sans délai de traitement garanti. Les
rapports sont traités au mieux, le plus grave d'abord. Il n'y a ni prime aux
bogues, ni engagement de service.

**Ne jamais joindre un calque extrait d'une base de production**, ni à un
rapport ni à quoi que ce soit d'autre. Si une reproduction en exige un, le
construire depuis `tests/ddl/` ou le réduire aux quelques objets qui
déclenchent le défaut.
