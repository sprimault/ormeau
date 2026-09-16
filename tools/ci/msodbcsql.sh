#!/usr/bin/env bash
# Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
# SPDX-License-Identifier: Apache-2.0
#
# Installe le pilote ODBC 18 de Microsoft sur un runner Ubuntu.
#
#   tools/ci/msodbcsql.sh
#
# pdo_sqlsrv, que setup-php installe, ne parle à SQL Server qu'à travers ce
# pilote : sans lui, l'extension se charge et refuse ensuite toute connexion.
# Les deux jobs d'aller-retour SQL Server l'appellent ; un script plutôt que
# deux copies des mêmes lignes dans ci.yml.
#
# Le paquet vient du dépôt APT de Microsoft, clé vérifiée, pour la version
# d'Ubuntu du runner. ACCEPT_EULA : le paquet refuse de s'installer sans.
# --batch --yes : l'image du runner porte déjà la clé, et gpg demanderait
# sinon, sur un terminal absent, s'il faut l'écraser.
set -euo pipefail

version=$(lsb_release -rs)
nom=$(lsb_release -cs)

curl -fsSL https://packages.microsoft.com/keys/microsoft.asc \
    | sudo gpg --batch --yes --dearmor -o /usr/share/keyrings/microsoft-prod.gpg
echo "deb [arch=amd64 signed-by=/usr/share/keyrings/microsoft-prod.gpg] https://packages.microsoft.com/ubuntu/${version}/prod ${nom} main" \
    | sudo tee /etc/apt/sources.list.d/mssql-release.list >/dev/null
sudo apt-get update -q
sudo ACCEPT_EULA=Y apt-get install -y -q msodbcsql18
