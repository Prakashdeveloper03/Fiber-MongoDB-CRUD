package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EmployeeInDB represents an employee in the database
type EmployeeInDB struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name   string             `bson:"name" json:"name"`
	Salary float64            `bson:"salary" json:"salary"`
	Age    float64            `bson:"age" json:"age"`
}

// Employee represents an employee
type Employee struct {
	Name   string  `json:"name"`
	Salary float64 `json:"salary"`
	Age    float64 `json:"age"`
}

var (
	client     *mongo.Client
	collection *mongo.Collection
)

func init() {
	var err error

	// Connect to MongoDB
	client, err = mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017/"))
	if err != nil {
		log.Fatal(err)
	}

	// Set database and collection
	collection = client.Database("hrms").Collection("employees")
}

func main() {
	app := fiber.New()

	// Middleware
	app.Use(logger.New())

	// Routes
	app.Get("/employee", getEmployees)
	app.Post("/employee", createEmployee)
	app.Put("/employee/:id", updateEmployee)
	app.Delete("/employee/:id", deleteEmployee)

	// Start server
	log.Fatal(app.Listen(":8000"))
}

func getEmployees(c *fiber.Ctx) error {
	// Get all employees from database
	cursor, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}
	defer cursor.Close(context.Background())

	// Parse employees into slice of EmployeeInDB
	var employees []EmployeeInDB
	if err := cursor.All(context.Background(), &employees); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	// Convert ObjectID to string and return as JSON
	var employeesJSON []byte
	if employeesJSON, err = json.Marshal(employees); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	return c.Send(employeesJSON)
}

func createEmployee(c *fiber.Ctx) error {
	// Parse request body into Employee struct
	var employee Employee
	if err := c.BodyParser(&employee); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	// Insert employee into database
	result, err := collection.InsertOne(context.Background(), employee)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	// Find newly created employee from database and return as JSON
	var createdEmployee EmployeeInDB
	if err := collection.FindOne(context.Background(), bson.M{"_id": result.InsertedID}).Decode(&createdEmployee); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	// Convert ObjectID to string and return as JSON
	var createdEmployeeJSON []byte
	if createdEmployeeJSON, err = json.Marshal(createdEmployee); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}
	return c.Send(createdEmployeeJSON)
}

func updateEmployee(c *fiber.Ctx) error {
	// Get employee ID from URL parameter
	id := c.Params("id")
	// Parse request body into Employee struct
	var employee Employee
	if err := c.BodyParser(&employee); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	// Convert ID string to ObjectID
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Invalid ID"})
	}

	// Update employee in database
	update := bson.M{
		"$set": bson.M{
			"name":   employee.Name,
			"salary": employee.Salary,
			"age":    employee.Age,
		},
	}
	if _, err := collection.UpdateOne(context.Background(), bson.M{"_id": oid}, update); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	// Find updated employee from database and return as JSON
	var updatedEmployee EmployeeInDB
	if err := collection.FindOne(context.Background(), bson.M{"_id": oid}).Decode(&updatedEmployee); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	// Convert ObjectID to string and return as JSON
	var updatedEmployeeJSON []byte
	if updatedEmployeeJSON, err = json.Marshal(updatedEmployee); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	return c.Send(updatedEmployeeJSON)
}

func deleteEmployee(c *fiber.Ctx) error {
	// Get employee ID from URL parameter
	id := c.Params("id")
	// Convert ID string to ObjectID
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"message": "Invalid ID"})
	}

	// Delete employee from database
	if _, err := collection.DeleteOne(context.Background(), bson.M{"_id": oid}); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{"message": "Employee deleted successfully"})
}
