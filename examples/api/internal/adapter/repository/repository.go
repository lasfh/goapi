package repository

// type filterByPagingOptions struct {
// 	Select      string
// 	CountSelect string
// }

// func filterByPaging[T any](
// 	query *gorm.DB,
// 	paging pagination.Pagination,
// 	filter queryfilter.FieldsAndSearch,
// 	options ...filterByPagingOptions,
// ) (*paginated.Paginated[T], error) {
// 	queryStr, args := filter.ToSQLQuery()
// 	if queryStr != "" {
// 		query = query.Where(queryStr, args...)
// 	}

// 	if len(options) > 0 {
// 		query = query.Select(options[0].CountSelect)
// 	}

// 	var total int64
// 	if err := query.Count(&total).Error; err != nil {
// 		return nil, err
// 	}

// 	if len(options) > 0 {
// 		query = query.Select(options[0].Select)
// 	}

// 	if paging.Sort != "" {
// 		query = query.Order(clause.OrderByColumn{
// 			Column: clause.Column{
// 				Name: paging.Sort,
// 			},
// 			Desc: paging.Order == pagination.DescendingOrder,
// 		})
// 	}

// 	size := paging.Limit
// 	if size <= 0 {
// 		size = 50
// 	}

// 	results := make([]T, 0, size)

// 	if err := query.
// 		Offset(int(paging.Skip)).
// 		Limit(int(paging.Limit)).
// 		Find(&results).Error; err != nil {
// 		return nil, err
// 	}

// 	return &paginated.Paginated[T]{
// 		Results:     results,
// 		Offset:      uint64(paging.Skip),
// 		Limit:       paging.Limit,
// 		TotalRecord: total,
// 	}, nil
// }
