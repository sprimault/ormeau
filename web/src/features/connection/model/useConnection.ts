// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useState } from 'react';

import { ErreurAPI } from '@/shared/api';
import { translate } from '@/shared/i18n';
import type { RequeteConnexion, ReponseConnexion } from '@/shared/model';
import { connecter, deconnecter } from '../api/connectionApi';

/** État de l'écran de connexion. */
export interface EtatConnexion {
  serveur: ReponseConnexion | null;
  enCours: boolean;
  erreur: string | null;
  ouvrir: (requete: RequeteConnexion) => Promise<void>;
  fermer: () => Promise<void>;
}

/**
 * Pilote l'écran de connexion.
 *
 * Rien de ce qui a été saisi n'est conservé ici : ni DSN, ni mot de passe. La
 * requête part et est oubliée, seule la réponse du serveur — qui n'en porte
 * rien — reste en mémoire. C'est aussi pourquoi il n'y a pas de store global :
 * un état persistant est précisément ce qu'on refuse à ces valeurs.
 */
export function useConnection(): EtatConnexion {
  const [serveur, setServeur] = useState<ReponseConnexion | null>(null);
  const [enCours, setEnCours] = useState(false);
  const [erreur, setErreur] = useState<string | null>(null);

  const ouvrir = useCallback(async (requete: RequeteConnexion) => {
    setEnCours(true);
    setErreur(null);
    try {
      setServeur(await connecter(requete));
    } catch (echec) {
      setErreur(echec instanceof ErreurAPI ? echec.message : translate('error.unknown'));
    } finally {
      setEnCours(false);
    }
  }, []);

  const fermer = useCallback(async () => {
    if (!serveur) {
      return;
    }
    try {
      await deconnecter(serveur.session);
    } catch {
      // La connexion est perdue de toute façon : l'écran revient au formulaire,
      // et le binaire referme ce qui reste à son arrêt.
    }
    setServeur(null);
  }, [serveur]);

  return { serveur, enCours, erreur, ouvrir, fermer };
}
