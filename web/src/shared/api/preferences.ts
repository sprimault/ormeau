// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import type { Preferences } from '@/shared/model';

import { postJSON } from './client';

/**
 * Enregistre les préférences d'affichage.
 *
 * Sans attendre la réponse, et sans la remonter : l'interface affiche déjà ce
 * qu'on vient de régler, et un panneau qui reviendrait en arrière parce que le
 * serveur a mis du temps serait pire que l'échec lui-même. Un refus est donc
 * silencieux — il ne peut venir que d'une valeur que le front ne produit pas,
 * ou d'un disque qui n'accepte plus rien, et ni l'un ni l'autre n'appelle à
 * interrompre ce qu'on est en train de faire.
 */
export function enregistrerPreferences(preferences: Preferences): void {
  void postJSON<Preferences, Preferences>('/api/preferences', preferences).catch(() => {});
}
