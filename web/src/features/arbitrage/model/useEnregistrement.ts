// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useState } from 'react';

import type { BrouillonDecisions } from '@/entities/decisions';
import { ErreurAPI } from '@/shared/api';
import { translate } from '@/shared/i18n';
import { CodeContenuManuel, type CodeRefus } from '@/shared/model';
import { ecrireDecisions } from '../api/arbitrageApi';

/** Ce que l'écriture du fichier demande à l'écran. */
export interface EtatEnregistrement {
  enCours: boolean;
  /** Refus en attente d'une réponse : relire, recharger ou confirmer. */
  refus: CodeRefus | null;
  erreur: string | null;
  /** Écrit le brouillon ; `ecraserManuel` confirme la perte d'un travail manuel. */
  enregistrer: (ecraserManuel?: boolean) => void;
  /** Referme le refus ou l'erreur en cours. */
  abandonner: () => void;
}

/**
 * Écrit le brouillon dans le fichier de décisions de la base.
 *
 * L'écriture porte les deux empreintes — celle du calque jugé, celle du fichier
 * lu —, et le serveur refuse si l'une ne correspond plus. Un fichier dont la
 * lecture a signalé un travail manuel demande confirmation avant tout envoi :
 * le serveur le refuserait de toute façon, et le dire d'abord évite un aller
 * retour qui ressemble à une panne.
 */
export function useEnregistrement(
  brouillon: Pick<BrouillonDecisions, 'base' | 'decisions' | 'fichier' | 'enregistre'>,
  empreinte: string,
): EtatEnregistrement {
  const { base, decisions, fichier, enregistre } = brouillon;
  const [enCours, setEnCours] = useState(false);
  const [refus, setRefus] = useState<CodeRefus | null>(null);
  const [erreur, setErreur] = useState<string | null>(null);

  const enregistrer = useCallback(
    (ecraserManuel = false) => {
      if (fichier?.manuel && !ecraserManuel) {
        setRefus(CodeContenuManuel);
        return;
      }
      setEnCours(true);
      setRefus(null);
      setErreur(null);

      ecrireDecisions({
        base,
        decisions,
        empreinte_physique: empreinte,
        empreinte_fichier: fichier?.empreinte_fichier,
        ecraser_manuel: ecraserManuel || undefined,
      })
        .then((reponse) => enregistre(decisions, reponse.empreinte_fichier))
        .catch((echec: unknown) => {
          if (echec instanceof ErreurAPI && echec.code) {
            setRefus(echec.code);
          } else {
            setErreur(echec instanceof ErreurAPI ? echec.message : translate('error.unknown'));
          }
        })
        .finally(() => setEnCours(false));
    },
    [base, decisions, fichier, enregistre, empreinte],
  );

  const abandonner = useCallback(() => {
    setRefus(null);
    setErreur(null);
  }, []);

  return { enCours, refus, erreur, enregistrer, abandonner };
}
