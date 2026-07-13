package dto

import "time"

type StatusJob string

const (
	StatusJobPendente    StatusJob = "pendente"
	StatusJobProcessando StatusJob = "processando"
	StatusJobConcluido   StatusJob = "concluido"
	StatusJobErro        StatusJob = "erro"
)

type JobStatus struct {
	ID           string    `json:"id"`
	Status       StatusJob `json:"status"`
	Erro         string    `json:"erro,omitempty"`
	CriadoEm     time.Time `json:"criado_em"`
	AtualizadoEm time.Time `json:"atualizado_em"`
}
