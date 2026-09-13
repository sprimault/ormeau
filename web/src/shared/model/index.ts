// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

export type {
  RequeteConnexion,
  ReponseConnexion,
  RequeteFermeture,
  ReponseContexte,
  RequeteRepertoire,
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
  CodeProfilExistant,
  type ProfilResume,
  type ReponseProfils,
  type RequeteProfil,
  type ReferenceProfil,
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
  CodeDecisionInvalide,
  CodeDecisionOrpheline,
  CodeDefautIncompatible,
  CodeFKImpliciteProbable,
  CodeHeritageDeduit,
  CodeJointurePure,
  CodePrefixeDetecte,
  CodeSequenceNonReconnue,
  CodeTableIgnoree,
  CodeTableSansClePrimaire,
  CodeTraitDeduit,
  CodeTypeNonReconnu,
} from './calque';
export {
  CodeActionInconnue,
  CodeAriteIncoherente,
  CodeChampRequisVide,
  CodeColonneDupliquee,
  CodeColonneIntrouvable,
  CodeEmpreinteMalformee,
  CodeGenreDefautInconnu,
  CodePositionInvalide,
  CodeSGBDInconnu,
  CodeStatistiquesOrphelines,
  CodeTableCibleIntrouvable,
  CodeTableDupliquee,
  CodeTableSansColonne,
  CodeTypeEnumereIntrouvable,
  CodeTypeHorsVocabulaire,
  CodeVersionInconnue,
} from './calque';
export type { TableSommaire, ColonneSommaire, Portee, Avancement } from './introspection';
export { initTheme, themeInitial, useThemeStore, type Theme } from './theme';
export type { Preferences, Profil, Session } from './config';
export type { ReponseSession, ReferenceSession } from './api';
export {
  brancherEnregistrement,
  usePreferencesStore,
  variableCSS,
  type ClePreference,
} from './preferences';
