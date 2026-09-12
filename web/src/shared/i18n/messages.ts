// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

/** Langues servies par l'interface. Le bilinguisme s'arrête ici : il ne remonte
 *  ni dans le calque, ni dans l'inférence, ni dans le code généré. */
export type Lang = 'fr' | 'en';

/**
 * Dictionnaire français, qui fait référence pour le typage des clés.
 *
 * Ce qui ne se traduit jamais n'y figure pas : noms de tables, de colonnes, de
 * schémas et types bruts s'affichent tels qu'ils sont en base, puisque c'est ce
 * que l'utilisateur retrouvera dans son SGBD.
 */
const fr = {
  'app.name': 'Ormeau',
  'profile.label': 'Profil',
  'profile.none': 'Nouvelle connexion',
  'profile.save': 'Enregistrer cette connexion',
  'profile.save.name': 'Nom du profil',
  'profile.save.confirm': 'Enregistrer',
  'profile.save.cancel': 'Annuler',
  'profile.delete': 'Supprimer ce profil',
  'profile.replace':
    'Un profil porte déjà ce nom. L’écraser remplacera ses paramètres de connexion, et son mot de passe s’il en avait un.',
  'profile.replace.confirm': 'Écraser',
  'profile.password.save': 'Enregistrer le mot de passe',
  // Deux messages distincts plutôt que le même deux fois : celui-ci explique le
  // champ, l'autre ce qu'un enregistrement ferait.
  'profile.password.stored':
    'Un mot de passe est enregistré pour ce profil : laissez le champ vide pour l’utiliser.',
  'profile.password.erase':
    'Réenregistrer sans cocher la case effacera le mot de passe enregistré pour ce profil.',
  // Ce texte est là où la décision se prend, et il ne dit pas « chiffré » tout
  // court : personne ne doit pouvoir dire qu'on lui a vendu mieux que ce qui
  // est livré.
  'profile.password.save.hint':
    'Il est chiffré sur ce poste, mais la clé y est aussi. Cela protège d’une lecture accidentelle, pas de quelqu’un qui a accès à votre session.',

  'app.workdir': 'Répertoire de travail',
  'app.workdir.copy': 'Copier le chemin complet',
  'app.workdir.change': 'Changer',
  'app.workdir.apply': 'Appliquer',
  'app.workdir.cancel': 'Annuler',
  'app.version': 'Version {version}',
  'app.title.busy': '({n}) Ormeau',

  'theme.label': 'Thème',
  'theme.clair': 'Clair',
  'theme.sombre': 'Sombre',
  'theme.systeme': 'Système',

  'lang.label': 'Langue',

  'connection.title': 'Connexion à la base',
  'connection.intro':
    'L’outil se connecte en lecture seule et ne lit aucune donnée tant que l’échantillonnage n’est pas demandé.',
  'connection.mode.composants': 'Composants',
  'connection.mode.dsn': 'Chaîne de connexion',
  'connection.dbms': 'SGBD',
  'connection.dbms.auto': 'Déduit du port',
  'connection.host': 'Hôte',
  'connection.port': 'Port',
  'connection.user': 'Utilisateur',
  'connection.password': 'Mot de passe',
  'connection.database': 'Base',
  'connection.database.hint': 'Laisser vide pour parcourir tout le serveur',
  'connection.dsn': 'DSN',
  'connection.dsn.hint':
    'Un DATABASE_URL de Symfony convient : les paramètres inutiles sont retirés.',
  'connection.submit': 'Se connecter',
  'connection.submitting': 'Connexion…',
  'connection.disconnect': 'Se déconnecter',
  'connection.connected': 'Connecté à {catalogue}',
  'connection.server': 'Serveur',
  'connection.schemas': 'Schémas',
  'connection.schemas.empty': 'Aucun schéma exploitable dans cette base.',
  'connection.password.hint':
    'Le mot de passe n’est jamais enregistré : il vit le temps de la session.',

  'inventory.title': 'Tables à extraire',
  'inventory.loading': 'Lecture du catalogue…',
  'inventory.empty': 'Aucune table dans les schémas retenus.',
  'inventory.search': 'Filtrer les tables',
  'inventory.noMatch': 'Aucune table ne correspond à « {terme} ».',
  'inventory.selected': '{n} sélectionnée(s) sur {total}',
  'inventory.tables': '{n} table(s)',
  'inventory.columns': '{n} col.',
  'inventory.rows': '{n} lignes',
  'inventory.noPrimaryKey': 'sans clé primaire',
  'inventory.singleDatabase':
    'Ce serveur n’expose qu’une base, ou n’en sait pas énumérer d’autres.',
  // Deux raisons, deux messages : annoncer un serveur à base unique alors qu'il
  // en porte vingt serait un écran qui ment.
  'inventory.profileDatabase':
    'Le profil de connexion désigne cette base. Choisir un autre profil, ou se connecter sans profil, donne accès aux autres.',
  'inventory.selectAll': 'Tout cocher',
  'inventory.clear': 'Tout décocher',
  'inventory.advice':
    'Extraire large, filtrer à l’inférence : le calque garde tout, et les tables ignorées se trient hors ligne, sans revenir en base.',
  'inventory.missing':
    '{n} table(s) référencée(s) par votre sélection n’en font pas partie. Sans elles, les clés étrangères pointent dans le vide et les associations disparaissent.',
  'inventory.addMissing': 'Ajouter les dépendances',

  'columns.loading': 'Lecture des colonnes…',
  'columns.primaryKey': 'clé primaire',
  'columns.primaryKeyKept':
    'Une colonne de clé primaire reste mappée : Doctrine refuse une entité sans identifiant.',
  'columns.excluded': '{n} colonne(s) écartée(s)',
  'columns.expand': 'Déplier les colonnes',

  'action.copy': 'Copier',
  'split.handle': 'Largeur du panneau — tirer, ou flèches gauche et droite',
  'split.handleVertical': 'Hauteur du panneau — tirer, ou flèches haut et bas',

  'scope.title': 'Ce que produira cet écran',
  'scope.extraction': 'Portée envoyée à l’extraction',
  'scope.decisions': 'Décisions — le calque garde tout',
  'scope.copy': 'Copier la portée',
  'scope.copyDecisions': 'Copier les décisions',
  'scope.all': 'toutes les tables des schémas retenus',
  'scope.count': '{n} table(s) sur {schemas} schéma(s)',

  'details.none': 'Choisir une table pour en voir le détail.',
  'details.columns': 'Colonnes',
  'details.rows': 'Lignes estimées',
  'details.primaryKey': 'Clé primaire',
  'details.yes': 'oui',
  'details.no': 'aucune',
  'details.noPrimaryKeyWarning': 'Aucune clé primaire : Doctrine refusera cette entité en l’état.',
  'details.column': 'Colonne',
  'details.type': 'Type',
  'details.nullable': 'Nullable',
  'details.excluded': 'écartée',
  'details.references': 'Références sortantes',
  'details.noReference': 'Aucune clé étrangère déclarée.',
  'details.outOfSelection': 'hors sélection',

  'extraction.launch': 'Extraire',
  'extraction.running': 'Extraction en cours',
  'extraction.target': 'Écrit dans le répertoire de travail',
  'extraction.indicator.running': '{n} extraction(s) en cours',
  'extraction.indicator.failed': '{n} extraction(s) en échec',
  'extraction.indicator.finished': '{n} extraction(s) finie(s)',
  'extraction.panel.title': 'Extractions',
  'extraction.stream.lost': 'Suivi interrompu, reconnexion…',
  'extraction.scope.all': 'toutes les tables',
  'extraction.scope.tables': '{n} table(s)',
  'extraction.state.en_attente': 'en attente',
  'extraction.state.en_cours': 'en cours',
  'extraction.state.terminee': 'terminée',
  'extraction.state.echouee': 'échouée',
  'extraction.state.annulee': 'annulée',
  'extraction.step.source': 'serveur',
  'extraction.step.tables': 'tables',
  'extraction.step.colonnes': 'colonnes',
  'extraction.step.contraintes': 'contraintes',
  'extraction.step.index': 'index',
  'extraction.step.sequences': 'séquences',
  'extraction.step.types_enumeres': 'types énumérés',
  'extraction.step.vues': 'vues',
  'extraction.progress': 'Lecture : {etape}, étape {rang} sur {total}',
  'extraction.progressLabel': 'Avancement par étape',
  'extraction.waiting': 'En attente d’une place : deux extractions tournent déjà.',
  'extraction.cancel': 'Annuler',
  'extraction.dismiss': 'Retirer',
  'extraction.result': '{tables} table(s), {colonnes} colonne(s)',
  'extraction.anomalies': 'Anomalies du calque : {n}',
  'extraction.at': 'à {heure}',
  'extraction.replaced':
    'Fichier réécrit depuis par l’extraction terminée à {heure} : ce résumé ne le décrit plus.',

  'anomaly.version_inconnue': 'Version de calque inconnue',
  'anomaly.sgbd_inconnu': 'SGBD hors du vocabulaire',
  'anomaly.champ_requis_vide': 'Champ requis vide',
  'anomaly.empreinte_malformee': 'Empreinte mal formée',
  'anomaly.type_hors_vocabulaire': 'Type hors du vocabulaire',
  'anomaly.genre_defaut_inconnu': 'Genre de défaut inconnu',
  'anomaly.action_inconnue': 'Action référentielle inconnue',
  'anomaly.position_invalide': 'Position de colonne invalide',
  'anomaly.table_sans_colonne': 'Table sans colonne',
  'anomaly.table_dupliquee': 'Table déclarée deux fois',
  'anomaly.colonne_dupliquee': 'Colonne déclarée deux fois',
  'anomaly.colonne_introuvable': 'Colonne introuvable',
  'anomaly.table_cible_introuvable': 'Table référencée hors du calque',
  'anomaly.arite_incoherente': 'Nombre de colonnes incohérent',
  'anomaly.type_enumere_introuvable': 'Type énuméré introuvable',
  'anomaly.statistiques_orphelines': 'Statistiques sans table',

  'calque.title': 'Dernier calque produit',
  'calque.extractedAt': 'extrait à {heure}',
  'calque.copy': 'Copier le calque',
  'calque.loading': 'Lecture du calque…',
  'calque.statsHidden':
    'Statistiques retirées : l’interface n’affiche aucune valeur échantillonnée.',

  'tabs.label': 'Écrans',
  'tabs.selection': 'Sélection des tables',
  'tabs.arbitrage': 'Arbitrage de {base}',

  'arbitrage.intro':
    'Les entités que la génération produira pour les tables extraites. « Écrire » enregistre dans {fichier} les noms de classes, les types forcés et les colonnes écartées à la sélection.',
  'arbitrage.loading': 'Inférence en cours…',
  'arbitrage.inferring': 'recalcul…',
  'arbitrage.save': 'Écrire {fichier}',
  'arbitrage.saving': 'Écriture…',
  'arbitrage.state.unsaved': 'Modifications non enregistrées',
  'arbitrage.state.saved': 'Enregistré',
  'arbitrage.state.absent': 'Aucun fichier de décisions pour cette base',
  'arbitrage.layerChanged':
    'Le calque a changé depuis le début de l’arbitrage : une extraction l’a réécrit. Les décisions en cours sont gardées ; celles qui ne visent plus rien seront signalées.',
  // La raison vient du serveur, seul à savoir laquelle des deux empreintes a
  // bougé.
  'arbitrage.draftDiscarded':
    'Le travail non enregistré de la dernière fois a été écarté : {raison}. L’écran repart du fichier de décisions.',
  'arbitrage.reload': 'Recharger le calque',
  'arbitrage.fileChanged':
    'Le fichier de décisions a changé sur disque depuis sa lecture. Le relire remplace les modifications en cours.',
  'arbitrage.reread': 'Relire le fichier',
  'arbitrage.manual':
    'Ce fichier porte des modifications faites à la main. L’écrire depuis l’interface le régénère entièrement : ses commentaires et sa mise en forme seront perdus.',
  'arbitrage.overwrite': 'Écraser quand même',
  'arbitrage.cancel': 'Annuler',

  'arbitrage.entities': 'Entités ({n})',
  'arbitrage.entities.filter': 'Filtrer les entités',
  'arbitrage.entities.none': 'Aucune entité à générer dans ce calque.',
  'arbitrage.pending': '{n} à traiter',

  'arbitrage.className': 'Nom de classe',
  'arbitrage.proposal': 'Proposé : {nom} (confiance {pct} %)',
  'arbitrage.useProposal': 'Reprendre',
  'arbitrage.warnings': 'Avertissements',
  'arbitrage.forceType': 'Forcer',
  'arbitrage.forceType.label': 'Type Doctrine pour {colonne}',
  'arbitrage.nameCases': 'Nommer les cas',
  'arbitrage.nameCases.label': 'Nom du cas {valeur}',
  'arbitrage.decided': 'Décisions de cette table',
  'arbitrage.decided.type': '{colonne} : type {type}',
  'arbitrage.decided.enumeration': '{colonne} : {nom} ({cas})',
  'arbitrage.decided.undo': 'Annuler',
  'arbitrage.decided.relation': '{colonne} → {cible}',
  'arbitrage.relation.open': 'Relier à une autre entité',
  'arbitrage.relation.title': 'Relier à une autre entité',
  'arbitrage.relation.column': 'Colonne',
  'arbitrage.relation.target': 'Entité liée',
  'arbitrage.relation.targetColumn': 'Colonne désignée',
  'arbitrage.relation.kind': 'Relation',
  'arbitrage.relation.manyToOne': 'plusieurs {source} pour un {cible} (ManyToOne)',
  'arbitrage.relation.oneToOne': 'un seul {source} par {cible} (OneToOne)',
  'arbitrage.relation.name': 'Nom de la propriété',
  'arbitrage.relation.submit': 'Relier',
  'arbitrage.relation.choose': 'Choisir…',

  'help.label': 'Aide',
  'arbitrage.help.entities':
    'Une entité par table extraite : la classe PHP que la génération produira. Le nombre en orange compte les avertissements à traiter.',
  'arbitrage.help.className':
    'Nom de la classe PHP générée pour cette table. Laissé vide, c’est le nom affiché en gris qui s’applique. Le nom saisi est enregistré dans le fichier de décisions et rejoué à chaque génération.',
  'arbitrage.help.warnings':
    'Ce que l’outil n’a pas pu décider seul pour cette entité. En orange, ce qui demande une action ; en gris, une simple information.',
  'arbitrage.help.forceType':
    'L’outil ne connaît pas le type de cette colonne. Le champ propose le type qu’il a retenu faute de mieux : « Forcer » le confirme et fait disparaître l’avertissement. Un autre type Doctrine peut être saisi, y compris un type propre au projet.',
  'arbitrage.help.nameCases':
    'La colonne n’accepte que ces valeurs, et l’outil en fait une énumération PHP. Faute de mieux, il nomme chaque cas d’après sa valeur : donner un nom lisible (FA → Facture), puis « Nommer les cas ». Confirmer sans rien changer garde les noms proposés.',
  'arbitrage.help.decided':
    'Ce qui est décidé pour cette table en plus du nom de classe : types forcés, cas d’énumération nommés, relations ajoutées. « Annuler » rend la main à l’inférence.',
  'arbitrage.help.relation':
    'Quand la base n’a jamais déclaré la clé étrangère, relier ici la colonne qui contient l’identifiant d’une autre entité — par exemple commande.client_ref vers client.id. La clé primaire de l’entité liée et le nom de la propriété sont proposés. L’entité liée reçoit l’autre côté, sa collection, sans rien faire de plus.',
  'arbitrage.help.entity':
    'Ce que la génération produira : propriétés, types PHP et Doctrine, associations. Recalculé à chaque modification.',
  'arbitrage.help.table':
    'La table telle que l’extraction l’a lue dans la base. Les colonnes écartées dans « Sélection des tables » ne deviennent pas des propriétés.',

  'arbitrage.detail.loading': 'Calcul de l’entité…',
  'arbitrage.detail.entity': 'Entité inférée',
  'arbitrage.detail.table': 'Table physique',
  'arbitrage.detail.noEntity':
    'Aucune entité : la table est ignorée, ou c’est une table de jointure.',
  'arbitrage.detail.identifier': 'Identifiant',
  'arbitrage.detail.property': 'Propriété',
  'arbitrage.detail.php': 'PHP',
  'arbitrage.detail.doctrine': 'Doctrine',
  'arbitrage.detail.associations': 'Associations',
  'arbitrage.association.gives': 'Contient',
  'arbitrage.association.relation': 'Relation',
  'arbitrage.association.one': 'un {cible}',
  'arbitrage.association.many': 'une collection de {cible}',
  'arbitrage.association.column': 'colonne {colonnes}',
  'arbitrage.association.joinTable': 'table de jointure {table}',
  'arbitrage.association.inverse': 'côté inverse de {cible}.{nom}',
  'arbitrage.help.associations':
    'Les liens vers les autres entités, tirés des clés étrangères. « un Plan » : la propriété contient un objet ; « une collection de Tenant » : elle en contient plusieurs. La relation dit d’où vient le lien : la colonne qui le porte, ou l’association dont c’est l’autre côté.',
  'arbitrage.detail.foreignKeys': 'Clés étrangères',
  'arbitrage.detail.copyEntity': 'Copier l’entité',
  'arbitrage.detail.copyTable': 'Copier la table',

  'warning.table_sans_cle_primaire': 'Table sans clé primaire',
  'warning.cle_primaire_composite': 'Clé primaire composite',
  'warning.fk_implicite_probable': 'Clé étrangère implicite probable',
  'warning.type_non_reconnu': 'Type non reconnu',
  'warning.nom_non_singularisable': 'Nom non singularisable',
  'warning.collision_de_nom': 'Collision de nom',
  'warning.decision_sans_cible': 'Décision sans cible',
  'warning.table_ignoree': 'Table ignorée',
  'warning.colonne_ignoree': 'Colonne ignorée',
  'warning.cle_primaire_non_ignorable': 'Clé primaire non ignorable',
  'warning.defaut_incompatible': 'Défaut incompatible',
  'warning.prefixe_detecte': 'Préfixe détecté',
  'warning.cible_hors_portee': 'Cible hors portée',
  'warning.cas_enumeration_opaque': 'Cas d’énumération opaque',
  'warning.trait_deduit': 'Trait déduit',
  'warning.table_de_jointure': 'Table de jointure',
  'warning.heritage_deduit': 'Héritage déduit',

  'error.network': 'Le serveur local ne répond pas.',
  'error.unknown': 'Échec inattendu.',
  'error.notJson': 'Réponse inattendue de {path} ({type}).',
  'error.noContentType': 'sans type de contenu',
} as const;

