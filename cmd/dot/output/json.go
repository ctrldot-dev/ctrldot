package output

import (
	"encoding/json"
	"io"

	"github.com/futurematic/kernel/internal/domain"
)

// JSONFormatter formats output as JSON
type JSONFormatter struct {
	w io.Writer
}

func (f *JSONFormatter) PrintPlan(plan interface{}) error {
	return json.NewEncoder(f.w).Encode(map[string]interface{}{"plan": plan})
}

func (f *JSONFormatter) PrintOperation(op interface{}) error {
	return json.NewEncoder(f.w).Encode(map[string]interface{}{"operation": op})
}

func (f *JSONFormatter) PrintExpand(result interface{}) error {
	return json.NewEncoder(f.w).Encode(map[string]interface{}{"result": result})
}

func (f *JSONFormatter) PrintHistory(ops []domain.Operation) error {
	return json.NewEncoder(f.w).Encode(map[string]interface{}{"result": ops})
}

func (f *JSONFormatter) PrintDiff(result interface{}) error {
	return json.NewEncoder(f.w).Encode(map[string]interface{}{"result": result})
}

func (f *JSONFormatter) PrintStatus(status interface{}) error {
	return json.NewEncoder(f.w).Encode(status)
}

func (f *JSONFormatter) PrintConfig(cfg interface{}) error {
	return json.NewEncoder(f.w).Encode(cfg)
}
