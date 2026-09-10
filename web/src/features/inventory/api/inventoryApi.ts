// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { getJSON } from '@/shared/api';
import type { ReponseColonnes, ReponseInventaire } from '@/shared/model';

/**
 * Lit l'inventaire d'une connexion ouverte.
 *
 * Une passe légère : le serveur interroge le catalogue et ne lit aucune donnée.
 * L'inventaire arrive entier, sans pagination — quelques dizaines de kilo-octets
 * pour quatre cents tables, contre un aller-retour par frappe si l'on filtrait
 * côté serveur.
 */
export function lireInventaire(
  session: string,
  schemas: string[],
  signal?: AbortSignal,
): Promise<ReponseInventaire> {
  const parametres = new URLSearchParams({ session });
  if (schemas.length > 0) {
    parametres.set('schemas', schemas.join(','));
  }
  return getJSON<ReponseInventaire>(`/api/inventaire?${parametres}`, signal);
}

/**
 * Lit les colonnes d'une table, au dépliement.
 *
 * Ce qui s'y coche ne restreint pas l'extraction : le calque garde toutes les
 * colonnes. Les colonnes décochées alimentent `colonnes_ignorees` dans le
 * fichier de décisions, qui les retire de l'entité générée.
 */
export function lireColonnes(
  session: string,
  schema: string,
  table: string,
  signal?: AbortSignal,
): Promise<ReponseColonnes> {
  const parametres = new URLSearchParams({ session, schema, table });
  return getJSON<ReponseColonnes>(`/api/colonnes?${parametres}`, signal);
}
