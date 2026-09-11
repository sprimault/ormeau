// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { useLangStore } from '@/shared/i18n';
import type { Portee } from '@/shared/model';
import type { EtatExclusions } from '../../model/useExclusions';
import { ScopePreview } from '../ScopePreview';

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

/** Portée telle qu'usePortee la compose. */
function portee(tables: string[] = [], schemas = ['public']): Portee {
  return { schemas, tables_incluses: tables };
}

describe('ScopePreview', () => {
  beforeEach(() => {
    useLangStore.setState({ lang: 'fr' });
  });

  it('dit qu’une sélection vide couvre tout, ce que le JSON seul ne montre pas', () => {
    render(<ScopePreview portee={portee()} exclusions={exclusions()} />);
    expect(screen.getByText('toutes les tables des schémas retenus')).toBeInTheDocument();
  });

  it('compte les tables et les schémas qui partiront', () => {
    render(
      <ScopePreview
        portee={portee(['public.clients', 'ventes.factures'], ['public', 'ventes'])}
        exclusions={exclusions()}
      />,
    );
    expect(screen.getByText('2 table(s) sur 2 schéma(s)')).toBeInTheDocument();
  });

  it('montre la portée exacte qui partira, une fois dépliée', async () => {
    render(
      <ScopePreview
        portee={portee(['public.clients', 'public.commandes'])}
        exclusions={exclusions()}
      />,
    );

    await userEvent.click(screen.getByRole('button', { name: /Ce que produira cet écran/ }));

    const json = screen.getByText(/tables_incluses/);
    expect(json.textContent).toBe(JSON.stringify(portee(['public.clients', 'public.commandes']), null, 2));
  });

  it('sépare ce qui part à l’extraction de ce qui va dans les décisions', async () => {
    render(
      <ScopePreview
        portee={portee(['public.clients'])}
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
      <ScopePreview portee={portee()} exclusions={exclusions({ 'public.clients': ['photo'] })} />,
    );
    expect(screen.getByText('1 colonne(s) écartée(s)')).toBeInTheDocument();
  });

  it('reste replié par défaut', () => {
    render(<ScopePreview portee={portee()} exclusions={exclusions()} />);
    expect(screen.queryByText(/tables_incluses/)).not.toBeInTheDocument();
  });

  it('place les actions fournies dans la barre, sans rien déplier', () => {
    render(
      <ScopePreview
        portee={portee()}
        exclusions={exclusions()}
        actions={<button type="button">Extraire</button>}
      />,
    );
    expect(screen.getByRole('button', { name: 'Extraire' })).toBeInTheDocument();
  });
});
