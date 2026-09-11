// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { render, screen } from '@testing-library/react';
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
    expect(screen.getByRole('tooltip', { hidden: true })).toHaveTextContent('Nom de la classe PHP générée.');
  });
});
