package alpsusersettings

import (
	"fmt"
	"html/template"
	"net/http"

	"git.sr.ht/~migadu/alps"
	"github.com/labstack/echo/v4"
	"github.com/nutsdb/nutsdb"
)

type userSettingsPlugin struct {
	db *nutsdb.DB
}

func (p *userSettingsPlugin) Name() string {
	return "usersettings"
}

func (p *userSettingsPlugin) LoadTemplate(t *template.Template) error {
	return nil
}

func (p *userSettingsPlugin) SetRoutes(group *echo.Group) {
	group.Add("GET", "/user-settings", func(ectx echo.Context) error {
		var json string
		ctx := ectx.(*alps.Context)

		err := p.db.View(func(tx *nutsdb.Tx) error {
			key := []byte(ctx.Session.Username())
			result, err := tx.Get("settings", key)

			if err == nil {
				json = string(result.Value)
			} else {
				json = "{}"
			}

			return nil
		})

		if err != nil {
			return err
		}

		return ctx.HTML(http.StatusOK, json)
	})

	group.Add("POST", "/user-settings", func(ectx echo.Context) error {
		ctx := ectx.(*alps.Context)

		err := p.db.Update(func(tx *nutsdb.Tx) error {
			formParams, _ := ctx.FormParams()
			json, _ := formParams["json"]
			key := []byte(ctx.Session.Username())
			value := []byte(json[0])
			err := tx.Put("settings", key, value, 0)

			return err
		})

		if err == nil {
			return ctx.HTML(http.StatusOK, "")
		} else {
			return ctx.HTML(http.StatusInternalServerError, fmt.Sprint(err))
		}
	})
}

func (p *userSettingsPlugin) Inject(ctx *alps.Context, name string, data alps.RenderData) error {
	return nil
}

func (p *userSettingsPlugin) Close() error {
	p.db.Close()
	return nil
}

func loader(s *alps.Server) ([]alps.Plugin, error) {
	p := &userSettingsPlugin{}
	plugins := make([]alps.Plugin, 0, 1)
	plugins = append(plugins, p)

	db, err := nutsdb.Open(
		nutsdb.DefaultOptions,
		nutsdb.WithDir("./db"),
	)
	p.db = db

	if err != nil {
		return plugins, err
	}

	err = db.Update(func(tx *nutsdb.Tx) error {
		key := []byte("{placeholder}")
		value := []byte("{}")
		err := tx.Put("settings", key, value, 0)

		return err
	})

	if err != nil {
		return plugins, err
	}

	return plugins, err
}

func init() {
	alps.RegisterPluginLoader(loader)
}
