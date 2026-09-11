// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useEffect, useState } from 'react';

import { ErreurAPI } from '@/shared/api';
import { translate } from '@/shared/i18n';
import { EtatTerminee, type Extraction, type ReponseCalque } from '@/shared/model';
import { lireCalque } from '../api/extractionApi';
import { useExtractions } from './contexte';

/** Ce que la colonne du calque affiche. */
export interface EtatCalque {
  calque: ReponseCalque | null;
  /** Message du serveur quand il n'y a rien à montrer, fichier absent compris. */
  message: string | null;
  enCours: boolean;
}

/**
 * Lit le calque d'une session, et le relit à chaque nouvelle version.
 *
 * `version` change quand le fichier a été réécrit : c'est le flux qui annonce
 * qu'un calque vient d'être écrit, le hook ne sonde pas le disque. Pendant la
 * relecture, le calque précédent reste affiché plutôt que de laisser la colonne
 * vide une fraction de seconde.
 */
export function useCalque(session: string, version: string): EtatCalque {
  const [calque, setCalque] = useState<ReponseCalque | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [enCours, setEnCours] = useState(true);

  useEffect(() => {
    const abandon = new AbortController();
    setEnCours(true);

    lireCalque(session, abandon.signal)
      .then((reponse) => {
        setCalque(reponse);
        setMessage(null);
        setEnCours(false);
      })
      .catch((echec: unknown) => {
        if (abandon.signal.aborted) {
          return;
        }
        setCalque(null);
        setMessage(echec instanceof ErreurAPI ? echec.message : translate('error.unknown'));
        setEnCours(false);
      });

    return () => abandon.abort();
  }, [session, version]);

  return { calque, message, enCours };
}

/**
 * Version du calque d'une base, lue dans le suivi des extractions.
 *
 * Ce que l'aperçu et l'arbitrage surveillent pour savoir que le fichier a pu
 * être réécrit, sans sonder le disque ni ouvrir un second flux.
 */
export function useVersionCalque(base: string): string {
  const { extractions } = useExtractions();
  return derniereFin(extractions, base);
}

/**
 * Fin de la dernière extraction terminée d'une base : la version du calque
 * qu'elle a écrit, vide tant qu'aucune ne l'a fait dans ce lancement.
 */
export function derniereFin(extractions: Extraction[], base: string): string {
  let derniere = '';
  for (const extraction of extractions) {
    if (
      extraction.base === base &&
      extraction.etat === EtatTerminee &&
      extraction.fin &&
      extraction.fin > derniere
    ) {
      derniere = extraction.fin;
    }
  }
  return derniere;
}
