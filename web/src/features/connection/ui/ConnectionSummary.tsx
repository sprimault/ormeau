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
 * Ce que l'écran affiche une fois la base atteinte.
 *
 * Le SGBD montré est celui que le serveur a annoncé, pas celui qui a été saisi :
 * viser un serveur MariaDB en écrivant « mysql » doit se voir ici.
 *
 * Les noms de schémas s'affichent tels qu'ils sont en base — jamais traduits,
 * jamais normalisés : c'est ce que l'utilisateur retrouvera dans son SGBD.
 */
export function ConnectionSummary({ serveur, onFermer }: ProprietesResume) {
  const t = useT();

  return (
    <section className="flex w-full max-w-xl flex-col gap-4">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-lg font-semibold">
            {t('connection.connected', { catalogue: serveur.catalogue })}
          </h1>
          <p className="mt-1 text-sm text-slate-600 dark:text-slate-400">
            {t('connection.server')} : <span className="font-mono">{serveur.sgbd}</span>{' '}
            <span className="font-mono">{serveur.version}</span>
          </p>
        </div>
        <Button variante="discret" onClick={onFermer}>
          {t('connection.disconnect')}
        </Button>
      </div>

      <div>
        <h2 className="text-xs font-medium text-slate-600 uppercase dark:text-slate-400">
          {t('connection.schemas')}
        </h2>
        {serveur.schemas.length === 0 ? (
          <p className="mt-1 text-sm text-slate-500">{t('connection.schemas.empty')}</p>
        ) : (
          <ul className="mt-1 flex flex-wrap gap-1">
            {serveur.schemas.map((schema) => (
              <li
                key={schema}
                className="rounded bg-slate-200 px-2 py-0.5 font-mono text-xs text-slate-800 dark:bg-slate-800 dark:text-slate-200"
              >
                {schema}
              </li>
            ))}
          </ul>
        )}
      </div>
    </section>
  );
}
