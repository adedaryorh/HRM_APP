package main

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type MongoInstance struct {
	Client *mongo.Client
	Db     *mongo.Database
}

var mg MongoInstance

const dbName = "fiber-hrm"
const MongoURL = "mongodb://localhost:27107/" + dbName

type Employee struct {
	ID     string  `json:"id,omitempty" bson:"_id,omitempty"`
	Name   string  `json:"name"`
	Salary float64 `json:"salary"`
	Age    float64 `json:"age"`
}

func connect() error {

}
func Connect() error {
	client, err := mongo.NewClient(options.Client().ApplyURI(MongoURL))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		return err
	}
	db := client.Database(dbName)

	if err != nil {
		return err
	}
	mg = MongoInstance{
		Client: client,
		Db:     db,
	}
}

func main() {
	if err := connect(); err != nil {
		log.Fatal()
	}
	app := fiber.New()

	app.Get("/employee", func(ctx *fiber.Ctx) error {
		query := bson.D{{}}

		cursor, err := mg.Db.Collection("employees").Find(ctx.Context(), query)
		if err != nil {
			return ctx.Status(500).SendString(err.Error())
		}

		var employees []Employee = make([]Employee, 0)
		if err := cursor.All(ctx.Context(), &employees); err != nil {
			return ctx.Status(500).SendString(err.Error())
		}
		return ctx.JSON(employees)
	})

	app.Post("/employee", func(ctx *fiber.Ctx) error {
		collection := mg.Db.Collection("employees")

		employee := new(Employee)
		if err := ctx.BodyParser(employee); err != nil {
			return ctx.Status(400).SendString(err.Error())
		}
		employee.ID = ""
		insertionResult, err := collection.InsertOne(ctx.Context(), employee)
		if err != nil {
			return ctx.Status(500).SendString(err.Error())
		}
		filter := bson.D{{Key: "_id", Value: insertionResult.InsertedID}}
		createdRecord := collection.FindOne(ctx.Context(), filter)

		createdEmployee := &Employee{}
		createdRecord.Decode(createdEmployee)

		return ctx.Status(201).JSON(createdEmployee)
	})

	app.Put("/employee/:id", func(ctx *fiber.Ctx) error {
		idParams := ctx.Params("id")
		employeeID, err := primitive.ObjectIDFromHex(idParams)
		if err != nil {
			return ctx.SendStatus(400)
		}
		employee := new(Employee)
		if err := ctx.BodyParser(employee); err != nil {
			return ctx.Status(400).SendString(err.Error())
		}
		query := bson.D{{Key: "_id", Value: employeeID}}
		update := bson.D{
			{Key: "$set",
				Value: bson.D{
					{Key: "name", Value: employee.Name},
					{Key: "age", Value: employee.Age},
					{Key: "salary", Value: employee.Salary},
				},
			},
		}
		err := mg.Db.Collection("employees").FindOneAndUpdate(ctx.Context(), query, update).Err()
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return ctx.SendStatus(400)
			}
			if errors.Is(err, mongo.ErrClientDisconnected) {
				return ctx.SendStatus(500)
			}
			return ctx.Status(200).JSON("recored updated successfully")

		}
		employee.ID = idParams
		return ctx.Status(200).JSON(employee)
	})

	app.Delete("/employee/:id", func(ctx *fiber.Ctx) error {
		idParams := ctx.Params("id")
		employeeID, err := primitive.ObjectIDFromHex(idParams)
		if err != nil {
			return ctx.SendStatus(400)
		}
		query := bson.D{{Key: "_id", Value: employeeID}}
		result, err := mg.Db.Collection("employees").DeleteOne(ctx.Context(), &query)
		if err != nil {
			return ctx.SendStatus(500)
		}
		if result.DeletedCount < 1 {
			return ctx.SendStatus(404)
		}
		return ctx.Status(200).JSON("recored deleted successfully")
	})

	log.Fatal(app.Listen(":3000"))
}
