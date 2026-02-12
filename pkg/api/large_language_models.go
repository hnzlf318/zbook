package api

import (
	"io"
	"strings"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/ocr"
	"github.com/mayswind/ezbookkeeping/pkg/services"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

// LargeLanguageModelsApi represents large language models api
type LargeLanguageModelsApi struct {
	ApiUsingConfig
	transactionCategories *services.TransactionCategoryService
	transactionTags       *services.TransactionTagService
	transactionItems      *services.TransactionItemService
	accounts              *services.AccountService
	users                 *services.UserService
}

// Initialize a large language models api singleton instance
var (
	LargeLanguageModels = &LargeLanguageModelsApi{
		ApiUsingConfig: ApiUsingConfig{
			container: settings.Container,
		},
		transactionCategories: services.TransactionCategories,
		transactionTags:       services.TransactionTags,
		transactionItems:      services.TransactionItems,
		accounts:              services.Accounts,
		users:                 services.Users,
	}
)

// RecognizeReceiptImageByOCRHandler returns recognized transactions from bill list screenshot using an external OCR service.
func (a *LargeLanguageModelsApi) RecognizeReceiptImageByOCRHandler(c *core.WebContext) (any, *errs.Error) {
	if !a.CurrentConfig().TransactionFromOCRImageRecognition {
		return nil, errs.ErrLargeLanguageModelProviderNotEnabled
	}

	clientTimezone, err := c.GetClientTimezone()
	if err != nil {
		log.Warnf(c, "[large_language_models.RecognizeReceiptImageByOCRHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)
	if err != nil {
		if !errs.IsCustomError(err) {
			log.Warnf(c, "[large_language_models.RecognizeReceiptImageByOCRHandler] failed to get user for user \"uid:%d\", because %s", uid, err.Error())
		}
		return nil, errs.ErrUserNotFound
	}

	if user.FeatureRestriction.Contains(core.USER_FEATURE_RESTRICTION_TYPE_CREATE_TRANSACTION_FROM_AI_IMAGE_RECOGNITION) {
		return nil, errs.ErrNotPermittedToPerformThisAction
	}

	form, err := c.MultipartForm()
	if err != nil {
		log.Errorf(c, "[large_language_models.RecognizeReceiptImageByOCRHandler] failed to get multi-part form data for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.ErrParameterInvalid
	}

	imageFiles := form.File["image"]
	if len(imageFiles) < 1 {
		log.Warnf(c, "[large_language_models.RecognizeReceiptImageByOCRHandler] there is no image in request for user \"uid:%d\"", uid)
		return nil, errs.ErrNoAIRecognitionImage
	}
	if imageFiles[0].Size < 1 {
		log.Warnf(c, "[large_language_models.RecognizeReceiptImageByOCRHandler] the size of image in request is zero for user \"uid:%d\"", uid)
		return nil, errs.ErrAIRecognitionImageIsEmpty
	}
	if imageFiles[0].Size > int64(a.CurrentConfig().MaxAIRecognitionPictureFileSize) {
		log.Warnf(c, "[large_language_models.RecognizeReceiptImageByOCRHandler] the upload file size \"%d\" exceeds the maximum size \"%d\" of image for user \"uid:%d\"", imageFiles[0].Size, a.CurrentConfig().MaxAIRecognitionPictureFileSize, uid)
		return nil, errs.ErrExceedMaxAIRecognitionImageFileSize
	}

	fileExtension := utils.GetFileNameExtension(imageFiles[0].Filename)
	contentType := utils.GetImageContentType(fileExtension)
	if contentType == "" {
		log.Warnf(c, "[large_language_models.RecognizeReceiptImageByOCRHandler] the file extension \"%s\" of image in request is not supported for user \"uid:%d\"", fileExtension, uid)
		return nil, errs.ErrImageTypeNotSupported
	}

	imageFile, err := imageFiles[0].Open()
	if err != nil {
		log.Errorf(c, "[large_language_models.RecognizeReceiptImageByOCRHandler] failed to get image file from request for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.ErrOperationFailed
	}
	defer imageFile.Close()

	imageData, err := io.ReadAll(imageFile)
	if err != nil {
		log.Errorf(c, "[large_language_models.RecognizeReceiptImageByOCRHandler] failed to read image file from request for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.ErrOperationFailed
	}

	if a.CurrentConfig().PaddleBillOCREndpoint == "" {
		log.Errorf(c, "[large_language_models.RecognizeReceiptImageByOCRHandler] paddle ocr endpoint is not configured for user \"uid:%d\"", uid)
		return nil, errs.ErrOperationFailed
	}

	rawItems, err := ocr.RunPaddleBillOCR(imageData, a.CurrentConfig().PaddleBillOCREndpoint)
	if err != nil {
		log.Warnf(c, "[large_language_models.RecognizeReceiptImageByOCRHandler] OCR failed for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.ErrOperationFailed
	}

	if len(rawItems) == 0 {
		return nil, errs.ErrNoTransactionInformationInImage
	}

	refTime := time.Now().In(clientTimezone)

	parsedList := make([]*models.RecognizedReceiptImageResult, 0, len(rawItems))
	for _, raw := range rawItems {
		// 基础清洗
		amountStr := strings.TrimSpace(raw.Amount)
		if amountStr == "" {
			// 没有金额就无法生成交易，跳过
			continue
		}

		result := &models.RecognizedReceiptImageResult{}

		// 金额 & 收支类型
		result.Amount = amountStr
		if amountStr[0] == '-' {
			result.Type = "expense"
		} else {
			result.Type = "income"
		}

		// 时间：优先用返回的 date；为空则用当前时间
		dateStr := strings.TrimSpace(raw.Date)
		if dateStr != "" {
			result.Time = dateStr
		} else {
			result.Time = refTime.Format("2006-01-02 15:04:05")
		}

		// 账户名称
		result.AccountName = strings.TrimSpace(raw.Account)

		// 分类名称：从 classify 中取最后一段，例如 "食品饮料-食品" => "食品"
		classify := strings.TrimSpace(raw.Classify)
		if classify != "" {
			parts := strings.Split(classify, "-")
			name := strings.TrimSpace(parts[len(parts)-1])
			result.CategoryName = name
		}

		// 标签：目前只支持单个 label
		label := strings.TrimSpace(raw.Label)
		if label != "" {
			result.TagNames = []string{label}
		}

		// 描述(comment)：使用 text 映射为内部 comment
		text := strings.TrimSpace(raw.Text)
		result.Description = text

		// 项目(transaction item)：使用 project 映射为内部交易项目
		project := strings.TrimSpace(raw.Project)
		if project != "" {
			result.ItemNames = []string{project}
		}

		parsedList = append(parsedList, result)
	}

	if len(parsedList) == 0 {
		return nil, errs.ErrNoTransactionInformationInImage
	}

	accounts, err := a.accounts.GetAllAccountsByUid(c, uid)
	if err != nil {
		log.Errorf(c, "[large_language_models.RecognizeReceiptImageByOCRHandler] failed to get all accounts for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}
	accountMap := a.accounts.GetVisibleAccountNameMapByList(accounts)

	categories, err := a.transactionCategories.GetAllCategoriesByUid(c, uid, 0, -1)
	if err != nil {
		log.Errorf(c, "[large_language_models.RecognizeReceiptImageByOCRHandler] failed to get categories for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}
	expenseCategoryMap := make(map[string]*models.TransactionCategory)
	incomeCategoryMap := make(map[string]*models.TransactionCategory)
	transferCategoryMap := make(map[string]*models.TransactionCategory)
	for i := 0; i < len(categories); i++ {
		cat := categories[i]
		if cat.Hidden || cat.ParentCategoryId == models.LevelOneTransactionCategoryParentId {
			continue
		}
		if cat.Type == models.CATEGORY_TYPE_EXPENSE {
			expenseCategoryMap[cat.Name] = cat
		} else if cat.Type == models.CATEGORY_TYPE_INCOME {
			incomeCategoryMap[cat.Name] = cat
		} else if cat.Type == models.CATEGORY_TYPE_TRANSFER {
			transferCategoryMap[cat.Name] = cat
		}
	}

	tags, err := a.transactionTags.GetAllTagsByUid(c, uid)
	if err != nil {
		log.Errorf(c, "[large_language_models.RecognizeReceiptImageByOCRHandler] failed to get tags for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}
	tagMap := a.transactionTags.GetVisibleTagNameMapByList(tags)

	items, err := a.transactionItems.GetAllItemsByUid(c, uid)
	if err != nil {
		log.Errorf(c, "[large_language_models.RecognizeReceiptImageByOCRHandler] failed to get transaction items for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}
	itemNameMap := make(map[string]*models.TransactionItem)
	for _, item := range items {
		if item != nil && !item.Hidden && item.Name != "" {
			if _, exists := itemNameMap[item.Name]; !exists {
				itemNameMap[item.Name] = item
			}
		}
	}

	transactions := make([]models.RecognizedReceiptImageResponse, 0, len(parsedList))
	for _, one := range parsedList {
		resp, parseErr := a.parseRecognizedReceiptImageResponse(c, uid, clientTimezone, one, accountMap, expenseCategoryMap, incomeCategoryMap, transferCategoryMap, tagMap, itemNameMap)
		if parseErr != nil {
			continue
		}
		transactions = append(transactions, *resp)
	}
	if len(transactions) == 0 {
		return nil, errs.ErrNoTransactionInformationInImage
	}

	return &models.RecognizedReceiptImageListResponse{Transactions: transactions}, nil
}

func (a *LargeLanguageModelsApi) parseRecognizedReceiptImageResponse(c *core.WebContext, uid int64, clientTimezone *time.Location, recognizedResult *models.RecognizedReceiptImageResult, accountMap map[string]*models.Account, expenseCategoryMap map[string]*models.TransactionCategory, incomeCategoryMap map[string]*models.TransactionCategory, transferCategoryMap map[string]*models.TransactionCategory, tagMap map[string]*models.TransactionTag, itemNameMap map[string]*models.TransactionItem) (*models.RecognizedReceiptImageResponse, *errs.Error) {
	recognizedReceiptImageResponse := &models.RecognizedReceiptImageResponse{
		Type: models.TRANSACTION_TYPE_EXPENSE,
	}

	if recognizedResult == nil {
		log.Errorf(c, "[large_language_models.parseRecognizedReceiptImageResponse] recoginzed result is null")
		return nil, errs.ErrNoTransactionInformationInImage
	}

	if recognizedResult.Type == "income" {
		recognizedReceiptImageResponse.Type = models.TRANSACTION_TYPE_INCOME

		if len(recognizedResult.CategoryName) > 0 {
			category, exists := incomeCategoryMap[recognizedResult.CategoryName]

			if exists {
				recognizedReceiptImageResponse.CategoryId = category.CategoryId
			}
		}
	} else if recognizedResult.Type == "expense" {
		recognizedReceiptImageResponse.Type = models.TRANSACTION_TYPE_EXPENSE

		if len(recognizedResult.CategoryName) > 0 {
			category, exists := expenseCategoryMap[recognizedResult.CategoryName]

			if exists {
				recognizedReceiptImageResponse.CategoryId = category.CategoryId
			}
		}
	} else if recognizedResult.Type == "transfer" {
		recognizedReceiptImageResponse.Type = models.TRANSACTION_TYPE_TRANSFER

		if len(recognizedResult.CategoryName) > 0 {
			category, exists := transferCategoryMap[recognizedResult.CategoryName]

			if exists {
				recognizedReceiptImageResponse.CategoryId = category.CategoryId
			}
		}
	} else if len(recognizedResult.Type) == 0 {
		return nil, errs.ErrNoTransactionInformationInImage
	} else {
		log.Errorf(c, "[large_language_models.parseRecognizedReceiptImageResponse] recoginzed transaction type \"%s\" is invalid", recognizedResult.Type)
		return nil, errs.ErrOperationFailed
	}

	if len(recognizedResult.Time) > 0 {
		longDateTime := a.getLongDateTime(recognizedResult.Time)
		timestamp, err := utils.ParseFromLongDateTimeInTimeZone(longDateTime, clientTimezone)

		if err != nil {
			log.Warnf(c, "[large_language_models.parseRecognizedReceiptImageResponse] recoginzed time \"%s\" is invalid", recognizedResult.Time)
		} else {
			recognizedReceiptImageResponse.Time = timestamp.Unix()
		}
	}

	if len(recognizedResult.Amount) > 0 {
		amount, err := utils.ParseAmount(recognizedResult.Amount)

		if err != nil {
			log.Errorf(c, "[large_language_models.parseRecognizedReceiptImageResponse] recoginzed amount \"%s\" is invalid", recognizedResult.Amount)
			return nil, errs.ErrOperationFailed
		}

		recognizedReceiptImageResponse.SourceAmount = amount

		if recognizedReceiptImageResponse.Type == models.TRANSACTION_TYPE_TRANSFER && len(recognizedResult.DestinationAmount) > 0 {
			destinationAmount, err := utils.ParseAmount(recognizedResult.DestinationAmount)

			if err != nil {
				log.Errorf(c, "[large_language_models.parseRecognizedReceiptImageResponse] recoginzed destination amount \"%s\" is invalid", recognizedResult.DestinationAmount)
				return nil, errs.ErrOperationFailed
			}

			recognizedReceiptImageResponse.DestinationAmount = destinationAmount
		}
	}

	if len(recognizedResult.AccountName) > 0 {
		account, exists := accountMap[recognizedResult.AccountName]

		if exists {
			recognizedReceiptImageResponse.SourceAccountId = account.AccountId
		}
	}

	if len(recognizedResult.DestinationAccountName) > 0 {
		account, exists := accountMap[recognizedResult.DestinationAccountName]

		if exists {
			recognizedReceiptImageResponse.DestinationAccountId = account.AccountId
		}
	}

	if len(recognizedResult.TagNames) > 0 {
		tagIds := make([]string, 0, len(recognizedResult.TagNames))

		for i := 0; i < len(recognizedResult.TagNames); i++ {
			tagName := recognizedResult.TagNames[i]
			tag, exists := tagMap[tagName]

			if exists {
				tagIds = append(tagIds, utils.Int64ToString(tag.TagId))
			}
		}

		recognizedReceiptImageResponse.TagIds = tagIds
	}

	if len(recognizedResult.Description) > 0 {
		recognizedReceiptImageResponse.Comment = recognizedResult.Description
	}

	if len(recognizedResult.ItemNames) > 0 {
		itemIds := make([]string, 0, len(recognizedResult.ItemNames))
		for _, name := range recognizedResult.ItemNames {
			if item, exists := itemNameMap[name]; exists {
				itemIds = append(itemIds, utils.Int64ToString(item.ItemId))
			}
		}
		recognizedReceiptImageResponse.ItemIds = itemIds
	}

	return recognizedReceiptImageResponse, nil
}

func (a *LargeLanguageModelsApi) getLongDateTime(dateTime string) string {
	if utils.IsValidLongDateTimeFormat(dateTime) {
		return dateTime
	}

	if utils.IsValidLongDateTimeWithoutSecondFormat(dateTime) {
		return dateTime + ":00"
	}

	if utils.IsValidLongDateFormat(dateTime) {
		return dateTime + " 00:00:00"
	}

	return dateTime
}
