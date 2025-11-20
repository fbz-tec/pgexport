package encoders

import (
	"github.com/fbz-tec/pgxport/core/formatters"
	"gopkg.in/yaml.v3"
)

type OrderedYamlEncoder struct {
	timeLayout string
	timezone   string
}

func NewOrderedYamlEncoder(timeFormat, timeZone string) OrderedYamlEncoder {
	return OrderedYamlEncoder{
		timeLayout: timeFormat,
		timezone:   timeZone,
	}
}

// EncodeRow builds a YAML mapping node (one record).
func (o OrderedYamlEncoder) EncodeRow(keys []string, dataTypes []uint32, values []interface{}) (*yaml.Node, error) {

	row := &yaml.Node{
		Kind: yaml.MappingNode,
	}

	for i, key := range keys {
		keyNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: key,
		}

		val := formatters.FormatYAMLValue(values[i], dataTypes[i], o.timeLayout, o.timezone)
		valueNode := &yaml.Node{}
		if err := valueNode.Encode(val); err != nil {
			return nil, err
		}

		row.Content = append(row.Content, keyNode, valueNode)
	}

	return row, nil
}
