// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import {
  CodeCasEnumerationOpaque,
  CodeCibleHorsPortee,
  CodeClePrimaireComposite,
  CodeClePrimaireGardee,
  CodeCollationNonReportee,
  CodeCollision,
  CodeColonneIgnoree,
  CodeDecisionInvalide,
  CodeDecisionOrpheline,
  CodeDefautIncompatible,
  CodeDefautNonReporte,
  CodeFuseauPrecisionNonLue,
  CodeHeritageDeduit,
  CodeJointurePure,
  CodeJSONSansUnicode,
  CodePrefixeDetecte,
  CodeReferenceHorsIdentifiant,
  CodeSequenceNonReconnue,
  CodeTableIgnoree,
  CodeTableSansClePrimaire,
  CodeTexteUnicodeJSONPropose,
  CodeTexteUnicodeSansEquivalent,
  CodeTraitDeduit,
  CodeTypeNonReconnu,
  CodeUUIDAFournir,
  type Avertissement,
} from '@/shared/model';

/**
 * Codes qui ne font que rappeler un choix fait dans l'onglet de sélection. Le
 * détail montre la colonne écartée ; le répéter en avertissement ferait croire
 * à un problème.
 */
const RAPPELS = new Set<string>([CodeTableIgnoree, CodeColonneIgnoree]);

/**
 * Où un avertissement se traite. Seul `ecran` a une action dans l'arbitrage ;
 * les autres disent où aller, et ne comptent pas comme à traiter ici.
 */
export type Lieu = 'ecran' | 'selection' | 'fichier' | 'base' | 'application' | 'information';

/**
 * Le lieu de chaque code d'avertissement.
 *
 * Écrit à la main et contrôlé par un test qui échoue sur un code sans lieu :
 * ranger un code, c'est dire où l'utilisateur agit, et la réponse ne se déduit
 * ni du code ni de son message.
 */
export const LIEUX: Readonly<Record<string, Lieu>> = {
  [CodeTypeNonReconnu]: 'ecran',
  [CodeCasEnumerationOpaque]: 'ecran',
  [CodeCollision]: 'ecran',
  [CodeTexteUnicodeJSONPropose]: 'ecran',

  // Une clé visée par colonnes_ignorees reste : la sélection se reprend.
  [CodeClePrimaireGardee]: 'selection',
  [CodeTableSansClePrimaire]: 'selection',
  [CodeCibleHorsPortee]: 'selection',

  [CodeDecisionOrpheline]: 'fichier',
  [CodeDecisionInvalide]: 'fichier',
  [CodePrefixeDetecte]: 'fichier',

  [CodeFuseauPrecisionNonLue]: 'base',
  [CodeReferenceHorsIdentifiant]: 'base',
  [CodeCollationNonReportee]: 'base',
  // Conséquence d'une décision prise : ce qui reste à faire, c'est ne pas
  // appliquer la migration que Doctrine proposera.
  [CodeJSONSansUnicode]: 'base',

  [CodeDefautNonReporte]: 'application',
  [CodeSequenceNonReconnue]: 'application',
  [CodeUUIDAFournir]: 'application',

  [CodeTraitDeduit]: 'information',
  [CodeJointurePure]: 'information',
  [CodeHeritageDeduit]: 'information',
  [CodeClePrimaireComposite]: 'information',
  [CodeDefautIncompatible]: 'information',
  [CodeTexteUnicodeSansEquivalent]: 'information',
  [CodeTableIgnoree]: 'information',
  [CodeColonneIgnoree]: 'information',
};

/**
 * Rend le lieu d'un avertissement. Un code inconnu, ajouté côté Go sans être
 * rangé ici, compte comme à traiter à l'écran : il se voit plutôt que de passer
 * inaperçu, et le test de rangement échoue de toute façon.
 */
export function lieu(avertissement: Avertissement): Lieu {
  return LIEUX[avertissement.code] ?? 'ecran';
}

/** Dit si un avertissement se traite dans l'écran d'arbitrage. */
export function estATraiter(avertissement: Avertissement): boolean {
  return lieu(avertissement) === 'ecran';
}

/**
 * Range les avertissements sous l'entité de la table qu'ils visent.
 *
 * La cible qualifie la table, puis éventuellement la colonne : `public.clients`
 * ou `public.clients.statut`. La table retenue est la plus longue qui la
 * préfixe, pour qu'un point dans un nom ne rattache pas une colonne à la
 * mauvaise table. Les rappels de choix sont écartés, et ce qui ne vise aucune
 * de ces tables n'apparaît pas. Chaque liste va du plus douteux au plus sûr.
 */
export function avertissementsParTable(
  avertissements: Avertissement[],
  tables: string[],
): Map<string, Avertissement[]> {
  const parLongueur = [...tables].sort((a, b) => b.length - a.length);
  const parTable = new Map<string, Avertissement[]>();

  for (const avertissement of trierAvertissements(avertissements)) {
    if (RAPPELS.has(avertissement.code)) {
      continue;
    }
    const { cible } = avertissement;
    const table = parLongueur.find((t) => cible === t || cible.startsWith(`${t}.`));
    if (table !== undefined) {
      parTable.set(table, [...(parTable.get(table) ?? []), avertissement]);
    }
  }
  return parTable;
}

/**
 * Trie les avertissements du plus douteux au plus sûr.
 *
 * Une confiance basse est ce que l'inférence a le moins bien tranché. À
 * confiance égale, code puis cible : l'ordre ne bouge pas d'un recalcul à
 * l'autre.
 */
export function trierAvertissements(avertissements: Avertissement[]): Avertissement[] {
  return [...avertissements].sort(
    (a, b) => a.confiance - b.confiance || comparer(a.code, b.code) || comparer(a.cible, b.cible),
  );
}

/** Compare deux chaînes sans dépendre de la langue du navigateur. */
function comparer(a: string, b: string): number {
  return a < b ? -1 : a > b ? 1 : 0;
}
