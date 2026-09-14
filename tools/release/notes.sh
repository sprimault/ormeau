#!/usr/bin/env bash
# Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
# SPDX-License-Identifier: Apache-2.0
#
# Lit dans le CHANGELOG la section d'une version et en écrit les notes.
#
#   tools/release/notes.sh <tag> <fichier de notes> [CHANGELOG]
#
# Écrit les notes dans le fichier, et sur la sortie standard la ligne
# « titre=… » que release.yml ajoute à GITHUB_OUTPUT. Échoue quand la section
# manque ou ne porte aucun texte : une version sans notes ne dit ni ce qui
# change, ni ce qu'un calque ou un fichier de décisions existant doit reprendre,
# et c'est ce contrôle qui arrête la publication.
#
# Deux jobs l'appellent sur le même commit, l'un pour contrôler avant le
# miroir, l'autre pour écrire les notes de la release : la même lecture, sans
# que le texte transite par une sortie de job, où une expression le recopierait
# dans un script.
set -euo pipefail

if [ "$#" -lt 2 ] || [ "$#" -gt 3 ]; then
  echo "usage : $0 <tag> <fichier de notes> [CHANGELOG]" >&2
  exit 2
fi
tag=$1
notes=$2
changelog=${3:-CHANGELOG.md}
version=${tag#v}

# index et non une expression régulière : les crochets du titre de section
# formeraient une classe de caractères, et « [0.4.0] » ne désignerait plus
# qu'un seul caractère.
awk -v v="$version" '
  index($0, "## [" v "]") == 1 { dedans = 1; next }
  dedans && index($0, "## ") == 1 { exit }
  dedans { print }
' "$changelog" > "$notes"

# Une section faite de lignes vides passerait un test de taille : c'est du
# texte qui est exigé, pas des octets.
if ! grep -q '[^[:space:]]' "$notes"; then
  echo "$changelog n'a pas de section pour la version $version" >&2
  exit 1
fi

# Le titre est la troisième partie du titre de section, après la version et la
# date. Absent, la release se nomme par son tag seul.
titre=$(grep -m1 -F "## [$version]" "$changelog" \
  | sed -n 's/^## \[[^]]*\][^—]*—[^—]*— *//p')
echo "titre=${tag}${titre:+ — $titre}"
