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

// Task - структура задачи
type Task struct {
	ID     string `json:"id"`
	Task   string `json:"task"`
	Status string `json:"status"` // "active", "completed", "archived"
}

// RequestBody - для POST /task (обратная совместимость)
type RequestBody struct {
	Task string `json:"task"`
}

// UpdateTaskRequest - для PATCH /tasks/:id
type UpdateTaskRequest struct {
	Task   *string `json:"task,omitempty"` // указатель для определения наличия поля
	Status *string `json:"status,omitempty"`
}

// Хранилище задач (заменяем глобальную переменную task на слайс задач)
var tasks = []Task{}

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

func patchCalculations(c echo.Context) error {
	id := c.Param("id")

	var req CalculationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	result, err := calculateExpression(req.Expression)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid expression"})
	}

	for i, calculation := range calculations {
		if calculation.ID == id {
			calculations[i].Expression = req.Expression
			calculations[i].Result = result
			return c.JSON(http.StatusOK, calculations[i])
		}
	}
	return c.JSON(http.StatusBadRequest, map[string]string{"error": "Calculation not found"})
}

func deleteCalculations(c echo.Context) error {
	id := c.Param("id")

	for i, calculation := range calculations {
		if calculation.ID == id {
			calculations = append(calculations[:i], calculations[i+1:]...)
			return c.NoContent(http.StatusNoContent)
		}
	}
	return c.JSON(http.StatusBadRequest, map[string]string{"error": "Calculation not found"})
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

// POST /tasks - создать новую задачу
func createTask(c echo.Context) error {
	var req RequestBody
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.Task == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Task field cannot be empty"})
	}

	newTask := Task{
		ID:     uuid.New().String(),
		Task:   req.Task,
		Status: "active", // статус по умолчанию
	}

	tasks = append(tasks, newTask)

	// Для обратной совместимости обновляем и старую переменную
	task = req.Task

	return c.JSON(http.StatusCreated, newTask)
}

func getTasks(c echo.Context) error {
	return c.JSON(http.StatusOK, tasks)
}

// 🔴 **ЗАДАНИЕ 1: PATCH /tasks/:id - обновить задачу**
func patchTask(c echo.Context) error {
	id := c.Param("id")

	var req UpdateTaskRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// Ищем задачу по ID
	for i, t := range tasks {
		if t.ID == id {
			// Обновляем только те поля, которые переданы
			if req.Task != nil {
				tasks[i].Task = *req.Task
				// Для обратной совместимости обновляем и старую переменную
				task = *req.Task
			}
			if req.Status != nil {
				// Валидация статуса
				status := *req.Status
				if status != "active" && status != "completed" && status != "archived" {
					return c.JSON(http.StatusBadRequest, map[string]string{
						"error": "Status must be 'active', 'completed', or 'archived'",
					})
				}
				tasks[i].Status = status
			}

			return c.JSON(http.StatusOK, tasks[i])
		}
	}

	return c.JSON(http.StatusNotFound, map[string]string{"error": "Task not found"})
}

// 🔴 **ЗАДАНИЕ 2: DELETE /tasks/:id - удалить задачу**
func deleteTask(c echo.Context) error {
	id := c.Param("id")

	for i, t := range tasks {
		if t.ID == id {
			// Удаляем задачу из слайса
			tasks = append(tasks[:i], tasks[i+1:]...)

			// Если удалили текущую task, обновляем переменную
			if t.Task == task {
				if len(tasks) > 0 {
					task = tasks[len(tasks)-1].Task // берём последнюю задачу
				} else {
					task = "world" // сбрасываем на дефолт
				}
			}

			return c.NoContent(http.StatusNoContent)
		}
	}

	return c.JSON(http.StatusNotFound, map[string]string{"error": "Task not found"})
}

// GET /tasks/:id - получить задачу по ID
func getTaskByID(c echo.Context) error {
	id := c.Param("id")

	for _, t := range tasks {
		if t.ID == id {
			return c.JSON(http.StatusOK, t)
		}
	}

	return c.JSON(http.StatusNotFound, map[string]string{"error": "Task not found"})
}

func main() {
	e := echo.New()

	e.Use(middleware.CORS())
	e.Use(middleware.RequestLogger())

	e.GET("/calculations", getCalculations)
	e.POST("/calculations", postCalculations)
	e.PATCH("/calculations/:id", patchCalculations)
	e.DELETE("/calculations/:id", deleteCalculations)

	// 🔴 НОВЫЕ ROUTES для tasks (полноценный CRUD)
	e.GET("/tasks", getTasks)          // Read all
	e.POST("/tasks", createTask)       // Create
	e.GET("/tasks/:id", getTaskByID)   // Read one
	e.PATCH("/tasks/:id", patchTask)   // Update
	e.DELETE("/tasks/:id", deleteTask) // Delete

	// Старые routes для обратной совместимости
	e.POST("/task", postTask)
	e.GET("/", getHello)

	e.Start("localhost:8080")
}
