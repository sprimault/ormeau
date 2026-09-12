// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useEffect, useState } from 'react';

import { ecrireSession, effacerSession, lireSession } from '@/entities/decisions';
import type { BrouillonDecisions } from '@/entities/decisions';
import type { Decisions } from '@/shared/model';

/**
 * Délai avant d'enregistrer le brouillon, en millisecondes.
 *
 * Plus long que celui des préférences : un brouillon se règle par gestes
 * espacés — nommer une classe, forcer un type —, et l'enregistrer à chaque
 * frappe écrirait le fichier pendant qu'on tape un nom.
 */
const DELAI_ENREGISTREMENT = 1000;

/** Ce que l'écran d'arbitrage tient du brouillon persisté. */
interface EtatSession {
  /** Raison pour laquelle un brouillon existant a été écarté, s'il y en a une. */
  ecartee: string | null;
  /** Efface le brouillon, après un enregistrement réussi du fichier. */
  oublier: () => void;
}

/** Ce que le hook a besoin de savoir de l'écran. */
interface OptionsSession {
  base: string;
  /**
   * Le brouillon, reçu et non lu depuis le contexte.
   *
   * L'écran l'a déjà sous la main, et un hook qui le relirait en dépendrait
   * pour rien — il serait aussi plus difficile à éprouver seul.
   */
  brouillon: BrouillonDecisions;
  /** Empreinte du calque jugé ; vide tant que l'inférence n'a pas répondu. */
  empreinteCalque: string;
  entiteOuverte: string | null;
  onRestaurerEntite: (qualifiee: string) => void;
}

/**
 * Fait survivre le travail non enregistré à un rechargement.
 *
 * Ce qui est retenu est ce qui n'est pas encore dans le fichier de décisions :
 * celui-ci fait foi. À la reprise, l'écran part du fichier — c'est le
 * fournisseur qui s'en charge —, et ce brouillon vient par-dessus, mais
 * seulement si le serveur confirme qu'il s'applique encore.
 *
 * Le serveur seul juge de cette validité : il a le calque et le fichier sous la
 * main. La page ne fait que dire ce qu'elle a et afficher ce qu'on lui répond.
 */
export function useSessionArbitrage({
  base,
  brouillon,
  empreinteCalque,
  entiteOuverte,
  onRestaurerEntite,
}: OptionsSession): EtatSession {
  const { decisions, fichier, modifie, modifier } = brouillon;
  const [ecartee, setEcartee] = useState<string | null>(null);

  // Tant que la restauration n'a pas eu lieu, rien n'est enregistré : sans
  // cela, le brouillon vide du premier rendu écraserait celui qu'on s'apprête à
  // relire.
  //
  // Un état et non une référence : c'est son passage à vrai qui doit relancer
  // l'effet d'enregistrement, et une référence mutée ne réveille rien.
  const [restauree, setRestauree] = useState(false);

  useEffect(() => {
    setRestauree(false);
    setEcartee(null);
  }, [base]);

  useEffect(() => {
    if (restauree || !fichier) {
      return;
    }
    const abandon = new AbortController();

    lireSession(base, abandon.signal)
      .then((reponse) => {
        if (reponse.ecartee) {
          setEcartee(reponse.ecartee);
        }
        if (reponse.session?.brouillon) {
          modifier(() => JSON.parse(reponse.session?.brouillon ?? '{}') as Decisions);
          if (reponse.session.entite_ouverte) {
            onRestaurerEntite(reponse.session.entite_ouverte);
          }
        }
      })
      .catch(() => {
        // Sans configuration accessible, l'écran fonctionne comme avant : le
        // brouillon ne survivra pas au rechargement, ce qui ne vaut pas
        // d'interrompre l'arbitrage.
      })
      .finally(() => {
        setRestauree(true);
      });

    return () => abandon.abort();
  }, [base, fichier, modifier, onRestaurerEntite, restauree]);

  useEffect(() => {
    if (!restauree || !empreinteCalque || !fichier) {
      return;
    }
    // Rien à retenir quand le brouillon ne dit rien de plus que le fichier :
    // le rouvrir repartira de lui, et un fichier de session vide n'aide
    // personne à comprendre ce qu'il contient.
    if (!modifie && !entiteOuverte) {
      return;
    }

    const minuterie = setTimeout(() => {
      void ecrireSession({
        base,
        empreinte_calque: empreinteCalque,
        empreinte_decisions: fichier.empreinte_fichier ?? '',
        brouillon: JSON.stringify(decisions),
        entite_ouverte: entiteOuverte ?? undefined,
      }).catch(() => {
        // Même raison qu'à la lecture : perdre un brouillon est regrettable,
        // interrompre l'arbitrage pour le dire le serait davantage.
      });
    }, DELAI_ENREGISTREMENT);

    return () => clearTimeout(minuterie);
  }, [base, decisions, empreinteCalque, entiteOuverte, fichier, modifie, restauree]);

  return {
    ecartee,
    oublier: () => {
      void effacerSession(base).catch(() => {});
    },
  };
}
