// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
)

// prefixeEmpreinte ouvre la ligne d'empreinte, en tête du fichier de décisions.
// Il est cherché tel quel à la relecture : le reformuler rendrait manuels tous
// les fichiers déjà écrits.
const prefixeEmpreinte = "# empreinte des décisions : "

// motifEmpreinte reconnaît la ligne d'empreinte.
var motifEmpreinte = regexp.MustCompile(`^` + regexp.QuoteMeta(prefixeEmpreinte) + `(sha256:[0-9a-f]{64})\s*$`)

// ContenuManuel dit si un fichier de décisions porte autre chose que ce que
// l'outil y a écrit, et donc si le réécrire perdrait un travail humain.
//
// La réponse ne vient pas d'une régénération : une partie du texte généré
// dépend du calque — une table ajoutée apporte sa proposition de renommage — et
// de la version du générateur. Elle vient de l'empreinte que le fichier porte,
// celle de ses décisions au moment où l'outil les a écrites.
//
// Un fichier sans empreinte, prérempli d'une version antérieure ou écrit à la
// main, n'est manuel que s'il décide quelque chose : sans paragraphe actif, il
// n'y a rien à perdre.
func ContenuManuel(base string, contenu []byte) bool {
	texte := normaliser(contenu)
	declaree, trouvee := empreinteDeclaree(texte)
	if !trouvee {
		return paragraphesActifs(texte) != ""
	}
	return declaree != empreinte(base, texte)
}

// lignesEmpreinte rend l'en-tête qui porte l'empreinte d'un fichier.
func lignesEmpreinte(base, texte string) string {
	return prefixeEmpreinte + empreinte(base, texte) + "\n" +
		"# Écrite par l'outil à chaque enregistrement ; ne pas la modifier ni la recopier.\n"
}

// empreinte hache les paragraphes actifs d'un fichier, précédés du nom de la
// base.
//
// Les paragraphes actifs seulement : la documentation et les propositions
// générées ne contiennent que des commentaires, et leur texte peut changer
// d'une version de l'outil à l'autre sans que rien de décidé ait bougé. Le nom
// de la base en fait partie pour qu'un fichier recopié d'une base à l'autre ne
// passe pas pour intact.
func empreinte(base, texte string) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(base+"\n"+paragraphesActifs(texte))))
}

// paragraphesActifs rend, joints, les paragraphes qui contiennent au moins une
// ligne hors commentaire.
//
// Un paragraphe est une suite de lignes non vides. Ce découpage range avec sa
// décision le commentaire humain écrit contre elle — sur sa ligne, indenté dans
// son bloc, ou juste au-dessus —, et laisse dehors celui qu'une ligne vide en
// sépare. Il porte sur le texte, sans analyse YAML : une montée de yaml.v3 n'y
// change rien.
func paragraphesActifs(texte string) string {
	var actifs, courant []string
	actif := false

	fermer := func() {
		if actif {
			actifs = append(actifs, strings.Join(courant, "\n"))
		}
		courant, actif = nil, false
	}

	for _, ligne := range strings.Split(texte, "\n") {
		nette := strings.TrimSpace(ligne)
		if nette == "" {
			fermer()
			continue
		}
		courant = append(courant, ligne)
		if !strings.HasPrefix(nette, "#") {
			actif = true
		}
	}
	fermer()

	return strings.Join(actifs, "\n\n")
}

// empreinteDeclaree rend l'empreinte que le fichier porte, s'il en porte une.
func empreinteDeclaree(texte string) (string, bool) {
	for _, ligne := range strings.Split(texte, "\n") {
		if correspondance := motifEmpreinte.FindStringSubmatch(ligne); correspondance != nil {
			return correspondance[1], true
		}
	}
	return "", false
}

// normaliser retire ce qu'un poste ajoute sans rien changer au fichier : le BOM
// d'un éditeur Windows, et les fins de ligne CRLF que git ressort avec
// core.autocrlf. Rien d'autre n'est toléré.
func normaliser(contenu []byte) string {
	texte := strings.TrimPrefix(string(contenu), bom)
	return strings.ReplaceAll(texte, "\r\n", "\n")
}

// bom est la marque d'ordre des octets qu'un éditeur Windows pose en tête d'un
// fichier UTF-8.
const bom = string(rune(0xFEFF))
