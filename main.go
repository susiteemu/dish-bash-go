package main

import (
	"dish-dash-go/db"
	dbinit "dish-dash-go/db_init"
	"dish-dash-go/model"
	"html/template"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func newTemplate() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("*.html")),
	}
}

func useSelections() []model.UsageOption {
	var useSelections []model.UsageOption
	now := time.Now()
	daysBefore := 7
	for i := 0; i < daysBefore; i++ {
		d := now.Add(time.Duration(i) * time.Hour * 24 * -1)
		useSelections = append(useSelections, model.UsageOption{
			Id:   d.Unix(),
			Name: d.Format("2.1"),
		})
	}
	return useSelections
}

func mapDish(v model.Dish) model.TemplateDish {

	useSelections := useSelections()

	now := time.Now()
	lastUsage := v.LastUsage
	daysSince := -1
	if !lastUsage.IsZero() {
		daysSince = int(now.Sub(lastUsage).Hours() / 24)
	}
	return model.TemplateDish{
		Id:      v.Id,
		Name:    v.Name,
		Url:     v.Url,
		Created: v.Created,
		UsageOptions: model.UsageOptions{
			Today: model.UsageOption{
				Id:   useSelections[0].Id,
				Name: "Today",
			},
			Yesterday: model.UsageOption{
				Id:   useSelections[1].Id,
				Name: "Yesterday",
			},
			WithinWeek: model.UsageOption{
				Id:   useSelections[6].Id,
				Name: "Within Week",
			},
		},
		UsageStats: model.UsageStats{
			Count:     v.UsedCount,
			DaysSince: daysSince,
		},
	}
}

func main() {
	dbinit.Init()

	repo, err := db.NewRepo()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	root := func(c echo.Context) error {
		path := c.Request().URL
		log.Printf("Root handler is going to handle this request to %v", path)
		allDishes, err := repo.SelectAllDishes()
		if err != nil {
			log.Fatal(err)
			return echo.NewHTTPError(500, "Failed to get dishes")
		}

		var templateDishes []model.TemplateDish
		for _, v := range allDishes {
			templateDishes = append(templateDishes, mapDish(v))
		}

		// TODO logic for suggestions
		dishes := map[string]interface{}{
			"Dishes": templateDishes,
		}
		return c.Render(200, "index.html", dishes)
	}

	addDish := func(c echo.Context) error {
		path := c.Request().URL
		log.Printf("Add dish handler is going to handle this request to %v", path)
		name := c.Request().PostFormValue("name")
		url := c.Request().PostFormValue("url")

		log.Printf("Inserting dish with name %s and url %s", name, url)

		dish, err := repo.InsertDish(model.Dish{
			Name:      name,
			Url:       url,
			UsedCount: 0,
		})

		templateDish := mapDish(dish)

		if err != nil {
			log.Fatal("Error occurred when inserting dish", err)
		}

		log.Printf("Inserted dish %v", dish)
		return c.Render(200, "dish-item", templateDish)
	}

	deleteDish := func(c echo.Context) error {
		path := c.Request().URL
		log.Printf("Delete dish handler is going to handle this request to %v", path)
		id, err := strconv.Atoi(c.Param("id"))

		if err != nil {
			log.Fatal(err)
			return echo.NewHTTPError(400, "Id is missing or invalid")
		}

		log.Printf("Deleting dish with id %d", id)

		ok, err := repo.DeleteDishById(id)
		if err != nil {
			log.Fatal("Error occurred when deleting dish", err)
			return echo.NewHTTPError(500, "Failed to delete dish")
		}

		if ok {
			log.Printf("Deleted dish by id %d", id)
			return c.NoContent(200)
		} else {
			log.Printf("Deleting dish by id %d failed", id)
			return echo.NewHTTPError(500, "Failed to delete dish")
		}

	}

	searchDishes := func(c echo.Context) error {
		path := c.Request().URL
		query := c.Request().PostFormValue("search")
		log.Printf("Search handler is going to handle this request to %v with query %s", path, query)
		allDishes, err := repo.Search(query)
		if err != nil {
			log.Fatal(err)
			return echo.NewHTTPError(500, "Failed to search dishes")
		}
		log.Printf("Got %d results", len(allDishes))

		var templateDishes []model.TemplateDish
		for _, v := range allDishes {
			templateDishes = append(templateDishes, mapDish(v))
		}

		dishes := map[string][]model.TemplateDish{
			"Dishes": templateDishes,
		}
		return c.Render(200, "dishes", dishes)
	}

	useDish := func(c echo.Context) error {
		path := c.Request().URL
		id, _ := strconv.Atoi(c.Param("id"))
		q := c.Request().URL.Query()
		ts := q.Get("ts")
		log.Printf("Use dish handler is going to handle this request to %v with id %d and ts %s", path, id, ts)
		return c.NoContent(200)
	}

	e := echo.New()
	e.Static("/static", "assets")
	e.Use(middleware.Logger())
	e.Renderer = newTemplate()
	e.GET("/", root)
	e.POST("/dish", addDish)
	e.DELETE("/dish/:id", deleteDish)
	e.POST("/search", searchDishes)
	e.POST("/dish/:id/use", useDish)
	e.Logger.Fatal(e.Start(":1323"))
}
