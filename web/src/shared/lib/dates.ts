// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

/**
 * Rend l'heure locale d'un instant RFC 3339 : « 11:42:17 ».
 *
 * Composée à la main plutôt que par toLocaleTimeString : le format ne change ni
 * avec la langue de l'interface ni avec les réglages du navigateur, et
 * vingt-quatre heures se lisent pareil en français et en anglais.
 */
export function heure(instant: string): string {
  const date = new Date(instant);
  return [date.getHours(), date.getMinutes(), date.getSeconds()]
    .map((valeur) => String(valeur).padStart(2, '0'))
    .join(':');
}
