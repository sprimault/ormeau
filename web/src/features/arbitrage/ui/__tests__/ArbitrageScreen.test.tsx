// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ErreurAPI } from '@/shared/api';
import { useLangStore } from '@/shared/i18n';
import {
  CodeCalqueModifie,
  CodeColonneIgnoree,
  CodeTableSansClePrimaire,
  CodeTraitDeduit,
  CodeTypeNonReconnu,
  type Decisions,
  type ReponseDecisions,
  type ReponseEntite,
  type ReponseInference,
} from '@/shared/model';
import { ecrireDecisions, inferer, lireEntite } from '../../api/arbitrageApi';
import { ArbitrageScreen } from '../ArbitrageScreen';

vi.mock('../../api/arbitrageApi', () => ({
  inferer: vi.fn(),
  lireEntite: vi.fn(),
  ecrireDecisions: vi.fn(),
}));

/** Le brouillon simulé : le fichier dont il part, et ce qu'il est devenu. */
const brouillon = vi.hoisted(() => ({
  fichier: null as ReponseDecisions | null,
  initial: {} as Decisions,
  courant: {} as Decisions,
}));

// Le brouillon a son fournisseur et ses tests : un état en mémoire suffit pour
// voir ce que l'écran y écrit.
vi.mock('@/entities/decisions', async () => {
  const { useCallback, useState } = await import('react');
  return {
    useDecisions: () => {
      const [decisions, setDecisions] = useState<Decisions>(brouillon.initial);
      const modifier = useCallback(
        (transformation: (d: Decisions) => Decisions) => setDecisions((d) => transformation(d)),
        [],
      );
      brouillon.courant = decisions;
      return {
        base: 'gescom',
        decisions,
        fichier: brouillon.fichier,
        modifie: false,
        modifier,
        relire: () => {},
        enregistre: () => {},
      };
    },
  };
});

/** Inférence d'un calque à deux entités : Clients, et TLog sans clé primaire. */
function inference(complement: Partial<ReponseInference> = {}): ReponseInference {
  return {
    empreinte_physique: 'sha256:1',
    avertissements: [
      {
        code: CodeTableSansClePrimaire,
        cible: 'public.t_log',
        message: 'aucune clé primaire',
        resolution: 'aucune',
        confiance: 1,
      },
      {
        code: CodeTypeNonReconnu,
        cible: 'public.clients.position',
        message: 'type point sans correspondance, rendu en chaîne',
        resolution: 'par_defaut',
        confiance: 0.3,
      },
      {
        code: CodeTraitDeduit,
        cible: 'public.clients',
        message: 'colonnes d’horodatage sorties dans un trait',
        resolution: 'par_defaut',
        confiance: 0.8,
      },
    ],
    propositions: [
      { cible: 'public.clients', nom: 'Client', raison: 'pluriel anglais', confiance: 0.6 },
    ],
    enumerations: [],
    entites: [
      { nom: 'Clients', table: { schema: 'public', nom: 'clients' }, nb_proprietes: 3, nb_associations: 0 },
      { nom: 'TLog', table: { schema: 'public', nom: 't_log' }, nb_proprietes: 1, nb_associations: 0 },
    ],
    types_doctrine: ['boolean', 'integer', 'string'],
    ...complement,
  };
}

/** Détail d'une entité, tel que le serveur le rend. */
function detail(table: string): ReponseEntite {
  if (table === 't_log') {
    return {
      entite: {
        nom: 'TLog',
        table: { schema: 'public', nom: 't_log' },
        proprietes: [
          { nom: 'ligne', colonne: 'ligne', type_php: 'string', type_doctrine: 'text', nullable: true },
        ],
      },
      table_physique: {
        schema: 'public',
        nom: 't_log',
        colonnes: [
          { nom: 'ligne', position: 1, type_brut: 'varchar(80)', type_normalise: 'texte', nullable: true },
        ],
      },
    };
  }
  return {
    entite: {
      nom: 'Clients',
      table: { schema: 'public', nom: 'clients' },
      proprietes: [
        { nom: 'id', colonne: 'id', type_php: 'int', type_doctrine: 'integer', nullable: false },
        { nom: 'position', colonne: 'position', type_php: 'string', type_doctrine: 'string', nullable: false },
      ],
    },
    table_physique: {
      schema: 'public',
      nom: 'clients',
      cle_primaire: { colonnes: ['id'] },
      colonnes: [
        { nom: 'id', position: 1, type_brut: 'int4', type_normalise: 'entier', nullable: false },
        { nom: 'position', position: 2, type_brut: 'point', type_normalise: 'inconnu', nullable: false },
        { nom: 'note', position: 3, type_brut: 'text', type_normalise: 'texte', nullable: true },
      ],
    },
  };
}

