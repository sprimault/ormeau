// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { useLangStore } from '@/shared/i18n';
import { Header } from '../Header';

/** Chemin long, pour vérifier la troncature ; remonté pour le `vi.mock` qui le sert. */
const { repertoire } = vi.hoisted(() => ({
  repertoire:
    'C:\\Users\\dev\\AppData\\Local\\Temp\\reprises\\gescom-2026\\extraction-initiale\\travail',
}));

vi.mock('@/shared/api', async (originel) => ({
  // ErreurAPI reste la vraie : WorkdirField la teste pour distinguer un refus
  // du serveur d'une panne quelconque.
  ...(await originel<typeof import('@/shared/api')>()),
  useContexte: () => ({
    contexte: { repertoire, version: 'test' },
    changerRepertoire: vi.fn(),
  }),
}));

// Le suivi des extractions a son propre fournisseur et ses propres tests : ce
// qui est sous test ici, c'est l'en-tête.
vi.mock('@/features/extraction', () => ({ ExtractionIndicator: () => null }));

describe('Header', () => {
  beforeEach(() => {
    useLangStore.setState({ lang: 'fr' });
  });

  it('tronque le répertoire à l’affichage', () => {
    render(<Header />);
    expect(screen.getByText(/…/)).toBeInTheDocument();
    expect(screen.queryByText(repertoire)).not.toBeInTheDocument();
  });

  it('copie le chemin complet, que le texte tronqué ne permet pas de retrouver', async () => {
    const utilisateur = userEvent.setup();
    render(<Header />);

    await utilisateur.click(screen.getByRole('button', { name: 'Copier le chemin complet' }));

    expect(await navigator.clipboard.readText()).toBe(repertoire);
  });
});
