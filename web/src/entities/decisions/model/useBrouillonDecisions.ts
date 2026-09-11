// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useEffect, useState } from 'react';

import { ErreurAPI } from '@/shared/api';
import { translate } from '@/shared/i18n';
import type { Decisions, ReponseDecisions } from '@/shared/model';
import { lireDecisions } from '../api/decisionsApi';

/** Le brouillon de décisions d'une base, et le fichier dont il part. */
export interface BrouillonDecisions {
  base: string;
  /** Les décisions en cours, telles qu'elles partiraient à l'enregistrement. */
  decisions: Decisions;
  /** Le fichier tel qu'il a été lu ; nul tant que la lecture n'a pas abouti. */
  fichier: ReponseDecisions | null;
  enCours: boolean;
  erreur: string | null;
  /** Applique une transformation au brouillon, jamais au fichier lu. */
  modifier: (transformation: (decisions: Decisions) => Decisions) => void;
}

/**
 * Tient le brouillon de décisions d'une base.
 *
 * Il part du fichier quand la base a déjà été arbitrée, d'un brouillon vide
 * sinon. Le fichier lu est gardé à part : c'est lui qui dira, à
 * l'enregistrement, si le disque a bougé entre-temps ou s'il porte un travail
 * écrit à la main.
 *
 * Changer de base repart de zéro et abandonne la lecture en cours : une réponse
 * arrivée après coup mettrait les décisions d'une base dans le brouillon d'une
 * autre.
 */
export function useBrouillonDecisions(base: string): BrouillonDecisions {
  const [decisions, setDecisions] = useState<Decisions>({});
  const [fichier, setFichier] = useState<ReponseDecisions | null>(null);
  const [enCours, setEnCours] = useState(true);
  const [erreur, setErreur] = useState<string | null>(null);

  useEffect(() => {
    const abandon = new AbortController();
    setDecisions({});
    setFichier(null);
    setEnCours(true);
    setErreur(null);

    lireDecisions(base, abandon.signal)
      .then((reponse) => {
        setFichier(reponse);
        setDecisions(reponse.decisions);
        setEnCours(false);
      })
      .catch((echec: unknown) => {
        if (abandon.signal.aborted) {
          return;
        }
        setErreur(echec instanceof ErreurAPI ? echec.message : translate('error.unknown'));
        setEnCours(false);
      });

    return () => abandon.abort();
  }, [base]);

  const modifier = useCallback((transformation: (decisions: Decisions) => Decisions) => {
    setDecisions((precedentes) => transformation(precedentes));
  }, []);

  return { base, decisions, fichier, enCours, erreur, modifier };
}
