// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

/**
 * Mise en forme des identifiants d'objets de base.
 *
 * C'est le seul endroit où l'on compose un nom de table. Les identifiants
 * s'affichent tels qu'ils sont en base — jamais traduits, jamais normalisés :
 * c'est ce que l'utilisateur retrouvera dans son SGBD.
 */

/**
 * Qualifie une table par son schéma, sous la forme que le serveur emploie dans
 * `reference_vers`.
 *
 * C'est aussi la clé de la sélection : deux tables homonymes dans deux schémas
 * sont deux tables, et les confondre en cocherait une pour l'autre.
 */
export function qualifier(schema: string, nom: string): string {
  return `${schema}.${nom}`;
}
