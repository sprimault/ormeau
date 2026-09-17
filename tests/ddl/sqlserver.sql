-- Base de référence SQL Server.
--
-- Même schéma logique que tests/ddl/postgres.sql, décliné dans les types de ce
-- dialecte : c'est ce qui rend les deux aller-retours comparables. Trois cas de
-- PostgreSQL n'ont pas d'équivalent ici et prennent la forme idiomatique de
-- SQL Server — le type énuméré devient un CHECK IN, jsonb un nvarchar(max)
-- contrôlé par ISJSON, et le tableau text[] disparaît : le remplacer par une
-- chaîne à séparateurs inventerait une donnée que la base ne porte pas.
--
-- En retour, ce fichier porte ce qu'aucun DDL du dépôt n'avait : un binaire de
-- longueur fixe, un identifiant unique, une colonne calculée non persistée, et
-- les traces d'une base reprise d'avant les ORM.

CREATE DATABASE gescom;
GO
USE gescom;
GO
-- Exigé par SQL Server pour créer un index filtré ou un index sur une colonne
-- calculée, et mémorisé avec eux : sans cette option, toute écriture ultérieure
-- sur la table serait refusée à son tour. sqlcmd ne la pose pas de lui-même.
SET QUOTED_IDENTIFIER ON;
SET ANSI_NULLS ON;
GO
CREATE SCHEMA ventes;
GO

-- Toutes les clés sont nommées, sous les noms que PostgreSQL génère pour le même
-- schéma. Un nom laissé au serveur porte un suffixe tiré à chaque création de la
-- base (PK__t_avoir__47184C0E7EAC9BA9) : le calque de référence ne se
-- reproduirait plus d'un conteneur à l'autre.

-- Cas nominal : entité simple, identité, commentaires, contrainte de contrôle
-- qui doit devenir une énumération.
CREATE TABLE ventes.t_commercial (
    com_id     int IDENTITY (1, 1) CONSTRAINT t_commercial_pkey PRIMARY KEY,
    -- collation par colonne : SQL Server en pose une sur chaque colonne texte,
    -- et « explicite » ne veut dire ici que « différente de celle de la base ».
    com_nom    nvarchar(80) COLLATE French_CI_AS NOT NULL,
    com_actif  bit NOT NULL CONSTRAINT DF_commercial_actif DEFAULT 1,
    com_email  nvarchar(120) NULL,
    -- Empreinte écrite hors ORM : l'entité la lit sans jamais la calculer.
    -- binary(n) est de longueur fixe, là où le bytea de PostgreSQL n'a pas de
    -- longueur : type_doctrine binary, avec l'option fixed.
    com_empreinte binary(64) NULL
);
GO
EXEC sp_addextendedproperty
     @name = N'MS_Description', @value = N'Force de vente',
     @level0type = N'SCHEMA', @level0name = N'ventes',
     @level1type = N'TABLE',  @level1name = N't_commercial';
GO
-- Index filtré : l'équivalent de l'index partiel. Recréé sans son prédicat, il
-- refuserait deux anciens commerciaux de même adresse que la base accepte.
--
-- Différence de dialecte à connaître : SQL Server tient deux NULL pour égaux
-- dans un index unique, quand PostgreSQL les tient pour distincts. La même
-- unicité y est donc plus stricte, et deux lignes sans adresse suffisent à la
-- violer — d'où les adresses données plus bas aux deux commerciaux.
CREATE UNIQUE INDEX uq_com_email_actif ON ventes.t_commercial (com_email)
    WHERE com_actif = 1;
GO

CREATE TABLE ventes.t_client (
    cli_id      int IDENTITY (1, 1) CONSTRAINT t_client_pkey PRIMARY KEY,
    cli_nom     nvarchar(120) NOT NULL,
    cli_siret   nchar(14) NULL,
    cli_statut  nvarchar(20) NOT NULL CONSTRAINT DF_client_statut DEFAULT 'ACTIF',
    cli_com_id  int NULL,
    cli_ca_ttc  decimal(12, 2) NULL,
    -- colonne calculée PERSISTED : le pendant du GENERATED ALWAYS … STORED.
    cli_ca_ht   AS (cli_ca_ttc / 1.2) PERSISTED,
    created_at  datetimeoffset NOT NULL CONSTRAINT DF_client_creation DEFAULT sysdatetimeoffset(),
    updated_at  datetimeoffset NULL,
    CONSTRAINT ck_cli_statut CHECK (cli_statut IN ('ACTIF', 'SUSPENDU', 'ARCHIVE')),
    CONSTRAINT uq_cli_siret UNIQUE (cli_siret),
    CONSTRAINT fk_client_commercial FOREIGN KEY (cli_com_id)
        REFERENCES ventes.t_commercial (com_id) ON DELETE SET NULL
);
GO
EXEC sp_addextendedproperty
     @name = N'MS_Description', @value = N'Nul tant que la fiche n''est pas validée',
     @level0type = N'SCHEMA', @level0name = N'ventes',
     @level1type = N'TABLE',  @level1name = N't_client',
     @level2type = N'COLUMN', @level2name = N'cli_siret';
