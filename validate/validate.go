package validate

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/locales/pt_BR"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	pt_translations "github.com/go-playground/validator/v10/translations/pt_BR"
	"github.com/lasfh/goapi/logs/resperr"
)

var (
	validate   *validator.Validate
	translator ut.Translator
)

func init() {
	portuguese := pt_BR.New()
	universalTranslator := ut.New(portuguese, portuguese)

	translator, _ = universalTranslator.GetTranslator("pt-br")

	validate = validator.New(validator.WithRequiredStructEnabled())
	pt_translations.RegisterDefaultTranslations(validate, translator)

	// Registra função para extrair o nome do campo no formato "json|label".
	// No loop de validação, split por "|" separa a chave JSON do nome formatado.
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		jsonKey := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if jsonKey == "" {
			jsonKey = fld.Name
		}

		label := fld.Tag.Get("label")
		if label == "" {
			label = jsonKey
		}

		return "{" + jsonKey + "|" + label + "}"
	})

	RegisterValidation(
		"is_valid",
		func(fl validator.FieldLevel) bool {
			if value, ok := fl.Field().Interface().(interface{ IsValid() bool }); ok {
				return value.IsValid()
			}

			return false
		},
		func(ut ut.Translator) error {
			return ut.Add("is_valid", "{0} não é válido", true)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("is_valid", fe.Field())
			return t
		},
	)
}

// RegisterValidation registra uma nova validação personalizada com tradução.
//
// Parâmetros:
//   - tag (string): Nome da tag de validação.
//   - validatorFunc (validator.Func): Função de validação personalizada.
//   - registerFunc (validator.RegisterTranslationsFunc): Função que registra a tradução da validação.
//   - translationFunc (validator.TranslationFunc): Função que traduz a mensagem de erro da validação.
//   - callValidationEvenIfNull (...bool): Opcional, indica se a validação deve ser chamada mesmo se o campo for nulo.
func RegisterValidation(
	tag string,
	validatorFunc validator.Func,
	registerFunc validator.RegisterTranslationsFunc,
	translationFunc validator.TranslationFunc,
	callValidationEvenIfNull ...bool,
) {
	validate.RegisterValidation(
		tag,
		validatorFunc,
		callValidationEvenIfNull...,
	)
	validate.RegisterTranslation(
		tag,
		translator,
		registerFunc,
		translationFunc,
	)
}

// Struct valida uma estrutura e retorna um erro contendo os detalhes da validação, se houver falha.
//
// Parâmetros:
//   - s (any): Estrutura a ser validada.
//
// Retorna:
//   - error: Erro de validação com detalhes dos campos inválidos, se houver.
func Struct(ctx context.Context, s any) error {
	if err := validate.Struct(s); err != nil {
		errGroup, ok := err.(validator.ValidationErrors)
		if ok {
			var (
				details = make(map[string]string, len(errGroup))
				fields  = make([]string, len(errGroup))
			)

			for index, err := range errGroup {
				rawField := strings.TrimPrefix(strings.TrimSuffix(err.Field(), "}"), "{")
				parts := strings.SplitN(rawField, "|", 2)
				jsonKey, label := parts[0], parts[0]
				if len(parts) == 2 {
					label = parts[1]
				}

				translated := strings.ReplaceAll(err.Translate(translator), err.Field(), label)
				details[jsonKey] = translated
				fields[index] = label
			}

			msg := fmt.Sprintf(
				"Os campos (%s) são inválidos. Verifique os valores informados.",
				strings.Join(fields, ", "),
			)
			if len(fields) == 1 {
				msg = fmt.Sprintf(
					"O campo (%s) é inválido. Verifique o valor informado.",
					fields[0],
				)
			}

			return &validateError{
				ResponseError: resperr.New(
					http.StatusBadRequest,
					msg,
				).WithDebugContext(
					ctx, errGroup,
				),
				details: details,
			}
		}

		return resperr.New(
			http.StatusInternalServerError,
			"Não foi possível validar os dados.",
		).WithErrorContext(ctx, err)
	}

	return nil
}
