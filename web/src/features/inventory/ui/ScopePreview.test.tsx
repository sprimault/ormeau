// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { useLangStore } from '@/shared/i18n';
import type { EtatExclusions } from '../model/useExclusions';
import { ScopePreview } from './ScopePreview';

/** Rend un état d'exclusions figé : ce composant les affiche, il ne les tient pas. */
function exclusions(ignorees: Record<string, string[]> = {}): EtatExclusions {
  return {
    ignorees,
    total: Object.values(ignorees).reduce((somme, colonnes) => somme + colonnes.length, 0),
    estIgnoree: () => false,
    compte: () => 0,
    basculer: vi.fn(),
  };
}

describe('ScopePreview', () => {
  beforeEach(() => {
    useLangStore.setState({ lang: 'fr' });
  });

  it('dit qu’une sélection vide couvre tout, ce que le JSON seul ne montre pas', () => {
    render(<ScopePreview schemas={['public']} selection={new Set()} exclusions={exclusions()} />);
    expect(screen.getByText('toutes les tables des schémas retenus')).toBeInTheDocument();
  });

  it('compte les tables retenues', () => {
    const selection = new Set(['public.clients', 'public.commandes']);
    render(<ScopePreview schemas={['public']} selection={selection} exclusions={exclusions()} />);
    expect(screen.getByText('2 table(s) sur 1 schéma(s)')).toBeInTheDocument();
  });

  it('montre la portée exacte qui partira, une fois dépliée', async () => {
    const selection = new Set(['public.commandes', 'public.clients']);
    render(<ScopePreview schemas={['public']} selection={selection} exclusions={exclusions()} />);

    await userEvent.click(screen.getByRole('button', { name: /Ce que produira cet écran/ }));

    // Triée : deux sélections identiques doivent donner le même texte, sinon
    // l'aperçu change sans que la portée ait bougé.
    const json = screen.getByText(/tables_incluses/);
    expect(json.textContent).toContain('"public.clients",\n    "public.commandes"');
  });

  it('sépare ce qui part à l’extraction de ce qui va dans les décisions', async () => {
    render(
      <ScopePreview
        schemas={['public']}
        selection={new Set(['public.clients'])}
        exclusions={exclusions({ 'public.clients': ['blob_import', 'photo'] })}
      />,
    );

    await userEvent.click(screen.getByRole('button', { name: /Ce que produira cet écran/ }));

    // La colonne écartée ne doit pas apparaître dans la portée : le calque
    // physique garde tout, c'est l'entité qui la perd.
    expect(screen.getByText(/tables_incluses/).textContent).not.toContain('photo');
    expect(screen.getByText(/colonnes_ignorees/).textContent).toContain(
      'public.clients: [blob_import, photo]',
    );
  });

  it('annonce le nombre de colonnes écartées sans rien déplier', () => {
    render(
      <ScopePreview
        schemas={['public']}
        selection={new Set()}
        exclusions={exclusions({ 'public.clients': ['photo'] })}
      />,
    );
    expect(screen.getByText('1 colonne(s) écartée(s)')).toBeInTheDocument();
  });

  it('reste replié par défaut', () => {
    render(<ScopePreview schemas={['public']} selection={new Set()} exclusions={exclusions()} />);
    expect(screen.queryByText(/tables_incluses/)).not.toBeInTheDocument();
  });
});
