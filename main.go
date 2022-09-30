package main

import (
	// "fmt"
	"log"
	"net/http"
	"os"

	"github.com/CodingMikel/go-fiber-postgres/models"
	"github.com/CodingMikel/go-fiber-postgres/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type Book struct {
	Author    string `json:"author"`
	Title     string `json:"title"`
	Publisher string `json:"publisher"`
}

type Repository struct {
	DB *gorm.DB
}

func (r *Repository) CreateBook(context *fiber.Ctx) error { //r is a var has Repository struct (a DB)
	book := Book{}                   // create var name book has type of Book struct
	err := context.BodyParser(&book) //take value from context, which is assigned after entering value using postman
	if err != nil {                  //check if entering value is good or not
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "request failed"})
		return err
	}

	err = r.DB.Create(&book).Error //assign status of Creating a new book to r (a DB)
	if err != nil {                // check the creating book status
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not create book"})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{ //if creating process has no problem, send success message
		"message": "book has been added"})
	return nil
}

func (r *Repository) GetBooks(context *fiber.Ctx) error {
	bookModels := &[]models.Books{}
	err := r.DB.Find(bookModels).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "could not get books"})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "book fetched succesfully",
		"data":    bookModels,
	})
	return nil
}

func (r *Repository) DeleteBook(context *fiber.Ctx) error {
	bookModel := models.Books{}
	id := context.Params("id")
	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "id cannot be empty"})
		return nil
	}

	err := r.DB.Delete(bookModel, id)
	if err.Error != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "could not delete book",
		})
		return err.Error
	}
	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "book has been deleted succesfully",
	})
	return nil
}

func (r *Repository) GetBookByID(context *fiber.Ctx) error {
	id := context.Params("id")
	bookModel := &models.Books{}
	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "id cannot be empty",
		})
		return nil
	}
	

	// fmt.Println("The ID is: ", id) //optional

	err := r.DB.Where("id = ?", id).First(bookModel).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "could not get the book",
		})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "book ID is fetched successfully",
		"data":    bookModel,
	})
	return nil
}

func (r *Repository) SetupRoutes(app *fiber.App) {
	api := app.Group("/api")
	api.Post("/create_books", r.CreateBook)
	api.Delete("/delete_book/:id", r.DeleteBook)
	api.Get("/get_books/:id", r.GetBookByID)
	api.Get("/books", r.GetBooks)
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	config := &storage.Config{
		Host:     os.Getenv("DB_HOSt"),
		Port:     os.Getenv("DB_PORT"),
		Password: os.Getenv("DB_PASS"),
		User:     os.Getenv("DB_USER"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
	}

	db, err := storage.NewConnection(config)

	if err != nil {
		log.Fatal("Could not load the database")
	}
	err = models.MigrateBooks(db)
	if err != nil {
		log.Fatal("Could not migrate db")
	}

	r := Repository{
		DB: db,
	}
	app := fiber.New()
	r.SetupRoutes(app)
	app.Listen(":8080")
}
