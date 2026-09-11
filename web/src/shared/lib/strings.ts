// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

/**
 * Mise en forme des chaînes, miroir de ce que le Go fait de son côté.
 *
 * Un formatage écrit en ligne dans un composant est une erreur à corriger :
 * c'est ici qu'il vit, et nulle part ailleurs.
 */

/**
 * Tronque au milieu plutôt qu'à la fin.
 *
 * Un chemin de travail se reconnaît par son début et par son dernier segment ;
 * couper la fin retirerait précisément le nom du projet, qui est ce qu'on
 * cherche à vérifier d'un coup d'œil.
 */
export function tronquerMilieu(valeur: string, max = 64): string {
  if (valeur.length <= max) {
    return valeur;
  }
  const garde = Math.floor((max - 1) / 2);
  return valeur.slice(0, garde) + '…' + valeur.slice(valeur.length - (max - 1 - garde));
}

/**
 * Met en minuscule la première lettre, et elle seule.
 *
 * C'est le passage d'un nom de classe à un nom de propriété, `TenantInvoice`
 * vers `tenantInvoice` : les majuscules internes restent, sans quoi le nom ne
 * se lirait plus.
 */
export function minusculeInitiale(valeur: string): string {
  return valeur.charAt(0).toLowerCase() + valeur.slice(1);
}

/**
 * Masque le mot de passe d'une chaîne de connexion avant affichage.
 *
 * Le front ne reçoit jamais de DSN du serveur, mais il en manipule un que
 * l'utilisateur vient de saisir : le réafficher tel quel dans un message
 * d'erreur le ferait apparaître dans une capture d'écran d'issue.
 */
export function masquerDSN(dsn: string): string {
  if (!dsn.includes('://')) {
    return dsn ? '***' : '';
  }
  return dsn.replace(/(:\/\/[^:@/]+):[^@]*@/, '$1:***@');
}
