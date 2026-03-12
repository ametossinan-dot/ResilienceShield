package storage

import (
	"time"
)

type Transaction struct {
	ID        int
	Method    string
	Endpoint  string
	Body      string
	Status    string
	CreatedAt time.Time
}

// Ajouter une transaction dans la file
func Enqueue(method, endpoint, body string) error {
	_, err := DB.Exec(
		`INSERT INTO transactions (method, endpoint, body, status)
		 VALUES (?, ?, ?, 'pending')`,
		method, endpoint, body,
	)
	return err
}

// Récupérer toutes les transactions en attente (ordre FIFO)
func GetPending() ([]Transaction, error) {
	rows, err := DB.Query(
		`SELECT id, method, endpoint, body, status, created_at
		 FROM transactions
		 WHERE status = 'pending'
		 ORDER BY created_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Transaction
	for rows.Next() {
		var t Transaction
		rows.Scan(&t.ID, &t.Method, &t.Endpoint,
			&t.Body, &t.Status, &t.CreatedAt)
		list = append(list, t)
	}
	return list, nil
}

// Marquer une transaction comme synchronisée
func MarkSynced(id int) error {
	_, err := DB.Exec(
		`UPDATE transactions
		 SET status = 'synced', synced_at = CURRENT_TIMESTAMP
		 WHERE id = ?`, id,
	)
	return err
}

// Compter les transactions en attente
func CountPending() int {
	var count int
	DB.QueryRow(
		`SELECT COUNT(*) FROM transactions WHERE status = 'pending'`,
	).Scan(&count)
	return count
}
