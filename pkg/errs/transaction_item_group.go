package errs

import "net/http"

// Error codes related to transaction item groups
var (
	ErrTransactionItemGroupIdInvalid            = NewNormalError(NormalSubcategoryItemGroup, 0, http.StatusBadRequest, "transaction item group id is invalid")
	ErrTransactionItemGroupNotFound             = NewNormalError(NormalSubcategoryItemGroup, 1, http.StatusBadRequest, "transaction item group not found")
	ErrTransactionItemGroupInUseCannotBeDeleted = NewNormalError(NormalSubcategoryItemGroup, 2, http.StatusBadRequest, "transaction item group is in use and cannot be deleted")
)
