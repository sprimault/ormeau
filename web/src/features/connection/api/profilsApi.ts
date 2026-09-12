// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { getJSON, postJSON, supprimer } from '@/shared/api';
import type { Profil, ReferenceProfil, ReponseProfils, RequeteProfil } from '@/shared/model';

/** Lit les connexions enregistrées. */
export function lireProfils(signal?: AbortSignal): Promise<ReponseProfils> {
  return getJSON<ReponseProfils>('/api/profils', signal);
}

/**
 * Enregistre une connexion.
 *
 * `remplacer` confirme l'écrasement d'un profil du même nom : sans lui, le
 * serveur refuse avec `profil_existant` et l'écran demande. Enregistrer
 * par-dessus est le geste qu'on fait sans y penser après avoir modifié un
 * champ.
 */
export function enregistrerProfil(
  profil: Profil,
  motDePasse: string,
  enregistrerMotDePasse: boolean,
  remplacer = false,
  dsn?: string,
): Promise<ReponseProfils> {
  return postJSON<RequeteProfil, ReponseProfils>('/api/profils', {
    profil,
    // Décomposée par le serveur : le front n'analyse jamais une chaîne de
    // connexion, et sans elle un profil enregistré depuis ce mode de saisie ne
    // garderait ni hôte ni base.
    dsn,
    mot_de_passe: enregistrerMotDePasse ? motDePasse : undefined,
    enregistrer_mot_de_passe: enregistrerMotDePasse,
    remplacer,
  });
}

/** Retire une connexion enregistrée. */
export function supprimerProfil(nom: string): Promise<void> {
  return supprimer<ReferenceProfil>('/api/profils', { nom });
}
