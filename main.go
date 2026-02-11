package main

import (
	"fmt"
	"net/http"

	"github.com/Knetic/govaluate"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Calcuation struct {
	ID         string `json:"id"`
	Expression string `json:"expression"`
	Result     string `json:"result"`
}

type CalculationRequest struct {
	Expression string `json:"expression"`
}

type RequestBody struct {
	Task string `json:"task"`
}

var task = "world"

var calculations = []Calcuation{}

func calculateExpression(expression string) (string, error) {
	expr, err := govaluate.NewEvaluableExpression(expression)
	if err != nil {
		return "", err
	}
	result, err := expr.Evaluate(nil)
	if err != nil {
		return "", nil
	}
	return fmt.Sprintf("%v", result), err
}

func getCalculations(c echo.Context) error {
	return c.JSON(http.StatusOK, calculations)
}

func postCalculations(c echo.Context) error {
	var req CalculationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	result, err := calculateExpression(req.Expression)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid expression"})
	}

	calc := Calcuation{
		ID:         uuid.NewString(),
		Expression: req.Expression,
		Result:     result,
	}

	calculations = append(calculations, calc)
	return c.JSON(http.StatusCreated, calc)
}

func postTask(c echo.Context) error {
	var req RequestBody

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.Task == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Task field cannot be empty"})
	}

	task = req.Task

	return c.JSON(http.StatusOK, map[string]string{"message": fmt.Sprintf("Task updated to %v", task)})
}

func getHello(c echo.Context) error {
	return c.JSON(http.StatusOK, fmt.Sprintf("hello %v", task))
}

func main() {
	e := echo.New()

	e.Use(middleware.CORS())
	e.Use(middleware.RequestLogger())

	e.GET("/calculations", getCalculations)
	e.POST("/calculations", postCalculations)

	e.POST("/task", postTask)
	e.GET("/", getHello)

	e.Start("localhost:8080")
}
