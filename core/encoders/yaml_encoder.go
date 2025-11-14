package encoders

import (
	"time"

	"github.com/fbz-tec/pgxport/core/formatters"
	"gopkg.in/yaml.v3"
)

type OrderedYamlEncoder struct {
	timeLayout string
	timezone   *time.Location
}

func NewOrderedYamlEncoder(timeFormat, timeZone string) OrderedYamlEncoder {
	layout, loc := formatters.UserTimeZoneFormat(timeFormat, timeZone)
	return OrderedYamlEncoder{
		timeLayout: layout,
		timezone:   loc,
	}
}

// EncodeRow builds a YAML mapping node (one record).
func (o OrderedYamlEncoder) EncodeRow(keys []string, values []interface{}) (*yaml.Node, error) {

	row := &yaml.Node{
		Kind: yaml.MappingNode,
	}

	for i, key := range keys {
		keyNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: key,
		}

		val := formatters.FormatYAMLValue(values[i], o.timeLayout, o.timezone)
		valueNode := &yaml.Node{}
		if err := valueNode.Encode(val); err != nil {
			return nil, err
		}

		row.Content = append(row.Content, keyNode, valueNode)
	}

	return row, nil
}
