// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import type { ButtonHTMLAttributes, ReactNode } from 'react';

/** Apparence du bouton. « discret » sert aux actions secondaires d'un en-tête. */
type Variante = 'primaire' | 'discret';

/** Propriétés du bouton, en plus de celles d'un bouton HTML. */
interface ProprietesBouton extends ButtonHTMLAttributes<HTMLButtonElement> {
  variante?: Variante;
  children: ReactNode;
}

const classes: Record<Variante, string> = {
  primaire:
    'bg-ormeau-600 text-white hover:bg-ormeau-700 disabled:bg-slate-400 dark:disabled:bg-slate-700',
  discret:
    'border border-slate-300 text-slate-700 hover:bg-slate-100 dark:border-slate-700 dark:text-slate-200 dark:hover:bg-slate-800',
};

/** Bouton de l'interface. */
export function Button({ variante = 'primaire', className = '', ...reste }: ProprietesBouton) {
  return (
    <button
      className={`rounded px-3 py-1.5 text-sm font-medium transition-colors disabled:cursor-not-allowed ${classes[variante]} ${className}`}
      {...reste}
    />
  );
}
