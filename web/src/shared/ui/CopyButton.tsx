// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useState } from 'react';
import { Check, Copy } from 'lucide-react';

import { useT } from '@/shared/i18n';

/** Propriétés du bouton. */
interface ProprietesCopie {
  valeur: string;
  libelle?: string;
}

/**
 * Bouton de copie.
 *
 * C'est le geste le plus répété pendant un arbitrage : ce qui s'affiche ici se
 * relit ailleurs — dans un terminal, dans une issue, dans un fichier de
 * décisions.
 */
export function CopyButton({ valeur, libelle }: ProprietesCopie) {
  const t = useT();
  const [copie, setCopie] = useState(false);

  async function copier() {
    try {
      await navigator.clipboard.writeText(valeur);
      setCopie(true);
      window.setTimeout(() => setCopie(false), 1500);
    } catch {
      // Presse-papiers refusé par le navigateur : le texte reste sélectionnable
      // à la main, et un message d'erreur n'apporterait rien de plus.
    }
  }

  return (
    <button
      type="button"
      onClick={() => void copier()}
      title={libelle ?? t('action.copy')}
      aria-label={libelle ?? t('action.copy')}
      className="rounded p-1 text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-slate-200"
    >
      {copie ? <Check size={14} aria-hidden /> : <Copy size={14} aria-hidden />}
    </button>
  );
}