GO
CREATE INDEX ix_cli_nom ON ventes.t_client (cli_nom);
CREATE INDEX ix_cli_actifs ON ventes.t_client (cli_com_id) WHERE cli_statut = 'ACTIF';
-- Index descendant : SQL Server le fait vraiment, et un index recréé ascendant
-- ne sert pas les mêmes tris.
CREATE INDEX ix_cli_nom_desc ON ventes.t_client (cli_nom DESC);
GO

CREATE TABLE ventes.t_tag (
    tag_id      int IDENTITY (1, 1) CONSTRAINT t_tag_pkey PRIMARY KEY,
    -- collation binaire : un tri octet par octet, que la collation de la base
    -- ne reproduit pas.
    tag_libelle nvarchar(40) COLLATE Latin1_General_BIN2 NOT NULL
);
GO

-- Table de jointure pure : doit produire une association, pas une entité.
CREATE TABLE ventes.t_client_tag (
    cli_id int NOT NULL CONSTRAINT t_client_tag_cli_id_fkey
        REFERENCES ventes.t_client (cli_id) ON DELETE CASCADE,
    tag_id int NOT NULL CONSTRAINT t_client_tag_tag_id_fkey REFERENCES ventes.t_tag (tag_id),
    CONSTRAINT t_client_tag_pkey PRIMARY KEY (cli_id, tag_id)
);
GO
EXEC sp_addextendedproperty
     @name = N'MS_Description', @value = N'Étiquettes posées sur un client',
     @level0type = N'SCHEMA', @level0name = N'ventes',
     @level1type = N'TABLE',  @level1name = N't_client_tag';
GO

-- Table de liaison portant une donnée propre : doit rester une entité. La clé
-- commence par la colonne d'identité dérivée.
CREATE TABLE ventes.t_client_contact (
    cli_id int NOT NULL CONSTRAINT t_client_contact_cli_id_fkey REFERENCES ventes.t_client (cli_id),
    ctc_id int NOT NULL,
    [role] nvarchar(30) NOT NULL,
    CONSTRAINT t_client_contact_pkey PRIMARY KEY (cli_id, ctc_id)
);
GO
EXEC sp_addextendedproperty
     @name = N'MS_Description', @value = N'Client du contact',
     @level0type = N'SCHEMA', @level0name = N'ventes',
     @level1type = N'TABLE',  @level1name = N't_client_contact',
     @level2type = N'COLUMN', @level2name = N'cli_id';
GO

-- Un-vers-un hors clé primaire : la clé étrangère porte une contrainte
-- d'unicité, dont SQL Server crée l'index de soutien du même nom.
CREATE TABLE ventes.t_client_adresse (
    adr_id  int IDENTITY (1, 1) CONSTRAINT t_client_adresse_pkey PRIMARY KEY,
    cli_id  int NOT NULL CONSTRAINT t_client_adresse_cli_id_fkey
        REFERENCES ventes.t_client (cli_id) ON DELETE CASCADE,
    adr_rue nvarchar(200) NOT NULL,
    CONSTRAINT uq_adresse_client UNIQUE (cli_id)
);
GO

-- Héritage : la clé primaire est aussi une clé étrangère.
CREATE TABLE ventes.t_client_grand_compte (
    cli_id      int CONSTRAINT t_client_grand_compte_pkey PRIMARY KEY
        CONSTRAINT t_client_grand_compte_cli_id_fkey REFERENCES ventes.t_client (cli_id),
    remise_taux decimal(4, 2) NOT NULL
);
GO