/** Clé du dictionnaire. Une clé absente fait échouer la compilation, pas
 *  seulement l'exécution. */
export type MessageKey = keyof typeof fr;

/** Dictionnaire anglais. Le type impose qu'il couvre exactement les mêmes clés. */
const en: Record<MessageKey, string> = {
  'app.name': 'Ormeau',
  'profile.label': 'Profile',
  'profile.none': 'New connection',
  'profile.save': 'Save this connection',
  'profile.save.name': 'Profile name',
  'profile.save.confirm': 'Save',
  'profile.save.cancel': 'Cancel',
  'profile.delete': 'Delete this profile',
  'profile.replace':
    'A profile already has this name. Overwriting it will replace its connection settings, and its password if it had one.',
  'profile.replace.confirm': 'Overwrite',
  'profile.password.save': 'Save the password',
  'profile.password.stored':
    'A password is saved for this profile: leave the field empty to use it.',
  'profile.password.erase':
    'Saving again without ticking the box will remove the password saved for this profile.',
  'profile.password.save.hint':
    'It is encrypted on this machine, but so is the key. This guards against accidental reading, not against someone with access to your account.',

  'app.workdir': 'Working directory',
  'app.workdir.copy': 'Copy the full path',
  'app.workdir.change': 'Change',
  'app.workdir.apply': 'Apply',
  'app.workdir.cancel': 'Cancel',
  'app.version': 'Version {version}',
  'app.title.busy': '({n}) Ormeau',

  'theme.label': 'Theme',
  'theme.clair': 'Light',
  'theme.sombre': 'Dark',
  'theme.systeme': 'System',

  'lang.label': 'Language',

  'connection.title': 'Database connection',
  'connection.intro': 'The tool connects read-only and reads no data unless sampling is requested.',
  'connection.mode.composants': 'Fields',
  'connection.mode.dsn': 'Connection string',
  'connection.dbms': 'DBMS',
  'connection.dbms.auto': 'From the port',
  'connection.host': 'Host',
  'connection.port': 'Port',
  'connection.user': 'User',
  'connection.password': 'Password',
  'connection.database': 'Database',
  'connection.database.hint': 'Leave empty to browse the whole server',
  'connection.dsn': 'DSN',
  'connection.dsn.hint': 'A Symfony DATABASE_URL works: unusable parameters are stripped.',
  'connection.submit': 'Connect',
  'connection.submitting': 'Connecting…',
  'connection.disconnect': 'Disconnect',
  'connection.connected': 'Connected to {catalogue}',
  'connection.server': 'Server',
  'connection.schemas': 'Schemas',
  'connection.schemas.empty': 'No usable schema in this database.',
  'connection.password.hint': 'The password is never stored: it lives for the session only.',

  'inventory.title': 'Tables to extract',
  'inventory.loading': 'Reading the catalogue…',
  'inventory.empty': 'No table in the selected schemas.',
  'inventory.search': 'Filter tables',
  'inventory.noMatch': 'No table matches “{terme}”.',
  'inventory.selected': '{n} of {total} selected',
  'inventory.tables': '{n} table(s)',
  'inventory.columns': '{n} col.',
  'inventory.rows': '{n} rows',
  'inventory.noPrimaryKey': 'no primary key',
  'inventory.singleDatabase': 'This server exposes a single database, or cannot enumerate others.',
  'inventory.profileDatabase':
    'The connection profile names this database. Pick another profile, or connect without one, to reach the others.',
  'inventory.selectAll': 'Select all',
  'inventory.clear': 'Clear selection',
  'inventory.advice':
    'Extract wide, filter at inference: the layer keeps everything, and ignored tables are sorted out offline, without going back to the database.',
  'inventory.missing':
    '{n} table(s) referenced by your selection are not part of it. Without them, foreign keys point nowhere and the associations vanish.',
  'inventory.addMissing': 'Add dependencies',

  'columns.loading': 'Reading columns…',
  'columns.primaryKey': 'primary key',
  'columns.primaryKeyKept':
    'A primary key column stays mapped: Doctrine rejects an entity without an identifier.',
  'columns.excluded': '{n} column(s) left out',
  'columns.expand': 'Expand columns',

  'action.copy': 'Copy',
  'split.handle': 'Panel width — drag, or left and right arrows',
  'split.handleVertical': 'Panel height — drag, or up and down arrows',

  'scope.title': 'What this screen will produce',
  'scope.extraction': 'Scope sent to extraction',
  'scope.decisions': 'Decisions — the layer keeps everything',
  'scope.copy': 'Copy the scope',
  'scope.copyDecisions': 'Copy the decisions',
  'scope.all': 'every table in the selected schemas',
  'scope.count': '{n} table(s) across {schemas} schema(s)',

  'details.none': 'Pick a table to see its details.',
  'details.columns': 'Columns',
  'details.rows': 'Estimated rows',
  'details.primaryKey': 'Primary key',
  'details.yes': 'yes',
  'details.no': 'none',
  'details.noPrimaryKeyWarning': 'No primary key: Doctrine will reject this entity as is.',
  'details.column': 'Column',
  'details.type': 'Type',
  'details.nullable': 'Nullable',
  'details.excluded': 'left out',
  'details.references': 'Outgoing references',
  'details.noReference': 'No foreign key declared.',
  'details.outOfSelection': 'outside the selection',

  'extraction.launch': 'Extract',
  'extraction.running': 'Extraction running',
  'extraction.target': 'Written to the working directory',
  'extraction.indicator.running': '{n} extraction(s) running',
  'extraction.indicator.failed': '{n} extraction(s) failed',
  'extraction.indicator.finished': '{n} extraction(s) finished',
  'extraction.panel.title': 'Extractions',
  'extraction.stream.lost': 'Tracking interrupted, reconnecting…',
  'extraction.scope.all': 'all tables',
  'extraction.scope.tables': '{n} table(s)',
  'extraction.state.en_attente': 'waiting',
  'extraction.state.en_cours': 'running',
  'extraction.state.terminee': 'done',
  'extraction.state.echouee': 'failed',
  'extraction.state.annulee': 'cancelled',
  'extraction.step.source': 'server',
  'extraction.step.tables': 'tables',
  'extraction.step.colonnes': 'columns',
  'extraction.step.contraintes': 'constraints',
  'extraction.step.index': 'indexes',
  'extraction.step.sequences': 'sequences',
  'extraction.step.types_enumeres': 'enumerated types',
  'extraction.step.vues': 'views',
  'extraction.progress': 'Reading: {etape}, step {rang} of {total}',
  'extraction.progressLabel': 'Progress by step',
  'extraction.waiting': 'Waiting for a slot: two extractions are already running.',
  'extraction.cancel': 'Cancel',
  'extraction.dismiss': 'Dismiss',
  'extraction.result': '{tables} table(s), {colonnes} column(s)',
  'extraction.anomalies': 'Layer anomalies: {n}',
  'extraction.at': 'at {heure}',
  'extraction.replaced':
    'File rewritten since by the extraction finished at {heure}: this summary no longer describes it.',

  'anomaly.version_inconnue': 'Unknown layer version',
  'anomaly.sgbd_inconnu': 'DBMS outside the vocabulary',
  'anomaly.champ_requis_vide': 'Required field empty',
  'anomaly.empreinte_malformee': 'Malformed fingerprint',
  'anomaly.type_hors_vocabulaire': 'Type outside the vocabulary',
  'anomaly.genre_defaut_inconnu': 'Unknown default kind',
  'anomaly.action_inconnue': 'Unknown referential action',
  'anomaly.position_invalide': 'Invalid column position',
  'anomaly.table_sans_colonne': 'Table without columns',
  'anomaly.table_dupliquee': 'Table declared twice',
  'anomaly.colonne_dupliquee': 'Column declared twice',
  'anomaly.colonne_introuvable': 'Column not found',
  'anomaly.table_cible_introuvable': 'Referenced table outside the layer',
  'anomaly.arite_incoherente': 'Column count mismatch',
  'anomaly.type_enumere_introuvable': 'Enumerated type not found',
  'anomaly.statistiques_orphelines': 'Statistics without a table',

  'calque.title': 'Last produced layer',
  'calque.extractedAt': 'extracted at {heure}',
  'calque.copy': 'Copy the layer',
  'calque.loading': 'Reading the layer…',
  'calque.statsHidden': 'Statistics removed: the interface shows no sampled value.',

  'tabs.label': 'Screens',
  'tabs.selection': 'Table selection',
  'tabs.arbitrage': 'Review of {base}',

  'arbitrage.intro':
    'The entities generation will produce for the extracted tables. “Write” saves class names, forced types and the columns excluded in selection to {fichier}.',
  'arbitrage.loading': 'Running the inference…',
  'arbitrage.inferring': 'recomputing…',
  'arbitrage.save': 'Write {fichier}',
  'arbitrage.saving': 'Writing…',
  'arbitrage.state.unsaved': 'Unsaved changes',
  'arbitrage.state.saved': 'Saved',
  'arbitrage.state.absent': 'No decisions file for this database',
  'arbitrage.layerChanged':
    'The layer changed since the review started: an extraction rewrote it. Current decisions are kept; those that no longer target anything will be reported.',
  'arbitrage.draftDiscarded':
    'Last time’s unsaved work was discarded: {raison}. The screen starts again from the decisions file.',
  'arbitrage.reload': 'Reload the layer',
  'arbitrage.fileChanged':
    'The decisions file changed on disk since it was read. Reading it again replaces the current changes.',
  'arbitrage.reread': 'Read the file again',
  'arbitrage.manual':
    'This file carries hand-made changes. Writing it from the interface regenerates it entirely: its comments and formatting will be lost.',
  'arbitrage.overwrite': 'Overwrite anyway',
  'arbitrage.cancel': 'Cancel',

  'arbitrage.entities': 'Entities ({n})',
  'arbitrage.entities.filter': 'Filter entities',
  'arbitrage.entities.none': 'No entity to generate in this layer.',
  'arbitrage.pending': '{n} to handle',

  'arbitrage.className': 'Class name',
  'arbitrage.proposal': 'Proposed: {nom} (confidence {pct}%)',
  'arbitrage.useProposal': 'Use it',
  'arbitrage.warnings': 'Warnings',
  'arbitrage.forceType': 'Force',
  'arbitrage.forceType.label': 'Doctrine type for {colonne}',
  'arbitrage.nameCases': 'Name the cases',
  'arbitrage.nameCases.label': 'Name of case {valeur}',
  'arbitrage.decided': 'Decisions for this table',
  'arbitrage.decided.type': '{colonne}: type {type}',
  'arbitrage.decided.enumeration': '{colonne}: {nom} ({cas})',
  'arbitrage.decided.undo': 'Undo',
  'arbitrage.decided.relation': '{colonne} → {cible}',
  'arbitrage.relation.open': 'Link to another entity',
  'arbitrage.relation.title': 'Link to another entity',
  'arbitrage.relation.column': 'Column',
  'arbitrage.relation.target': 'Linked entity',
  'arbitrage.relation.targetColumn': 'Referenced column',
  'arbitrage.relation.kind': 'Relation',
  'arbitrage.relation.manyToOne': 'many {source} for one {cible} (ManyToOne)',
  'arbitrage.relation.oneToOne': 'a single {source} per {cible} (OneToOne)',
  'arbitrage.relation.name': 'Property name',
  'arbitrage.relation.submit': 'Link',
  'arbitrage.relation.choose': 'Choose…',

  'help.label': 'Help',
  'arbitrage.help.entities':
    'One entity per extracted table: the PHP class generation will produce. The orange number counts the warnings to handle.',
  'arbitrage.help.className':
    'Name of the PHP class generated for this table. Left empty, the name shown in grey applies. A typed name is saved in the decisions file and replayed at each generation.',
  'arbitrage.help.warnings':
    'What the tool could not decide on its own for this entity. Orange needs an action; grey is information only.',
  'arbitrage.help.forceType':
    'The tool does not know this column’s type. The field suggests the type it fell back on: “Force” confirms it and clears the warning. Another Doctrine type can be typed, including one specific to the project.',
  'arbitrage.help.nameCases':
    'The column only accepts these values, and the tool turns them into a PHP enum. Lacking anything better, it names each case after its value: give a readable name (FA → Invoice), then “Name the cases”. Confirming without changes keeps the suggested names.',
  'arbitrage.help.decided':
    'What is decided for this table besides the class name: forced types, named enumeration cases, added relations. “Undo” gives control back to the inference.',
  'arbitrage.help.relation':
    'When the database never declared the foreign key, link here the column that holds another entity’s identifier — for example commande.client_ref to client.id. The linked entity’s primary key and the property name are suggested. The linked entity gets the other side, its collection, with nothing more to do.',
  'arbitrage.help.entity':
    'What generation will produce: properties, PHP and Doctrine types, associations. Recomputed at each change.',
  'arbitrage.help.table':
    'The table as the extraction read it from the database. Columns excluded in “Table selection” do not become properties.',

  'arbitrage.detail.loading': 'Computing the entity…',
  'arbitrage.detail.entity': 'Inferred entity',
  'arbitrage.detail.table': 'Physical table',
  'arbitrage.detail.noEntity': 'No entity: the table is ignored, or it is a join table.',
  'arbitrage.detail.identifier': 'Identifier',
  'arbitrage.detail.property': 'Property',
  'arbitrage.detail.php': 'PHP',
  'arbitrage.detail.doctrine': 'Doctrine',
  'arbitrage.detail.associations': 'Associations',
  'arbitrage.association.gives': 'Holds',
  'arbitrage.association.relation': 'Relation',
  'arbitrage.association.one': 'one {cible}',
  'arbitrage.association.many': 'a collection of {cible}',
  'arbitrage.association.column': 'column {colonnes}',
  'arbitrage.association.joinTable': 'join table {table}',
  'arbitrage.association.inverse': 'inverse side of {cible}.{nom}',
  'arbitrage.help.associations':
    'Links to other entities, taken from foreign keys. “one Plan”: the property holds an object; “a collection of Tenant”: it holds several. The relation says where the link comes from: the column that carries it, or the association it is the other side of.',
  'arbitrage.detail.foreignKeys': 'Foreign keys',
  'arbitrage.detail.copyEntity': 'Copy the entity',
  'arbitrage.detail.copyTable': 'Copy the table',

  'warning.table_sans_cle_primaire': 'Table without primary key',
  'warning.cle_primaire_composite': 'Composite primary key',
  'warning.fk_implicite_probable': 'Probable implicit foreign key',
  'warning.type_non_reconnu': 'Unrecognized type',
  'warning.nom_non_singularisable': 'Name cannot be singularized',
  'warning.collision_de_nom': 'Name collision',
  'warning.decision_sans_cible': 'Decision without target',
  'warning.table_ignoree': 'Ignored table',
  'warning.colonne_ignoree': 'Ignored column',
  'warning.cle_primaire_non_ignorable': 'Primary key cannot be ignored',
  'warning.defaut_incompatible': 'Incompatible default',
  'warning.prefixe_detecte': 'Prefix detected',
  'warning.cible_hors_portee': 'Target out of scope',
  'warning.cas_enumeration_opaque': 'Opaque enumeration case',
  'warning.trait_deduit': 'Inferred trait',
  'warning.table_de_jointure': 'Join table',
  'warning.heritage_deduit': 'Inferred inheritance',

  'error.network': 'The local server is not responding.',
  'error.unknown': 'Unexpected failure.',
  'error.notJson': 'Unexpected response from {path} ({type}).',
  'error.noContentType': 'no content type',
};

/** Dictionnaires servis à l'interface. */
export const messages: Record<Lang, Record<MessageKey, string>> = { fr, en };

/**
 * Dit si une chaîne est une clé du dictionnaire.
 *
 * Sert aux codes venus du serveur — étapes, anomalies —, dont la clé se
 * construit à l'exécution et que le typage ne peut donc pas vérifier.
 */
export function estCle(cle: string): cle is MessageKey {
  return Object.prototype.hasOwnProperty.call(fr, cle);
}
