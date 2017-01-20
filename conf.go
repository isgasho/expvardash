package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"io/ioutil"
	"fmt"
)

var Template = template.Must(template.ParseFiles("templates/index.html"))

func LoadConf(path string) (*Config, error) {
	raw, err := ReadConf(path)
	if err != nil {
		return nil, err
	}

	return raw.ParseConf()
}

func ReadConf(path string) (*RawConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var conf RawConfig
	err = json.Unmarshal(data, &conf)
	if err != nil {
		return nil, err
	}

	return &conf, nil
}

type RawConfig struct {
	Rows []RawRow `json:"rows"`
}

type RawRow struct {
	Items []RawItem `json:"items"`
}

type RawItem struct {
	Type string           `json:"type"`
	Size int              `json:"size"`
	Conf *json.RawMessage `json:"conf"`
}

type Config struct {
	Layout *Layout
	Charts *Charts
}

type Layout struct {
	Rows []*Row
}

type Row struct {
	Cols []*Col
}

type Col struct {
	ID   string
	Size int
}

func (c *RawConfig) ParseConf() (*Config, error) {
	config := &Config{
		Layout: &Layout{
			Rows: []*Row{},
		},
		Charts: &Charts{
			LineCharts: []*LineChart{},
		},
	}

	for _, row := range c.Rows {
		cols := []*Col{}

		for _, item := range row.Items {
			c, err := ReadChart(item)
			if err != nil {
				return nil, err
			}

			c.SetID(config.Charts.NextID())

			err = config.Charts.Append(c)
			if err != nil {
				return nil, err
			}

			cols = append(cols, &Col{
				ID: c.ID(),
				Size: item.Size,
			})
		}

		config.Layout.Rows = append(config.Layout.Rows, &Row{
			Cols: cols,
		})
	}

	return config, nil
}

func ReadChart(item RawItem) (Chart, error) {
	if item.Conf == nil {
		return nil, fmt.Errorf("Missing configuration for: %s", item.Type)
	}

	switch item.Type {
	case LineChartType:
		return ReadLineChart(item.Conf)
	default:
		return nil, fmt.Errorf("Unknown chart type: %s", item.Type)
	}
}

func ReadLineChart(data *json.RawMessage) (*LineChart, error) {
	var chart LineChart
	err := json.Unmarshal(*data, &chart)
	if err != nil {
		return nil, err
	}
	return &chart, nil
}

func (l *Layout) Render() (string, error) {
	buf := new(bytes.Buffer)

	err := l.RenderTo(buf)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (l *Layout) RenderTo(w io.Writer) error {
	return Template.Execute(w, *l)
}
