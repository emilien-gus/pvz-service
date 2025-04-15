package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"pvz/internal/models"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

type PVZRepositoryInterface interface {
	InsertPVZ(ctx context.Context, city string) (*models.PVZ, error)
	GetPVZList(ctx context.Context, startDate, endDate *time.Time, page, limit int) ([]models.PVZWithReceptions, error)
}

type PVZRepository struct {
	db *sql.DB
}

func NewPWZRepository(db *sql.DB) *PVZRepository {
	return &PVZRepository{db: db}
}

func (p *PVZRepository) InsertPVZ(ctx context.Context, city string) (*models.PVZ, error) {
	id := uuid.New()
	query, args, err := sq.Insert("pvz").Columns("id, city").Values(id, city).Suffix("RETURNING id, registration_date, city").PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var pvz models.PVZ
	err = p.db.QueryRowContext(ctx, query, args...).Scan(&pvz.ID, &pvz.RegistrationDate, &pvz.City)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &pvz, nil
}

func (p *PVZRepository) GetPVZList(ctx context.Context, startDate, endDate *time.Time, page, limit int) ([]models.PVZWithReceptions, error) {
	query := sq.Select(
		"p.id AS pvz_id",
		"p.registration_date",
		"p.city",
		"r.id AS reception_id",
		"r.date_time AS reception_dateTime",
		"r.status AS reception_status",
		"pr.id AS product_id",
		"pr.date_time AS product_dateTime",
		"pr.type AS product_type",
	).
		From("pvz p").
		LeftJoin("receptions r ON p.id = r.pvz_id").
		LeftJoin("products pr ON r.id = pr.reception_id").
		PlaceholderFormat(sq.Dollar).
		Limit(uint64(limit)).
		Offset(uint64((page-1)*limit)).
		OrderBy("p.id", "r.date_time", "pr.date_time")

	if startDate != nil {
		query = query.Where(sq.GtOrEq{"r.date_time": *startDate})
	}
	if endDate != nil {
		query = query.Where(sq.LtOrEq{"r.date_time": *endDate})
	}

	sqlQuery, args, err := query.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := p.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var rawResult []*models.PVZWithReceptions
	pvzMap := make(map[uuid.UUID]*models.PVZWithReceptions)

	for rows.Next() {
		var (
			pvzID             uuid.UUID
			city              string
			registrationDate  time.Time
			receptionID       sql.NullString
			receptionDateTime sql.NullTime
			receptionStatus   sql.NullString
			productID         sql.NullString
			productDateTime   sql.NullTime
			productType       sql.NullString
		)

		err := rows.Scan(
			&pvzID,
			&registrationDate,
			&city,
			&receptionID,
			&receptionDateTime,
			&receptionStatus,
			&productID,
			&productDateTime,
			&productType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		pvzResp, exists := pvzMap[pvzID]
		if !exists {
			pvzResp = &models.PVZWithReceptions{
				ID:               pvzID,
				RegistrationDate: registrationDate,
				City:             city,
				Receptions:       []models.ReceptionWithProducts{},
			}
			pvzMap[pvzID] = pvzResp
			rawResult = append(rawResult, pvzResp)
		}

		if receptionID.Valid {
			receptionUUID, err := uuid.Parse(receptionID.String)
			log.Println("Reception:", receptionUUID)

			if err != nil {
				return nil, fmt.Errorf("error while parsing pvz data reception uuid: %w", err)
			}
			var existingReception *models.ReceptionWithProducts
			for i := range pvzResp.Receptions {
				if pvzResp.Receptions[i].Reception.ID == receptionUUID {
					existingReception = &pvzResp.Receptions[i]
					break
				}
			}
			if existingReception == nil {
				newReception := models.ReceptionWithProducts{
					Reception: models.Reception{
						ID:       receptionUUID,
						DateTime: receptionDateTime.Time,
						PVZID:    pvzID,
						Status:   receptionStatus.String,
					},
					Products: []models.Product{},
				}
				pvzResp.Receptions = append(pvzResp.Receptions, newReception)
				existingReception = &newReception
			}

			if productID.Valid {
				productUUID, err := uuid.Parse(productID.String)
				log.Println("Product:", productUUID)
				if err != nil {
					return nil, fmt.Errorf("error while parsing pvz data product uuid: %w", err)
				}
				product := models.Product{
					ID:          productUUID,
					DateTime:    productDateTime.Time,
					ProductType: productType.String,
					ReceptionID: receptionUUID,
				}
				existingReception.Products = append(existingReception.Products, product)
			}
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error after row iteration: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error after row iteration: %w", err)
	}

	result := make([]models.PVZWithReceptions, len(rawResult))
	for i := range rawResult {
		result[i] = *rawResult[i]
	}

	log.Println(rawResult)
	return result, nil
}
