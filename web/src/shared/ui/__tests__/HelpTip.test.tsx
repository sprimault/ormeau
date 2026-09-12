// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it } from 'vitest';

import { useLangStore } from '@/shared/i18n';
import { HelpTip } from '../HelpTip';

describe('HelpTip', () => {
  beforeEach(() => {
    useLangStore.setState({ lang: 'fr' });
  });

  it('décrit le bouton d’aide par son explication, lisible au clavier', () => {
    render(<HelpTip texte="Nom de la classe PHP générée." />);

    const bouton = screen.getByRole('button', { name: 'Aide' });
    expect(bouton).toHaveAccessibleDescription('Nom de la classe PHP générée.');
    expect(screen.getByRole('tooltip', { hidden: true })).toHaveTextContent(
      'Nom de la classe PHP générée.',
    );
  });

  it('sort la bulle du flux pour qu’un panneau défilant ne la découpe pas', async () => {
    render(<HelpTip texte="Une explication longue." />);

    // En absolute, un parent avec overflow la coupait : dans l'arbre des
    // tables, le texte le plus utile de l'écran ne se lisait pas.
    await userEvent.tab();

    expect(screen.getByRole('tooltip')).toHaveStyle({ position: 'fixed' });
  });
});
