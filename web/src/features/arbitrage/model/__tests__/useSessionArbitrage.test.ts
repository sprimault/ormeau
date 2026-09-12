// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { renderHook, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import type { Decisions, ReponseSession } from '@/shared/model';
import { useSessionArbitrage } from '../useSessionArbitrage';

const etat = vi.hoisted(() => ({
  /** Ce que le serveur rend à la lecture du brouillon. */
  lue: {} as ReponseSession,
  /** Ce que le hook a fait enregistrer. */
  ecrites: [] as unknown[],
  /** Ce que le hook a appliqué au brouillon de décisions. */
  appliquees: null as Decisions | null,
  /** Le brouillon décide-t-il autre chose que le fichier. */
  modifie: false,
}));

vi.mock('@/entities/decisions', () => ({
  lireSession: () => Promise.resolve(etat.lue),
  ecrireSession: (session: unknown) => {
    etat.ecrites.push(session);
    return Promise.resolve({ session });
  },
  effacerSession: () => Promise.resolve(),
}));

/** Le brouillon tel que l'écran le passe au hook. */
function brouillonDouble() {
  return {
    base: 'gescom',
    decisions: {} as Decisions,
    fichier: { existe: true, decisions: {}, empreinte_fichier: 'sha256:bb', manuel: false },
    pret: true,
    enCours: false,
    erreur: null,
    modifie: etat.modifie,
    modifier: (transformation: (d: Decisions) => Decisions) => {
      etat.appliquees = transformation({});
    },
    relire: () => {},
    enregistre: () => {},
  };
}

/** Monte le hook avec l'écran minimal dont il a besoin. */
function monter(entiteOuverte: string | null, onRestaurerEntite = vi.fn()) {
  return renderHook(() =>
    useSessionArbitrage({
      base: 'gescom',
      brouillon: brouillonDouble(),
      empreinteCalque: 'sha256:aa',
      entiteOuverte,
      onRestaurerEntite,
    }),
  );
}

describe('useSessionArbitrage', () => {
  beforeEach(() => {
    etat.lue = {};
    etat.ecrites = [];
    etat.appliquees = null;
    etat.modifie = false;
    vi.useFakeTimers({ shouldAdvanceTime: true });
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('n’enregistre rien tant que le brouillon ne dit rien de plus', async () => {
    monter(null);
    await vi.advanceTimersByTimeAsync(3000);

    // Un fichier de session vide n'aide personne à comprendre ce qu'il
    // contient, et repartir du fichier de décisions donne le même écran.
    expect(etat.ecrites).toHaveLength(0);
  });

  it('enregistre le brouillon après un délai', async () => {
    // Minuteries réelles : la restauration passe par une promesse, et faire
    // avancer un temps factice au travers demande plus de précautions que le
    // test n'en vaut. Une seconde d'attente, une fois.
    vi.useRealTimers();
    etat.modifie = true;
    monter('public.client');

    await waitFor(() => expect(etat.ecrites).toHaveLength(1), { timeout: 3000 });
    expect(etat.ecrites[0]).toMatchObject({
      base: 'gescom',
      empreinte_calque: 'sha256:aa',
      empreinte_decisions: 'sha256:bb',
      entite_ouverte: 'public.client',
    });
  });

  it('signale un brouillon écarté par le serveur', async () => {
    etat.lue = { ecartee: 'le calque a été réextrait depuis ce brouillon' };
    const { result } = monter('public.client');

    await waitFor(() => expect(result.current.ecartee).not.toBeNull());
    expect(result.current.ecartee).toContain('réextrait');
  });

  it('restaure le brouillon qui s’applique encore', async () => {
    const restaurer = vi.fn();
    etat.lue = {
      session: {
        base: 'gescom',
        empreinte_calque: 'sha256:aa',
        empreinte_decisions: 'sha256:bb',
        brouillon: JSON.stringify({ renommages: { 'public.client': 'Client' } }),
        entite_ouverte: 'public.facture',
      },
    };
    monter(null, restaurer);

    await waitFor(() => expect(etat.appliquees).not.toBeNull());
    expect(etat.appliquees).toEqual({ renommages: { 'public.client': 'Client' } });
    expect(restaurer).toHaveBeenCalledWith('public.facture');
  });
});
