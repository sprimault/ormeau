// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useEffect, useState } from 'react';

import type { EtatExtractions, Extraction, ReferenceExtraction } from '@/shared/model';
import { ouvrirFlux } from '../api/extractionApi';

/** Ce que le flux dit des tâches. */
export interface SuiviExtractions {
  /** Toutes les tâches que le serveur connaît, dans leur ordre d'arrivée. */
  extractions: Extraction[];
  /** Faux tant que le flux n'est pas ouvert, et pendant une reconnexion. */
  connecte: boolean;
}

/**
 * Suit les tâches d'extraction par le flux du serveur.
 *
 * Un seul flux pour toute l'interface : un navigateur n'ouvre que six connexions
 * par origine, et un flux par écran ou par tâche les épuiserait. C'est pourquoi
 * ce hook n'est appelé que par le fournisseur de la racine, et lu ailleurs par
 * contexte.
 *
 * Chaque événement porte l'état complet d'une tâche : l'appliquer, c'est la
 * remplacer. Un instantané remplace la liste entière — le serveur l'envoie à
 * l'ouverture, et quand il ne sait pas reprendre là où le navigateur en était.
 */
export function useSuiviExtractions(): SuiviExtractions {
  const [extractions, setExtractions] = useState<Extraction[]>([]);
  const [connecte, setConnecte] = useState(false);

  useEffect(() => {
    const source = ouvrirFlux();
    source.addEventListener('open', () => setConnecte(true));
    // EventSource se reconnecte seul : l'erreur signale une coupure, pas un
    // abandon, et l'instantané ou la reprise suivront.
    source.addEventListener('error', () => setConnecte(false));

    ecouter<EtatExtractions>(source, 'etat', (etat) => setExtractions(etat.extractions));
    ecouter<Extraction>(source, 'extraction', (extraction) =>
      setExtractions((precedentes) => remplacer(precedentes, extraction)),
    );
    ecouter<ReferenceExtraction>(source, 'retrait', ({ id }) =>
      setExtractions((precedentes) => precedentes.filter((e) => e.id !== id)),
    );

    return () => source.close();
  }, []);

  return { extractions, connecte };
}

/** Abonne un traitement à un événement nommé du flux, données décodées. */
function ecouter<T>(source: EventSource, nom: string, traiter: (donnees: T) => void) {
  source.addEventListener(nom, (evenement) => {
    traiter(JSON.parse((evenement as MessageEvent<string>).data) as T);
  });
}

/** Remplace une tâche à sa place, ou l'ajoute en fin de liste. */
function remplacer(extractions: Extraction[], extraction: Extraction): Extraction[] {
  const rang = extractions.findIndex((e) => e.id === extraction.id);
  if (rang === -1) {
    return [...extractions, extraction];
  }
  return extractions.map((e, i) => (i === rang ? extraction : e));
}
