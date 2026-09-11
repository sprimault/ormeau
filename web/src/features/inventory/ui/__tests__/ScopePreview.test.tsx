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
    window.localStorage.clear();
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

  it('montre dès l’ouverture la portée exacte qui partira', () => {
    const attendue = portee(['public.clients', 'public.commandes']);
    render(<ScopePreview portee={attendue} exclusions={exclusions()} />);

    expect(screen.getByText(/tables_incluses/).textContent).toBe(JSON.stringify(attendue, null, 2));
  });

  it('sépare ce qui part à l’extraction de ce qui va dans les décisions', () => {
    render(
      <ScopePreview
        portee={portee(['public.clients'])}
        exclusions={exclusions({ 'public.clients': ['blob_import', 'photo'] })}
      />,
    );

    // La colonne écartée ne doit pas apparaître dans la portée : le calque
    // physique garde tout, c'est l'entité qui la perd.
    expect(screen.getByText(/tables_incluses/).textContent).not.toContain('photo');
    expect(screen.getByText(/colonnes_ignorees/).textContent).toContain(
      'public.clients: [blob_import, photo]',
    );
  });

  it('annonce le nombre de colonnes écartées dans la barre', () => {
    render(
      <ScopePreview portee={portee()} exclusions={exclusions({ 'public.clients': ['photo'] })} />,
    );
    expect(screen.getByText('1 colonne(s) écartée(s)')).toBeInTheDocument();
  });

  it('place le calque produit à côté de ce qui part', () => {
    render(
      <ScopePreview
        portee={portee()}
        exclusions={exclusions()}
        produit={<p>calque produit</p>}
      />,
    );

    expect(screen.getByText('calque produit')).toBeInTheDocument();
    expect(screen.getByText(/tables_incluses/)).toBeInTheDocument();
  });

  it('se replie, et le retient au lancement suivant', async () => {
    const { unmount } = render(<ScopePreview portee={portee()} exclusions={exclusions()} />);

    await userEvent.click(screen.getByRole('button', { name: /Ce que produira cet écran/ }));
    expect(screen.queryByText(/tables_incluses/)).not.toBeInTheDocument();

    unmount();
    render(<ScopePreview portee={portee()} exclusions={exclusions()} />);
    expect(screen.queryByText(/tables_incluses/)).not.toBeInTheDocument();
  });

  it('place les actions juste après le compte, pas au bout de la barre', () => {
    render(
      <ScopePreview
        portee={portee()}
        exclusions={exclusions({ 'public.clients': ['photo'] })}
        actions={<button type="button">Extraire</button>}
      />,
    );

    const compte = screen.getByText('toutes les tables des schémas retenus');
    const bouton = screen.getByRole('button', { name: 'Extraire' });
    const ecartees = screen.getByText('1 colonne(s) écartée(s)');

    expect(compte.compareDocumentPosition(bouton) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    expect(bouton.compareDocumentPosition(ecartees) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
  });
});
