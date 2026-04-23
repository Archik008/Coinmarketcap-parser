package ports

import "crypto_parser/internal/reporting/domain/valueobject"

type ReportGenerator interface {
	Generate(data valueobject.ReportData) (string, error)
}
