package errs

import "net/http"

// Error codes related to transaction items
var (
	ErrTransactionItemIdInvalid            = NewNormalError(NormalSubcategoryItem, 0, http.StatusBadRequest, "transaction item id is invalid")
	ErrTransactionItemNotFound             = NewNormalError(NormalSubcategoryItem, 1, http.StatusBadRequest, "transaction item not found")
	ErrTransactionItemNameIsEmpty          = NewNormalError(NormalSubcategoryItem, 2, http.StatusBadRequest, "transaction item name is empty")
	ErrTransactionItemNameAlreadyExists     = NewNormalError(NormalSubcategoryItem, 3, http.StatusBadRequest, "transaction item name already exists")
	ErrTransactionItemInUseCannotBeDeleted = NewNormalError(NormalSubcategoryItem, 4, http.StatusBadRequest, "transaction item is in use and cannot be deleted")
	ErrTransactionItemIndexNotFound        = NewNormalError(NormalSubcategoryItem, 5, http.StatusBadRequest, "transaction item index not found")
)
