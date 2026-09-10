// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useT } from '@/shared/i18n';
import { Button } from '@/shared/ui';
import type { ReponseConnexion } from '@/shared/model';

/** Propriétés du résumé. */
interface ProprietesResume {
  serveur: ReponseConnexion;
  onFermer: () => void;
}

/**
 * Bandeau de la connexion en cours.
 *
 * Sur une ligne, parce qu'il reste affiché pendant tout le travail de sélection
 * et que la place appartient à l'arbre.
 *
 * Le SGBD montré est celui que le serveur a annoncé, pas celui qui a été saisi :
 * viser un serveur MariaDB en écrivant « mysql » doit se voir ici. Les noms de
 * schémas s'affichent tels qu'ils sont en base, jamais traduits.
 */
export function ConnectionSummary({ serveur, onFermer }: ProprietesResume) {
  const t = useT();

  return (
    <section className="flex flex-wrap items-center gap-x-4 gap-y-1">
      <h1 className="text-sm font-semibold">
        {t('connection.connected', { catalogue: serveur.catalogue })}
      </h1>
      <span className="font-mono text-xs text-slate-500">
        {serveur.sgbd} {serveur.version}
      </span>

      <span className="min-w-0 text-xs text-slate-500">
        {t('connection.schemas')} :{' '}
        {serveur.schemas.length === 0 ? (
          <span className="text-amber-700 dark:text-amber-500">
            {t('connection.schemas.empty')}
          </span>
        ) : (
          <span className="font-mono">{serveur.schemas.join(', ')}</span>
        )}
      </span>

      <div className="ml-auto shrink-0">
        <Button variante="discret" onClick={onFermer}>
          {t('connection.disconnect')}
        </Button>
      </div>
    </section>
  );
}
