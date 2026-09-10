// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useId } from 'react';
import type { InputHTMLAttributes, ReactNode } from 'react';

/** Propriétés du champ, en plus de celles d'un champ HTML. */
interface ProprietesChamp extends InputHTMLAttributes<HTMLInputElement> {
  label: string;
  aide?: ReactNode;
}

/**
 * Champ de saisie étiqueté.
 *
 * L'identifiant est tiré par React plutôt que passé en propriété : c'est ce qui
 * lie l'étiquette au champ pour un lecteur d'écran, et l'oublier ne se voit pas
 * à l'affichage.
 */
export function Field({ label, aide, className = '', ...reste }: ProprietesChamp) {
  const id = useId();

  return (
    <div className="flex flex-col gap-1">
      <label htmlFor={id} className="text-xs font-medium text-slate-600 dark:text-slate-400">
        {label}
      </label>
      <input
        id={id}
        className={`focus:border-ormeau-500 rounded border border-slate-300 bg-white px-2 py-1.5 text-sm text-slate-900 outline-none dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100 ${className}`}
        {...reste}
      />
      {aide ? <p className="text-xs text-slate-500 dark:text-slate-500">{aide}</p> : null}
    </div>
  );
}
