package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/lib/pq"
)

func (m *DbModel) GetAllMovies() ([]*Movie, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

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
   		 ORDER BY m.id ASC
`

	fullQuery := baseQuery + joinClause + groupAndOrderClause

	rows, err := m.Db.QueryContext(ctx, fullQuery)
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
		movie.Image = "https://res.cloudinary.com/dvc85iwpj/image/upload/v1720247654/download_i0205y.png"
		movie.MovieGenre = make(map[int]string)

		// get genres, if any
		genreQuery := `SELECT
                mg.id, mg.movie_id, mg.genre_id, g.genre_name
            FROM
                movies_genres mg
                LEFT JOIN genres g ON (g.id = mg.genre_id)
            WHERE
                mg.movie_id = $1`

		genreRows, err := m.Db.QueryContext(ctx, genreQuery, movie.ID)
		if err != nil {
			return nil, err
		}

		for genreRows.Next() {
			var genreID int
			var genreName string
			err := genreRows.Scan(
				new(int), // We don't need the id from movies_genres
				new(int), // We don't need the movie_id
				&genreID,
				&genreName)

			if err != nil {
				genreRows.Close()
				return nil, err
			}

			movie.MovieGenre[genreID] = genreName
		}
		genreRows.Close()

		movies = append(movies, &movie)
	}

	return movies, nil
}

func (m *DbModel) CheckGenre(Genreid int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	query := `select id from genres where id = $1`

	var id int

	err := m.Db.QueryRowContext(ctx, query, Genreid).Scan(&id)

	if err != nil {
		return false, err
	}

	if id <= 0 {
		return false, errors.New("Genre not found")
	}

	return true, nil
}

func (m *DbModel) InsertGenre(Genrename string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	query := `insert into genres (genre_name, created_at, updated_at) values ($1,$2,$3) returning id`

	var id int

	err := m.Db.QueryRowContext(ctx, query, Genrename, time.Now(), time.Now()).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (m *DbModel) UpdateGenre(id int, GenreName string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	query := `update genres set genre_name = $1, updated_at = $2 where id = $3`

	_, err := m.Db.ExecContext(ctx, query, GenreName, time.Now(), id)

	if err != nil {
		return 0, err
	}

	// rowsAffected,err := result.RowsAffected()
	// if err!=nil {
	// 	return 0,err
	// }

	return id, nil
}

func (m *DbModel) DeleteGenre(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	query := `delete from genres where id = $1`

	_, err := m.Db.ExecContext(ctx, query, id)

	if err != nil {
		return err
	}

	return nil
}

// Check rating
func (m *DbModel) CheckRating(movieID, userID int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select id from ratings where movie_id = $1 and user_id = $2`

	ratingID := 0

	err := m.Db.QueryRowContext(ctx, query, movieID, userID).Scan(&ratingID)
	if err != nil {
		return 0, errors.New("Rating not found")
	}

	return ratingID, nil
}

// InsertRating inserts a new rating into the database
func (m *DbModel) InsertRating(rating *Rating) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `insert into ratings (movie_id, user_id, rating, created_at, updated_at) values ($1, $2, $3, $4, $5) returning id`

	var id int
	err := m.Db.QueryRowContext(ctx, query, rating.MovieID, rating.UserID, rating.Rating, rating.CreatedAt, rating.UpdatedAt).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

// UpdateRating updates a rating in the database
func (m *DbModel) UpdateRating(rating *Rating) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `update ratings set rating = $1, updated_at = $2 where id = $3`

	_, err := m.Db.ExecContext(ctx, query, rating.Rating, rating.UpdatedAt, rating.ID)
	if err != nil {
		return rating.ID, errors.New("Rating not found")
	}

	return rating.ID, nil
}