-- Auto-référence.
CREATE TABLE ventes.t_categorie (
    cat_id      int IDENTITY (1, 1) CONSTRAINT t_categorie_pkey PRIMARY KEY,
    cat_libelle nvarchar(60) NOT NULL,
    cat_parent  int NULL CONSTRAINT t_categorie_cat_parent_fkey REFERENCES ventes.t_categorie (cat_id)
);
GO

-- Aucune clé primaire : doit produire un avertissement, pas une exception.
CREATE TABLE ventes.t_log_import (
    horodatage datetime2 NOT NULL,
    message    nvarchar(max) NULL
);
GO

-- Clé étrangère implicite : aucune contrainte déclarée, mais toutes les valeurs
-- existent dans t_client. Détectable seulement avec --echantillonner.
CREATE TABLE ventes.t_facture (
    fac_id     int IDENTITY (1, 1) CONSTRAINT t_facture_pkey PRIMARY KEY,
    fac_cli_id int NOT NULL,
    fac_total  decimal(12, 2) NOT NULL,
    -- défauts calculés : getdate() sur une date vaut la date du jour, et
    -- CONVERT(date, getdate()) est la forme qu'écrit Doctrine sous DBAL 3.
    fac_date   date NOT NULL CONSTRAINT DF_facture_date DEFAULT getdate(),
    fac_saisie date NULL CONSTRAINT DF_facture_saisie DEFAULT CONVERT(date, getdate()),
    -- simple précision : real existe aussi ici, et se recrée en float.
    fac_taux   real NULL,
    -- money : propre à SQL Server, sans équivalent Doctrine direct.
    fac_remise money NULL,
    -- binaire de longueur déclarée mais variable : recréé en varbinary(16),
    -- quand binary(n) garde sa longueur fixe.
    fac_jeton  varbinary(16) NULL,
    -- tinyint : aucun type Doctrine, smallint est le plus proche.
    fac_niveau tinyint NULL
);
GO

-- Clé par séquence, la forme héritée : un DEFAULT qui lit NEXT VALUE FOR,
-- plutôt qu'une colonne IDENTITY.
CREATE SEQUENCE ventes.sq_avoir AS int START WITH 1 INCREMENT BY 1;
GO
CREATE TABLE ventes.t_avoir (
    avo_id     int CONSTRAINT t_avoir_pkey PRIMARY KEY CONSTRAINT DF_avoir_id DEFAULT (NEXT VALUE FOR ventes.sq_avoir),
    avo_fac_id int NOT NULL
);
GO

-- Clé en longueur fixe, visée par une clé étrangère du même type. Recréé en
-- nvarchar, un nchar(n) ne complète plus ses valeurs d'espaces.
CREATE TABLE ventes.t_pays (
    pay_code    nchar(2) CONSTRAINT t_pays_pkey PRIMARY KEY,
    pay_libelle nvarchar(60) COLLATE French_CI_AS NOT NULL
);
GO

CREATE TABLE ventes.t_commande (
    cmd_id    int IDENTITY (1, 1) CONSTRAINT t_commande_pkey PRIMARY KEY,
    -- Pas de type énuméré nommé sous SQL Server : la forme idiomatique est un
    -- CHECK IN, que l'inférence sait déjà lire pour produire une énumération.
    cmd_canal varchar(10) NOT NULL,
    -- uniqueidentifier avec son défaut calculé : newid() est reconnu, mais
    -- Doctrine ne sait pas l'écrire, et l'application fournit la valeur.
    cmd_ref   uniqueidentifier NOT NULL CONSTRAINT DF_commande_ref DEFAULT newid(),
    cmd_heure time NULL CONSTRAINT DF_commande_heure DEFAULT getdate(),
    cmd_pay_code nchar(2) NULL CONSTRAINT t_commande_cmd_pay_code_fkey REFERENCES ventes.t_pays (pay_code),
    -- Pas de type JSON avant SQL Server 2025 : du texte, et une contrainte qui
    -- dit ce qu'il contient.
    cmd_options nvarchar(max) NULL,
    -- Texte illimité sans Unicode : Doctrine le recrée à l'identique en text,
    -- là où nvarchar(max) n'a pas d'équivalent.
    cmd_notes varchar(max) NULL,
    -- JSON sans Unicode : rendu en json, que Doctrine recrée à l'identique.
    -- cmd_options, Unicode, reste en chaîne.
    cmd_trace varchar(max) NULL,
    CONSTRAINT ck_cmd_canal CHECK (cmd_canal IN ('web', 'telephone', 'agence')),
    CONSTRAINT ck_cmd_options CHECK (cmd_options IS NULL OR ISJSON(cmd_options) = 1),
    CONSTRAINT ck_cmd_trace CHECK (ISJSON(cmd_trace) = 1)
);
GO

