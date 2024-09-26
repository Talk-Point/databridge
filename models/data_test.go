package models

import (
	"testing"
)

func TestLoadModel(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid model",
			data: map[string]interface{}{
				"columns": []interface{}{
					map[interface{}]interface{}{
						"name": "id",
						"type": "bigint",
					},
					map[interface{}]interface{}{
						"name": "name",
						"type": "string",
					},
				},
				"unique_key": []interface{}{
					"id",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid column type",
			data: map[string]interface{}{
				"columns": []interface{}{
					map[interface{}]interface{}{
						"name": "id",
						"type": "bigstr", // Invalid type
					},
				},
				"unique_key": []interface{}{
					"id",
				},
			},
			wantErr: true,
		},
		{
			name: "missing column name",
			data: map[string]interface{}{
				"columns": []interface{}{
					map[interface{}]interface{}{
						"type": "string",
					},
				},
				"unique_key": []interface{}{
					"id",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid unique key format",
			data: map[string]interface{}{
				"columns": []interface{}{
					map[interface{}]interface{}{
						"name": "id",
						"type": "bigint",
					},
				},
				"unique_key": []interface{}{
					123, // Invalid format for unique key
				},
			},
			wantErr: true,
		},
		{
			name: "no columns provided",
			data: map[string]interface{}{
				"columns":    []interface{}{}, // Empty columns
				"unique_key": []interface{}{"id"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadModel(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadModel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
