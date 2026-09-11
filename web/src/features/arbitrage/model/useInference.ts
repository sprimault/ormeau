// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useEffect, useRef, useState } from 'react';

import { ErreurAPI } from '@/shared/api';
import { translate } from '@/shared/i18n';
import { useDiffere } from '@/shared/lib';
import { CodeCalqueModifie, type Decisions, type ReponseInference } from '@/shared/model';
import { inferer } from '../api/arbitrageApi';

/** Attente après la dernière modification du brouillon avant de recalculer. */
const DELAI = 300;

/** Ce que l'écran d'arbitrage sait de l'inférence. */
export interface EtatInference {
  /** La dernière inférence aboutie, gardée affichée pendant un recalcul. */
  resultat: ReponseInference | null;
  /** Le brouillon dont `resultat` est l'effet. */
  decisionsJugees: Decisions;
  enCours: boolean;
  erreur: string | null;
  /** Le calque a été réécrit depuis la première inférence. */
  calqueModifie: boolean;
  /** Oublie le calque jugé et repart de celui du répertoire. */
  recharger: () => void;
}

/**
 * Rejoue l'inférence à chaque modification du brouillon.
 *
 * La première réponse apprend l'empreinte du calque, et chaque appel suivant la
 * renvoie : si une extraction a réécrit le fichier entre-temps, le serveur
 * refuse, et l'écran le dit au lieu d'afficher l'effet de décisions prises sur
 * un autre schéma.
 *
 * `versionCalque` change quand une extraction de la base se termine. Il relance
 * l'inférence avec l'empreinte connue : le refus arrive aussitôt, sans attendre
 * la prochaine modification. Une extraction qui réécrit un calque identique ne
 * change rien.
 */
export function useInference(
  base: string,
  decisions: Decisions,
  versionCalque: string,
): EtatInference {
  const differees = useDiffere(decisions, DELAI);
  const empreinte = useRef('');
  const [resultat, setResultat] = useState<ReponseInference | null>(null);
  const [decisionsJugees, setDecisionsJugees] = useState<Decisions>(differees);
  const [enCours, setEnCours] = useState(true);
  const [erreur, setErreur] = useState<string | null>(null);
  const [calqueModifie, setCalqueModifie] = useState(false);
  const [chargement, setChargement] = useState(0);

  useEffect(() => {
    const abandon = new AbortController();
    setEnCours(true);

    inferer(
      { base, decisions: differees, empreinte_physique: empreinte.current || undefined },
      abandon.signal,
    )
      .then((reponse) => {
        empreinte.current = reponse.empreinte_physique;
        setResultat(reponse);
        setDecisionsJugees(differees);
        setErreur(null);
        setCalqueModifie(false);
        setEnCours(false);
      })
      .catch((echec: unknown) => {
        if (abandon.signal.aborted) {
          return;
        }
        if (echec instanceof ErreurAPI && echec.code === CodeCalqueModifie) {
          setCalqueModifie(true);
        } else {
          setErreur(echec instanceof ErreurAPI ? echec.message : translate('error.unknown'));
        }
        setEnCours(false);
      });

    return () => abandon.abort();
  }, [base, differees, versionCalque, chargement]);

  const recharger = useCallback(() => {
    empreinte.current = '';
    setChargement((n) => n + 1);
  }, []);

  return { resultat, decisionsJugees, enCours, erreur, calqueModifie, recharger };
}
