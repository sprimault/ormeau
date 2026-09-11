// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

export type {
  RequeteConnexion,
  ReponseConnexion,
  RequeteFermeture,
  ReponseContexte,
  ReponseErreur,
  ReponseInventaire,
  ReponseBases,
  RequeteBase,
  ReponseColonnes,
} from './api';
export {
  EtatAnnulee,
  EtatEchouee,
  EtatEnAttente,
  EtatEnCours,
  EtatTerminee,
  type EtatExtraction,
  type EtatExtractions,
  type Extraction,
  type ReferenceExtraction,
  type RequeteExtraction,
  type ReponseCalque,
  type ReponseDecisions,
  type ResultatExtraction,
} from './api';
export {
  CodeCalqueModifie,
  CodeContenuManuel,
  CodeDecisionsModifiees,
  type CodeRefus,
  type EnumerationInferee,
  type RequeteEcritureDecisions,
  type ReponseEcritureDecisions,
  type RequeteEntite,
  type ReponseEntite,
  type RequeteInference,
  type ReponseInference,
  type ResumeEntite,
} from './api';
export type { Decisions, EnumerationForcee, Proposition, RelationForcee } from './inference';
export type {
  Anomalie,
  Association,
  Avertissement,
  Entite,
  ReferenceTable,
  Table,
} from './calque';
export {
  CodeCasEnumerationOpaque,
  CodeCibleHorsPortee,
  CodeClePrimaireComposite,
  CodeClePrimaireGardee,
  CodeCollision,
  CodeColonneIgnoree,
  CodeDecisionOrpheline,
  CodeDefautIncompatible,
  CodeFKImpliciteProbable,
  CodeHeritageDeduit,
  CodeJointurePure,
  CodeNomNonSingularisable,
  CodePrefixeDetecte,
  CodeTableIgnoree,
  CodeTableSansClePrimaire,
  CodeTraitDeduit,
  CodeTypeNonReconnu,
} from './calque';
export type { TableSommaire, ColonneSommaire, Portee, Avancement } from './introspection';
export { CLE_THEME, initTheme, themeEnregistre, useThemeStore, type Theme } from './theme';
