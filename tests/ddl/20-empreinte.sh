#!/bin/sh
# Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
# SPDX-License-Identifier: Apache-2.0

# Pose l'empreinte du DDL en commentaire de la base, à la création du volume.
#
# L'entrée du conteneur n'exécute ses scripts que sur un volume vide : un DDL
# modifié après coup laisse la base sur l'ancien schéma, sans rien signaler.
# Les tests d'intégration relisent ce commentaire et refusent de tourner s'il
# ne correspond pas au fichier du dépôt. Un commentaire de base n'entre ni dans
# le calque ni dans la liste des bases : le contrôle ne touche pas ce qu'il
# protège.
#
# L'entrée l'exécute s'il est exécutable, le source sinon : pas de set -u, qui
# survivrait dans son propre shell.
set -e

empreinte=$(sha256sum /docker-entrypoint-initdb.d/10-gescom.sql | cut -d ' ' -f 1)
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" \
    -c "COMMENT ON DATABASE \"$POSTGRES_DB\" IS 'ddl-sha256:$empreinte'"
