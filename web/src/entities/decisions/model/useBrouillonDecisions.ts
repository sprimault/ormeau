// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useEffect, useMemo, useState } from 'react';

import { ErreurAPI } from '@/shared/api';
import { translate } from '@/shared/i18n';
import type { Decisions, ReponseDecisions } from '@/shared/model';
import { lireDecisions } from '../api/decisionsApi';
import { decisionsEgales } from './egalite';

/** Le brouillon de décisions d'une base, et le fichier dont il part. */
export interface BrouillonDecisions {
  base: string;
  /** Les décisions en cours, telles qu'elles partiraient à l'enregistrement. */
  decisions: Decisions;
  /** Le fichier tel qu'il a été lu ou écrit en dernier ; nul tant qu'aucune
   *  lecture n'a abouti. */
  fichier: ReponseDecisions | null;
  /** Vrai dès que la première lecture a abouti ou échoué. */
  pret: boolean;
  enCours: boolean;
  erreur: string | null;
  /** Le brouillon décide autre chose que le fichier. */
  modifie: boolean;
  /** Applique une transformation au brouillon, jamais au fichier lu. */
  modifier: (transformation: (decisions: Decisions) => Decisions) => void;
  /** Relit le fichier, et remplace le brouillon par ce qu'il contient. */
  relire: () => void;
  /** Retient que ces décisions sont désormais celles du fichier. */
  enregistre: (decisions: Decisions, empreinte: string) => void;
}

/**
 * Tient le brouillon de décisions d'une base.
 *
 * Il part du fichier quand la base a déjà été arbitrée, d'un brouillon vide
 * sinon. Le fichier est gardé à part : c'est lui qui dira, à l'enregistrement,
 * si le disque a bougé entre-temps ou s'il porte un travail écrit à la main.
 *
 * Une base par brouillon : le fournisseur est remonté quand elle change. Une
 * lecture en cours est abandonnée dès qu'une autre la remplace, pour qu'une
 * réponse arrivée après coup n'écrase pas la plus récente.
 */
export function useBrouillonDecisions(base: string): BrouillonDecisions {
  const [decisions, setDecisions] = useState<Decisions>({});
  const [fichier, setFichier] = useState<ReponseDecisions | null>(null);
  const [pret, setPret] = useState(false);
  const [enCours, setEnCours] = useState(true);
  const [erreur, setErreur] = useState<string | null>(null);
  const [lecture, setLecture] = useState(0);

  useEffect(() => {
    const abandon = new AbortController();
    setEnCours(true);
    setErreur(null);

    lireDecisions(base, abandon.signal)
      .then((reponse) => {
        setFichier(reponse);
        setDecisions(reponse.decisions);
        setEnCours(false);
        setPret(true);
      })
      .catch((echec: unknown) => {
        if (abandon.signal.aborted) {
          return;
        }
        setErreur(echec instanceof ErreurAPI ? echec.message : translate('error.unknown'));
        setEnCours(false);
        setPret(true);
      });

    return () => abandon.abort();
  }, [base, lecture]);

  const modifier = useCallback((transformation: (decisions: Decisions) => Decisions) => {
    setDecisions((precedentes) => transformation(precedentes));
  }, []);

  const relire = useCallback(() => setLecture((n) => n + 1), []);

  // Le serveur rend l'empreinte du fichier écrit, pas son contenu : le
  // brouillon envoyé est ce que le fichier décide désormais, et aucun
  // commentaire manuel n'a survécu à la réécriture.
  const enregistre = useCallback((ecrites: Decisions, empreinte: string) => {
    setFichier({ existe: true, decisions: ecrites, empreinte_fichier: empreinte, manuel: false });
  }, []);

  const modifie = useMemo(
    () => !decisionsEgales(decisions, fichier?.decisions ?? {}),
    [decisions, fichier],
  );

  return { base, decisions, fichier, pret, enCours, erreur, modifie, modifier, relire, enregistre };
}