describe('ArbitrageScreen', () => {
  beforeEach(() => {
    useLangStore.setState({ lang: 'fr' });
    brouillon.fichier = null;
    brouillon.initial = {};
    vi.mocked(inferer).mockReset();
    vi.mocked(lireEntite).mockReset();
    vi.mocked(lireEntite).mockImplementation(async (requete) => detail(requete.table));
    vi.mocked(ecrireDecisions).mockReset();
  });

  it('ouvre la première entité sans attendre de clic, à côté de sa table', async () => {
    vi.mocked(inferer).mockResolvedValue(inference());
    render(<ArbitrageScreen base="gescom" versionCalque="" />);

    expect(await screen.findByText('Entité inférée')).toBeInTheDocument();
    expect(screen.getByText('int4')).toBeInTheDocument();
    expect(vi.mocked(lireEntite).mock.calls[0][0]).toMatchObject({ schema: 'public', table: 'clients' });
  });

  it('compte dans la liste les avertissements à traiter, sans les informations', async () => {
    vi.mocked(inferer).mockResolvedValue(inference());
    render(<ArbitrageScreen base="gescom" versionCalque="" />);

    expect(await screen.findByRole('button', { name: /Clients.*1 à traiter/ })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /TLog.*1 à traiter/ })).toBeInTheDocument();
  });

  it('montre les avertissements au-dessus de l’entité, et confirme d’un clic le type retenu', async () => {
    vi.mocked(inferer).mockResolvedValue(inference());
    render(<ArbitrageScreen base="gescom" versionCalque="" />);

    expect(await screen.findByText(/type point sans correspondance/)).toBeInTheDocument();
    const saisie = screen.getByRole('combobox', { name: 'Type Doctrine pour position' });
    await waitFor(() => expect(saisie).toHaveValue('string'));
    await userEvent.click(screen.getByRole('button', { name: 'Forcer' }));

    expect(brouillon.courant.types_forces).toEqual({ 'public.clients.position': 'string' });
    expect(screen.getByText('position → string')).toBeInTheDocument();

    await userEvent.click(screen.getByRole('button', { name: 'Annuler' }));
    expect(brouillon.courant.types_forces).toBeUndefined();
  });

  it('saisit un autre type que celui retenu', async () => {
    vi.mocked(inferer).mockResolvedValue(inference());
    render(<ArbitrageScreen base="gescom" versionCalque="" />);

    const saisie = await screen.findByRole('combobox', { name: 'Type Doctrine pour position' });
    await waitFor(() => expect(saisie).toHaveValue('string'));
    await userEvent.clear(saisie);
    await userEvent.type(saisie, 'point');
    await userEvent.click(screen.getByRole('button', { name: 'Forcer' }));

    expect(brouillon.courant.types_forces).toEqual({ 'public.clients.position': 'point' });
  });

  it('ouvre une autre entité en cliquant sur sa ligne', async () => {
    vi.mocked(inferer).mockResolvedValue(inference());
    render(<ArbitrageScreen base="gescom" versionCalque="" />);

    await userEvent.click(await screen.findByRole('button', { name: /TLog/ }));

    expect(await screen.findByText('varchar(80)')).toBeInTheDocument();
    expect(screen.getByText(/aucune clé primaire/)).toBeInTheDocument();
  });

  it('marque la colonne écartée à la sélection, sans la répéter en avertissement', async () => {
    brouillon.initial = { colonnes_ignorees: { 'public.clients': ['note'] } };
    vi.mocked(inferer).mockResolvedValue(
      inference({
        avertissements: [
          {
            code: CodeColonneIgnoree,
            cible: 'public.clients.note',
            message: 'note est retirée de l’entité',
            resolution: 'ignoree',
            confiance: 1,
          },
        ],
      }),
    );
    render(<ArbitrageScreen base="gescom" versionCalque="" />);

    expect(await screen.findByText('(écartée)')).toBeInTheDocument();
    expect(screen.queryByText(/retirée de l’entité/)).not.toBeInTheDocument();
    expect(screen.queryByText('Avertissements')).not.toBeInTheDocument();
  });

  it('renomme la classe en reprenant la proposition', async () => {
    vi.mocked(inferer).mockResolvedValue(inference());
    render(<ArbitrageScreen base="gescom" versionCalque="" />);

    expect(await screen.findByText(/pluriel anglais/)).toBeInTheDocument();
    await userEvent.click(screen.getByRole('button', { name: 'Reprendre' }));

    expect(brouillon.courant.renommages).toEqual({ 'public.clients': 'Client' });
    expect(screen.getByRole('textbox', { name: 'Nom de classe' })).toHaveValue('Client');
  });

  it('annonce un calque réécrit par une extraction, et le recharge', async () => {
    vi.mocked(inferer)
      .mockResolvedValueOnce(inference())
      .mockRejectedValueOnce(new ErreurAPI(409, 'le calque a changé', CodeCalqueModifie))
      .mockResolvedValueOnce(inference({ empreinte_physique: 'sha256:2' }));
    const { rerender } = render(<ArbitrageScreen base="gescom" versionCalque="" />);
    await screen.findByText('Entité inférée');

    rerender(<ArbitrageScreen base="gescom" versionCalque="2026-09-11T10:00:00Z" />);
    await userEvent.click(await screen.findByRole('button', { name: 'Recharger le calque' }));

    await waitFor(() => expect(inferer).toHaveBeenCalledTimes(3));
    expect(vi.mocked(inferer).mock.calls[2][0].empreinte_physique).toBeUndefined();
    await waitFor(() =>
      expect(screen.queryByRole('button', { name: 'Recharger le calque' })).not.toBeInTheDocument(),
    );
  });

  it('demande confirmation avant d’écraser un fichier écrit à la main', async () => {
    brouillon.fichier = { existe: true, decisions: {}, empreinte_fichier: 'sha256:f1', manuel: true };
    vi.mocked(inferer).mockResolvedValue(inference());
    vi.mocked(ecrireDecisions).mockResolvedValue({ empreinte_fichier: 'sha256:f2' });
    render(<ArbitrageScreen base="gescom" versionCalque="" />);
    await screen.findByText('Entité inférée');

    await userEvent.click(screen.getByRole('button', { name: 'Écrire gescom.decisions.yaml' }));
    expect(ecrireDecisions).not.toHaveBeenCalled();

    await userEvent.click(screen.getByRole('button', { name: 'Écraser quand même' }));

    await waitFor(() => expect(ecrireDecisions).toHaveBeenCalled());
    expect(vi.mocked(ecrireDecisions).mock.calls[0][0]).toMatchObject({
      base: 'gescom',
      empreinte_physique: 'sha256:1',
      empreinte_fichier: 'sha256:f1',
      ecraser_manuel: true,
    });
  });
});
