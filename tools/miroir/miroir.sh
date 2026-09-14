#!/usr/bin/env bash
# Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
# SPDX-License-Identifier: Apache-2.0
#
# Pousse le split de php/ vers le dépôt miroir du paquet Composer.
#
#   tools/miroir/miroir.sh <dépôt distant> <référence> <commit>
#
# La référence est refs/heads/master ou refs/tags/vX.Y.Z : celle que le
# miroir reçoit, qui porte le split du commit donné.
#
# git subtree split est déterministe — le même commit donne le même split à
# chaque passage — et le split d'un commit plus ancien est ancêtre de celui
# d'un commit plus récent, y compris à travers un commit de fusion. Le miroir
# avance donc toujours en avance rapide, et rien ne se pousse jamais en force :
# un refus signifie que le miroir a divergé, et c'est ce qu'il faut examiner,
# pas écraser. Seule exception, reconnue plus bas : le split plus ancien d'un
# run relancé.
set -euo pipefail

if [ "$#" -ne 3 ]; then
  echo "usage : $0 <dépôt distant> <référence> <commit>" >&2
  exit 2
fi
distant=$1
reference=$2
commit=$3

case "$reference" in
  refs/heads/master | refs/tags/v*) ;;
  *)
    echo "référence refusée : $reference ; le miroir ne reçoit que master et les tags v*" >&2
    exit 2
    ;;
esac

split=$(git subtree split -q --prefix=php "$commit")
echo "split de php/ à $commit : $split"

procedure='La marche à suivre est dans docs/construction.fr.md, « Si la publication échoue ».'

# La sortie du push est gardée pour nommer le refus. LC_ALL=C fixe les messages
# de git ; ceux de ssh ne sont jamais traduits.
journal=$(mktemp)
trap 'rm -f "$journal"' EXIT
if ! LC_ALL=C git push "$distant" "$split:$reference" 2>"$journal"; then
  cat "$journal" >&2

  # Hôte injoignable et clé refusée finissent tous deux sur « Could not read
  # from remote repository » : seule la ligne de ssh qui précède les sépare. Le
  # premier est passager, le second ne l'est pas.
  if grep -qE 'Could not resolve hostname|Connection (refused|timed out)' "$journal"; then
    echo "Le miroir est injoignable, rien n'a été poussé : cause passagère, relancer le run." >&2
    exit 1
  fi
  if grep -qF 'Permission denied (publickey)' "$journal"; then
    cat >&2 <<EOF
Le miroir refuse la clé, rien n'a été poussé : MIROIR_CLE ne correspond à
aucune clé de déploiement du miroir. La remplacer des deux côtés, comme à la
mise en place (docs/construction.fr.md, « Ce qui protège l'écriture »).
EOF
    exit 1
  fi

  case "$reference" in
    refs/heads/*)
      # Relancer un run dépassé pousse un split plus ancien que le miroir : refus
      # d'avance rapide, sans rien d'étranger. Le split doit être ancêtre de la
      # tête du miroir, et cette tête ancêtre du split de la même branche du
      # dépôt principal (origin/…, que le checkout à fetch-depth: 0 fournit) :
      # l'ascendance seule accepterait un commit étranger posé au-dessus d'un
      # split.
      #
      # Les branches seulement. Un tag du miroir ne se déplace pas : il porte
      # exactement le split de son commit, jamais un descendant.
      branche=${reference#refs/heads/}
      if git fetch -q "$distant" "$reference" &&
        git merge-base --is-ancestor "$split" FETCH_HEAD &&
        tete=$(git subtree split -q --prefix=php "refs/remotes/origin/$branche") &&
        git merge-base --is-ancestor FETCH_HEAD "$tete"; then
        echo "$reference du miroir : $(git rev-parse FETCH_HEAD), plus récent, dont le split $split est ancêtre ; rien à pousser"
        exit 0
      fi
      if grep -qE '\((non-fast-forward|fetch first)\)' "$journal"; then
        cat >&2 <<EOF

Le miroir refuse $reference : il porte un commit qui ne vient pas du split, et
il ne se corrige pas en force. $procedure
EOF
      else
        echo "Le miroir refuse $reference. $procedure" >&2
      fi
      ;;
    *)
      if grep -qF '(already exists)' "$journal"; then
        cat >&2 <<EOF

Le miroir refuse $reference : il existe déjà sur un autre split, laissé par une
tentative précédente. $procedure
EOF
      else
        echo "Le miroir refuse $reference. $procedure" >&2
      fi
      ;;
  esac
  exit 1
fi
cat "$journal" >&2

# Le push a pu réussir sans rien écrire, si la référence existait déjà : ce qui
# compte est ce que le miroir porte.
recu=$(git ls-remote "$distant" "$reference" | cut -f1)
if [ "$recu" != "$split" ]; then
  echo "le miroir porte $recu pour $reference, le split attendu est $split" >&2
  exit 1
fi
echo "$reference du miroir : $split"
