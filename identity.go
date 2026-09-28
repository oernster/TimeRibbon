package main

// What Help shows (FR-607, FR-608). Everything here is built into the binary, so About and Licence
// cannot disagree with what was built: the version is the one build.ps1 stamped, the credits come
// from internal/product and the terms are the LICENSE file itself.

import (
	_ "embed"

	"github.com/oernster/timeribbon/internal/product"
)

//go:embed LICENSE
var licenceText string

// About answers what the About panel shows (FR-607).
func (a *App) About() aboutDTO {
	credits := make([]creditDTO, 0, len(product.Credits))
	for _, credit := range product.Credits {
		credits = append(credits, creditDTO{Name: credit.Name, Licence: credit.Licence, Role: credit.Role})
	}
	return aboutDTO{
		Name: product.Name, Version: product.Version, Author: product.Author, Copyright: product.Copyright,
		Credits: credits,
	}
}

// Licence answers the whole of the terms the application is released under (FR-608).
func (a *App) Licence() string { return licenceText }
