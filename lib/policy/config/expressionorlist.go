package config

import (
	"encoding/json"
	"errors"
	"slices"
)

var (
	ErrExpressionOrListMustBeStringOrObject = errors.New("config: this must be a string or an object")
	ErrExpressionEmpty                      = errors.New("config: this expression is empty")
	ErrExpressionCantHaveBoth               = errors.New("config: expression block can't contain multiple expression types")
)

type ExpressionOrList struct {
	Expression string   `json:"-"`
	And        []string `json:"and"`
	Or         []string `json:"or"`
}

func (eol ExpressionOrList) Equal(rhs *ExpressionOrList) bool {
	if eol.Expression != rhs.Expression {
		return false
	}

	if !slices.Equal(eol.And, rhs.And) {
		return false
	}

	if !slices.Equal(eol.Or, rhs.Or) {
		return false
	}

	return true
}

func (eol *ExpressionOrList) UnmarshalJSON(data []byte) error {
	switch string(data[0]) {
	case `"`: // string
		return json.Unmarshal(data, &eol.Expression)
	case "{": // object
		type RawExpressionOrList ExpressionOrList
		var val RawExpressionOrList
		if err := json.Unmarshal(data, &val); err != nil {
			return err
		}
		eol.And = val.And
		eol.Or = val.Or

		return nil
	}

	return ErrExpressionOrListMustBeStringOrObject
}

func (eol *ExpressionOrList) Valid() error {
	if len(eol.And) != 0 && len(eol.Or) != 0 {
		return ErrExpressionCantHaveBoth
	}

	return nil
}
