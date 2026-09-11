// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import type { InputHTMLAttributes } from 'react';

/**
 * Champ de saisie compact, pour les listes où chaque ligne se modifie sur
 * place. Le formulaire de connexion garde `Field`, qui porte son libellé.
 */
export function TextInput({ className = '', ...reste }: InputHTMLAttributes<HTMLInputElement>) {
  return (
    <input
      type="text"
      className={`rounded border border-slate-300 bg-transparent px-1.5 py-0.5 text-xs placeholder:text-slate-400 disabled:opacity-60 dark:border-slate-700 ${className}`}
      {...reste}
    />
  );
}
