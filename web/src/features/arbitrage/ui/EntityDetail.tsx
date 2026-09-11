// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { useT } from '@/shared/i18n';
import { qualifier } from '@/shared/lib';
import type { Entite, ReponseEntite, Table } from '@/shared/model';
import { CopyButton, HelpTip } from '@/shared/ui';

/** Propriétés du détail. */
interface ProprietesDetail {
  detail: ReponseEntite;
  /** Colonnes écartées dans l'onglet de sélection. */
  ecartees: string[];
}

/**
 * L'entité inférée à côté de la table dont elle vient.
 *
 * C'est là qu'une décision se vérifie : un nom de classe, un type forcé ou une
 * colonne écartée s'y lisent dans leur effet, avant qu'aucune classe PHP
 * n'existe. Les deux se copient, pour les relire ailleurs.
 */
export function EntityDetail({ detail, ecartees }: ProprietesDetail) {
  const t = useT();
  const { entite, table_physique: table } = detail;

  return (
    <div className="grid gap-4 xl:grid-cols-2">
      <section className="min-w-0">
        <Titre
          libelle={t('arbitrage.detail.entity')}
          aide={t('arbitrage.help.entity')}
          copie={entite ? JSON.stringify(entite, null, 2) : undefined}
          libelleCopie={t('arbitrage.detail.copyEntity')}
        />
        {entite ? (
          <EntityView entite={entite} />
        ) : (
          <p className="text-slate-500">{t('arbitrage.detail.noEntity')}</p>
        )}
      </section>
      <section className="min-w-0">
        <Titre
          libelle={t('arbitrage.detail.table')}
          aide={t('arbitrage.help.table')}
          copie={JSON.stringify(table, null, 2)}
          libelleCopie={t('arbitrage.detail.copyTable')}
        />
        <TableView table={table} ecartees={ecartees} />
      </section>
    </div>
  );
}

/** Propriétés d'un titre de colonne. */
interface ProprietesTitre {
  libelle: string;
  aide: string;
  copie?: string;
  libelleCopie: string;
}

/** Titre d'une des deux colonnes, avec son aide et sa copie. */
function Titre({ libelle, aide, copie, libelleCopie }: ProprietesTitre) {
  return (
    <div className="mb-2 flex items-center gap-2">
      <h4 className="flex items-center gap-1 font-semibold text-slate-600 dark:text-slate-400">
        {libelle}
        <HelpTip texte={aide} />
      </h4>
      {copie ? <CopyButton valeur={copie} libelle={libelleCopie} /> : null}
    </div>
  );
}

/** L'entité : identifiant, propriétés avec leurs types, associations. */
function EntityView({ entite }: { entite: Entite }) {
  const t = useT();

  return (
    <>
      <p className="mb-1 font-mono text-sm font-semibold">{entite.nom}</p>
      {entite.identifiant ? (
        <p className="mb-2 text-slate-500">
          {t('arbitrage.detail.identifier')} :{' '}
          <span className="font-mono">{entite.identifiant.proprietes.join(', ')}</span> (
          {entite.identifiant.strategie})
        </p>
      ) : null}
      <div className="overflow-x-auto">
        <table className="w-full text-left">
          <thead className="text-slate-500">
            <tr>
              <th className="pr-2 font-normal">{t('arbitrage.detail.property')}</th>
              <th className="pr-2 font-normal">{t('details.column')}</th>
              <th className="pr-2 font-normal">{t('arbitrage.detail.php')}</th>
              <th className="font-normal">{t('arbitrage.detail.doctrine')}</th>
            </tr>
          </thead>
          <tbody className="font-mono">
            {entite.proprietes.map((p) => (
              <tr key={p.nom}>
                <td className="pr-2">{p.nom}</td>
                <td className="pr-2 text-slate-500">{p.colonne}</td>
                <td className="pr-2">
                  {p.nullable ? '?' : ''}
                  {p.type_php}
                </td>
                <td>{p.type_doctrine}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {entite.associations?.length ? (
        <>
          <h5 className="mt-3 mb-1 text-slate-500">{t('arbitrage.detail.associations')}</h5>
          <ul className="font-mono">
            {entite.associations.map((a) => (
              <li key={a.nom}>
                {a.nom} : {a.genre} → {a.cible}
              </li>
            ))}
          </ul>
        </>
      ) : null}
    </>
  );
}

/** Propriétés de la table. */
interface ProprietesTable {
  table: Table;
  ecartees: string[];
}

/** La table telle que le calque la décrit : colonnes, types bruts, clés. */
function TableView({ table, ecartees }: ProprietesTable) {
  const t = useT();
  const cle = new Set(table.cle_primaire?.colonnes ?? []);

  return (
    <>
      <p className="mb-1 font-mono text-sm font-semibold">{qualifier(table.schema, table.nom)}</p>
      {table.commentaire ? <p className="mb-2 text-slate-500">{table.commentaire}</p> : null}
      <div className="overflow-x-auto">
        <table className="w-full text-left">
          <thead className="text-slate-500">
            <tr>
              <th className="pr-2 font-normal">{t('details.column')}</th>
              <th className="pr-2 font-normal">{t('details.type')}</th>
              <th className="font-normal">{t('details.nullable')}</th>
            </tr>
          </thead>
          <tbody className="font-mono">
            {table.colonnes.map((c) => (
              <tr key={c.nom} className={ecartees.includes(c.nom) ? 'text-slate-400' : ''}>
                <td className="pr-2">
                  {c.nom}
                  {cle.has(c.nom) ? (
                    <span className="ml-1 font-sans text-slate-500">({t('columns.primaryKey')})</span>
                  ) : null}
                  {ecartees.includes(c.nom) ? (
                    <span className="ml-1 font-sans">({t('details.excluded')})</span>
                  ) : null}
                </td>
                <td className="pr-2">{c.type_brut}</td>
                <td>{c.nullable ? t('details.yes') : ''}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {table.cles_etrangeres?.length ? (
        <>
          <h5 className="mt-3 mb-1 text-slate-500">{t('arbitrage.detail.foreignKeys')}</h5>
          <ul className="font-mono">
            {table.cles_etrangeres.map((fk, rang) => (
              <li key={fk.nom ?? rang}>
                ({fk.colonnes.join(', ')}) → {qualifier(fk.schema_cible, fk.table_cible)} (
                {fk.colonnes_cibles.join(', ')})
              </li>
            ))}
          </ul>
        </>
      ) : null}
    </>
  );
}
