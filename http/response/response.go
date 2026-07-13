package response

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ResponseError interface {
	StatusCode() int
	Message() string
	Details() map[string]string
}

type response struct {
	Message string `json:"message"`
}

type responseError struct {
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// Encode escreve a resposta no formato JSON no http.ResponseWriter.
//
// Parâmetros:
//   - w (http.ResponseWriter): O writer HTTP onde a resposta será escrita.
//   - statusCode (int): Código de status HTTP a ser retornado na resposta.
//   - data (T): Dados genéricos a serem codificados como JSON.
func Encode[T any](w http.ResponseWriter, statusCode int, data T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		_ = json.NewEncoder(w).Encode(responseError{
			Message: "Não foi possível codificar a resposta.",
		})
	}
}

// EncodeOK serializa os dados fornecidos como resposta HTTP com status 200 (OK).
//
// Parâmetros:
//   - w (http.ResponseWriter): O objeto de resposta HTTP.
//   - data (T): Os dados a serem codificados e enviados na resposta.
func EncodeOK[T any](w http.ResponseWriter, data T) {
	Encode(
		w,
		http.StatusOK,
		data,
	)
}

// SendOK envia uma resposta JSON de sucesso (status HTTP 200 OK) para o cliente com a mensagem especificada.
//
// Parâmetros:
//   - w (http.ResponseWriter): O objeto de resposta HTTP.
//   - message (string): A mensagem a ser incluída na resposta.
func SendOK(w http.ResponseWriter, message string) {
	Encode(
		w,
		http.StatusOK,
		response{
			Message: message,
		},
	)
}

// SendOkf envia uma resposta JSON com status HTTP 200 (OK) e uma mensagem formatada.
//
// Parâmetros:
//   - w (http.ResponseWriter): O objeto de resposta HTTP.
//   - format (string): Uma string de formato que especifica a mensagem a ser enviada.
//   - a (...any): Argumentos opcionais que serão formatados na string de formato.
func SendOkf(w http.ResponseWriter, format string, a ...any) {
	Encode(
		w,
		http.StatusOK,
		response{
			Message: fmt.Sprintf(format, a...),
		},
	)
}

// Created envia uma resposta HTTP 201 (Created) com mensagem e cabeçalho opcional Location.
//
// Parâmetros:
//   - w (http.ResponseWriter): Writer da resposta HTTP.
//   - message (string): Mensagem a ser retornada no corpo da resposta.
//   - location (string, opcional): URL para o cabeçalho Location.
func Created(w http.ResponseWriter, message string, location ...string) {
	if location != nil {
		w.Header().Set("Location", location[0])
	}

	Encode(
		w,
		http.StatusCreated,
		response{
			Message: message,
		},
	)
}

// SendError envia uma resposta de erro com base no erro fornecido. Se o erro for do tipo HTTPError,
// ele será enviado como uma resposta JSON com o status e a mensagem apropriados. Caso contrário, uma resposta JSON
// de status 500 será enviada com a descrição do erro.
//
// Parâmetros:
//   - w (http.ResponseWriter): O objeto de resposta HTTP.
//   - err (error): O erro a ser enviado como resposta.
func SendError(w http.ResponseWriter, err error) {
	if httpErr, ok := err.(ResponseError); ok {
		Encode(
			w,
			httpErr.StatusCode(),
			responseError{
				Message: httpErr.Message(),
				Details: httpErr.Details(),
			},
		)

		return
	}

	Encode(
		w,
		http.StatusInternalServerError,
		responseError{
			Message: "Erro interno desconhecido.",
		},
	)
}

// NoContent envia uma resposta sem conteúdo com o status HTTP No Content (204).
//
// Parâmetros:
//   - w (http.ResponseWriter): O objeto de resposta HTTP.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
