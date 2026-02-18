package helpers

// SearchHelper manages search execution and query saving.
type SearchHelper struct {
	c *HelperCommon
}

// NewSearchHelper creates a new SearchHelper.
func NewSearchHelper(c *HelperCommon) *SearchHelper {
	return &SearchHelper{c: c}
}
