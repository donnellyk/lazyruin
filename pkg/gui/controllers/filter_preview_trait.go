package controllers

import "kvnd/lazyruin/pkg/gui/types"

// FilterablePreviewTrait provides shared filter bindings for preview controllers
// that support card-list filtering (CardList, PickResults).
type FilterablePreviewTrait struct {
	c *ControllerCommon
}

func (t *FilterablePreviewTrait) openFilter() error {
	return t.c.Helpers().CardListFilter().OpenFilterDialog()
}

func (t *FilterablePreviewTrait) clearFilter() error {
	return t.c.Helpers().CardListFilter().ClearFilter()
}

func (t *FilterablePreviewTrait) filterNotActive() *types.DisabledReason {
	if !t.c.Helpers().CardListFilter().FilterActive() {
		return &types.DisabledReason{Text: "No active filter"}
	}
	return nil
}
