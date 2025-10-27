package sqlX

import (
	"database/sql/driver"
	"encoding/json"
	"reflect"
	"testing"
)

// 测试用的结构体
type TestConfig struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func TestJsonColumn_Value(t *testing.T) {
	tests := []struct {
		name    string
		input   JsonColumn[TestConfig]
		want    driver.Value
		wantErr bool
	}{
		{
			name: "valid data",
			input: JsonColumn[TestConfig]{
				Val:   TestConfig{Name: "test", Count: 42},
				Valid: true,
			},
			want:    []byte(`{"name":"test","count":42}`),
			wantErr: false,
		},
		{
			name: "invalid (nil)",
			input: JsonColumn[TestConfig]{
				Valid: false,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "valid but zero value",
			input: JsonColumn[TestConfig]{
				Val:   TestConfig{}, // zero value
				Valid: true,
			},
			want:    []byte(`{"name":"","count":0}`),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.Value()
			if (err != nil) != tt.wantErr {
				t.Errorf("Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Value() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestJsonColumn_Scan(t *testing.T) {
	tests := []struct {
		name      string
		src       interface{}
		wantVal   TestConfig
		wantValid bool
		wantErr   bool
	}{
		{
			name:      "nil src",
			src:       nil,
			wantVal:   TestConfig{}, // zero value
			wantValid: false,
			wantErr:   false,
		},
		{
			name: "[]byte valid JSON",
			src:  []byte(`{"name":"alice","count":100}`),
			wantVal: TestConfig{
				Name:  "alice",
				Count: 100,
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "string valid JSON",
			src:  `{"name":"bob","count":200}`,
			wantVal: TestConfig{
				Name:  "bob",
				Count: 200,
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name:      "empty object",
			src:       []byte(`{}`),
			wantVal:   TestConfig{},
			wantValid: true,
			wantErr:   false,
		},
		{
			name:      "invalid JSON",
			src:       []byte(`{invalid}`),
			wantVal:   TestConfig{},
			wantValid: false, // 注意：Scan 失败时，j.Val 可能未修改，但测试中我们只关心 error
			wantErr:   true,
		},
		{
			name:      "unsupported type",
			src:       123,
			wantVal:   TestConfig{},
			wantValid: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var jc JsonColumn[TestConfig]
			err := jc.Scan(tt.src)

			if (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return // 不校验值
			}

			if !reflect.DeepEqual(jc.Val, tt.wantVal) {
				t.Errorf("Scan() jc.Val = %+v, want %+v", jc.Val, tt.wantVal)
			}
			if jc.Valid != tt.wantValid {
				t.Errorf("Scan() jc.Valid = %v, want %v", jc.Valid, tt.wantValid)
			}
		})
	}
}

// 测试 Scan(nil) 后 Val 是否被重置为零值
func TestJsonColumn_ScanNilResetsValue(t *testing.T) {
	jc := JsonColumn[TestConfig]{
		Val:   TestConfig{Name: "old", Count: 999},
		Valid: true,
	}

	err := jc.Scan(nil)
	if err != nil {
		t.Fatalf("Scan(nil) failed: %v", err)
	}

	if jc.Valid {
		t.Error("Expected Valid = false after Scan(nil)")
	}
	if jc.Val != (TestConfig{}) {
		t.Errorf("Expected zero value after Scan(nil), got %+v", jc.Val)
	}
}

// 测试 Marshal/Unmarshal 循环一致性
func TestJsonColumn_RoundTrip(t *testing.T) {
	original := JsonColumn[TestConfig]{
		Val: TestConfig{
			Name:  "roundtrip",
			Count: 42,
		},
		Valid: true,
	}

	// Value → []byte
	data, err := original.Value()
	if err != nil {
		t.Fatalf("Value() failed: %v", err)
	}

	// Scan back
	var restored JsonColumn[TestConfig]
	err = restored.Scan(data)
	if err != nil {
		t.Fatalf("Scan() failed: %v", err)
	}

	if !reflect.DeepEqual(original, restored) {
		t.Errorf("Round-trip failed: original=%+v, restored=%+v", original, restored)
	}
}

// 测试嵌入到 struct 中的 JSON 序列化（可选，验证与 json.Marshal 兼容）
func TestJsonColumn_JSONMarshal(t *testing.T) {
	type Wrapper struct {
		Config JsonColumn[TestConfig] `json:"config"`
	}

	w := Wrapper{
		Config: JsonColumn[TestConfig]{
			Val:   TestConfig{Name: "api", Count: 1},
			Valid: true,
		},
	}

	b, err := json.Marshal(w)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	expected := `{"config":{"name":"api","count":1}}`
	if string(b) != expected {
		t.Errorf("json.Marshal = %s, want %s", b, expected)
	}
}
