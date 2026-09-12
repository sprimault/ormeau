// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useEffect, useState } from 'react';

import { ErreurAPI } from '@/shared/api';
import { CodeProfilExistant, type Profil, type ProfilResume } from '@/shared/model';

import { enregistrerProfil, lireProfils, supprimerProfil } from '../api/profilsApi';

/** Ce que l'écran de connexion tient des profils. */
interface EtatProfils {
  profils: ProfilResume[];
  /** Ce que le fichier portait d'inutilisable, dit une fois au chargement. */
  avertissement: string;
  /** Rend vrai quand c'est enregistré, faux quand le nom est déjà pris. */
  enregistrer: (
    profil: Profil,
    motDePasse: string,
    avecMotDePasse: boolean,
    remplacer?: boolean,
    dsn?: string,
  ) => Promise<boolean>;
  retirer: (nom: string) => Promise<void>;
  erreur: string | null;
}

/**
 * Charge les connexions enregistrées et les tient à jour.
 *
 * Le serveur rend la liste complète après chaque écriture : pas de mise à jour
 * locale à tenir d'accord avec le fichier, et l'ordre — trié par nom — vient de
 * lui.
 */
export function useProfils(): EtatProfils {
  const [profils, setProfils] = useState<ProfilResume[]>([]);
  const [avertissement, setAvertissement] = useState('');
  const [erreur, setErreur] = useState<string | null>(null);

  useEffect(() => {
    const abandon = new AbortController();
    lireProfils(abandon.signal)
      .then((reponse) => {
        setProfils(reponse.profils);
        setAvertissement(reponse.avertissement ?? '');
      })
      .catch(() => {
        // Sans configuration accessible, l'écran s'ouvre sans profil et on
        // saisit comme avant : ce n'est pas une raison de bloquer la connexion.
      });
    return () => abandon.abort();
  }, []);

  const enregistrer = useCallback(
    async (
      profil: Profil,
      motDePasse: string,
      avecMotDePasse: boolean,
      remplacer = false,
      dsn?: string,
    ) => {
      setErreur(null);
      try {
        const reponse = await enregistrerProfil(
          profil,
          motDePasse,
          avecMotDePasse,
          remplacer,
          dsn,
        );
        setProfils(reponse.profils);
        setAvertissement(reponse.avertissement ?? '');
        return true;
      } catch (echec) {
        if (echec instanceof ErreurAPI && echec.code === CodeProfilExistant) {
          return false;
        }
        setErreur(echec instanceof ErreurAPI ? echec.message : String(echec));
        return false;
      }
    },
    [],
  );

  const retirer = useCallback(async (nom: string) => {
    setErreur(null);
    try {
      await supprimerProfil(nom);
      const reponse = await lireProfils();
      setProfils(reponse.profils);
    } catch (echec) {
      setErreur(echec instanceof ErreurAPI ? echec.message : String(echec));
    }
  }, []);

  return { profils, avertissement, enregistrer, retirer, erreur };
}
