package sql_filter

import (
	"fmt"
	"intmax2-node/internal/sql_filter/models"
	"strings"

	"github.com/rs/xid"
)

type SQLFilter struct{}

func (c *SQLFilter) FilterDataToWhereQuery(
	filterData models.FiltersList,
) (query string, params map[string]interface{}) {
	const (
		int0Key  = 0
		emptyKey = ""
	)

	params = make(map[string]interface{})
	for i := range filterData {
		filterQuery, filterParams := c.filterToQueryWithParams(filterData[i])

		for key := range filterParams {
			params[key] = filterParams[key]
		}

		if i > int0Key {
			switch filterData[i].Relation {
			case models.RelationAnd:
				query += ") and ("
			case models.RelationOr:
				query += " or "
			}
		}

		query += filterQuery
	}

	if query != emptyKey {
		query = fmt.Sprintf("(%s)", query)
	}

	return query, params
}

func (c *SQLFilter) dataFieldToColumn(dataField models.DataField) (column string) {
	switch dataField {
	case models.DataFieldBlockNumber:
		column = "block_number"
	}

	return column
}

func (c *SQLFilter) dataFieldToValue(dataField models.DataField, input string) (value string) {
	value = input

	return value
}

func (c *SQLFilter) conditionToExpression(
	condition models.Condition,
	value string,
) (cond string, val []interface{}) {
	switch condition {
	case models.ConditionContains:
		cond = "ilike"
		const percent = "%"
		val = []interface{}{percent + value + percent}
	case models.ConditionDoesNotContain:
		cond = "not ilike"
		const percent = "%"
		val = []interface{}{percent + value + percent}
	case models.ConditionIs:
		cond = "="
		val = []interface{}{value}
	case models.ConditionIsNot:
		cond = "!="
		val = []interface{}{value}
	case models.ConditionIsTrue:
		cond = "="
		val = []interface{}{true}
	case models.ConditionIsFalse:
		cond = "="
		val = []interface{}{false}
	case models.ConditionIsEmpty:
		cond = "is null"
	case models.ConditionIsNotEmpty:
		cond = "is not null"
	case models.ConditionGreaterThan:
		cond = ">"
		val = []interface{}{value}
	case models.ConditionLessThan:
		cond = "<"
		val = []interface{}{value}
	case models.ConditionGreaterThanOrEqualTo:
		cond = ">="
		val = []interface{}{value}
	case models.ConditionLessThanOrEqualTo:
		cond = "<="
		val = []interface{}{value}
	}

	return cond, val
}

func (c *SQLFilter) generatePlaceholderWithKey() (value, placeholder string) {
	placeholder = xid.New().String()
	value = fmt.Sprintf("@%s", placeholder)
	return value, placeholder
}

// filterToQueryWithParams convert filters.Filter to sql where string
func (c *SQLFilter) filterToQueryWithParams(
	filter *models.Filter,
) (query string, params map[string]interface{}) {
	params = make(map[string]interface{})
	column := c.dataFieldToColumn(filter.DataField)
	value := c.dataFieldToValue(filter.DataField, filter.Value)
	expression, newArguments := c.conditionToExpression(filter.Condition, value)
	placeholder, placeholderKey := c.generatePlaceholderWithKey()

	switch filter.DataField {
	case
		models.DataFieldBlockNumber:
		const mask = "(%s %s %s)"
		query = fmt.Sprintf(mask, column, expression, placeholder)
	}

	const (
		int0 = 0
		int1 = 1
	)
	if len(newArguments) == int1 {
		params[placeholderKey] = newArguments[int0]
	} else {
		params[placeholderKey] = newArguments
	}

	query = strings.TrimSpace(query)

	return query, params
}

func (c *SQLFilter) PrepareWhereString(rowWhere string, needWhere bool) (where string) {
	if rowWhere == "" {
		return
	}

	rowWhere = strings.Trim(rowWhere, " ")
	rowWhere = strings.TrimLeft(rowWhere, "and")
	rowWhere = strings.TrimLeft(rowWhere, "or")

	where = " (" + rowWhere + ") " // nolint:goconst
	if needWhere {
		where = " where (" + where + ") " // nolint:goconst
	}

	return where
}
