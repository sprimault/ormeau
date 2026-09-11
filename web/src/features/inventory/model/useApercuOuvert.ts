// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useState } from 'react';

import { usePreferencesStore } from '@/shared/model';

/**
 * Tient l'ouverture de l'aperçu du bas, retenue d'un lancement à l'autre.
 *
 * Ouvert par défaut : c'est le résultat du travail de l'écran, il n'a pas à se
 * chercher. Tenue au-dessus de l'aperçu parce que le séparateur qui le
 * dimensionne en dépend — replié, la zone se réduit à sa barre et la poignée
 * disparaît.
 *
 * Le repli est une préférence d'affichage, comme la largeur de l'arbre : c'est
 * le serveur qui l'enregistre, le navigateur n'en retenant rien d'un lancement
 * à l'autre.
 */
export function useApercuOuvert(): [boolean, () => void] {
  const [ouvert, setOuvert] = useState(ouvertureInjectee);

  const basculer = useCallback(() => {
    const suivant = !ouvert;
    setOuvert(suivant);
    usePreferencesStore.getState().regler({ apercu_ouvert: suivant });
  }, [ouvert]);

  return [ouvert, basculer];
}

/** Relit le repli injecté dans la page ; ouvert quand rien n'a été retenu. */
function ouvertureInjectee(): boolean {
  return usePreferencesStore.getState().preferences.apercu_ouvert ?? true;
}
