// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useEffect, useState } from 'react';

import { ErreurAPI } from '@/shared/api';
import { translate } from '@/shared/i18n';
import type { RequeteConnexion, ReponseConnexion } from '@/shared/model';
import { changerDeBase, connecter, deconnecter, lireBases } from '../api/connectionApi';

/** État de l'écran de connexion. */
export interface EtatConnexion {
  serveur: ReponseConnexion | null;
  /** Bases du serveur atteint, vide tant que rien n'est connecté ou quand le
   *  dialecte ne sait pas les énumérer. */
  bases: string[];
  enCours: boolean;
  erreur: string | null;
  ouvrir: (requete: RequeteConnexion) => Promise<void>;
  fermer: () => Promise<void>;
  changerBase: (base: string) => Promise<void>;
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
  const [bases, setBases] = useState<string[]>([]);
  const [enCours, setEnCours] = useState(false);
  const [erreur, setErreur] = useState<string | null>(null);

  // Les bases arrivent après la connexion, et non à la demande : l'arbre les
  // affiche toutes dès l'ouverture, comme le ferait un client de base.
  const session = serveur?.session;
  useEffect(() => {
    if (!session) {
      setBases([]);
      return;
    }
    const abandon = new AbortController();
    lireBases(session, abandon.signal)
      .then((reponse) => setBases(reponse.bases))
      .catch(() => {
        // Le dialecte ne sait pas énumérer, ou la connexion vient de tomber :
        // l'arbre montre alors la seule base ouverte.
        setBases([]);
      });
    return () => abandon.abort();
  }, [session]);

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

  const changerBase = useCallback(
    async (base: string) => {
      if (!serveur) {
        return;
      }
      setEnCours(true);
      setErreur(null);
      try {
        setServeur(await changerDeBase(serveur.session, base));
      } catch (echec) {
        // La connexion précédente reste ouverte côté serveur : l'écran garde la
        // base où l'on était plutôt que de renvoyer au formulaire.
        setErreur(echec instanceof ErreurAPI ? echec.message : translate('error.unknown'));
      } finally {
        setEnCours(false);
      }
    },
    [serveur],
  );

  return { serveur, bases, enCours, erreur, ouvrir, fermer, changerBase };
}
