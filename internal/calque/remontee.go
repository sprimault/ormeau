// Copyright 2026 Stéphane Primault <sprimault@users.noreply.github.com>
// SPDX-License-Identifier: Apache-2.0

package calque

import "slices"

// remonter aligne en mémoire un calque physique d'une version antérieure sur
// le sens de la version courante, sans rien inventer : seul ce que l'ancienne
// version disait deux fois est retiré.
//
// VersionRI garde la version lue. Elle dit ce qu'est le fichier, pas ce qu'on
// en a fait, et l'empreinte de source reste celle de son contenu d'origine.
func (p *Physique) remonter() {
	if p.VersionRI < 2 {
		for i := range p.Tables {
			p.Tables[i].Index = sansIndexDUnicite(p.Tables[i])
		}
	}
}

// sansIndexDUnicite retire l'index qu'une version 1 reportait en plus de la
// contrainte d'unicité qu'il soutient : même nom, mêmes colonnes dans le même
// ordre, unique. Un index qui ne partage que le nom — une base où les deux
// coexistent sous des colonnes différentes — reste.
func sansIndexDUnicite(t Table) []Index {
	if len(t.Unicites) == 0 || len(t.Index) == 0 {
		return t.Index
	}
	return slices.DeleteFunc(slices.Clone(t.Index), func(idx Index) bool {
		return idx.Unique && idx.Predicat == "" && slices.ContainsFunc(t.Unicites, func(u Contrainte) bool {
			return u.Nom == idx.Nom && slices.Equal(u.Colonnes, idx.Colonnes)
		})
	})
}
