// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import type { ReactNode } from 'react';

import { ContexteDecisions } from '../model/contexte';
import { useBrouillonDecisions } from '../model/useBrouillonDecisions';

/** Propriétés du fournisseur. */
interface ProprietesFournisseur {
  base: string;
  children: ReactNode;
}

/**
 * Tient le brouillon de décisions d'une base pour tout ce qui l'arbitre.
 *
 * Un seul brouillon, lu par l'arbre qui écarte des colonnes comme par l'écran
 * qui enregistre le fichier : deux copies finiraient par se contredire, et
 * c'est la mauvaise qui partirait sur disque.
 *
 * Rien n'est monté avant que la première lecture ait abouti : une colonne
 * écartée pendant ce temps serait écrasée par le fichier à son arrivée, sans
 * que rien ne le signale. Le fichier est local, l'attente ne se voit pas. Une
 * relecture demandée ensuite ne démonte rien : les tables cochées et la place
 * dans les listes survivent.
 */
export function DecisionsProvider({ base, children }: ProprietesFournisseur) {
  const brouillon = useBrouillonDecisions(base);
  return (
    <ContexteDecisions.Provider value={brouillon}>
      {brouillon.pret ? children : null}
    </ContexteDecisions.Provider>
  );
}