// getting gerne by id
func (m *DbModel) GetGenreByID(id int) (*Genre, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select id, genre_name, created_at, updated_at from genres where id = $1`

	var genre Genre

	err := m.Db.QueryRowContext(ctx, query, id).Scan(&genre.ID, &genre.GenreName, &genre.CreatedAt, &genre.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &genre, nil
}

// getting all genres
func (m *DbModel) GetAllGenres() ([]*Genre, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select id, genre_name, created_at, updated_at from genres`

	rows, err := m.Db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []*Genre
	for rows.Next() {
		var genre Genre
		err := rows.Scan(&genre.ID, &genre.GenreName, &genre.CreatedAt, &genre.UpdatedAt)
		if err != nil {
			return nil, err
		}
		genres = append(genres, &genre)
	}

	return genres, nil
}

// get latest Movies Featured on the website
func (m *DbModel) GetLatestMovies(userID ...int) ([]*Movie, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT m.id, m.title, m.image, m.description, m.year, m.release_date,
		COALESCE(TRUNC(AVG(r.rating)::numeric, 1), 1.0) AS rating,
		m.runtime, m.created_at, m.updated_at
		FROM movies m
		LEFT JOIN ratings r ON r.movie_id = m.id
		GROUP BY m.id
		ORDER BY m.updated_at DESC
		LIMIT 5
	`

	rows, err := m.Db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []*Movie
	var image sql.NullString
	for rows.Next() {
		var movie Movie
		err := rows.Scan(
			&movie.ID,
			&movie.Title,
			&image,
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
		if !image.Valid || image.String == "" {
			movie.Image = "https://res.cloudinary.com/dvc85iwpj/image/upload/v1720247654/download_i0205y.png"
		} else {
			movie.Image = fmt.Sprintf("https://res.cloudinary.com/%s/image/upload/%s", os.Getenv("CLOUD_NAME"), image.String)
		}
		movie.MovieGenre = make(map[int]string)

		// get genres, if any
		genreQuery := `SELECT
                mg.id, mg.movie_id, mg.genre_id, g.genre_name
            FROM
                movies_genres mg
                LEFT JOIN genres g ON (g.id = mg.genre_id)
            WHERE
                mg.movie_id = $1`

		genreRows, err := m.Db.QueryContext(ctx, genreQuery, movie.ID)
		if err != nil {
			return nil, err
		}

		for genreRows.Next() {
			var genreID int
			var genreName string
			err := genreRows.Scan(
				new(int), // We don't need the id from movies_genres
				new(int), // We don't need the movie_id
				&genreID,
				&genreName)

			if err != nil {
				genreRows.Close()
				return nil, err
			}

			movie.MovieGenre[genreID] = genreName
		}
		genreRows.Close()

		if len(userID) > 0 {
			// check if movie is favorite
			favoriteQuery := `select id from favorites where movie_id = $1 and user_id = $2`
			_ = m.Db.QueryRowContext(ctx, favoriteQuery, movie.ID, userID[0]).Scan(&movie.IsFavorite)
		}

		movies = append(movies, &movie)
	}

	return movies, nil
}

func (m *DbModel) GetMoviesByGenre(genreID int) ([]*Movie, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
    SELECT DISTINCT ON (m.id)
        m.id, 
        m.title, 
		m.image,
        m.description, 
        m.year, 
        m.release_date, 
        COALESCE(trunc(AVG(r.rating)::numeric, 1), 1.0) AS rating, 
        m.runtime, 
        m.created_at, 
        m.updated_at,
        array_agg(DISTINCT g.id) AS genre_ids,
        array_agg(DISTINCT g.genre_name) AS genre_names
    FROM movies m
    JOIN movies_genres mg ON (mg.movie_id = m.id)
    JOIN genres g ON (g.id = mg.genre_id)
    LEFT JOIN ratings r ON (r.movie_id = m.id)
    WHERE mg.genre_id = $1
    GROUP BY 
        m.id, 
        m.title, 
        m.description, 
        m.year, 
        m.release_date, 
        m.runtime, 
        m.created_at, 
        m.updated_at
    ORDER BY m.id, rating DESC
    `

	rows, err := m.Db.QueryContext(ctx, query, genreID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []*Movie
	for rows.Next() {
		var image sql.NullString
		var movie Movie
		var genreIDs []int64 // Changed to int64
		var genreNames []string

		err := rows.Scan(
			&movie.ID,
			&movie.Title,
			&image,
			&movie.Description,
			&movie.Year,
			&movie.ReleaseDate,
			&movie.Rating,
			&movie.Runtime,
			&movie.CreatedAt,
			&movie.UpdatedAt,
			pq.Array(&genreIDs),
			pq.Array(&genreNames),
		)
		if err != nil {
			return nil, err
		}
		if !image.Valid || image.String == "" {
			movie.Image = "https://res.cloudinary.com/dvc85iwpj/image/upload/v1720247654/download_i0205y.png"
		} else {
			movie.Image = fmt.Sprintf("https://res.cloudinary.com/%s/image/upload/%s", os.Getenv("CLOUD_NAME"), image.String)
		}

		movie.MovieGenre = make(map[int]string)
		for i, id := range genreIDs {
			movie.MovieGenre[int(id)] = genreNames[i] // Convert int64 to int
		}

		movies = append(movies, &movie)
	}

	return movies, nil
}

func (m *DbModel) GetMovie(id int) (*Movie, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT m.id, m.title, m.description, m.year, m.release_date, m.runtime, m.image, m.created_at, m.updated_at,
    COALESCE(TRUNC(AVG(r.rating)::numeric, 1), 1.0) AS rating,
		COUNT(DISTINCT f.id) AS favorites_count
FROM movies m
LEFT JOIN ratings r ON r.movie_id = m.id
LEFT JOIN favorites f ON f.movie_id = m.id
WHERE m.id = $1
GROUP BY m.id;
`

	row := m.Db.QueryRowContext(ctx, query, id)

	var movie Movie
	var image sql.NullString

	err := row.Scan(
		&movie.ID,
		&movie.Title,
		&movie.Description,
		&movie.Year,
		&movie.ReleaseDate,
		&movie.Runtime,
		&image,
		&movie.CreatedAt,
		&movie.UpdatedAt,
		&movie.Rating,
		&movie.TotalFavorites,
	)
	if err != nil {
		return nil, err
	}

	// Check if the Image value is NULL or empty, and if it is, assign a default value
	if !image.Valid || image.String == "" {
		movie.Image = "https://res.cloudinary.com/dvc85iwpj/image/upload/v1720247654/download_i0205y.png"
	} else {
		movie.Image = fmt.Sprintf("https://res.cloudinary.com/%s/image/upload/%s", os.Getenv("CLOUD_NAME"), image.String)
	}

	movie.MovieGenre = make(map[int]string)
	// get genres, if any
	genreQuery := `select
	mg.id, mg.movie_id, mg.genre_id, g.genre_name
from
	movies_genres mg
	left join genres g on (g.id = mg.genre_id)
where
	mg.movie_id = $1
`

	genreRows, _ := m.Db.QueryContext(ctx, genreQuery, movie.ID)

	for genreRows.Next() {
		var mg MovieGenre
		err := genreRows.Scan(
			&mg.ID,
			&mg.MovieID,
			&mg.GenreID,
			&mg.Genre.GenreName,
		)
		if err != nil {
			return nil, err
		}
		movie.MovieGenre[mg.GenreID] = mg.Genre.GenreName
	}
	defer genreRows.Close()

	// Get comments ordered by recent update
	query = `SELECT
    c.id, c.user_id, c.comment, c.created_at, c.updated_at, u.name
    FROM
    	comments c
    LEFT JOIN users u ON (u.id = c.user_id)
    WHERE
    	c.movie_id = $1
  	ORDER BY c.created_at DESC
    `

	rows, _ := m.Db.QueryContext(ctx, query, id)
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		err := rows.Scan(
			&comment.ID,
			&comment.UserID,
			&comment.Comment,
			&comment.CreatedAt,
			&comment.UpdatedAt,
			&comment.UserName,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	movie.Comments = comments
	movie.TotalComments = len(comments)

	return &movie, nil
}
