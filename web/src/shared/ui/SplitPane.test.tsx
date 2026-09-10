// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it } from 'vitest';

import { useLangStore } from '@/shared/i18n';
import { SplitPane } from './SplitPane';

const CLE = 'ormeau-test-largeur';

/** Monte les deux panneaux avec un contenu reconnaissable. */
function monter(defaut?: number) {
  return render(
    <SplitPane cleStockage={CLE} defaut={defaut} gauche={<p>arbre</p>} droite={<p>détail</p>} />,
  );
}

describe('SplitPane', () => {
  beforeEach(() => {
    useLangStore.setState({ lang: 'fr' });
    window.localStorage.clear();
  });

  it('affiche les deux panneaux', () => {
    monter();
    expect(screen.getByText('arbre')).toBeInTheDocument();
    expect(screen.getByText('détail')).toBeInTheDocument();
  });

  it('expose un séparateur atteignable au clavier', () => {
    monter();
    const poignee = screen.getByRole('separator');
    expect(poignee).toHaveAttribute('aria-orientation', 'vertical');
    expect(poignee).toHaveAttribute('tabindex', '0');
  });

  it('élargit et rétrécit aux flèches', async () => {
    monter(400);
    const poignee = screen.getByRole('separator');

    poignee.focus();
    await userEvent.keyboard('{ArrowRight}');
    expect(poignee).toHaveAttribute('aria-valuenow', '424');

    await userEvent.keyboard('{ArrowLeft}{ArrowLeft}');
    expect(poignee).toHaveAttribute('aria-valuenow', '376');
  });

  it('refuse de descendre sous la largeur lisible', async () => {
    monter(260);
    const poignee = screen.getByRole('separator');

    poignee.focus();
    await userEvent.keyboard('{ArrowLeft}{ArrowLeft}{ArrowLeft}');
    expect(poignee).toHaveAttribute('aria-valuenow', '240');
  });

  it('garde la largeur d’une session à l’autre', async () => {
    monter(400);
    screen.getByRole('separator').focus();
    await userEvent.keyboard('{ArrowRight}');

    expect(window.localStorage.getItem(CLE)).toBe('424');
  });

  it('relit la largeur enregistrée', () => {
    window.localStorage.setItem(CLE, '512');
    monter();
    expect(screen.getByRole('separator')).toHaveAttribute('aria-valuenow', '512');
  });

  it('ignore une largeur enregistrée hors bornes', () => {
    window.localStorage.setItem(CLE, '9000');
    monter(384);
    expect(screen.getByRole('separator')).toHaveAttribute('aria-valuenow', '384');
  });

  it('revient au défaut au double-clic', async () => {
    monter(400);
    const poignee = screen.getByRole('separator');

    poignee.focus();
    await userEvent.keyboard('{ArrowRight}{ArrowRight}');
    await userEvent.dblClick(poignee);

    expect(poignee).toHaveAttribute('aria-valuenow', '400');
  });
});
