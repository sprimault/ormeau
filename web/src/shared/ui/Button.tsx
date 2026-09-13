// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import type { ButtonHTMLAttributes, ReactNode } from 'react';

/** Apparence du bouton. « discret » sert aux actions secondaires d'un en-tête. */
type Variante = 'primaire' | 'discret';

/** Taille du bouton. « petite » sert aux actions posées sur une ligne de liste. */
type Taille = 'normale' | 'petite';

/** Propriétés du bouton, en plus de celles d'un bouton HTML. */
interface ProprietesBouton extends ButtonHTMLAttributes<HTMLButtonElement> {
  variante?: Variante;
  taille?: Taille;
  children: ReactNode;
}

/** Classes Tailwind de chaque variante, état désactivé et thème sombre compris. */
const classes: Record<Variante, string> = {
  primaire:
    'bg-ormeau-600 text-white hover:bg-ormeau-700 disabled:bg-slate-400 dark:disabled:bg-slate-700',
  discret:
    'border border-slate-300 text-slate-700 hover:bg-slate-100 disabled:opacity-50 dark:border-slate-700 dark:text-slate-200 dark:hover:bg-slate-800',
};

/** Classes Tailwind de chaque taille. */
const tailles: Record<Taille, string> = {
  normale: 'px-3 py-1.5 text-sm',
  petite: 'px-2 py-0.5 text-xs',
};

/** Bouton de l'interface. */
export function Button({
  variante = 'primaire',
  taille = 'normale',
  className = '',
  ...reste
}: ProprietesBouton) {
  return (
    <button
      className={`shrink-0 rounded font-medium transition-colors disabled:cursor-not-allowed ${tailles[taille]} ${classes[variante]} ${className}`}
      {...reste}
    />
  );
}
