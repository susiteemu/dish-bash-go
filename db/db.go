package db

import (
	"database/sql"
	"dish-dash-go/model"
	"fmt"
	"log"
	"time"
)

type Repo struct {
	db *sql.DB
}

func NewRepo() (*Repo, error) {
	db, err := openDB()
	if err != nil {
		return nil, err
	}
	return &Repo{
		db: db,
	}, nil
}

func openDB() (*sql.DB, error) {

	db, err := sql.Open("sqlite3", "./dishbashgo.db")
	if err != nil {
		fmt.Println("db problem")
		log.Fatal(err)
	}

	return db, err
}

func (repo Repo) SelectAllDishes() ([]model.Dish, error) {

	var dishes []model.Dish
	db := repo.db
	rows, err := db.Query("SELECT id, name, url, created, usedCount, lastUsage FROM dish ORDER BY created DESC")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			id        int
			name      string
			url       string
			created   time.Time
			usedCount int
			lastUsage time.Time
		)
		err = rows.Scan(&id, &name, &url, &created, &usedCount, &lastUsage)
		if err != nil {
			log.Fatal(err)
		} else {
			dishes = append(dishes, model.Dish{
				Id:        id,
				Name:      name,
				Url:       url,
				Created:   created,
				UsedCount: usedCount,
				LastUsage: lastUsage,
			})
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return dishes, nil
}

func (repo Repo) SelectDishById(id int) (model.Dish, error) {

	log.Printf("Selecting dish by id %d", id)

	db := repo.db
	row := db.QueryRow("SELECT name, url, created, usedCount, lastUsage FROM dish WHERE id = $1", id)

	var (
		name      string
		url       string
		created   time.Time
		usedCount int
		lastUsage time.Time
	)
	err := row.Scan(&name, &url, &created, &usedCount, &lastUsage)
	if err != nil {
		log.Fatal(err)
		return model.Dish{}, err
	}

	log.Printf("Got dish name=%s, url=%s, created=%v, usedCount=%d, lastUsage=%v with id=%d", name, url, created, usedCount, lastUsage, id)

	return model.Dish{
		Id:        id,
		Name:      name,
		Url:       url,
		Created:   created,
		UsedCount: usedCount,
		LastUsage: lastUsage,
	}, nil
}

func (repo Repo) Search(query string) ([]model.Dish, error) {
	var dishes []model.Dish
	db := repo.db
	rows, err := db.Query("SELECT id, name, url, created, usedCount, lastUsage FROM dish WHERE name LIKE '%' || $1 || '%' ORDER BY created DESC", query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			id        int
			name      string
			url       string
			created   time.Time
			usedCount int
			lastUsage time.Time
		)
		err = rows.Scan(&id, &name, &url, &created, &usedCount, &lastUsage)
		if err != nil {
			log.Fatal(err)
		} else {
			dishes = append(dishes, model.Dish{
				Id:        id,
				Name:      name,
				Url:       url,
				Created:   created,
				UsedCount: usedCount,
				LastUsage: lastUsage,
			})
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return dishes, nil
}

func (repo Repo) SortByLastUsageDesc() ([]model.Dish, error) {
	return repo.selectAll("SELECT id, name, url, created, usedCount, lastUsage FROM dish ORDER BY lastUsage DESC")
}

func (repo Repo) SortByLastUsageAsc() ([]model.Dish, error) {
	return repo.selectAll("SELECT id, name, url, created, usedCount, lastUsage FROM dish ORDER BY lastUsage ASC")
}

func (repo Repo) SortByUsedCountDesc() ([]model.Dish, error) {
	return repo.selectAll("SELECT id, name, url, created, usedCount, lastUsage FROM dish ORDER BY usedCount DESC")
}

func (repo Repo) SortByUsedCountAsc() ([]model.Dish, error) {
	return repo.selectAll("SELECT id, name, url, created, usedCount, lastUsage FROM dish ORDER BY usedCount ASC")
}

func (repo Repo) selectAll(sql string) ([]model.Dish, error) {
	var dishes []model.Dish
	db := repo.db
	rows, err := db.Query(sql)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			id        int
			name      string
			url       string
			created   time.Time
			usedCount int
			lastUsage time.Time
		)
		err = rows.Scan(&id, &name, &url, &created, &usedCount, &lastUsage)
		if err != nil {
			log.Fatal(err)
		} else {
			dishes = append(dishes, model.Dish{
				Id:        id,
				Name:      name,
				Url:       url,
				Created:   created,
				UsedCount: usedCount,
				LastUsage: lastUsage,
			})
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return dishes, nil

}

func (repo Repo) InsertDish(dish model.Dish) (model.Dish, error) {
	db := repo.db
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
		return model.Dish{}, err
	}

	stmt, err := tx.Prepare("INSERT INTO dish(name, url, created, usedCount, lastUsage) VALUES(?, ?, ?, ?, ?) RETURNING id")
	if err != nil {
		log.Fatal(err)
		return model.Dish{}, err
	}
	defer stmt.Close()
	var id int
	err = stmt.QueryRow(dish.Name, dish.Url, time.Now(), dish.UsedCount, time.Time{}).Scan(&id)
	if err != nil {
		return model.Dish{}, err
	}
	err = tx.Commit()
	if err != nil {
		return model.Dish{}, err
	}

	log.Printf("Inserted dish which got id %d", id)

	newDish, err := repo.SelectDishById(id)
	return newDish, err
}

func (repo Repo) UpdateDish(dish model.Dish) (model.Dish, error) {
	db := repo.db
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
		return model.Dish{}, err
	}

	log.Printf("Updating dish")

	stmt, err := tx.Prepare("UPDATE dish set name=?, url=?, created=?, usedCount=?, lastUsage=? WHERE id = ?")
	if err != nil {
		log.Fatal(err)
		return model.Dish{}, err
	}
	defer stmt.Close()
	_, err = stmt.Exec(dish.Name, dish.Url, dish.Created, dish.UsedCount, dish.LastUsage, dish.Id)
	if err != nil {
		return model.Dish{}, err
	}
	err = tx.Commit()
	if err != nil {
		return model.Dish{}, err
	}

	updatedDish, err := repo.SelectDishById(dish.Id)
	return updatedDish, err
}

func (repo Repo) DeleteDishById(id int) (bool, error) {
	db := repo.db
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
		return false, err
	}

	stmt, err := tx.Prepare("DELETE FROM dish WHERE id = ?")
	if err != nil {
		log.Fatal(err)
		return false, err
	}
	defer stmt.Close()
	_, err = stmt.Exec(id)
	if err != nil {
		return false, err
	}
	err = tx.Commit()
	if err != nil {
		return false, err
	}

	return true, err
}
