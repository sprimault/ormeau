// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { describe, expect, it } from 'vitest';

import { messages } from './messages';
import { translate } from './useT';
import { useLangStore } from './store';

describe('dictionnaire', () => {
  it('couvre les mêmes clés dans les deux langues', () => {
    expect(Object.keys(messages.en).sort()).toEqual(Object.keys(messages.fr).sort());
  });

  it('ne laisse aucun libellé vide', () => {
    for (const [langue, dictionnaire] of Object.entries(messages)) {
      for (const [cle, libelle] of Object.entries(dictionnaire)) {
        expect(libelle, `${langue}/${cle}`).not.toBe('');
      }
    }
  });
});

describe('translate', () => {
  it('substitue les paramètres', () => {
    useLangStore.setState({ lang: 'fr' });
    expect(translate('connection.connected', { catalogue: 'gescom' })).toBe('Connecté à gescom');
  });

  it('suit la langue courante', () => {
    useLangStore.setState({ lang: 'en' });
    expect(translate('connection.submit')).toBe('Connect');
    useLangStore.setState({ lang: 'fr' });
    expect(translate('connection.submit')).toBe('Se connecter');
  });
});
