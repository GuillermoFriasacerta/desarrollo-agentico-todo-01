package actions

import (
	"net/http"
	"strings"

	"todo/models"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

// TasksIndex muestra la lista de tareas y el formulario de creación.
// Con ?edit=<id> muestra el formulario de edición de esa tarea.
func TasksIndex(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	editing, editTitle := "", ""
	if id, err := uuid.FromString(c.Param("edit")); err == nil {
		task := &models.Task{}
		if err := tx.Find(task, id); err == nil {
			editing, editTitle = task.ID.String(), task.Title
		}
	}

	return renderIndex(c, http.StatusOK, &models.Task{}, validate.NewErrors(), editing, editTitle, validate.NewErrors())
}

// renderIndex pinta la lista de tareas con el estado de los formularios
// de creación y de edición.
func renderIndex(c buffalo.Context, status int, task *models.Task, errors *validate.Errors, editing, editTitle string, editErrors *validate.Errors) error {
	tx := c.Value("tx").(*pop.Connection)

	tasks := models.Tasks{}
	if err := tx.Order("created_at asc, id asc").All(&tasks); err != nil {
		return err
	}

	c.Set("tasks", tasks)
	c.Set("task", task)
	c.Set("errors", errors)
	c.Set("editing", editing)
	c.Set("editTitle", editTitle)
	c.Set("editErrors", editErrors)
	return c.Render(status, r.HTML("tasks/index.plush.html"))
}

// TasksCreate crea una tarea nueva a partir del formulario.
func TasksCreate(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := c.Bind(task); err != nil {
		return err
	}
	task.Title = strings.TrimSpace(task.Title)

	verrs, err := tx.ValidateAndCreate(task)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		return renderIndex(c, http.StatusUnprocessableEntity, task, verrs, "", "", validate.NewErrors())
	}

	c.Flash().Add("success", "Tarea creada.")
	return c.Redirect(http.StatusSeeOther, "/")
}

// TasksUpdate cambia el título de una tarea.
func TasksUpdate(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := tx.Find(task, c.Param("task_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	submitted := c.Request().FormValue("title")
	task.Title = strings.TrimSpace(submitted)

	verrs, err := tx.ValidateAndUpdate(task)
	if err != nil {
		return err
	}

	if verrs.HasAny() {
		return renderIndex(c, http.StatusUnprocessableEntity, &models.Task{}, validate.NewErrors(), task.ID.String(), submitted, verrs)
	}

	c.Flash().Add("success", "Tarea actualizada.")
	return c.Redirect(http.StatusSeeOther, "/")
}

// TasksToggle alterna una tarea entre completada y pendiente.
func TasksToggle(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := tx.Find(task, c.Param("task_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	task.Completed = !task.Completed
	if err := tx.Update(task); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/")
}

// TasksDestroy elimina una tarea.
func TasksDestroy(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	task := &models.Task{}
	if err := tx.Find(task, c.Param("task_id")); err != nil {
		return c.Error(http.StatusNotFound, err)
	}

	if err := tx.Destroy(task); err != nil {
		return err
	}

	c.Flash().Add("success", "Tarea eliminada.")
	return c.Redirect(http.StatusSeeOther, "/")
}
