// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

/**
 * Rend une volumétrie sous forme compacte : 48 210 devient « 48 k ».
 *
 * Un arbre de quatre cents tables affiche autant de nombres, alignés en colonne
 * étroite. L'ordre de grandeur est ce qu'on y cherche — savoir si une table
 * pèse mille ou un million de lignes —, jamais le chiffre exact, qui n'est de
 * toute façon qu'une estimation du catalogue.
 */
export function compact(valeur: number): string {
  if (valeur < 1000) {
    return String(valeur);
  }
  const unites = ['k', 'M', 'G', 'T'];
  let reste = valeur;
  let rang = -1;
  while (reste >= 1000 && rang < unites.length - 1) {
    reste /= 1000;
    rang += 1;
  }
  // Une décimale en dessous de dix, aucune au-delà : « 4,8 k » est utile,
  // « 482,1 k » ne l'est pas.
  const arrondi = reste < 10 ? Math.round(reste * 10) / 10 : Math.round(reste);
  return `${String(arrondi).replace('.', ',')} ${unites[rang]}`;
}
