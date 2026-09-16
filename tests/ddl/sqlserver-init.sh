#!/bin/sh
# Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
# SPDX-License-Identifier: Apache-2.0

# Contrôle de santé du conteneur SQL Server, et son initialisation.
#
# L'image n'a pas de répertoire d'initialisation : le DDL se charge au premier
# contrôle où le serveur répond, et le conteneur ne se dit prêt qu'une fois
# l'empreinte du DDL posée. C'est elle, et non l'existence de la base, qui dit
# que le chargement est allé au bout — gescom existe dès la première
# instruction du script.
#
# L'empreinte est une propriété étendue de la base, pendant du commentaire de
# base de 20-empreinte.sh : les tests la relisent et refusent de tourner si
# elle ne correspond pas au fichier du dépôt. Une propriété de niveau base
# n'entre pas dans le calque, qui ne lit que celles des tables et des colonnes.
set -e

sqlcmd() {
    /opt/mssql-tools18/bin/sqlcmd -C -S localhost -U sa -P "$MSSQL_SA_PASSWORD" -b "$@"
}

pret="IF NOT EXISTS (SELECT 1 FROM sys.extended_properties WHERE class = 0 AND name = N'ormeau_ddl') RAISERROR('ddl absent', 16, 1)"
if sqlcmd -d gescom -Q "$pret" >/dev/null 2>&1; then
    exit 0
fi

# Le serveur ne répond pas encore : rien à charger, contrôle suivant.
sqlcmd -Q "SELECT 1" >/dev/null 2>&1 || exit 1

sqlcmd -i /ddl/sqlserver.sql
empreinte=$(sha256sum /ddl/sqlserver.sql | cut -d ' ' -f 1)
sqlcmd -d gescom -Q "EXEC sp_addextendedproperty @name = N'ormeau_ddl', @value = N'ddl-sha256:$empreinte'"
exit 1
