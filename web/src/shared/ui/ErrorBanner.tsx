// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

/** Propriétés du bandeau. */
interface ProprietesBandeau {
  message: string;
}

/**
 * Bandeau d'échec.
 *
 * `role="alert"` parce qu'un échec de connexion arrive après une action et doit
 * être annoncé, pas seulement affiché.
 */
export function ErrorBanner({ message }: ProprietesBandeau) {
  return (
    <p
      role="alert"
      className="rounded border border-red-300 bg-red-50 px-3 py-2 text-sm text-red-800 dark:border-red-900 dark:bg-red-950 dark:text-red-200"
    >
      {message}
    </p>
  );
}
