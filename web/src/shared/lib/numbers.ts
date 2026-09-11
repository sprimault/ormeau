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

/**
 * Rend une durée lisible : « < 1 s », « 42 s », « 3 min 05 s », « 1 h 12 min ».
 *
 * Sous la seconde, « < 1 s » plutôt que « 0 s » : une extraction de quelques
 * tables tient dans la seconde, et un zéro laisserait croire qu'il ne s'est rien
 * passé. Les secondes disparaissent au-delà de l'heure : elles ne disent plus
 * rien et font danser la largeur du texte.
 */
export function duree(secondes: number): string {
  if (secondes < 1) {
    return '< 1 s';
  }
  const entieres = Math.floor(secondes);
  if (entieres < 60) {
    return `${entieres} s`;
  }
  if (entieres < 3600) {
    return `${Math.floor(entieres / 60)} min ${String(entieres % 60).padStart(2, '0')} s`;
  }
  const minutes = Math.floor(entieres / 60) % 60;
  return `${Math.floor(entieres / 3600)} h ${String(minutes).padStart(2, '0')} min`;
}
