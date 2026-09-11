// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useState } from 'react';

/** Clé de persistance du repli : une préférence d'affichage, comme la largeur
 *  de l'arbre, jamais rien qui touche à la base. */
const CLE_OUVERTURE = 'ormeau-apercu-ouvert';

/**
 * Tient l'ouverture de l'aperçu du bas, retenue d'un lancement à l'autre.
 *
 * Ouvert par défaut : c'est le résultat du travail de l'écran, il n'a pas à se
 * chercher. Tenue au-dessus de l'aperçu parce que le séparateur qui le
 * dimensionne en dépend — replié, la zone se réduit à sa barre et la poignée
 * disparaît.
 */
export function useApercuOuvert(): [boolean, () => void] {
  const [ouvert, setOuvert] = useState(ouvertureEnregistree);

  const basculer = useCallback(() => {
    const suivant = !ouvert;
    setOuvert(suivant);
    try {
      window.localStorage.setItem(CLE_OUVERTURE, suivant ? '1' : '0');
    } catch {
      // Stockage indisponible : le repli vaut pour cette session.
    }
  }, [ouvert]);

  return [ouvert, basculer];
}

/** Relit le repli choisi ; ouvert quand rien n'a été retenu. */
function ouvertureEnregistree(): boolean {
  try {
    return window.localStorage.getItem(CLE_OUVERTURE) !== '0';
  } catch {
    return true;
  }
}
