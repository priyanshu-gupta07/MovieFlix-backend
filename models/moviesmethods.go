package models

import (
	"context"
	"time"
)

func (m *DbModel) GetAllMovies(title string) ([]*Movie, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	var dbArgumens []interface{}

	where := "WHERE title ILIKE $1 OR description ILIKE $2"
	dbArgumens = append(dbArgumens, "%"+title+"%", "%"+title+"%")

	baseQuery := `
    SELECT 
        m.id, 
        m.title, 
        m.description, 
        m.year, 
        m.release_date, 
        COALESCE(trunc(AVG(r.rating)::numeric, 1), 1.0) AS rating, 
        m.runtime, 
        m.created_at, 
        m.updated_at 
    FROM movies m
`

joinClause := `LEFT JOIN ratings r ON (r.movie_id = m.id)`

groupAndOrderClause := `
    GROUP BY 
        m.id, 
        m.title, 
        m.description, 
        m.year, 
        m.release_date, 
        m.runtime, 
        m.created_at, 
        m.updated_at
    ORDER BY rating DESC 
    LIMIT 2 OFFSET 1
`

	fullQuery := baseQuery + joinClause + where + groupAndOrderClause


	rows, err := m.Db.QueryContext(ctx, fullQuery, dbArgumens...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []*Movie
	for rows.Next() {
		var movie Movie
		err := rows.Scan(
			&movie.ID,
			&movie.Title,
			&movie.Description,
			&movie.Year,
			&movie.ReleaseDate,
			&movie.Rating,
			&movie.Runtime,
			&movie.CreatedAt,
			&movie.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		movies = append(movies, &movie)
	}

	return movies, nil
}
