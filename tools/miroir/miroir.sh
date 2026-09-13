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
# pas écraser.
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

if ! git push "$distant" "$split:$reference"; then
  cat >&2 <<EOF
Le miroir refuse $reference.

Pour une branche : le miroir a reçu un commit qui ne vient pas du split, et
il ne se corrige pas en force. Pour un tag : il existe déjà sur un autre
commit. La marche à suivre est dans docs/construction.fr.md, « Le miroir du
paquet PHP ».
EOF
  exit 1
fi

# Le push a pu réussir sans rien écrire, si la référence existait déjà : ce qui
# compte est ce que le miroir porte.
recu=$(git ls-remote "$distant" "$reference" | cut -f1)
if [ "$recu" != "$split" ]; then
  echo "le miroir porte $recu pour $reference, le split attendu est $split" >&2
  exit 1
fi
echo "$reference du miroir : $split"
