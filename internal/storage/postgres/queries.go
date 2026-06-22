package postgres

const (
	saveLinkQuery      = `INSERT INTO links (code, original_url) VALUES ($1, $2)`
	getByCodeQuery     = `SELECT code, original_url FROM links WHERE code = $1`
	getByOriginalQuery = `SELECT code, original_url FROM links WHERE original_url = $1`
)