-- Identifiants réservés et accents, pour éprouver l'échappement. SQL Server les
-- écrit entre crochets, jamais entre guillemets doubles par défaut.
CREATE TABLE ventes.[t_référence] (
    [id]      int IDENTITY (1, 1) CONSTRAINT [t_référence_pkey] PRIMARY KEY,
    [order]   int NULL,
    [select]  nvarchar(10) NULL,
    [Libellé] nvarchar(30) NULL,
    -- Nom à espaces : courant sur une base reprise, et refusé par DBAL comme
    -- nom d'index.
    [reprise an] int NULL
);
GO
CREATE INDEX ix_reference_order ON ventes.[t_référence] ([order]);
CREATE INDEX ix_reference_libelle ON ventes.[t_référence] ([Libellé]);
GO

-- Reprise d'avant les ORM : une base écrite pour ASP classic, portée telle
-- quelle sur SQL Server. Les contraintes de défaut y portent les noms
-- automatiques de l'époque, à côté de noms choisis.
CREATE TABLE ventes.users (
    id_user       int IDENTITY (1, 1) CONSTRAINT PK_users PRIMARY KEY,
    nom           nvarchar(30) NULL,
    prenom        nvarchar(30) NULL,
    email         nvarchar(100) NULL,
    login         nvarchar(15) NULL,
    nivhab        smallint NULL,
    -- Un nom tel que le serveur l'a généré sur la base d'origine, recopié par
    -- le script de reprise : le calque ne sait pas encore le porter.
    mem_montant   bit CONSTRAINT DF__users__mem_monta__67A95F59 DEFAULT 0,
    -- Le même cas, nommé à la main : les deux formes cohabitent sur une base
    -- reprise, et rien ne les distingue à la lecture du catalogue.
    Change_mdp    bit NOT NULL CONSTRAINT DF_users_Change_mdp DEFAULT 0,
    Actif         bit NULL,
    DT_cre_mdp    datetime NULL,
    -- uniqueidentifier tiré par la base à chaque insertion.
    Salt          uniqueidentifier CONSTRAINT DF_users_Salt DEFAULT newid(),
    -- Empreinte SHA2_512 écrite par une procédure, jamais par l'ORM.
    PasswordHash  binary(64) NULL,
    -- Colonne calculée NON persistée : SQL Server la calcule à la lecture.
    -- PostgreSQL ne sait pas l'exprimer — GENERATED ALWAYS … STORED y est la
    -- seule forme —, donc stockee: false n'a jamais été éprouvé.
    DT_Change_mdp AS dateadd(month, 3, [DT_cre_mdp]),
    DT_creation   datetime NOT NULL CONSTRAINT DF_users_DT_creation DEFAULT getdate(),
    session_id    varchar(255) NULL
);
GO

CREATE VIEW ventes.v_client_actif AS
SELECT cli_id, cli_nom FROM ventes.t_client WHERE cli_statut = 'ACTIF';
GO

INSERT INTO ventes.t_commercial (com_nom, com_email)
VALUES ('Durand', 'durand@example.test'), ('Nguyen', 'nguyen@example.test');
GO
-- Un SIRET par client : sous SQL Server, l'unicité de cli_siret refuse un
-- second NULL (voir uq_com_email_actif plus haut).
INSERT INTO ventes.t_client (cli_nom, cli_siret, cli_statut, cli_com_id, cli_ca_ttc)
VALUES ('Alpha', '11111111111111', 'ACTIF', 1, 1200.00),
       ('Beta', '22222222222222', 'SUSPENDU', 1, 0.00),
       ('Gamma', '33333333333333', 'ARCHIVE', 2, 640.50);
GO
INSERT INTO ventes.t_pays (pay_code, pay_libelle) VALUES ('FR', 'France'), ('BE', 'Belgique');
GO
INSERT INTO ventes.t_facture (fac_cli_id, fac_total) VALUES (1, 100.00), (2, 250.00);
GO
