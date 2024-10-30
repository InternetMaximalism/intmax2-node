package models

type OrderBy string

const (
	LessSymbol = "<"
	MoreSymbol = ">"
)

const (
	DateCreate OrderBy = "date_create"
)

type Sorting string

const (
	SortingASC  Sorting = "asc"
	SortingDESC Sorting = "desc"
)

type Direction string

const (
	DirectionPrev Direction = "prev"
	DirectionNext Direction = "next"
)

type Relation string

const (
	RelationAnd Relation = "and"
	RelationOr  Relation = "or"
)

type DataField string

const (
	DataFieldBlockNumber     DataField = "block_number"
	DataFieldStartBackupTime DataField = "start_backup_time"
)

type Condition string

const (
	ConditionContains             Condition = "contains"
	ConditionDoesNotContain       Condition = "doesNotContain"
	ConditionIs                   Condition = "is"
	ConditionIsNot                Condition = "isNot"
	ConditionIsTrue               Condition = "isTrue"
	ConditionIsFalse              Condition = "isFalse"
	ConditionIsEmpty              Condition = "isEmpty"
	ConditionIsNotEmpty           Condition = "isNotEmpty"
	ConditionGreaterThan          Condition = "greaterThan"
	ConditionLessThan             Condition = "lessThan"
	ConditionGreaterThanOrEqualTo Condition = "greaterThanOrEqualTo"
	ConditionLessThanOrEqualTo    Condition = "lessThanOrEqualTo"
)

type Filter struct {
	Relation  Relation  `json:"relation"`
	DataField DataField `json:"data_field"`
	Condition Condition `json:"condition"`
	Value     string    `json:"value"`
}

type FiltersList []*Filter
