// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import type { BrouillonDecisions } from '../../model/useBrouillonDecisions';
import { ContexteDecisions } from '../../model/contexte';
import { DecisionsError } from '../DecisionsError';

/** Brouillon figé, avec ou sans erreur de lecture. */
function brouillon(erreur: string | null): BrouillonDecisions {
  return {
    base: 'gescom',
    decisions: {},
    fichier: null,
    pret: true,
    enCours: false,
    erreur,
    modifie: false,
    modifier: vi.fn(),
    relire: vi.fn(),
    enregistre: vi.fn(),
  };
}

describe('DecisionsError', () => {
  it('annonce un fichier de décisions illisible', () => {
    render(
      <ContexteDecisions.Provider value={brouillon('gescom.decisions.yaml illisible : ligne 3')}>
        <DecisionsError />
      </ContexteDecisions.Provider>,
    );
    expect(screen.getByRole('alert')).toHaveTextContent('gescom.decisions.yaml illisible : ligne 3');
  });

  it('ne dit rien quand le fichier a été lu', () => {
    const { container } = render(
      <ContexteDecisions.Provider value={brouillon(null)}>
        <DecisionsError />
      </ContexteDecisions.Provider>,
    );
    expect(container).toBeEmptyDOMElement();
  });
});
