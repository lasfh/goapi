package validate

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/lasfh/goapi/http/response"
)

type customValidatorIsValid string

func (c customValidatorIsValid) IsValid() bool {
	switch c {
	case "1", "2", "3":
		return true
	}

	return false
}

func TestFields(t *testing.T) {
	type TestStruct struct {
		Name   string                 `json:"name" validate:"required" label:"nome"`
		Email  string                 `json:"email" validate:"required,email"`
		Age    int                    `json:"age,omitempty" validate:"gte=18" label:"idade"`
		Status customValidatorIsValid `json:"status" validate:"is_valid"`
		Phone  string                 `json:"phone" validate:"omitempty,min=10,max=15" label:"telefone"`
		Score  int                    `json:"score" validate:"omitempty,lt=100"`
		Role   string                 `json:"role" validate:"required,oneof=admin user guest" label:"perfil"`
	}

	base := &TestStruct{
		Name:   "John Doe",
		Email:  "teste@mpmg.mp.br",
		Age:    25,
		Status: "1",
		Role:   "admin",
	}

	tests := []struct {
		name               string
		input              *TestStruct
		expectErr          bool
		expectedStatusCode int
		expectedMsg        string
		expectedDetails    map[string]string
	}{
		{
			name:      "Estrutura válida",
			input:     base,
			expectErr: false,
		},
		{
			name:               "Estrutura nil (erro interno)",
			input:              nil,
			expectErr:          true,
			expectedStatusCode: http.StatusInternalServerError,
			expectedMsg:        "Não foi possível validar os dados.",
		},
		{
			name:               "Nome ausente",
			input:              &TestStruct{Email: base.Email, Age: base.Age, Status: base.Status, Role: base.Role},
			expectErr:          true,
			expectedStatusCode: http.StatusBadRequest,
			expectedMsg:        "O campo (nome) é inválido. Verifique o valor informado.",
			expectedDetails: map[string]string{
				"name": "nome é um campo obrigatório",
			},
		},
		{
			name:      "E-mail inválido",
			input:     &TestStruct{Name: base.Name, Email: "mpmg.mp.br", Age: base.Age, Status: base.Status, Role: base.Role},
			expectErr: true,
			expectedDetails: map[string]string{
				"email": "email deve ser um endereço de e-mail válido",
			},
		},
		{
			name:      "Idade abaixo de 18 anos",
			input:     &TestStruct{Name: base.Name, Email: base.Email, Age: 17, Status: base.Status, Role: base.Role},
			expectErr: true,
			expectedDetails: map[string]string{
				"age": "idade deve ser 18 ou superior",
			},
		},
		{
			name:      "Status inválido (is_valid)",
			input:     &TestStruct{Name: base.Name, Email: base.Email, Age: base.Age, Status: "10", Role: base.Role},
			expectErr: true,
			expectedDetails: map[string]string{
				"status": "status não é válido",
			},
		},
		{
			name:      "Telefone muito curto (min=10)",
			input:     &TestStruct{Name: base.Name, Email: base.Email, Age: base.Age, Status: base.Status, Role: base.Role, Phone: "123"},
			expectErr: true,
			expectedDetails: map[string]string{
				"phone": "telefone deve ter pelo menos 10 caracteres",
			},
		},
		{
			name:      "Telefone muito longo (max=15)",
			input:     &TestStruct{Name: base.Name, Email: base.Email, Age: base.Age, Status: base.Status, Role: base.Role, Phone: "1234567890123456"},
			expectErr: true,
			expectedDetails: map[string]string{
				"phone": "telefone deve ter no máximo 15 caracteres",
			},
		},
		{
			name:      "Pontuação maior ou igual a 100 (lt=100)",
			input:     &TestStruct{Name: base.Name, Email: base.Email, Age: base.Age, Status: base.Status, Role: base.Role, Score: 100},
			expectErr: true,
			expectedDetails: map[string]string{
				"score": "score deve ser menor que 100",
			},
		},
		{
			name:      "Perfil inválido (oneof)",
			input:     &TestStruct{Name: base.Name, Email: base.Email, Age: base.Age, Status: base.Status, Role: "superadmin"},
			expectErr: true,
			expectedDetails: map[string]string{
				"role": "perfil deve ser um de [admin user guest]",
			},
		},
		{
			name:               "Múltiplos campos inválidos",
			input:              &TestStruct{Age: 17, Status: "10"},
			expectErr:          true,
			expectedStatusCode: http.StatusBadRequest,
			expectedMsg:        "Os campos (nome, email, idade, status, perfil) são inválidos. Verifique os valores informados.",
			expectedDetails: map[string]string{
				"name":   "nome é um campo obrigatório",
				"email":  "email é um campo obrigatório",
				"age":    "idade deve ser 18 ou superior",
				"status": "status não é válido",
				"role":   "perfil é um campo obrigatório",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Struct(context.TODO(), tt.input)

			if !tt.expectErr {
				if err != nil {
					t.Fatalf("Fields() retornou erro inesperado: %v", err)
				}

				return
			}

			if err == nil {
				t.Fatal("Fields() deveria retornar erro, mas retornou nil")
			}

			var ve response.ResponseError
			if !errors.As(err, &ve) {
				t.Fatalf("erro deveria ser response.ResponseError, mas é %T", err)
			}

			if tt.expectedStatusCode != 0 && ve.StatusCode() != tt.expectedStatusCode {
				t.Errorf("StatusCode() = %d, esperado %d", ve.StatusCode(), tt.expectedStatusCode)
			}

			if tt.expectedMsg != "" && ve.Message() != tt.expectedMsg {
				t.Errorf("Message() = %q, esperado %q", ve.Message(), tt.expectedMsg)
			}

			details := ve.Details()
			for key, expectedMsg := range tt.expectedDetails {
				if details[key] != expectedMsg {
					t.Errorf("details[%q] = %q, esperado %q", key, details[key], expectedMsg)
				}
			}
		})
	}
}

func TestFieldsWithoutJSONTag(t *testing.T) {
	type TestStruct struct {
		Name string `validate:"required"`
	}

	err := Struct(context.TODO(), &TestStruct{})
	if err == nil {
		t.Fatal("Fields() deveria retornar erro, mas retornou nil")
	}

	var ve *validateError
	if !errors.As(err, &ve) {
		t.Fatalf("erro deveria ser *validateError, mas é %T", err)
	}

	expectedDetail := "Name é um campo obrigatório"
	if ve.Details()["Name"] != expectedDetail {
		t.Errorf("details[\"Name\"] = %q, esperado %q", ve.Details()["Name"], expectedDetail)
	}

	expectedMsg := "O campo (Name) é inválido. Verifique o valor informado."
	if ve.Message() != expectedMsg {
		t.Errorf("Message() = %q, esperado %q", ve.Message(), expectedMsg)
	}
}

func TestIsValidWithUnsupportedType(t *testing.T) {
	type TestStruct struct {
		Status string `json:"status" validate:"is_valid"`
	}

	err := Struct(context.TODO(), &TestStruct{Status: "1"})
	if err == nil {
		t.Fatal("Fields() deveria retornar erro para tipo sem IsValid(), mas retornou nil")
	}

	var ve *validateError
	if !errors.As(err, &ve) {
		t.Fatalf("erro deveria ser *validateError, mas é %T", err)
	}

	expectedDetail := "status não é válido"
	if ve.Details()["status"] != expectedDetail {
		t.Errorf("details[\"status\"] = %q, esperado %q", ve.Details()["status"], expectedDetail)
	}
}
