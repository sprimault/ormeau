// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it } from 'vitest';

import { useLangStore } from '@/shared/i18n';
import { SplitPane } from '../SplitPane';

const CLE = 'ormeau-test-largeur';

/** Monte les deux panneaux côte à côte, avec un contenu reconnaissable. */
function monter(defaut?: number) {
  return render(
    <SplitPane cleStockage={CLE} defaut={defaut} premier={<p>arbre</p>} second={<p>détail</p>} />,
  );
}

/** Monte les deux panneaux empilés. */
function monterEmpiles(defaut?: number, replie?: boolean) {
  return render(
    <SplitPane
      sens="vertical"
      cleStockage={CLE}
      defaut={defaut}
      replie={replie}
      premier={<p>arbre</p>}
      second={<p>aperçu</p>}
    />,
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

  it('empilés, règle la hauteur du bas aux flèches haut et bas', async () => {
    monterEmpiles(320);
    const poignee = screen.getByRole('separator', { name: /Hauteur du panneau/ });
    expect(poignee).toHaveAttribute('aria-orientation', 'horizontal');

    poignee.focus();
    await userEvent.keyboard('{ArrowUp}');
    expect(poignee).toHaveAttribute('aria-valuenow', '344');

    await userEvent.keyboard('{ArrowDown}{ArrowDown}');
    expect(poignee).toHaveAttribute('aria-valuenow', '296');
  });

  it('empilés, descend plus bas que côte à côte, sans passer sous une barre', async () => {
    monterEmpiles(144);
    const poignee = screen.getByRole('separator');

    poignee.focus();
    await userEvent.keyboard('{ArrowDown}{ArrowDown}');
    expect(poignee).toHaveAttribute('aria-valuenow', '120');
  });

  it('replié, retire la poignée sans démonter les panneaux', () => {
    monterEmpiles(320, true);

    expect(screen.queryByRole('separator')).not.toBeInTheDocument();
    expect(screen.getByText('arbre')).toBeInTheDocument();
    expect(screen.getByText('aperçu')).toBeInTheDocument();
  });
});
