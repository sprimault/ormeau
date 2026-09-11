// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package inference

import (
	"fmt"
	"strconv"
	"strings"
)

// motsReserves sont les écritures qu'un lecteur YAML prendrait pour autre chose
// qu'une chaîne : nul, booléens des deux versions de la norme, infinis, clé de
// fusion. Écrites nues, elles se reliraient mal — une clé null devient la
// chaîne vide.
var motsReserves = map[string]bool{
	"": true, "~": true, "null": true, "true": true, "false": true,
	"y": true, "n": true, "yes": true, "no": true, "on": true, "off": true,
	".inf": true, "-.inf": true, "+.inf": true, ".nan": true, "<<": true,
}

// scalaire rend une chaîne sous la forme YAML qui se relit à l'identique.
//
// Nue quand rien ne prête à confusion, entre apostrophes sinon, entre
// guillemets doubles pour un caractère de contrôle, que les apostrophes ne
// savent pas porter. Écrit à la main plutôt que confié à l'émetteur de
// yaml.v3 : ses règles de guillemets peuvent changer d'une version à l'autre,
// et le fichier vit dans le dépôt de l'utilisateur, où une ligne qui change sans
// décision nouvelle est un diff que personne ne comprend.
//
// flux signale une valeur écrite dans une liste en ligne, où virgules et
// crochets deviennent significatifs.
func scalaire(valeur string, flux bool) string {
	switch {
	case strings.IndexFunc(valeur, estControle) >= 0:
		return guillemets(valeur)
	case apostrophesRequises(valeur, flux):
		return "'" + strings.ReplaceAll(valeur, "'", "''") + "'"
	default:
		return valeur
	}
}

// apostrophesRequises dit si une chaîne écrite nue changerait de sens ou de
// structure.
func apostrophesRequises(valeur string, flux bool) bool {
	switch {
	case motsReserves[strings.ToLower(valeur)]:
		return true
	case strings.TrimSpace(valeur) != valeur:
		return true
	case strings.ContainsRune("-?:,[]{}#&*!|>'\"%@`", rune(valeur[0])):
		return true
	case strings.Contains(valeur, ": "), strings.Contains(valeur, " #"), strings.HasSuffix(valeur, ":"):
		return true
	case flux && strings.ContainsAny(valeur, ",[]{}"):
		return true
	}
	// Un nombre se relirait comme une chaîne dans une structure de chaînes, mais
	// pas forcément ailleurs : 007 ou 1e3 restent entre apostrophes.
	_, err := strconv.ParseFloat(valeur, 64)
	return err == nil
}

// estControle reconnaît un caractère de contrôle.
func estControle(r rune) bool {
	return r < 0x20 || r == 0x7f
}

// guillemets écrit une chaîne entre guillemets doubles, contrôles échappés.
func guillemets(valeur string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range valeur {
		switch {
		case r == '"' || r == '\\':
			b.WriteByte('\\')
			b.WriteRune(r)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\t':
			b.WriteString(`\t`)
		case estControle(r):
			fmt.Fprintf(&b, `\x%02X`, r)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}
