package query

import (
	"context"
	"database/sql"
	"strings"
)

var partsQuery = `SELECT 
	parts.id, 
	parts.name, 
	parts.comment, 
	parts.description, 
	parts.instock, 
	storelocations.name AS Lagerplatz
FROM 
    parts
INNER JOIN storelocations ON 
    parts.id_storelocation = storelocations.id
WHERE parts.id >= ?
ORDER BY parts.id`

var locationQuery = `SELECT
    a.id, a.name, a.comment,
    b.name AS "Lagerort"
FROM
    storelocations a
        INNER JOIN storelocations b ON
            b.id = a.parent_id
WHERE a.id >= ?
ORDER BY a.id`

type Querier struct {
	db *sql.DB
}

type PartData struct {
	ID          int64
	Name        string
	Comment     string
	Description string
	Instock     int64
	Lagerplatz  string
}

type LocationData struct {
	ID       int64
	Name     string
	Comment  string
	Lagerort string
}

func NewQuerier(db *sql.DB) *Querier {
	return &Querier{
		db: db,
	}
}

func (q *Querier) GetPartData(ctx context.Context, startId int64) ([]PartData, error) {
	rows, err := q.db.QueryContext(ctx, partsQuery, startId)
	if err != nil {
		return nil, err
	}

	var partRows []PartData
	for rows.Next() {
		var partData PartData
		err = rows.Scan(&partData.ID, &partData.Name, &partData.Comment, &partData.Description, &partData.Instock, &partData.Lagerplatz)
		if err != nil {
			return nil, err
		}
		partData.Comment = strings.ReplaceAll(partData.Comment, "\r\n", "\n")
		partData.Description = strings.ReplaceAll(partData.Description, "\r\n", "\n")
		partRows = append(partRows, partData)
	}

	return partRows, nil
}

func (q *Querier) GetLocationData(ctx context.Context, startID int64) ([]LocationData, error) {
	rows, err := q.db.QueryContext(ctx, locationQuery, startID)
	if err != nil {
		return nil, err
	}

	var locationRows []LocationData
	for rows.Next() {
		var locationData LocationData
		err = rows.Scan(&locationData.ID, &locationData.Name, &locationData.Comment, &locationData.Lagerort)
		if err != nil {
			return nil, err
		}
		locationData.Comment = strings.ReplaceAll(locationData.Comment, "\r\n", "\n")
		locationRows = append(locationRows, locationData)
	}
	return locationRows, nil
}
